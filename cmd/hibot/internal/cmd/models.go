package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newModelsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "models",
		Aliases: []string{"model"},
		Short:   "Manage model registrations",
	}
	cmd.AddCommand(newModelsListCmd(v))
	cmd.AddCommand(newModelsGetCmd(v))
	cmd.AddCommand(newModelsCreateCmd(v))
	cmd.AddCommand(newModelsDeleteCmd(v))
	cmd.AddCommand(newModelsProvidersCmd(v))
	return cmd
}

func newModelsListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List models",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			result, err := client.V1.Models.List(context.Background(), hibot.V1ModelListParams{})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(result.Items))
			for _, m := range result.Items {
				rows = append(rows, []string{m.ID, m.Name, m.Type, m.Provider, m.ModelName})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"ID", "NAME", "TYPE", "PROVIDER", "MODEL_NAME"}, rows)
		},
	}
}

func newModelsGetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-name>",
		Short: "Get a model by ID (or by Name when no exact ID)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			query := args[0]
			model, err := client.V1.Models.Get(context.Background(), hibot.V1ModelGetParams{ID: query})
			if err != nil {
				model, err = client.V1.Models.Get(context.Background(), hibot.V1ModelGetParams{Name: query})
				if err != nil {
					return err
				}
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(model,
				[]string{"ID", "NAME", "TYPE", "PROVIDER", "MODEL_NAME"},
				[][]string{{model.ID, model.Name, model.Type, model.Provider, model.ModelName}})
		},
	}
}

func newModelsCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		name        string
		modelType   string
		provider    string
		modelName   string
		spec        string
		description string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a model",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			if modelType == "" {
				return newUserError("--type is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibotv1.V1ModelNewParams{
				Name:        name,
				Type:        modelType,
				Provider:    provider,
				ModelName:   modelName,
				Spec:        spec,
				Description: description,
			}
			model, err := client.V1.Models.New(context.Background(), params)
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(model,
				[]string{"ID", "NAME", "TYPE", "PROVIDER", "MODEL_NAME"},
				[][]string{{model.ID, model.Name, model.Type, model.Provider, model.ModelName}})
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "Model name (required)")
	f.StringVar(&modelType, "type", "", "Model type (required)")
	f.StringVar(&provider, "provider", "", "Provider")
	f.StringVar(&modelName, "model-name", "", "Provider-side model name")
	f.StringVar(&spec, "spec", "", "Model spec")
	f.StringVar(&description, "description", "", "Model description")
	return cmd
}

func newModelsDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Models.Delete(context.Background(), hibotv1.V1ModelDeleteParams{ID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "model %s deleted\n", args[0])
			return nil
		},
	}
}

func newModelsProvidersCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "Discover model providers",
	}
	cmd.AddCommand(newModelsProvidersListCmd(v))
	cmd.AddCommand(newModelsProvidersListModelsCmd(v))
	return cmd
}

func newModelsProvidersListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List provider names",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			providers, err := client.V1.Models.ListProviders(context.Background())
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(providers))
			for _, p := range providers {
				rows = append(rows, []string{p})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(providers, []string{"PROVIDER"}, rows)
		},
	}
}

func newModelsProvidersListModelsCmd(v *viper.Viper) *cobra.Command {
	var (
		provider  string
		modelType string
	)
	cmd := &cobra.Command{
		Use:   "list-models",
		Short: "List provider-supported models",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			result, err := client.V1.Models.ListModelProviders(context.Background(), hibotv1.V1ModelProviderListParams{
				Provider: provider,
				Type:     modelType,
			})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(result.Items))
			for _, m := range result.Items {
				rows = append(rows, []string{m.ID, m.Provider, m.Type, m.ModelName})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"ID", "PROVIDER", "TYPE", "MODEL_NAME"}, rows)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "Filter by provider")
	cmd.Flags().StringVar(&modelType, "type", "", "Filter by model type")
	return cmd
}
