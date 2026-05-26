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
	SkillFile      string
	ResourceFile   string
	MCPEndpoint    string
	CredentialName string
	Input          string
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
	skillFile, err := exampleutil.RequiredEnv("HIBOT_SKILL_FILE")
	if err != nil {
		return err
	}
	resourceFile, err := exampleutil.RequiredEnv("HIBOT_RESOURCE_FILE")
	if err != nil {
		return err
	}
	mcpEndpoint, err := exampleutil.RequiredEnv("HIBOT_GITHUB_MCP_ENDPOINT")
	if err != nil {
		return err
	}
	return runScenario(ctx, client, scenarioOptions{
		SkillFile:      skillFile,
		ResourceFile:   resourceFile,
		MCPEndpoint:    mcpEndpoint,
		CredentialName: exampleutil.EnvOrDefault("HIBOT_GITHUB_CREDENTIAL_NAME", "github-token"),
		Input:          exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请综合本地 Skill、资料和 GitHub MCP 给出排障方案。"),
	})
}

func runScenario(ctx context.Context, client *hibot.Client, opts scenarioOptions) error {
	model, err := exampleutil.DefaultModel(ctx, client)
	if err != nil {
		return err
	}
	prompt, err := client.V1.Prompts.New(ctx, hibot.V1PromptNewParams{
		Name:    fmt.Sprintf("managed-agent-prompt-%d", time.Now().UnixNano()),
		Content: "你是一个企业研发排障助手。回答时先说明依据，再给出可执行步骤。",
	})
	if err != nil {
		return fmt.Errorf("create prompt: %w", err)
	}
	skillBlobID, err := uploadBlob(ctx, client, opts.SkillFile, "application/zip")
	if err != nil {
		return fmt.Errorf("upload skill: %w", err)
	}
	enabled := true
	skill, err := client.V1.Skills.New(ctx, hibot.V1SkillNewParams{
		Name:        fmt.Sprintf("comprehensive-skill-%d", time.Now().UnixNano()),
		Description: "Skill uploaded by the comprehensive SDK example.",
		BlobID:      skillBlobID,
		Enabled:     &enabled,
		Version:     "1.0.0",
	})
	if err != nil {
		return fmt.Errorf("create skill: %w", err)
	}
	resourceBlobID, err := uploadBlob(ctx, client, opts.ResourceFile, "application/octet-stream")
	if err != nil {
		return fmt.Errorf("upload resource: %w", err)
	}
	resource, err := client.V1.Resources.New(ctx, hibot.V1ResourceNewParams{
		Name:   fmt.Sprintf("comprehensive-resource-%d", time.Now().UnixNano()),
		Type:   hibot.V1ResourceTypeDocumentCollection,
		BlobID: resourceBlobID,
	})
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}
	mcp, err := client.V1.MCPs.New(ctx, hibot.V1MCPNewParams{
		Name:      fmt.Sprintf("comprehensive-github-%d", time.Now().UnixNano()),
		Transport: hibot.V1MCPTransportStreamableHTTP,
		Endpoint:  opts.MCPEndpoint,
		Credential: &hibot.V1CredentialRefParams{
			Name: opts.CredentialName,
		},
	})
	if err != nil {
		return fmt.Errorf("create mcp: %w", err)
	}
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name:   fmt.Sprintf("comprehensive-agent-%d", time.Now().UnixNano()),
		Model:  hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
		System: hibot.String(prompt.Content),
		Tools: []hibot.V1AgentNewParamsToolUnion{
			{OfSkill: &hibot.V1ManagedAgentSkillToolParams{Type: hibot.V1ManagedAgentSkillToolParamsTypeSkill, SkillVersionID: skill.ID}},
			{OfMCP: &hibot.V1ManagedAgentMCPToolParams{Type: hibot.V1ManagedAgentMCPToolParamsTypeMCP, ID: mcp.ID}},
		},
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

func uploadBlob(ctx context.Context, client *hibot.Client, filePath, contentType string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	resp, err := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
		Filename:    filepath.Base(filePath),
		ContentType: contentType,
	}, f)
	if err != nil {
		return "", err
	}
	return resp.BlobID, nil
}
