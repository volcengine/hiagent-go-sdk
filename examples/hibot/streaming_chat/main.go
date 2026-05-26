package main

import (
	"context"
	"fmt"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/exampleutil"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

type scenarioOptions struct {
	AgentName string
	Input     string
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
	return runScenario(ctx, client, scenarioOptions{
		AgentName: exampleutil.EnvOrDefault("HIBOT_AGENT_NAME", "hibot-streaming-agent"),
		Input:     exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请流式输出一个三步排障计划。"),
	})
}

func runScenario(ctx context.Context, client *hibot.Client, opts scenarioOptions) error {
	model, err := exampleutil.DefaultModel(ctx, client)
	if err != nil {
		return err
	}
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name: opts.AgentName,
		Model: hibot.V1ManagedAgentModelConfigParams{
			ID: model.ID,
		},
		System: hibot.String("你是一个擅长流式解释的 Hibot Agent。"),
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
