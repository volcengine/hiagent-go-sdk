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
	SkillFile string
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
	skillFile, err := exampleutil.RequiredEnv("HIBOT_SKILL_FILE")
	if err != nil {
		return err
	}
	return runScenario(ctx, client, scenarioOptions{
		SkillFile: skillFile,
		Input:     exampleutil.EnvOrDefault("HIBOT_CHAT_INPUT", "请使用绑定的本地 Skill 总结当前任务。"),
	})
}

func runScenario(ctx context.Context, client *hibot.Client, opts scenarioOptions) error {
	model, err := exampleutil.DefaultModel(ctx, client)
	if err != nil {
		return err
	}
	blobID, err := uploadSkillBlob(ctx, client, opts.SkillFile)
	if err != nil {
		return err
	}
	enabled := true
	skill, err := client.V1.Skills.New(ctx, hibot.V1SkillNewParams{
		Name:        fmt.Sprintf("local-skill-%d", time.Now().UnixNano()),
		Description: "SDK example skill uploaded from a local zip file.",
		BlobID:      blobID,
		Enabled:     &enabled,
		Version:     "1.0.0",
	})
	if err != nil {
		return fmt.Errorf("create skill: %w", err)
	}
	agent, err := client.V1.Agents.New(ctx, hibot.V1AgentNewParams{
		Name:  fmt.Sprintf("skill-agent-%d", time.Now().UnixNano()),
		Model: hibot.V1ManagedAgentModelConfigParams{ID: model.ID},
		Tools: []hibot.V1AgentNewParamsToolUnion{{
			OfSkill: &hibot.V1ManagedAgentSkillToolParams{
				Type:           hibot.V1ManagedAgentSkillToolParamsTypeSkill,
				SkillVersionID: skill.ID,
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

func uploadSkillBlob(ctx context.Context, client *hibot.Client, skillFile string) (string, error) {
	f, err := os.Open(skillFile)
	if err != nil {
		return "", err
	}
	defer f.Close()
	resp, err := client.V1.Uploads.UploadBlob(ctx, hibot.V1UploadBlobParams{
		Filename:    filepath.Base(skillFile),
		ContentType: "application/zip",
	}, f)
	if err != nil {
		return "", fmt.Errorf("upload skill: %w", err)
	}
	return resp.BlobID, nil
}
