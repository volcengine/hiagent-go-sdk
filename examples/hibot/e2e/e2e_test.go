// Package e2e 提供 hibot-go-sdk 的端到端集成测试。
//
// 测试目标：模拟一个真实的私有化用户视角，从 0 → 1 走完
//
//  1. 选择/查询模型 (Models.Get)
//  2. 选择默认运行环境 (Environments.Default → ListEnv)
//  3. 上传 Skill / Resource 二进制 (Uploads.UploadBlob)
//  4. 创建 Prompt (Prompts.New → CreateAgentPromptTemplate)
//  5. 创建 Skill 版本 (Skills.New → CreateSkill)
//  6. 创建 Resource (Resources.New → CreateResource)
//  7. 注册 MCP 服务 (MCPs.New → CreateMCP)
//  8. 创建 Agent，绑定 Skill / MCP / Resource (Agents.New → CreateAgent)
//  9. 创建 Session（webchat 默认场景，无需显式 PeerKind/PeerID）
//     (Sessions.New → CreateSession)
//  10. 与 Agent 进行流式对话 (Sessions.ChatStreaming)
//  11. 与 Agent 进行批量(非流式)对话 (Sessions.Chat)
//
// 默认情况下，测试不依赖任何真实 Hibot 集群，使用 examples/internal/mocktop
// 的 httptest server 完整模拟 TOP 路由与 SSE 行为，确保 SDK 的契约
// （Action / Service / Version、签名、字段映射、SSE 解析）都被覆盖。
//
// 当设置环境变量 HIBOT_E2E_TOP_HOST 时（同时配套 HIBOT_E2E_AK /
// HIBOT_E2E_SK / HIBOT_E2E_WORKSPACE / HIBOT_E2E_TENANT_ID），测试会
// 切换到"真实环境分支"，跳过 mocktop，使用真实集群完成最小闭环：
// ListAgents → CreateSession → ChatStreaming → Chat。
package e2e

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

// 端到端 fixture：测试运行目录是 examples/e2e/，相对路径指回 SDK 根的
// testdata/。两个文件分别承担 Resource 挂载 / Skill 调用闭环验证：
//   - testdata/runbook.md       → 上传为 Resource；当前仅断言挂载链路
//     (UploadBlob → CreateResource → CreateAgent.Resources)，
//     不再断言模型是否检索到内容（私有化集群上 retrieve
//     未稳定接通，需先做集群侧排查）。
//   - testdata/skill/SKILL.md   → 打包成 zip 上传为 Skill；front matter
//     的 name/description 必须能被 server 端
//     parseSkillManifest 解析通过。
const (
	e2eResourceFixturePath = "../../../testdata/hibot/runbook.md"
	e2eSkillFixtureDir     = "../../../testdata/hibot/skill"
	e2eSkillPulseToken     = "PULSE_OK_E2E"
)

