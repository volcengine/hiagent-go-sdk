package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newAgentsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agents",
		Aliases: []string{"agent"},
		Short:   "Manage Hibot Agents",
	}
	cmd.AddCommand(newAgentsCreateCmd(v))
	cmd.AddCommand(newAgentsListCmd(v))
	cmd.AddCommand(newAgentsGetCmd(v))
	cmd.AddCommand(newAgentsUpdateCmd(v))
	cmd.AddCommand(newAgentsDeleteCmd(v))
	return cmd
}

func newAgentsCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		name            string
		modelID         string
		system          string
		envID           string
		skillVersionIDs []string
		mcpIDs          []string
		resourceIDs     []string
		directoryIDs    []string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an Agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			if modelID == "" {
				return newUserError("--model-id is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibot.V1AgentNewParams{
				Name:  name,
				EnvID: envID,
				Model: hibot.V1ManagedAgentModelConfigParams{ID: modelID},
			}
			if system != "" {
				resolved, rerr := readContentArg(system)
				if rerr != nil {
					return rerr
				}
				params.System = hibot.String(resolved)
			}
			for _, id := range skillVersionIDs {
				params.Tools = append(params.Tools, hibot.V1AgentNewParamsToolUnion{
					OfSkill: &hibot.V1ManagedAgentSkillToolParams{
						Type:           hibot.V1ManagedAgentSkillToolParamsTypeSkill,
						SkillVersionID: id,
					},
				})
			}
			for _, id := range mcpIDs {
				params.Tools = append(params.Tools, hibot.V1AgentNewParamsToolUnion{
					OfMCP: &hibot.V1ManagedAgentMCPToolParams{
						Type: hibot.V1ManagedAgentMCPToolParamsTypeMCP,
						ID:   id,
					},
				})
			}
			for _, id := range resourceIDs {
				params.Resources = append(params.Resources, hibot.V1ManagedAgentResourceRefParams{ID: id})
			}
			for _, id := range directoryIDs {
				params.Resources = append(params.Resources, hibot.V1ManagedAgentResourceRefParams{DirectoryID: id})
			}

			ctx := context.Background()
			agent, err := client.V1.Agents.New(ctx, params)
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(agent,
				[]string{"ID", "NAME", "MODEL_ID", "ENV_ID"},
				[][]string{{agent.ID, agent.Name, agent.ModelID, agent.EnvID}})
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "Agent name (required)")
	f.StringVar(&modelID, "model-id", "", "Model ID (required)")
	f.StringVar(&system, "system", "", "System prompt (use @path/to/file to read from a file)")
	f.StringVar(&envID, "env-id", "", "Environment ID")
	f.StringArrayVar(&skillVersionIDs, "skill-version-id", nil, "Skill version ID to attach (repeatable)")
	f.StringArrayVar(&mcpIDs, "mcp-id", nil, "MCP ID to attach (repeatable)")
	f.StringArrayVar(&resourceIDs, "resource-id", nil, "Resource ID to attach (repeatable)")
	f.StringArrayVar(&directoryIDs, "directory-id", nil, "Resource directory ID to attach (repeatable)")
	return cmd
}

func newAgentsListCmd(v *viper.Viper) *cobra.Command {
	var keyword string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			items, err := client.V1.Agents.List(context.Background(), hibot.V1AgentListParams{Keyword: keyword})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, a := range items {
				rows = append(rows, []string{a.ID, a.Name, a.ModelID, a.EnvID})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(items, []string{"ID", "NAME", "MODEL_ID", "ENV_ID"}, rows)
		},
	}
	cmd.Flags().StringVar(&keyword, "keyword", "", "Filter agents by name keyword")
	return cmd
}

func newAgentsGetCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <agent-id>",
		Short: "Get an Agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			agent, err := client.V1.Agents.Get(context.Background(), hibotv1.V1AgentGetParams{AgentID: args[0]})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(agent,
				[]string{"ID", "NAME", "MODEL_ID", "ENV_ID", "SYSTEM"},
				[][]string{{agent.ID, agent.Name, agent.ModelID, agent.EnvID, truncate(agent.SystemPrompt, 60)}})
		},
	}
	return cmd
}

func newAgentsUpdateCmd(v *viper.Viper) *cobra.Command {
	var (
		description string
		modelID     string
		system      string
	)
	cmd := &cobra.Command{
		Use:   "update <agent-id>",
		Short: "Update an Agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibotv1.V1AgentUpdateParams{AgentID: args[0]}
			if cmd.Flags().Changed("description") {
				params.Description = hibot.String(description)
			}
			if cmd.Flags().Changed("model-id") {
				params.ModelID = hibot.String(modelID)
			}
			if cmd.Flags().Changed("system") {
				resolved, rerr := readContentArg(system)
				if rerr != nil {
					return rerr
				}
				params.System = hibot.String(resolved)
			}
			if err := client.V1.Agents.Update(context.Background(), params); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "agent %s updated\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Agent description")
	cmd.Flags().StringVar(&modelID, "model-id", "", "Model ID")
	cmd.Flags().StringVar(&system, "system", "", "System prompt (use @path/to/file)")
	return cmd
}

func newAgentsDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <agent-id>",
		Short: "Delete an Agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Agents.Delete(context.Background(), hibot.V1AgentDeleteParams{AgentID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "agent %s deleted\n", args[0])
			return nil
		},
	}
}
