package main

import (
	"context"
	"fmt"
	"time"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/exampleutil"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

type scenarioOptions struct {
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
	endpoint, err := exampleutil.RequiredEnv("HIBOT_GITHUB_MCP_ENDPOINT")
	if err != nil {
		return err
	}
	return runScenario(ctx, client, scenarioOptions{
		MCPEndpoint:    endpoint,
		CredentialName: exampleutil.EnvOrDefault("HIBOT_GITHUB_CREDENTIAL_NAME", "github-token"),
		Input:          exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请通过 MCP 说明如何查看仓库最近变更。"),
	})
}

func runScenario(ctx context.Context, client *hibot.Client, opts scenarioOptions) error {
	model, err := exampleutil.DefaultModel(ctx, client)
	if err != nil {
		return err
	}
	mcp, err := client.V1.MCPs.New(ctx, hibot.V1MCPNewParams{
		Name:      fmt.Sprintf("github-mcp-%d", time.Now().UnixNano()),
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
		Name:  fmt.Sprintf("mcp-agent-%d", time.Now().UnixNano()),
		Model: hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
		Tools: []hibot.V1AgentNewParamsToolUnion{{
			OfMCP: &hibot.V1ManagedAgentMCPToolParams{
				Type: hibot.V1ManagedAgentMCPToolParamsTypeMCP,
				ID:   mcp.ID,
			},
		}},
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
