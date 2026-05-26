package main

import (
	"context"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func TestMCPAgentScenario(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t)
	if err := runScenario(context.Background(), client, scenarioOptions{MCPEndpoint: "http://mcp.local/mcp", CredentialName: "github-token", Input: "hello"}); err != nil {
		t.Fatalf("run scenario: %v", err)
	}
	server.RequireActions("CreateMCP", "CreateAgent", "CreateSession", "Chat")
	if got := server.Body("CreateMCP")["URL"]; got != "http://mcp.local/mcp" {
		t.Fatalf("CreateMCP URL = %v", got)
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
