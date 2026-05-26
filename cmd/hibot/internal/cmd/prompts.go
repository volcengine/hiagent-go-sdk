package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newPromptsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prompts",
		Aliases: []string{"prompt"},
		Short:   "Manage Prompt templates",
	}
	cmd.AddCommand(newPromptsListCmd(v))
	cmd.AddCommand(newPromptsCreateCmd(v))
	cmd.AddCommand(newPromptsUpdateCmd(v))
	cmd.AddCommand(newPromptsDeleteCmd(v))
	return cmd
}

func newPromptsListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List prompt templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			items, err := client.V1.Prompts.List(context.Background(), hibotv1.V1PromptListParams{})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, p := range items {
				rows = append(rows, []string{p.ID, p.Name, truncate(p.Content, 60)})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(items, []string{"ID", "NAME", "CONTENT"}, rows)
		},
	}
}

func newPromptsCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		name    string
		content string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a prompt template",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			if content == "" {
				return newUserError("--content is required")
			}
			resolved, err := readContentArg(content)
			if err != nil {
				return err
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			p, err := client.V1.Prompts.New(context.Background(), hibot.V1PromptNewParams{
				Name:    name,
				Content: resolved,
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(p,
				[]string{"ID", "NAME", "CONTENT"},
				[][]string{{p.ID, p.Name, truncate(p.Content, 60)}})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Prompt name (required)")
	cmd.Flags().StringVar(&content, "content", "", "Prompt content (use @path/to/file)")
	return cmd
}

func newPromptsUpdateCmd(v *viper.Viper) *cobra.Command {
	var (
		name    string
		content string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prompt template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibotv1.V1PromptUpdateParams{ID: args[0]}
			if cmd.Flags().Changed("name") {
				params.Name = hibot.String(name)
			}
			if cmd.Flags().Changed("content") {
				resolved, rerr := readContentArg(content)
				if rerr != nil {
					return rerr
				}
				params.Content = hibot.String(resolved)
			}
			if err := client.V1.Prompts.Update(context.Background(), params); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "prompt %s updated\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New name")
	cmd.Flags().StringVar(&content, "content", "", "New content (use @path/to/file)")
	return cmd
}

func newPromptsDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a prompt template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Prompts.Delete(context.Background(), hibotv1.V1PromptDeleteParams{ID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "prompt %s deleted\n", args[0])
			return nil
		},
	}
}
