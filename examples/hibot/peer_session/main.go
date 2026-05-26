package main

import (
	"context"
	"fmt"

	"github.com/volcengine/hiagent-go-sdk/examples/hibot/internal/exampleutil"
	"github.com/volcengine/hiagent-go-sdk/hibot"
)

// 本示例演示如何把 Hibot 会话挂到 IM 渠道（飞书 / 企微 等）上：
// 主流程的 WebChat 场景不需要传 Peer，SDK 会自动注入 webchat 渠道。
// 当需要按 IM 用户/群隔离会话历史时，必须显式传 Channel + PeerKind + PeerID，
// 三者共同决定服务端的 SessionKey 唯一性。
type scenarioOptions struct {
	AgentName string
	Channel   string
	PeerKind  string
	PeerID    string
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
		AgentName: exampleutil.EnvOrDefault("HIBOT_AGENT_NAME", "hibot-peer-agent"),
		Channel:   exampleutil.EnvOrDefault("HIBOT_PEER_CHANNEL", "feishu"),
		PeerKind:  exampleutil.EnvOrDefault("HIBOT_PEER_KIND", "user"),
		PeerID:    exampleutil.EnvOrDefault("HIBOT_PEER_ID", "ou_feishu_user_xxx"),
		Input:     exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请记住这是来自飞书的独立用户会话。"),
	})
}

func runScenario(ctx context.Context, client *hibot.Client, opts scenarioOptions) error {
	model, err := exampleutil.DefaultModel(ctx, client)
	if err != nil {
		return err
	}
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name:  opts.AgentName,
		Model: hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
	})
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	session, err := client.V1.Sessions.New(ctx, hibot.V1SessionNewParams{
		AgentID: agent.ID,
		Peer: &hibot.V1SessionPeerParams{
			Channel:  opts.Channel,
			PeerKind: opts.PeerKind,
			PeerID:   opts.PeerID,
		},
	})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return exampleutil.PrintChatStream(ctx, client, session.ID, hibot.V1SessionChatParams{Input: opts.Input})
}
