package main

import (
	"context"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func TestStreamingChatScenario(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t)
	if err := runScenario(context.Background(), client, scenarioOptions{AgentName: "stream", Input: "stream hello"}); err != nil {
		t.Fatalf("run scenario: %v", err)
	}
	server.RequireActions("GetModel", "CreateAgent", "CreateSession", "Chat")
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
