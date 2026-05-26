package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newEnvironmentsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environments",
		Aliases: []string{"environment", "env"},
		Short:   "Manage execution environments",
	}
	cmd.AddCommand(newEnvironmentsListCmd(v))
	cmd.AddCommand(newEnvironmentsGetCmd(v))
	cmd.AddCommand(newEnvironmentsCreateCmd(v))
	cmd.AddCommand(newEnvironmentsDeleteCmd(v))
	return cmd
}

func newEnvironmentsListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			items, err := client.V1.Environments.List(context.Background(), hibot.V1EnvironmentListParams{})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, e := range items {
				rows = append(rows, []string{e.ID, e.Name, e.ImageType, e.CPULimit, e.MemoryLimit})
			}
			format := resolveOutputFormat(cmd)
			em := newEmitter(format, cmd.OutOrStdout())
			return em.emitObject(items, []string{"ID", "NAME", "IMAGE_TYPE", "CPU", "MEM"}, rows)
		},
	}
}

func newEnvironmentsGetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "get <env-id>",
		Short: "Get an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			env, err := client.V1.Environments.Get(context.Background(), hibotv1.V1EnvironmentGetParams{EnvID: args[0]})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			em := newEmitter(format, cmd.OutOrStdout())
			return em.emitObject(env,
				[]string{"ID", "NAME", "IMAGE_TYPE", "CPU", "MEM"},
				[][]string{{env.ID, env.Name, env.ImageType, env.CPULimit, env.MemoryLimit}})
		},
	}
}

func newEnvironmentsCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		name        string
		description string
		imageType   string
		envVarsJSON string
		cpuLimit    string
		memLimit    string
		pvcSize     string
		dataPath    string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			if imageType == "" {
				return newUserError("--image-type is required")
			}
			params := hibot.V1EnvironmentNewParams{
				Name:        name,
				Description: description,
				ImageType:   imageType,
				CPULimit:    cpuLimit,
				MemoryLimit: memLimit,
				PVCSize:     pvcSize,
				DataPath:    dataPath,
			}
			if envVarsJSON != "" {
				resolved, rerr := readContentArg(envVarsJSON)
				if rerr != nil {
					return rerr
				}
				if !json.Valid([]byte(resolved)) {
					return newUserError("--env-vars must be valid JSON")
				}
				params.EnvVars = json.RawMessage(resolved)
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			env, err := client.V1.Environments.New(context.Background(), params)
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			em := newEmitter(format, cmd.OutOrStdout())
			return em.emitObject(env,
				[]string{"ID", "NAME", "IMAGE_TYPE"},
				[][]string{{env.ID, env.Name, env.ImageType}})
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "Environment name (required)")
	f.StringVar(&description, "description", "", "Description")
	f.StringVar(&imageType, "image-type", "", "Image type (required)")
	f.StringVar(&envVarsJSON, "env-vars", "", "Env vars JSON (or @file)")
	f.StringVar(&cpuLimit, "cpu", "", "CPU limit")
	f.StringVar(&memLimit, "memory", "", "Memory limit")
	f.StringVar(&pvcSize, "pvc-size", "", "PVC size")
	f.StringVar(&dataPath, "data-path", "", "Data path")
	return cmd
}

func newEnvironmentsDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <env-id>",
		Short: "Delete an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Environments.Delete(context.Background(), hibotv1.V1EnvironmentDeleteParams{EnvID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "environment %s deleted\n", args[0])
			return nil
		},
	}
}
