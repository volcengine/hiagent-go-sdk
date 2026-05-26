package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func TestComprehensiveManagedAgentScenario(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t)
	dir := t.TempDir()
	skillFile := filepath.Join(dir, "skill.zip")
	resourceFile := filepath.Join(dir, "runbook.md")
	if err := os.WriteFile(skillFile, []byte("skill"), 0o600); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	if err := os.WriteFile(resourceFile, []byte("resource"), 0o600); err != nil {
		t.Fatalf("write resource: %v", err)
	}
	if err := runScenario(context.Background(), client, scenarioOptions{
		SkillFile:      skillFile,
		ResourceFile:   resourceFile,
		MCPEndpoint:    "http://mcp.local/mcp",
		CredentialName: "github-token",
		Input:          "hello",
	}); err != nil {
		t.Fatalf("run scenario: %v", err)
	}
	server.RequireActions("GetModel", "CreateAgentPromptTemplate", "UploadBlob", "CreateSkill", "CreateResource", "CreateMCP", "CreateAgent", "CreateSession", "Chat")
	if got := server.Body("CreateSession")["Payload"].(map[string]any)["Channel"]; got != "webchat" {
		t.Fatalf("Channel = %v, want webchat", got)
	}
	if got := server.Body("CreateSession")["Payload"].(map[string]any)["PeerKind"]; got != "system" {
		t.Fatalf("PeerKind = %v, want system", got)
	}
}

func newMockClient(t *testing.T) (*hibot.Client, *mocktop.Server) {
	t.Helper()
	server := mocktop.New(t)
	t.Cleanup(server.Close)
	client, err := hibot.NewClient(hibot.Config{
		Endpoint:    server.URL(),
		AccessKey:   "ak",
		SecretKey:   "sk",
		WorkspaceID: "workspace-test",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client, server
}
