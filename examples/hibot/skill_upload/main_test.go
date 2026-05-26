package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/mocktop"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func TestSkillUploadScenario(t *testing.T) {
	t.Parallel()
	client, server := newMockClient(t)
	skillFile := filepath.Join(t.TempDir(), "skill.zip")
	if err := os.WriteFile(skillFile, []byte("skill bytes"), 0o600); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	if err := runScenario(context.Background(), client, scenarioOptions{SkillFile: skillFile, Input: "hello"}); err != nil {
		t.Fatalf("run scenario: %v", err)
	}
	server.RequireActions("UploadBlob", "CreateSkill", "CreateAgent", "CreateSession", "Chat")
	tools := server.Body("CreateAgent")["Skills"].([]any)
	if len(tools) != 1 {
		t.Fatalf("Skills = %#v, want one skill binding", tools)
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
