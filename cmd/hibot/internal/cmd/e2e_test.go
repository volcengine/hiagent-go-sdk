package cmd

// CLI 端到端集成测试：与 go/examples/e2e/e2e_test.go 一一对应。
//
// 目标：模拟一个用户从空白工作空间出发，完整跑完
//
//  1. hibot models get            → GetModel
//  2. hibot prompts create        → CreateAgentPromptTemplate
//  3. hibot skills upload         → UploadBlob + CreateSkill
//  4. hibot resources create      → UploadBlob + CreateResource
//  5. hibot mcps create           → CreateMCP
//  6. hibot agents create         → ListEnv (auto-resolve) + CreateAgent
//  7. hibot sessions create       → CreateSession
//  8. hibot chat <sid> --stream   → Chat (SSE delta+completed)
//  9. hibot chat <sid>            → Chat (batch / 同一 SSE 端点同步消费)
//
// 不依赖任何真实 Hibot 集群：使用 httptest.Server 完整模拟 TOP 路由、
// Action 路由、SSE 行为，确保 CLI 的每一条命令都路由到正确的 TOP Action，
// 并把字段透传到请求体里。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// e2eMock 是 go/examples/internal/mocktop 在 CLI 测试包内的对等实现。
// 因为 cli/ 与 go/ 是两个独立模块，CLI 不能直接 import go SDK 的
// examples/internal/mocktop（internal 可见性 + 跨模块），所以在这里复刻一份。
type e2eMock struct {
	t      testing.TB
	server *httptest.Server

	mu     sync.Mutex
	seen   map[string]int
	bodies map[string]map[string]any
}

func newE2EMock(t testing.TB) *e2eMock {
	t.Helper()
	m := &e2eMock{t: t, seen: map[string]int{}, bodies: map[string]map[string]any{}}
	m.server = httptest.NewServer(http.HandlerFunc(m.handle))
	t.Cleanup(m.server.Close)
	return m
}

func (m *e2eMock) URL() string { return m.server.URL }

func (m *e2eMock) Body(action string) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.bodies[action]
}

func (m *e2eMock) requireActions(actions ...string) {
	m.t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range actions {
		if m.seen[a] == 0 {
			m.t.Fatalf("action %q was not called (seen=%v)", a, m.seen)
		}
	}
}

func (m *e2eMock) handle(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("Action")

	m.mu.Lock()
	m.seen[action]++
	m.mu.Unlock()

	// UploadBlob 走 multipart，不解析 JSON；其余 action 都是 JSON。
	if action != "UploadBlob" {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			m.t.Fatalf("decode %s body: %v", action, err)
		}
		m.mu.Lock()
		m.bodies[action] = body
		m.mu.Unlock()
	}

	switch action {
	case "GetModel":
		writeE2EResult(w, `{"Items":[{"ID":"model-1","Name":"e2e-model","Type":"chat","Provider":"doubao","ModelName":"doubao-seed"}]}`)
	case "CreateAgentPromptTemplate":
		writeE2EResult(w, `{"ID":"prompt-1","Name":"e2e-prompt","Content":"hello"}`)
	case "UploadBlob":
		if r.URL.Query().Get("Filename") == "" {
			m.t.Fatalf("UploadBlob Filename is empty")
		}
		writeE2EResult(w, `{"BlobID":"blob-1"}`)
	case "CreateSkill":
		writeE2EResult(w, `{"ID":"skill-version-1","SkillID":"skill-1","Name":"e2e-skill","Version":"1.0.0"}`)
	case "CreateResource":
		writeE2EResult(w, `{"ID":"resource-1","Name":"e2e-resource","Type":"document_collection"}`)
	case "CreateMCP":
		writeE2EResult(w, `{"ID":"mcp-1","Name":"e2e-mcp","Transport":"streamable-http","Endpoint":"http://mcp.local/mcp"}`)
	case "ListEnv":
		writeE2EResult(w, `{"Items":[{"ID":"env-1","Name":"default-env","ImageType":"hermes","CreatedAt":"2026-01-01T00:00:00Z"}]}`)
	case "CreateAgent":
		writeE2EResult(w, `{"ID":"agent-1","Name":"e2e-agent","ModelID":"model-1","EnvID":"env-1"}`)
	case "CreateSession":
		writeE2EResult(w, `{"ID":"session-1","AgentID":"agent-1","PeerKind":"system","PeerID":"agent-1"}`)
	case "Chat":
		// SDK 的 Sessions.Chat / Sessions.ChatStreaming 共用同一个 Action=Chat 入口。
		// 流式分支由 CLI --stream 决定如何展示；mock 始终回 SSE。
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: delta\ndata: {\"request_id\":\"req-test\",\"delta\":{\"text\":\"ok\"}}\n\n")
		_, _ = fmt.Fprint(w, "event: completed\ndata: {\"request_id\":\"req-test\",\"message\":{\"ID\":\"message-1\",\"Content\":\"ok\"}}\n\n")
	default:
		m.t.Fatalf("unexpected action %q", action)
	}
}

