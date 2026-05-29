package hibot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 验证创建 MCP 时 CredentialConfig 中的 SecretValue 字段被正确序列化到请求 body。
func TestCreateMCPSerializesCredentialConfigSecretValue(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "CreateMCP" {
			t.Fatalf("Action = %q, want CreateMCP", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"Result":{"ID":"mcp-1"}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.V1.MCPs.New(context.Background(), V1MCPNewParams{
		WorkspaceID: "ws-1",
		Name:        "demo",
		Transport:   V1MCPTransportStreamableHTTP,
		Endpoint:    "https://example.com/mcp",
		CredentialConfig: &V1MCPCredentialInputParams{
			Name:         "demo-cred",
			ProviderType: "basic",
			Secrets: []V1CredentialSecretInputParams{
				{KeyName: "token", SecretType: "string", SecretValue: "s3cr3t"},
			},
		},
	})
	if err != nil {
		t.Fatalf("create mcp: %v", err)
	}

	cfg, ok := captured["CredentialConfig"].(map[string]any)
	if !ok {
		t.Fatalf("CredentialConfig missing or wrong type: %#v", captured["CredentialConfig"])
	}
	secrets, ok := cfg["Secrets"].([]any)
	if !ok || len(secrets) != 1 {
		t.Fatalf("Secrets missing: %#v", cfg["Secrets"])
	}
	first := secrets[0].(map[string]any)
	if first["SecretValue"] != "s3cr3t" {
		t.Fatalf("SecretValue = %v, want s3cr3t", first["SecretValue"])
	}
	if first["KeyName"] != "token" {
		t.Fatalf("KeyName = %v, want token", first["KeyName"])
	}
	// 旧字段名 Value 不应再出现
	if _, exists := first["Value"]; exists {
		t.Fatalf("legacy Value field should not be present: %#v", first)
	}
}

// 验证更新 MCP 时 CredentialConfig 通过手工构建的 body 透传。
func TestUpdateMCPPassesCredentialConfig(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "UpdateMCP" {
			t.Fatalf("Action = %q, want UpdateMCP", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"Result":{}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	if err := client.V1.MCPs.Update(context.Background(), V1MCPUpdateParams{
		WorkspaceID: "ws-1",
		ID:          "mcp-1",
		CredentialConfig: &V1MCPCredentialInputParams{
			Name: "demo-cred",
			Secrets: []V1CredentialSecretInputParams{
				{KeyName: "token", SecretValue: "new-secret"},
			},
		},
	}); err != nil {
		t.Fatalf("update mcp: %v", err)
	}

	cfg, ok := captured["CredentialConfig"].(map[string]any)
	if !ok {
		t.Fatalf("CredentialConfig missing: %#v", captured)
	}
	secrets := cfg["Secrets"].([]any)
	first := secrets[0].(map[string]any)
	if first["SecretValue"] != "new-secret" {
		t.Fatalf("SecretValue = %v, want new-secret", first["SecretValue"])
	}
}
