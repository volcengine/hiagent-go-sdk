package main

import (
	"context"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func TestPeerSessionScenario(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t)
	if err := runScenario(context.Background(), client, scenarioOptions{AgentName: "peer", Channel: "feishu", PeerKind: "user", PeerID: "ou_feishu_user_001", Input: "hello"}); err != nil {
		t.Fatalf("run scenario: %v", err)
	}
	server.RequireActions("CreateSession", "Chat")
	payload := server.Body("CreateSession")["Payload"].(map[string]any)
	if payload["Channel"] != "feishu" || payload["PeerKind"] != "user" || payload["PeerID"] != "ou_feishu_user_001" {
		t.Fatalf("peer payload = %#v, want feishu/user/ou_feishu_user_001", payload)
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
