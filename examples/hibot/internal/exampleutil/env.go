package exampleutil

import (
	"context"
	"fmt"
	"os"

	"github.com/volcengine/hiagent-go-sdk/hibot"
)

// ClientFromEnv builds a Hibot client from the environment variables used by all examples.
func ClientFromEnv() (*hibot.Client, error) {
	return hibot.NewClient(hibot.Config{
		Endpoint:    EnvOrDefault("HIBOT_ENDPOINT", EnvOrDefault("HIBOT_E2E_TOP_HOST", "https://hibot.internal")),
		AccessKey:   EnvOrDefault("HIBOT_AK", os.Getenv("HIBOT_E2E_AK")),
		SecretKey:   EnvOrDefault("HIBOT_SK", os.Getenv("HIBOT_E2E_SK")),
		WorkspaceID: EnvOrDefault("HIBOT_WORKSPACE_ID", EnvOrDefault("HIBOT_E2E_WORKSPACE", "workspace-ops-prod")),
	})
}

func DefaultModel(ctx context.Context, client *hibot.Client) (*hibot.V1Model, error) {
	model, err := client.V1.Models.Get(ctx, hibot.V1ModelGetParams{
		ID: EnvOrDefault("HIBOT_MODEL_ID", hibot.V1ManagedAgentModelDoubaoSeedPro),
	})
	if err != nil {
		return nil, fmt.Errorf("get model: %w", err)
	}
	return model, nil
}

func RequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func EnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