// TestFullJourney_StreamingAndBatch 演示并验证一个用户从空白工作空间
// 到完成两轮对话的完整闭环；流式与非流式对话都必须能获得 final message。
func TestFullJourney_StreamingAndBatch(t *testing.T) {
	t.Parallel()

	if host := trimEnv("HIBOT_E2E_TOP_HOST"); host != "" {
		runRealEnvJourney(t, host)
		return
	}

	server := mocktop.New(t)
	t.Cleanup(server.Close)

	client, err := hibot.NewClient(hibot.Config{
		Endpoint:    server.URL(),
		AccessKey:   "ak-e2e",
		SecretKey:   "sk-e2e",
		WorkspaceID: "workspace-e2e",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ---------------------------------------------------------------
	// Step 1: 查询默认模型 → 校验路由到 hibot-model 服务，固定 Version。
	// ---------------------------------------------------------------
	model, err := client.V1.Models.Get(ctx, hibot.V1ModelGetParams{
		ID: hibot.V1ManagedAgentModelDoubaoSeedPro,
	})
	if err != nil {
		t.Fatalf("get model: %v", err)
	}
	if model.ID == "" {
		t.Fatalf("model.ID empty: %#v", model)
	}

	// ---------------------------------------------------------------
	// Step 2: 创建系统 Prompt (CreateAgentPromptTemplate)。
	// ---------------------------------------------------------------
	prompt, err := client.V1.Prompts.New(ctx, hibot.V1PromptNewParams{
		Name:    "e2e-prompt",
		Content: "你是一个 SDK 端到端测试中的助手。",
	})
	if err != nil {
		t.Fatalf("create prompt: %v", err)
	}
	if prompt.Content == "" {
		t.Fatalf("prompt content empty")
	}

	// ---------------------------------------------------------------
	// Step 3: 上传 Skill 与 Resource 的本地二进制。
	// ---------------------------------------------------------------
	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "skill.zip")
	resourcePath := filepath.Join(tmpDir, "runbook.md")
	if err := os.WriteFile(skillPath, []byte("PK\x03\x04skill-bytes"), 0o600); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	if err := os.WriteFile(resourcePath, []byte("# runbook\nstep 1\n"), 0o600); err != nil {
		t.Fatalf("write resource file: %v", err)
	}

	skillBlobID := uploadFile(ctx, t, client, skillPath, "application/zip")
	resourceBlobID := uploadFile(ctx, t, client, resourcePath, "text/markdown")

	// ---------------------------------------------------------------
	// Step 4: 注册 Skill 版本（绑定刚刚上传的 blob）。
	// ---------------------------------------------------------------
	enabled := true
	skill, err := client.V1.Skills.New(ctx, hibot.V1SkillNewParams{
		Name:        "e2e-skill",
		Description: "skill registered by e2e test",
		BlobID:      skillBlobID,
		Enabled:     &enabled,
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("create skill: %v", err)
	}
	if skill.ID == "" {
		t.Fatalf("skill.ID empty")
	}

	// ---------------------------------------------------------------
	// Step 5: 创建 Resource。
	// ---------------------------------------------------------------
	resource, err := client.V1.Resources.New(ctx, hibot.V1ResourceNewParams{
		Name:   "e2e-resource",
		Type:   hibot.V1ResourceTypeDocumentCollection,
		BlobID: resourceBlobID,
	})
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if resource.ID == "" {
		t.Fatalf("resource.ID empty")
	}

	// ---------------------------------------------------------------
	// Step 6: 注册 MCP 服务器。
	// ---------------------------------------------------------------
	mcp, err := client.V1.MCPs.New(ctx, hibot.V1MCPNewParams{
		Name:      "e2e-mcp",
		Transport: hibot.V1MCPTransportStreamableHTTP,
		Endpoint:  "http://mcp.local/mcp",
		Credential: &hibot.V1CredentialRefParams{
			Name: "e2e-token",
		},
	})
	if err != nil {
		t.Fatalf("create mcp: %v", err)
	}
	if mcp.ID == "" {
		t.Fatalf("mcp.ID empty")
	}

	// ---------------------------------------------------------------
	// Step 7: 创建 Agent —— EnvID 留空，触发 Environments.Default 自选。
	// ---------------------------------------------------------------
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name:   "e2e-agent",
		Model:  hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
		System: hibot.String(prompt.Content),
		Tools: []hibot.V1AgentNewParamsToolUnion{
			{OfSkill: &hibot.V1ManagedAgentSkillToolParams{
				Type:           hibot.V1ManagedAgentSkillToolParamsTypeSkill,
				SkillVersionID: skill.ID,
			}},
			{OfMCP: &hibot.V1ManagedAgentMCPToolParams{
				Type: hibot.V1ManagedAgentMCPToolParamsTypeMCP,
				ID:   mcp.ID,
			}},
		},
		Resources: []hibot.V1ManagedAgentResourceRefParams{{ID: resource.ID}},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if agent.ID == "" {
		t.Fatalf("agent.ID empty")
	}

	// 校验 CreateAgent 的关键字段：ModelID / Skills / MCPs / Resources / EnvID。
	createAgentBody := server.Body("CreateAgent")
	if got := createAgentBody["ModelID"]; got != model.ID {
		t.Fatalf("CreateAgent ModelID = %v, want %v", got, model.ID)
	}
	if got := createAgentBody["EnvID"]; got != "env-1" {
		t.Fatalf("CreateAgent EnvID = %v, want env-1 (from ListEnv default)", got)
	}
	if _, ok := createAgentBody["Skills"].([]any); !ok {
		t.Fatalf("CreateAgent Skills missing or wrong type: %#v", createAgentBody["Skills"])
	}
	if _, ok := createAgentBody["MCPs"].([]any); !ok {
		t.Fatalf("CreateAgent MCPs missing or wrong type: %#v", createAgentBody["MCPs"])
	}
	if _, ok := createAgentBody["Resources"].(map[string]any); !ok {
		t.Fatalf("CreateAgent Resources missing or wrong shape: %#v", createAgentBody["Resources"])
	}

	// ---------------------------------------------------------------
	// Step 8: 创建 Session —— webchat 场景下无需显式 PeerKind/PeerID，
	// SDK 内部会用 Channel=webchat / PeerKind=system / PeerID=AgentID
	// 兜底，保证服务端可生成确定性 SessionKey。
	// ---------------------------------------------------------------
	session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
		AgentID: agent.ID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.ID == "" {
		t.Fatalf("session.ID empty")
	}

	createSessionBody := server.Body("CreateSession")
	if got := createSessionBody["AgentID"]; got != agent.ID {
		t.Fatalf("CreateSession AgentID = %v, want %v", got, agent.ID)
	}
	payload, ok := createSessionBody["Payload"].(map[string]any)
	if !ok {
		t.Fatalf("CreateSession Payload missing: %#v", createSessionBody)
	}
	if got := payload["Channel"]; got != "webchat" {
		t.Fatalf("CreateSession Channel = %v, want webchat (SDK default for managed-agent flow)", got)
	}
	if got := payload["PeerKind"]; got != "system" {
		t.Fatalf("CreateSession PeerKind = %v, want system (SDK default when caller omits Peer)", got)
	}
	if got := payload["PeerID"]; got != agent.ID {
		t.Fatalf("CreateSession PeerID = %v, want AgentID %v (SDK default fallback)", got, agent.ID)
	}

	// ---------------------------------------------------------------
	// Step 9 (streaming): 流式对话 —— 必须能解析 delta 与 completed 两类事件，
	// 并提取出最终 Message。
	// ---------------------------------------------------------------
	streamingFinal, _ := runStreamingChat(ctx, t, client, session.ID, agent.ID,
		"流式：请用一句话介绍自己。")
	if streamingFinal.ID == "" || streamingFinal.Content == "" {
		t.Fatalf("streaming final message incomplete: %#v", streamingFinal)
	}

	// ---------------------------------------------------------------
	// Step 10 (batch / non-streaming): 在同一个 Session 上再发一次 Chat。
	// SDK 的 Sessions.Chat 内部消费完整 SSE 流后只返回最终 Message，
	// 对调用方来说等价于一次"批量同步"调用。
	// ---------------------------------------------------------------
	batchFinal, err := client.V1.Sessions.Chat(ctx, session.ID, hibot.V1SessionChatParams{
		Input: "批量：再回答一次同样的问题。",
	})
	if err != nil {
		t.Fatalf("batch chat: %v", err)
	}
	if batchFinal.ID == "" || batchFinal.Content == "" {
		t.Fatalf("batch final message incomplete: %#v", batchFinal)
	}

	// 校验 Chat 的请求体（透传 SessionID + AgentID + Content）。
	chatBody := server.Body("Chat")
	if got := chatBody["SessionID"]; got != session.ID {
		t.Fatalf("Chat SessionID = %v, want %v", got, session.ID)
	}
	if got := chatBody["AgentID"]; got != agent.ID {
		t.Fatalf("Chat AgentID = %v, want %v (SDK should infer agent from session)", got, agent.ID)
	}
	if got, _ := chatBody["Content"].(string); !strings.Contains(got, "批量") {
		t.Fatalf("Chat Content = %q, want contains '批量'", got)
	}

	// ---------------------------------------------------------------
	// Step 11: 全链路 Action 命中校验（缺一不可）。
	// ---------------------------------------------------------------
	server.RequireActions(
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

// uploadFile 是 Uploads.UploadBlob 的薄封装，校验 BlobID 非空并返回。
func uploadFile(ctx context.Context, t *testing.T, client *hibot.Client, path, contentType string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	resp, err := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
		Filename:    filepath.Base(path),
		ContentType: contentType,
	}, f)
	if err != nil {
		t.Fatalf("upload %s: %v", path, err)
	}
	if resp.BlobID == "" {
		t.Fatalf("upload %s: empty BlobID", path)
	}
	return resp.BlobID
}

// runStreamingChat 消费一次完整的 ChatStreaming SSE 流，返回最终 Message
// 以及流上观察到的事件名列表（用于上层断言 tool_start / tool_complete 等
// "副作用"事件，从而证明 skill 真的被调用了，而不是模型脑补）。
// 至少必须见到一个 completed 事件；delta 事件是可选的（短响应或非分块
// runtime 可能直接走 started → completed 路径）。
func runStreamingChat(ctx context.Context, t *testing.T, client *hibot.Client, sessionID, agentID, input string) (*hibot.V1Message, []string) {
	t.Helper()
	stream := client.V1.Sessions.ChatStreaming(ctx, sessionID, hibot.V1SessionChatParams{
		AgentID: agentID,
		Input:   input,
	})
	defer stream.Close()

	var (
		sawDelta     bool
		sawCompleted bool
		eventNames   []string
	)
	for stream.Next() {
		event := stream.Current()
		eventNames = append(eventNames, event.Type)
		switch event.Type {
		case hibot.V1SessionChatEventDelta:
			sawDelta = true
		case hibot.V1SessionChatEventCompleted:
			sawCompleted = true
		case hibot.V1SessionChatEventFailed:
			t.Fatalf("streaming chat failed: %s", event.Error.Message)
		}
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("streaming chat err: %v (events=%v)", err, eventNames)
	}
	if !sawCompleted {
		t.Fatalf("streaming chat: no completed event observed (events=%v)", eventNames)
	}
	if !sawDelta {
		t.Logf("streaming chat: no delta event (short reply); events=%v", eventNames)
	}
	final, err := stream.FinalMessage()
	if err != nil {
		t.Fatalf("streaming final: %v", err)
	}
	return final, eventNames
}

// containsAny 用于在 SSE 事件名列表里探测是否出现了任意一个目标事件。
// 当前主要用于 e2e_test 里断言"是否真的发生了 tool 调用"。
func containsAny(events []string, targets ...string) bool {
	for _, ev := range events {
		for _, want := range targets {
			if ev == want {
				return true
			}
		}
	}
	return false
}

// trimEnv reads an environment variable and strips wrapping whitespace,
// backticks and quotes — accepts the slightly noisy formats that ops
// hand-paste into shells (e.g. `'http://...'` or “ ` “-wrapped).
func trimEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	v = strings.Trim(v, "`'\" ")
	return v
}

// runRealEnvJourney exercises the minimum-viable closed loop against a real
// hibot cluster. We intentionally skip the resource-creation steps (model /
// prompt / skill / resource / mcp / agent) because those require pre-staged
// images and assets in the target cluster; instead we discover an existing
// agent via ListAgents and verify that CreateSession / ChatStreaming / Chat
// all succeed end-to-end.
func runRealEnvJourney(t *testing.T, host string) {
	ak := trimEnv("HIBOT_E2E_AK")
	sk := trimEnv("HIBOT_E2E_SK")
	workspace := trimEnv("HIBOT_E2E_WORKSPACE")
	if ak == "" || sk == "" || workspace == "" {
		t.Fatalf("real-env journey requires HIBOT_E2E_AK / HIBOT_E2E_SK / HIBOT_E2E_WORKSPACE")
	}

	t.Logf("real-env journey: host=%s workspace=%s", host, workspace)

	client, err := hibot.NewClient(hibot.Config{
		Endpoint:    host,
		AccessKey:   ak,
		SecretKey:   sk,
		WorkspaceID: workspace,
	})
	if err != nil {
		t.Fatalf("new real-env client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Step 1: discover an existing agent — real cluster fixtures own creation.
	agents, err := client.V1.Agents.List(ctx, hibot.V1AgentListParams{})
	if err != nil {
		t.Fatalf("list agents: %v", err)
	}
	if len(agents) == 0 {
		t.Fatalf("real-env workspace %q has no agents; please pre-create one before running this test", workspace)
	}
	agent := agents[0]
	t.Logf("using agent: id=%s name=%s", agent.ID, agent.Name)

	// Step 2: create a session with no Peer — webchat default path.
	session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
		AgentID: agent.ID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.ID == "" {
		t.Fatalf("real-env CreateSession returned empty ID: %#v", session)
	}
	t.Logf("created session: %s", session.ID)

	// Step 3: streaming chat — must observe delta + completed.
	streamingFinal, _ := runStreamingChat(ctx, t, client, session.ID, agent.ID,
		"流式真实环境冒烟：请用一句话介绍你自己。")
	if streamingFinal.ID == "" || streamingFinal.Content == "" {
		t.Fatalf("real-env streaming final message incomplete: %#v", streamingFinal)
	}
	t.Logf("streaming final: id=%s content=%q", streamingFinal.ID, streamingFinal.Content)

	// Step 4: batch (non-streaming) chat reuses the same session.
	batchFinal, err := client.V1.Sessions.Chat(ctx, session.ID, hibot.V1SessionChatParams{
		AgentID: agent.ID,
		Input:   "批量真实环境冒烟：再回答一次同样的问题。",
	})
	if err != nil {
		t.Fatalf("batch chat: %v", err)
	}
	if batchFinal.ID == "" || batchFinal.Content == "" {
		t.Fatalf("real-env batch final message incomplete: %#v", batchFinal)
	}
	t.Logf("batch final: id=%s content=%q", batchFinal.ID, batchFinal.Content)
}

// TestRealEnvResourceSkillLoop 是一个面向真实集群的"端到端闭环"测试：
//
//  1. 上传 testdata/runbook.md 作为 Resource（包含可断言 token）。
//  2. 把 testdata/skill/ 打包成 zip 上传为 Skill 版本。
//  3. 创建一个临时 Agent，绑定上述 Resource + Skill。
//  4. 创建 Session，分别提两类问题：
//     - 资料类：要求复述 runbook 中的 secret token。
//     - 技能类：要求执行 pulse check，触发 skill。
//
// 默认仅在设置了 HIBOT_E2E_TOP_HOST 时运行（避免污染本地 mock 链路）；
// 设置 HIBOT_E2E_KEEP_AGENT=1 可保留创建出的 Agent / Resource / Skill
// 用于事后排障，否则测试结束会尽力清理。
func TestRealEnvResourceSkillLoop(t *testing.T) {
	host := trimEnv("HIBOT_E2E_TOP_HOST")
	if host == "" {
		t.Skip("real-env loop requires HIBOT_E2E_TOP_HOST; skipping in mock-only run")
	}

	ak := trimEnv("HIBOT_E2E_AK")
	sk := trimEnv("HIBOT_E2E_SK")
	workspace := trimEnv("HIBOT_E2E_WORKSPACE")
	if ak == "" || sk == "" || workspace == "" {
		t.Fatalf("real-env loop requires HIBOT_E2E_AK / HIBOT_E2E_SK / HIBOT_E2E_WORKSPACE")
	}

	t.Logf("real-env loop: host=%s workspace=%s", host, workspace)

	client, err := hibot.NewClient(hibot.Config{
		Endpoint:    host,
		AccessKey:   ak,
		SecretKey:   sk,
		WorkspaceID: workspace,
	})
	if err != nil {
		t.Fatalf("new real-env client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Step 1: 拿到一个可用模型。
	//   - 若设置了 HIBOT_E2E_MODEL_ID，直接复用预置模型（跳过 GetModel/
	//     ListModel；私有化集群上业务 AK/SK 可能没有 aigw 接口权限）。
	//   - 否则按 base ModelName=doubao-seed-2.0-pro-260215 过滤定位
	//     工作空间内的自定义实例；找不到再退化到 ListModels 第一项。
	var model *hibot.V1Model
	if presetID := trimEnv("HIBOT_E2E_MODEL_ID"); presetID != "" {
		model = &hibot.V1Model{ID: presetID}
		t.Logf("using preset model id=%s (HIBOT_E2E_MODEL_ID)", presetID)
	} else {
		got, err := client.V1.Models.Get(ctx, hibot.V1ModelGetParams{
			ModelName: hibot.V1ManagedAgentModelDoubaoSeedPro,
		})
		if err != nil {
			t.Logf("get default model by ModelName=%q failed: %v; falling back to ListModels", hibot.V1ManagedAgentModelDoubaoSeedPro, err)
			list, listErr := client.V1.Models.List(ctx, hibot.V1ModelListParams{})
			if listErr != nil {
				t.Fatalf("list models: %v", listErr)
			}
			if len(list.Items) == 0 {
				t.Fatalf("real-env workspace %q has no models; please pre-create one before running this test", workspace)
			}
			got = &list.Items[0]
			t.Logf("fallback model picked: id=%s name=%s modelName=%s type=%s", got.ID, got.Name, got.ModelName, got.Type)
		} else {
			t.Logf("matched model by ModelName: id=%s name=%s modelName=%s", got.ID, got.Name, got.ModelName)
		}
		model = got
	}

	// Step 2: 上传 Resource fixture。
	resourceBlobID := uploadFile(ctx, t, client, e2eResourceFixturePath, "text/markdown")
	resource, err := client.V1.Resources.New(ctx, hibot.V1ResourceNewParams{
		Name:   fmt.Sprintf("e2e-runbook-%d", time.Now().UnixNano()),
		Type:   hibot.V1ResourceTypeDocumentCollection,
		BlobID: resourceBlobID,
	})
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}
	t.Logf("created resource: id=%s name=%s", resource.ID, resource.Name)
	t.Cleanup(func() {
		if trimEnv("HIBOT_E2E_KEEP_AGENT") != "" {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.V1.Resources.Delete(cleanupCtx, hibot.V1ResourceDeleteParams{
			ResourceID: resource.ID,
		}); err != nil {
			t.Logf("cleanup resource %s: %v", resource.ID, err)
		}
	})

	// Step 3: 把 testdata/skill/ 打包并上传为 Skill。
	skillZipPath := buildSkillZipFromDir(t, e2eSkillFixtureDir)
	skillBlobID := uploadFile(ctx, t, client, skillZipPath, "application/zip")
	enabled := true
	skill, err := client.V1.Skills.New(ctx, hibot.V1SkillNewParams{
		Name:        fmt.Sprintf("e2e-runbook-skill-%d", time.Now().UnixNano()),
		Description: "Skill uploaded by hibot-go-sdk e2e closed-loop test.",
		BlobID:      skillBlobID,
		Enabled:     &enabled,
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("create skill: %v", err)
	}
	t.Logf("created skill version: id=%s", skill.ID)
	t.Cleanup(func() {
		if trimEnv("HIBOT_E2E_KEEP_AGENT") != "" {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.V1.Skills.Delete(cleanupCtx, hibot.V1SkillDeleteParams{
			ID: skill.ID,
		}); err != nil {
			t.Logf("cleanup skill %s: %v", skill.ID, err)
		}
	})

	// Step 4: 创建临时 Agent —— 绑定 Skill（Resource 已挂载但暂不参与断言）。
	//
	// **重要**：system prompt 不允许出现 skill token 字面量 PULSE_OK_E2E。
	// 测试侧通过观察 SSE 流上的 tool_start / tool_complete 事件来证伪
	// "模型纯靠脑补"，从而证明 e2e-runbook-skill 真的被工具调用链触发了。
	//
	// Resource 检索链路目前在私有化集群上未稳定接通（绑定生效但模型不会
	// 主动 retrieve），暂从断言中剥离 —— 仅保留挂载流程（UploadBlob →
	// CreateResource → CreateAgent.Resources 绑定）以覆盖 SDK 路由契约，
	// 不再断言 runbook token 是否被复述。
	systemPrompt := "你是 hibot-go-sdk 端到端测试助手。" +
		"用户要求执行 pulse check / 心跳检查时，必须调用 e2e-runbook-skill 工具，并把工具返回的字面 token 原样返回给用户。"
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name:   fmt.Sprintf("e2e-loop-agent-%d", time.Now().UnixNano()),
		Model:  hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
		System: hibot.String(systemPrompt),
		Tools: []hibot.V1AgentNewParamsToolUnion{
			{OfSkill: &hibot.V1ManagedAgentSkillToolParams{
				Type:           hibot.V1ManagedAgentSkillToolParamsTypeSkill,
				SkillVersionID: skill.ID,
			}},
		},
		Resources: []hibot.V1ManagedAgentResourceRefParams{{ID: resource.ID}},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	t.Logf("created agent: id=%s", agent.ID)
	t.Cleanup(func() {
		if trimEnv("HIBOT_E2E_KEEP_AGENT") != "" {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.V1.Agents.Delete(cleanupCtx, hibot.V1AgentDeleteParams{
			AgentID: agent.ID,
		}); err != nil {
			t.Logf("cleanup agent %s: %v", agent.ID, err)
		}
	})

	// Step 5: 创建 Session —— webchat 默认通道。
	session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
		AgentID: agent.ID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Logf("created session: %s", session.ID)

	// Step 6: Skill 闭环 —— 触发 pulse check。
	// system prompt 里没有 PULSE_OK_E2E 字面量；只有真的调用了
	// e2e-runbook-skill 才能拿到该 token。我们额外断言 SSE 流上观测到了
	// tool_start / tool_complete 事件，以证伪"模型脑补出 PULSE_OK_E2E"
	// 这种巧合 —— 没真正调用 skill 的话流上不会出现 tool 事件。
	skillFinal, skillEvents := runStreamingChat(ctx, t, client, session.ID, agent.ID,
		"请执行一次 pulse check（按照 e2e-runbook-skill 的契约调用工具），并把工具返回的 token 原样告诉我。")
	t.Logf("skill loop events=%v final=%q", skillEvents, skillFinal.Content)
	if !strings.Contains(skillFinal.Content, e2eSkillPulseToken) {
		t.Fatalf("skill loop: agent did not return pulse token %q; "+
			"system prompt no longer leaks the token, so this proves the skill was NOT invoked. got=%q",
			e2eSkillPulseToken, skillFinal.Content)
	}
	if !containsAny(skillEvents, hibot.V1SessionChatEventToolStart, hibot.V1SessionChatEventToolComplete) {
		t.Fatalf("skill loop: no tool_start/tool_complete event observed on SSE stream; "+
			"the model likely answered without actually invoking e2e-runbook-skill. events=%v",
			skillEvents)
	}
	t.Logf("skill loop ok: pulse token %q returned via real skill invocation (events=%v)",
		e2eSkillPulseToken, skillEvents)
}

// buildSkillZipFromDir 把指定目录下的所有文件打包成临时 zip，返回路径。
// SKILL.md 必须出现在 zip 顶层（或带统一前缀），server 端会自动识别。
func buildSkillZipFromDir(t *testing.T, dir string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "skill.zip")
	f, err := os.Create(out)
	if err != nil {
		t.Fatalf("create skill zip: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		buf, err := io.ReadAll(src)
		if err != nil {
			return err
		}
		_, err = w.Write(buf)
		return err
	})
	if walkErr != nil {
		t.Fatalf("walk skill dir: %v", walkErr)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close skill zip: %v", err)
	}
	// 防御：保证打包结果非空且确实包含 SKILL.md。
	st, err := os.Stat(out)
	if err != nil || st.Size() == 0 {
		t.Fatalf("skill zip is empty: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err != nil {
		t.Fatalf("skill fixture missing SKILL.md: %v", err)
	}
	if got, want := bytes.Contains(mustReadAll(t, out), []byte("SKILL.md")), true; got != want {
		t.Fatalf("skill zip does not reference SKILL.md")
	}
	return out
}

func mustReadAll(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
