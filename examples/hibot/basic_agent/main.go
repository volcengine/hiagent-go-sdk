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
		AgentName: exampleutil.EnvOrDefault("HIBOT_AGENT_NAME", "hibot-basic-agent"),
		Input:     exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请用一句话介绍 Hibot Agent。"),
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
		System: hibot.String("你是一个简洁的 Hibot SDK 示例助手。"),
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
	message, err := client.V1.Sessions.Chat(ctx, session.ID, hibot.V1SessionChatParams{Input: opts.Input})
	if err != nil {
		return fmt.Errorf("chat: %w", err)
	}
	fmt.Printf("message_id=%s content=%s\n", message.ID, message.Content)
	return nil
}
