package hibot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientInjectsWorkspaceAndRoutesAction(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "CreateAgent" {
			t.Fatalf("Action = %q, want CreateAgent", got)
		}
		if got := r.URL.Query().Get("Version"); got != defaultServerVersion {
			t.Fatalf("Version = %q, want %s", got, defaultServerVersion)
		}
		if got := r.Header.Get("X-Top-Service"); got != defaultServerService {
			t.Fatalf("X-Top-Service = %q, want %s", got, defaultServerService)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["WorkspaceID"]; got != "workspace-ops-prod" {
			t.Fatalf("WorkspaceID = %v, want workspace-ops-prod", got)
		}
		if got := body["ModelID"]; got != V1ManagedAgentModelDoubaoSeedPro {
			t.Fatalf("ModelID = %v, want %s", got, V1ManagedAgentModelDoubaoSeedPro)
		}
		if got := body["EnvID"]; got != "env-1" {
			t.Fatalf("EnvID = %v, want env-1", got)
		}
		if _, ok := body["Skills"].([]any); !ok {
			t.Fatalf("Skills missing or wrong type: %#v", body["Skills"])
		}
		if _, ok := body["MCPs"].([]any); !ok {
			t.Fatalf("MCPs missing or wrong type: %#v", body["MCPs"])
		}

		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"ID":"agent-1"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	agent, err := client.V1.Agents.New(context.Background(), V1AgentNewParams{
		Name:  "ops-troubleshooter",
		EnvID: "env-1",
		Model: V1ManagedAgentModelConfigParams{ID: V1ManagedAgentModelDoubaoSeedPro},
		Tools: []V1AgentNewParamsToolUnion{
			{OfSkill: &V1ManagedAgentSkillToolParams{Type: V1ManagedAgentSkillToolParamsTypeSkill, SkillVersionID: "skill-version-1"}},
			{OfMCP: &V1ManagedAgentMCPToolParams{Type: V1ManagedAgentMCPToolParamsTypeMCP, ID: "mcp-1"}},
		},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if agent.ID != "agent-1" {
		t.Fatalf("agent.ID = %q, want agent-1", agent.ID)
	}
}

func TestRequestWorkspaceOverridesClientDefault(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["WorkspaceID"]; got != "workspace-staging" {
			t.Fatalf("WorkspaceID = %v, want workspace-staging", got)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"ID":"agent-2"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.V1.Agents.New(context.Background(), V1AgentNewParams{
		WorkspaceID: "workspace-staging",
		Name:        "staging-agent",
		EnvID:       "env-staging",
		Model:       V1ManagedAgentModelConfigParams{ID: V1ManagedAgentModelDoubaoSeedPro},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
}

func TestUploadBlobRoutesToUpService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "UploadBlob" {
			t.Fatalf("Action = %q, want UploadBlob", got)
		}
		if got := r.URL.Query().Get("Version"); got != defaultUpVersion {
			t.Fatalf("Version = %q, want %s", got, defaultUpVersion)
		}
		if got := r.URL.Query().Get("Filename"); got != "skill.zip" {
			t.Fatalf("Filename = %q, want skill.zip", got)
		}
		if got := r.Header.Get("X-Top-Service"); got != defaultUpService {
			t.Fatalf("X-Top-Service = %q, want %s", got, defaultUpService)
		}
		if got := r.Header.Get("Content-Type"); got != "application/zip" {
			t.Fatalf("Content-Type = %q, want application/zip", got)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"BlobID":"blob-1"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	blob, err := client.V1.Uploads.UploadBlob(context.Background(), V1UploadBlobParams{
		Filename:    "skill.zip",
		ContentType: "application/zip",
	}, strings.NewReader("zip bytes"))
	if err != nil {
		t.Fatalf("upload blob: %v", err)
	}
	if blob.BlobID != "blob-1" {
		t.Fatalf("BlobID = %q, want blob-1", blob.BlobID)
	}
}

func TestGetModelUsesFixedModelVersion(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "GetModel" {
			t.Fatalf("Action = %q, want GetModel", got)
		}
		if got := r.URL.Query().Get("Version"); got != defaultModelVersion {
			t.Fatalf("Version = %q, want %s", got, defaultModelVersion)
		}
		if got := r.Header.Get("X-Top-Service"); got != defaultModelService {
			t.Fatalf("X-Top-Service = %q, want %s", got, defaultModelService)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"Items":[{"ID":"model-1"}]}}`)
	}))
	defer server.Close()

	client, err := NewClient(Config{
		Endpoint:    server.URL,
		AccessKey:   "ak",
		SecretKey:   "sk",
		WorkspaceID: "workspace-ops-prod",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	model, err := client.V1.Models.Get(context.Background(), V1ModelGetParams{ID: "model-1"})
	if err != nil {
		t.Fatalf("get model: %v", err)
	}
	if model.ID != "model-1" {
		t.Fatalf("model.ID = %q, want model-1", model.ID)
	}
}

func TestEnvironmentDefaultSelectsEarliest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "ListEnv" {
			t.Fatalf("Action = %q, want ListEnv", got)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"Items":[{"ID":"env-new","CreatedAt":"2026-02-01T00:00:00Z"},{"ID":"env-old","CreatedAt":"2026-01-01T00:00:00Z"}]}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	env, err := client.V1.Environments.Default(context.Background(), V1EnvironmentListParams{})
	if err != nil {
		t.Fatalf("default env: %v", err)
	}
	if env.ID != "env-old" {
		t.Fatalf("env.ID = %q, want env-old", env.ID)
	}
}

func TestAgentsNewSelectsDefaultEnvironment(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("Action")
		seen[action] = true

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode %s body: %v", action, err)
		}
		switch action {
		case "ListEnv":
			if got := body["WorkspaceID"]; got != "workspace-ops-prod" {
				t.Fatalf("ListEnv WorkspaceID = %v, want workspace-ops-prod", got)
			}
			_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"Items":[{"ID":"env-new","CreatedAt":"2026-02-01T00:00:00Z"},{"ID":"env-old","CreatedAt":"2026-01-01T00:00:00Z"}]}}`)
		case "CreateAgent":
			if got := body["EnvID"]; got != "env-old" {
				t.Fatalf("CreateAgent EnvID = %v, want env-old", got)
			}
			_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-2"},"Result":{"ID":"agent-1"}}`)
		default:
			t.Fatalf("unexpected action %q", action)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	agent, err := client.V1.Agents.New(context.Background(), V1AgentNewParams{
		Name:  "ops-troubleshooter",
		Model: V1ManagedAgentModelConfigParams{ID: V1ManagedAgentModelDoubaoSeedPro},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if agent.ID != "agent-1" {
		t.Fatalf("agent.ID = %q, want agent-1", agent.ID)
	}
	for _, action := range []string{"ListEnv", "CreateAgent"} {
		if !seen[action] {
			t.Fatalf("action %s was not called", action)
		}
	}
}

func TestChatStreamingParsesSSE(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "Chat" {
			t.Fatalf("Action = %q, want Chat", got)
		}
		if got := r.URL.Query().Get("Version"); got != defaultChatVersion {
			t.Fatalf("Version = %q, want %s", got, defaultChatVersion)
		}
		if got := r.Header.Get("X-Top-Service"); got != defaultGatewayService {
			t.Fatalf("X-Top-Service = %q, want %s", got, defaultGatewayService)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("Accept = %q, want text/event-stream", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["SessionID"]; got != "session-1" {
			t.Fatalf("SessionID = %v, want session-1", got)
		}
		if got := body["AgentID"]; got != "agent-1" {
			t.Fatalf("AgentID = %v, want agent-1", got)
		}
		if got := body["Content"]; got != "hello" {
			t.Fatalf("Content = %v, want hello", got)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: message.started\ndata: {\"request_id\":\"req-1\"}\n\n")
		_, _ = fmt.Fprint(w, "event: delta\ndata: {\"delta\":{\"text\":\"hi\"}}\n\n")
		_, _ = fmt.Fprint(w, "event: completed\ndata: {\"message\":{\"ID\":\"msg-1\",\"Content\":\"hi\"}}\n\n")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	stream := client.V1.Sessions.ChatStreaming(context.Background(), "session-1", V1SessionChatParams{Input: "hello", AgentID: "agent-1"})
	defer stream.Close()

	var events []V1SessionChatEvent
	for stream.Next() {
		events = append(events, stream.Current())
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream err: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3", len(events))
	}
	if events[0].RequestID != "req-1" {
		t.Fatalf("request id = %q, want req-1", events[0].RequestID)
	}
	if events[1].Delta.Text != "hi" {
		t.Fatalf("delta text = %q, want hi", events[1].Delta.Text)
	}
	final, err := stream.FinalMessage()
	if err != nil {
		t.Fatalf("final message: %v", err)
	}
	if final.ID != "msg-1" {
		t.Fatalf("final.ID = %q, want msg-1", final.ID)
	}
}

func TestChatStreamingNormalizesLegacyMessageEvents(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: message.chunk\ndata: {\"request_id\":\"req-1\",\"message_id\":\"msg-1\",\"content\":\"hi\"}\n\n")
		_, _ = fmt.Fprint(w, "event: message.completed\ndata: {\"request_id\":\"req-1\",\"message_id\":\"msg-1\",\"content\":\"hi\"}\n\n")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	stream := client.V1.Sessions.ChatStreaming(context.Background(), "session-1", V1SessionChatParams{Input: "hello", AgentID: "agent-1"})
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("first event missing: %v", stream.Err())
	}
	if got := stream.Current().Type; got != V1SessionChatEventDelta {
		t.Fatalf("first event type = %q, want %q", got, V1SessionChatEventDelta)
	}
	if got := stream.Current().Delta.Text; got != "hi" {
		t.Fatalf("delta text = %q, want hi", got)
	}
	if !stream.Next() {
		t.Fatalf("second event missing: %v", stream.Err())
	}
	if got := stream.Current().Type; got != V1SessionChatEventCompleted {
		t.Fatalf("second event type = %q, want %q", got, V1SessionChatEventCompleted)
	}
	final, err := stream.FinalMessage()
	if err != nil {
		t.Fatalf("final message: %v", err)
	}
	if final.ID != "msg-1" || final.Content != "hi" {
		t.Fatalf("final = %#v, want msg-1/hi", final)
	}
}

func TestChatStreamingNormalizesLegacyMessageFailed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: message.failed\ndata: {\"request_id\":\"req-1\",\"code\":\"RuntimeProviderError\",\"message\":\"runtime failed\"}\n\n")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	stream := client.V1.Sessions.ChatStreaming(context.Background(), "session-1", V1SessionChatParams{Input: "hello", AgentID: "agent-1"})
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("event missing: %v", stream.Err())
	}
	event := stream.Current()
	if event.Type != V1SessionChatEventFailed {
		t.Fatalf("event type = %q, want %q", event.Type, V1SessionChatEventFailed)
	}
	if event.Error.Code != "RuntimeProviderError" || event.Error.Message != "runtime failed" {
		t.Fatalf("event error = %#v", event.Error)
	}
}

func TestResolveSkillVersionUsesSkillID(t *testing.T) {
	t.Parallel()

	var listVersionsBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("Action")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		switch action {
		case "ListSkills":
			if got := body["Name"]; got != "k8s-diagnose" {
				t.Fatalf("ListSkills Name = %v, want k8s-diagnose", got)
			}
			_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"Items":[{"SkillID":"skill-1","Name":"k8s-diagnose"}]}}`)
		case "ListSkillVersions":
			listVersionsBody = body
			_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-2"},"Result":{"Items":[{"ID":"skill-version-1","Version":"1.0.0"}]}}`)
		default:
			t.Fatalf("unexpected action %q", action)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	version, err := client.V1.Skills.ResolveVersion(context.Background(), V1SkillResolveVersionParams{
		Name:       "k8s-diagnose",
		Constraint: ">=1.0.0",
	})
	if err != nil {
		t.Fatalf("resolve version: %v", err)
	}
	if version.ID != "skill-version-1" {
		t.Fatalf("version.ID = %q, want skill-version-1", version.ID)
	}
	if got := listVersionsBody["SkillID"]; got != "skill-1" {
		t.Fatalf("ListSkillVersions SkillID = %v, want skill-1", got)
	}
}

func TestResolveSkillVersionByIDDoesNotCallAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected action %q", r.URL.Query().Get("Action"))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	version, err := client.V1.Skills.ResolveVersion(context.Background(), V1SkillResolveVersionParams{
		ID:         "skill-version-1",
		Name:       "k8s-diagnose",
		Constraint: ">=1.0.0",
	})
	if err != nil {
		t.Fatalf("resolve version: %v", err)
	}
	if version.ID != "skill-version-1" {
		t.Fatalf("version.ID = %q, want skill-version-1", version.ID)
	}
}

func TestCreateSkill(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "CreateSkill" {
			t.Fatalf("Action = %q, want CreateSkill", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["Name"]; got != "k8s-diagnose" {
			t.Fatalf("Name = %v, want k8s-diagnose", got)
		}
		if got := body["Source"]; got != "manual" {
			t.Fatalf("Source = %v, want manual", got)
		}
		if got := body["BlobID"]; got != "blob-1" {
			t.Fatalf("BlobID = %v, want blob-1", got)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"ID":"skill-version-1"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	enabled := true
	version, err := client.V1.Skills.New(context.Background(), V1SkillNewParams{
		Name:    "k8s-diagnose",
		BlobID:  "blob-1",
		Enabled: &enabled,
		Version: "1.0.0",
	})
	if err != nil {
		t.Fatalf("create skill: %v", err)
	}
	if version.ID != "skill-version-1" {
		t.Fatalf("version.ID = %q, want skill-version-1", version.ID)
	}
}

func TestResolveMCPByIDDoesNotCallAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected action %q", r.URL.Query().Get("Action"))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	mcp, err := client.V1.MCPs.Resolve(context.Background(), V1MCPResolveParams{
		ID:   "mcp-1",
		Name: "gitlab",
	})
	if err != nil {
		t.Fatalf("resolve mcp: %v", err)
	}
	if mcp.ID != "mcp-1" {
		t.Fatalf("mcp.ID = %q, want mcp-1", mcp.ID)
	}
}

func TestResolveMCPByName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "ListMCPs" {
			t.Fatalf("Action = %q, want ListMCPs", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["Keyword"]; got != "gitlab" {
			t.Fatalf("Keyword = %v, want gitlab", got)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"Items":[{"ID":"mcp-1","Name":"gitlab","Transport":"streamable_http","URL":"http://gitlab-mcp.internal:8080"}]}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	mcp, err := client.V1.MCPs.Resolve(context.Background(), V1MCPResolveParams{Name: "gitlab"})
	if err != nil {
		t.Fatalf("resolve mcp: %v", err)
	}
	if mcp.ID != "mcp-1" {
		t.Fatalf("mcp.ID = %q, want mcp-1", mcp.ID)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-error","Error":{"Code":"InvalidArgument","Message":"bad request"}}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.V1.Models.Get(context.Background(), V1ModelGetParams{ID: "missing"})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "InvalidArgument" || apiErr.RequestID != "req-error" {
		t.Fatalf("api error = %#v", apiErr)
	}
}

func newTestClient(t *testing.T, endpoint string) *Client {
	t.Helper()
	client, err := NewClient(Config{
		Endpoint:    endpoint,
		AccessKey:   "ak",
		SecretKey:   "sk",
		WorkspaceID: "workspace-ops-prod",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client
}
