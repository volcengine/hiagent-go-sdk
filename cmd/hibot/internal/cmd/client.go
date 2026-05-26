package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
)

// buildClient assembles a Hibot SDK client from the merged configuration in v.
// Missing required fields produce a UserError with a hint on how to configure them.
func buildClient(v *viper.Viper) (*hibot.Client, error) {
	endpoint := v.GetString(keyEndpoint)
	ak := v.GetString(keyAccessKey)
	sk := v.GetString(keySecretKey)
	wsID := v.GetString(keyWorkspaceID)

	missing := make([]string, 0, 4)
	if endpoint == "" {
		missing = append(missing, "endpoint")
	}
	if ak == "" {
		missing = append(missing, "ak")
	}
	if sk == "" {
		missing = append(missing, "sk")
	}
	if wsID == "" {
		missing = append(missing, "workspace-id")
	}
	if len(missing) > 0 {
		return nil, newUserError(
			"hibot: missing required config: %v\n"+
				"Configure via flags (--endpoint --ak --sk --workspace-id),\n"+
				"environment variables (HIBOT_ENDPOINT / HIBOT_AK / HIBOT_SK / HIBOT_WORKSPACE_ID),\n"+
				"or `hibot config init` to write $HOME/.hibot/config.yaml.",
			missing,
		)
	}

	cfg := hibot.Config{
		Endpoint:       endpoint,
		AccessKey:      ak,
		SecretKey:      sk,
		WorkspaceID:    wsID,
		Region:         v.GetString(keyRegion),
		ServerService:  v.GetString(keyServerService),
		GatewayService: v.GetString(keyGatewayService),
		ModelService:   v.GetString(keyModelService),
		UpService:      v.GetString(keyUpService),
	}
	client, err := hibot.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// formatError produces a human-friendly representation of err. APIError gets
// special treatment so request_id / code / message stand out.
func formatError(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *hibot.APIError
	if errors.As(err, &apiErr) {
		return fmt.Sprintf("hibot API error: status=%d request_id=%s code=%s message=%s",
			apiErr.StatusCode, apiErr.RequestID, apiErr.Code, apiErr.Message)
	}
	return err.Error()
}
