package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// UserError signals a user-facing argument/flag problem (exit code 2).
type UserError struct{ Msg string }

func (e *UserError) Error() string { return e.Msg }

func newUserError(format string, args ...any) error {
	return &UserError{Msg: fmt.Sprintf(format, args...)}
}

// ExitCodeFor maps an error to the appropriate process exit code.
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var ue *UserError
	if errors.As(err, &ue) {
		return 2
	}
	return 1
}

const (
	flagConfigFile     = "config-file"
	flagEndpoint       = "endpoint"
	flagAccessKey      = "ak"
	flagSecretKey      = "sk"
	flagWorkspaceID    = "workspace-id"
	flagRegion         = "region"
	flagServerService  = "server-service"
	flagGatewayService = "gateway-service"
	flagModelService   = "model-service"
	flagUpService      = "up-service"
	flagOutput         = "output"
	flagVerbose        = "verbose"
)

// keys used inside viper / config file
const (
	keyEndpoint       = "endpoint"
	keyAccessKey      = "ak"
	keySecretKey      = "sk"
	keyWorkspaceID    = "workspace_id"
	keyRegion         = "region"
	keyServerService  = "server_service"
	keyGatewayService = "gateway_service"
	keyModelService   = "model_service"
	keyUpService      = "up_service"
)

var (
	configFlagToKey = map[string]string{
		flagEndpoint:       keyEndpoint,
		flagAccessKey:      keyAccessKey,
		flagSecretKey:      keySecretKey,
		flagWorkspaceID:    keyWorkspaceID,
		flagRegion:         keyRegion,
		flagServerService:  keyServerService,
		flagGatewayService: keyGatewayService,
		flagModelService:   keyModelService,
		flagUpService:      keyUpService,
	}
	configKeyToEnv = map[string]string{
		keyEndpoint:       "HIBOT_ENDPOINT",
		keyAccessKey:      "HIBOT_AK",
		keySecretKey:      "HIBOT_SK",
		keyWorkspaceID:    "HIBOT_WORKSPACE_ID",
		keyRegion:         "HIBOT_REGION",
		keyServerService:  "HIBOT_SERVER_SERVICE",
		keyGatewayService: "HIBOT_GATEWAY_SERVICE",
		keyModelService:   "HIBOT_MODEL_SERVICE",
		keyUpService:      "HIBOT_UP_SERVICE",
	}
)

// NewRootCmd builds the cobra command tree for `hibot`.
func NewRootCmd() *cobra.Command {
	v := viper.New()
	v.SetConfigType("yaml")

	root := &cobra.Command{
		Use:   "hibot",
		Short: "Hibot CLI - manage agents, sessions, and resources on the Hibot platform",
		Long: `hibot is the command-line interface for the Hibot Managed Agent platform.

It is a thin wrapper around the official Go SDK and reuses the same TOP API
surface: Agents, Sessions, Skills, MCPs, Resources, Prompts, Environments,
Models, and Uploads.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := root.PersistentFlags()
	flags.String(flagConfigFile, "", "Path to config file (default $HOME/.hibot/config.yaml)")
	flags.String(flagEndpoint, "", "TOP endpoint (env HIBOT_ENDPOINT)")
	flags.String(flagAccessKey, "", "Access key (env HIBOT_AK)")
	flags.String(flagSecretKey, "", "Secret key (env HIBOT_SK)")
	flags.String(flagWorkspaceID, "", "Workspace ID (env HIBOT_WORKSPACE_ID)")
	flags.String(flagRegion, "", "Region (env HIBOT_REGION)")
	flags.String(flagServerService, "", "Override hibot-server TOP service name (env HIBOT_SERVER_SERVICE)")
	flags.String(flagGatewayService, "", "Override hibot-gateway TOP service name (env HIBOT_GATEWAY_SERVICE)")
	flags.String(flagModelService, "", "Override aigw-server TOP service name (env HIBOT_MODEL_SERVICE)")
	flags.String(flagUpService, "", "Override up TOP service name (env HIBOT_UP_SERVICE)")
	flags.StringP(flagOutput, "o", "table", "Output format: json|yaml|table")
	flags.BoolP(flagVerbose, "v", false, "Verbose output (e.g. show stream tool events)")

	cobra.OnInitialize(func() { initConfig(root, v) })

	root.AddCommand(newVersionCmd())
	root.AddCommand(newConfigCmd(v))
	root.AddCommand(newAgentsCmd(v))
	root.AddCommand(newSessionsCmd(v))
	root.AddCommand(newChatCmd(v))
	root.AddCommand(newModelsCmd(v))
	root.AddCommand(newSkillsCmd(v))
	root.AddCommand(newMCPsCmd(v))
	root.AddCommand(newResourcesCmd(v))
	root.AddCommand(newPromptsCmd(v))
	root.AddCommand(newEnvironmentsCmd(v))
	root.AddCommand(newUploadsCmd(v))

	return root
}

// Execute runs the CLI.
func Execute() error {
	return NewRootCmd().Execute()
}

// initConfig loads config file (when present), then merges in env vars and flags.
// Precedence (high -> low): flag > env > config file.
func initConfig(root *cobra.Command, v *viper.Viper) {
	// Bind every config flag so viper sees flag overrides automatically.
	for flagName, key := range configFlagToKey {
		_ = v.BindPFlag(key, root.PersistentFlags().Lookup(flagName))
	}
	// Env bindings.
	for key, env := range configKeyToEnv {
		_ = v.BindEnv(key, env)
	}

	cfgPath, _ := root.PersistentFlags().GetString(flagConfigFile)
	if cfgPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfgPath = filepath.Join(home, ".hibot", "config.yaml")
		}
	}
	if cfgPath != "" {
		if _, err := os.Stat(cfgPath); err == nil {
			v.SetConfigFile(cfgPath)
			_ = v.ReadInConfig()
		}
	}
}

// resolveOutputFormat reads --output and returns one of "json"/"yaml"/"table".
func resolveOutputFormat(cmd *cobra.Command) string {
	val, _ := cmd.Flags().GetString(flagOutput)
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "json", "yaml", "table":
		return val
	case "":
		return "table"
	default:
		return val
	}
}

// readContentArg implements `@file` semantics: if value starts with `@`, read
// the file at the remaining path and return its contents.
func readContentArg(value string) (string, error) {
	if !strings.HasPrefix(value, "@") {
		return value, nil
	}
	path := strings.TrimPrefix(value, "@")
	if path == "" {
		return "", newUserError("expected file path after @")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}