func writeE2EResult(w http.ResponseWriter, result string) {
	_, _ = fmt.Fprintf(w, `{"ResponseMetadata":{"RequestId":"req-test"},"Result":%s}`, result)
}

// runCLI 在新创建的 root 上执行一次命令；--config-file 指向一个空文件以
// 屏蔽宿主机上 ~/.hibot/config.yaml 可能带来的污染。TOP 连接配置通过
// 环境变量传入（HIBOT_ENDPOINT 等），避免与子命令 local flag 同名冲突
// （例如 `mcps create --endpoint=...` 中的 --endpoint 是 MCP 服务地址，
// 与 root 持久化的 --endpoint 同名）。
func runCLI(t *testing.T, cfgPath string, args ...string) string {
	t.Helper()
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	full := append([]string{"--config-file=" + cfgPath}, args...)
	root.SetArgs(full)
	if err := root.Execute(); err != nil {
		t.Fatalf("hibot %v: %v\noutput: %s", args, err, buf.String())
	}
	return buf.String()
}

// TestCLIFullJourney_StreamingAndBatch 是 CLI 全链路 e2e，对应
// go/examples/e2e/e2e_test.go::TestFullJourney_StreamingAndBatch。
func TestCLIFullJourney_StreamingAndBatch(t *testing.T) {
	mock := newE2EMock(t)

	// TOP 连接配置走环境变量，避开 mcps 子命令本地 --endpoint 名字冲突。
	t.Setenv("HIBOT_ENDPOINT", mock.URL())
	t.Setenv("HIBOT_AK", "AK")
	t.Setenv("HIBOT_SK", "SK")
	t.Setenv("HIBOT_WORKSPACE_ID", "ws-1")
	t.Setenv("HIBOT_REGION", "cn")

	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte{}, 0o600); err != nil {
		t.Fatalf("write empty config: %v", err)
	}

	// 本地 fixtures：skill bundle + resource markdown。内容随意，server 端
	// 只验证 Filename 非空。
	skillFile := filepath.Join(tmp, "skill.zip")
	if err := os.WriteFile(skillFile, []byte("PK\x03\x04skill-bytes"), 0o600); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	resourceFile := filepath.Join(tmp, "runbook.md")
	if err := os.WriteFile(resourceFile, []byte("# runbook\nstep 1\n"), 0o600); err != nil {
		t.Fatalf("write resource: %v", err)
	}

	// Step 1: models get → GetModel
	out := runCLI(t, cfgPath, "--output=json", "models", "get", "model-1")
	if !strings.Contains(out, "model-1") {
		t.Fatalf("models get output missing model-1: %q", out)
	}

	// Step 2: prompts create → CreateAgentPromptTemplate
	out = runCLI(t, cfgPath, "--output=json", "prompts", "create",
		"--name=e2e-prompt", "--content=hello")
	if !strings.Contains(out, "prompt-1") {
		t.Fatalf("prompts create output missing prompt-1: %q", out)
	}

	// Step 3: skills upload → UploadBlob + CreateSkill
	out = runCLI(t, cfgPath, "--output=json", "skills", "upload",
		"--name=e2e-skill", "--version=1.0.0", "--file="+skillFile)
	if !strings.Contains(out, "skill-version-1") {
		t.Fatalf("skills upload output missing skill-version-1: %q", out)
	}

	// Step 4: resources create → UploadBlob + CreateResource
	out = runCLI(t, cfgPath, "--output=json", "resources", "create",
		"--name=e2e-resource", "--type=document_collection", "--file="+resourceFile)
	if !strings.Contains(out, "resource-1") {
		t.Fatalf("resources create output missing resource-1: %q", out)
	}

	// Step 5: mcps create → CreateMCP
	out = runCLI(t, cfgPath, "--output=json", "mcps", "create",
		"--name=e2e-mcp", "--endpoint=http://mcp.local/mcp", "--credential-name=tok")
	if !strings.Contains(out, "mcp-1") {
		t.Fatalf("mcps create output missing mcp-1: %q", out)
	}

	// Step 6: agents create —— 故意省略 --env-id，触发 SDK 的 Environments.Default
	// 自动解析（→ ListEnv），与 go SDK e2e 一致。
	out = runCLI(t, cfgPath, "--output=json", "agents", "create",
		"--name=e2e-agent",
		"--model-id=model-1",
		"--system=you are a test bot",
		"--skill-version-id=skill-version-1",
		"--mcp-id=mcp-1",
		"--resource-id=resource-1",
	)
	if !strings.Contains(out, "agent-1") {
		t.Fatalf("agents create output missing agent-1: %q", out)
	}
	body := mock.Body("CreateAgent")
	if got := body["ModelID"]; got != "model-1" {
		t.Fatalf("CreateAgent ModelID = %v, want model-1", got)
	}
	if got := body["EnvID"]; got != "env-1" {
		t.Fatalf("CreateAgent EnvID = %v, want env-1 (auto-resolved via ListEnv)", got)
	}
	if _, ok := body["Skills"].([]any); !ok {
		t.Fatalf("CreateAgent Skills missing or wrong type: %#v", body["Skills"])
	}
	if _, ok := body["MCPs"].([]any); !ok {
		t.Fatalf("CreateAgent MCPs missing or wrong type: %#v", body["MCPs"])
	}
	if _, ok := body["Resources"].(map[string]any); !ok {
		t.Fatalf("CreateAgent Resources missing or wrong shape: %#v", body["Resources"])
	}

	// Step 7: sessions create → CreateSession（不传 peer，走 SDK 默认 webchat 兜底）
	out = runCLI(t, cfgPath, "--output=json", "sessions", "create",
		"--agent-id=agent-1")
	if !strings.Contains(out, "session-1") {
		t.Fatalf("sessions create output missing session-1: %q", out)
	}

	// Step 8: streaming chat —— 必须能写出 delta 文本 "ok" 与 completed marker。
	out = runCLI(t, cfgPath, "chat", "session-1",
		"--stream", "--input=streaming hello")
	if !strings.Contains(out, "ok") || !strings.Contains(out, "[completed message_id=message-1]") {
		t.Fatalf("streaming chat output: %q", out)
	}

	// Step 9: batch chat —— SDK 内部消费完整 SSE 流并返回 final message。
	// 用 --output=json 方便断言 message-1 字段存在。
	out = runCLI(t, cfgPath, "--output=json", "chat", "session-1",
		"--input=batch hello")
	if !strings.Contains(out, "message-1") {
		t.Fatalf("batch chat output missing message-1: %q", out)
	}

	// Step 10: 全链路 Action 命中校验（缺一不可）。
	mock.requireActions(
		"GetModel",
		"CreateAgentPromptTemplate",
		"UploadBlob",
		"CreateSkill",
		"CreateResource",
		"CreateMCP",
		"ListEnv",
		"CreateAgent",
		"CreateSession",
		"Chat",
	)
}
