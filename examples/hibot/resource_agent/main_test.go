package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func TestResourceAgentScenario(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t)
	resourceFile := filepath.Join(t.TempDir(), "runbook.md")
	if err := os.WriteFile(resourceFile, []byte("# runbook"), 0o600); err != nil {
		t.Fatalf("write resource: %v", err)
	}
	if err := runScenario(context.Background(), client, scenarioOptions{ResourceFile: resourceFile, Input: "hello"}); err != nil {
		t.Fatalf("run scenario: %v", err)
	}
	server.RequireActions("UploadBlob", "CreateResource", "CreateAgent", "CreateSession", "Chat")
	resources := server.Body("CreateAgent")["Resources"].(map[string]any)
	if len(resources["ResourceIDs"].([]any)) != 1 {
		t.Fatalf("Resources = %#v, want one ResourceID", resources)
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
