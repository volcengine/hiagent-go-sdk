package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/exampleutil"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

type scenarioOptions struct {
	ResourceFile string
	Input        string
}

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	client, err := exampleutil.ClientFromEnv()
	if err != nil {
		return err
	}
	resourceFile, err := exampleutil.RequiredEnv("HIBOT_RESOURCE_FILE")
	if err != nil {
		return err
	}
	return runScenario(ctx, client, scenarioOptions{
		ResourceFile: resourceFile,
		Input:        exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请基于绑定资料回答：这份资料适合解决什么问题？"),
	})
}

func runScenario(ctx context.Context, client *hibot.Client, opts scenarioOptions) error {
	model, err := exampleutil.DefaultModel(ctx, client)
	if err != nil {
		return err
	}
	blobID, err := uploadResourceBlob(ctx, client, opts.ResourceFile)
	if err != nil {
		return err
	}
	resource, err := client.V1.Resources.New(ctx, hibot.V1ResourceNewParams{
		Name:   fmt.Sprintf("resource-%d", time.Now().UnixNano()),
		Type:   hibot.V1ResourceTypeDocumentCollection,
		BlobID: blobID,
	})
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name:      fmt.Sprintf("resource-agent-%d", time.Now().UnixNano()),
		Model:     hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
		Resources: []hibot.V1ManagedAgentResourceRefParams{{ID: resource.ID}},
	})
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
		AgentID: agent.ID,
	})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return exampleutil.PrintChatStream(ctx, client, session.ID, hibot.V1SessionChatParams{Input: opts.Input})
}

func uploadResourceBlob(ctx context.Context, client *hibot.Client, resourceFile string) (string, error) {
	f, err := os.Open(resourceFile)
	if err != nil {
		return "", err
	}
	defer f.Close()
	resp, err := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
		Filename:    filepath.Base(resourceFile),
		ContentType: "application/octet-stream",
	}, f)
	if err != nil {
		return "", fmt.Errorf("upload resource: %w", err)
	}
	return resp.BlobID, nil
}
