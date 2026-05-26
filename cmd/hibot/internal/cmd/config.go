package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newConfigCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage hibot CLI configuration",
	}
	cmd.AddCommand(newConfigInitCmd(v))
	cmd.AddCommand(newConfigViewCmd(v))
	cmd.AddCommand(newConfigSetCmd(v))
	return cmd
}

// configValuesFromFlags collects only the config-related flags that the user
// actually passed; missing flags fall through to env / file values already
// loaded into v.
func configValuesFromFlags(cmd *cobra.Command) map[string]string {
	out := map[string]string{}
	for flag, key := range configFlagToKey {
		if f := cmd.Flags().Lookup(flag); f != nil && f.Changed {
			out[key] = f.Value.String()
		}
	}
	return out
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hibot", "config.yaml"), nil
}

func resolveConfigPath(cmd *cobra.Command) (string, error) {
	root := cmd.Root()
	cfgPath, _ := root.PersistentFlags().GetString(flagConfigFile)
	if cfgPath != "" {
		return cfgPath, nil
	}
	return defaultConfigPath()
}

// loadConfigFile reads the YAML at path. Missing file returns empty map.
func loadConfigFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	out := map[string]any{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func saveConfigFile(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}

func newConfigInitCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Write the current --endpoint/--ak/--sk/--workspace-id flags to the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(cmd)
			if err != nil {
				return err
			}
			existing, err := loadConfigFile(path)
			if err != nil {
				return err
			}
			values := configValuesFromFlags(cmd)
			if len(values) == 0 {
				return newUserError("no config flags provided; pass --endpoint/--ak/--sk/--workspace-id (etc.) to write")
			}
			for k, val := range values {
				existing[k] = val
			}
			if err := saveConfigFile(path, existing); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote config to %s (%d key(s))\n", path, len(values))
			return nil
		},
	}
}

func newConfigViewCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show resolved configuration (file + env + flags)",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := map[string]string{}
			for _, key := range []string{
				keyEndpoint, keyAccessKey, keySecretKey, keyWorkspaceID, keyRegion,
				keyServerService, keyGatewayService, keyModelService, keyUpService,
			} {
				val := v.GetString(key)
				if (key == keySecretKey || key == keyAccessKey) && val != "" {
					val = maskSecret(val)
				}
				out[key] = val
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			rows := make([][]string, 0, len(out))
			for _, k := range []string{
				keyEndpoint, keyAccessKey, keySecretKey, keyWorkspaceID, keyRegion,
				keyServerService, keyGatewayService, keyModelService, keyUpService,
			} {
				rows = append(rows, []string{k, out[k]})
			}
			return e.emitObject(out, []string{"KEY", "VALUE"}, rows)
		},
	}
}

func newConfigSetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a single key in the config file (e.g. workspace_id, endpoint)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			if _, ok := configKeyToEnv[key]; !ok {
				return newUserError("unknown config key %q (allowed: endpoint, ak, sk, workspace_id, region, server_service, gateway_service, model_service, up_service)", key)
			}
			path, err := resolveConfigPath(cmd)
			if err != nil {
				return err
			}
			existing, err := loadConfigFile(path)
			if err != nil {
				return err
			}
			existing[key] = value
			if err := saveConfigFile(path, existing); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set %s in %s\n", key, path)
			return nil
		},
	}
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "***"
	}
	return s[:2] + "***" + s[len(s)-2:]
}
