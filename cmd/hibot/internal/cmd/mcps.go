package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newMCPsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mcps",
		Aliases: []string{"mcp"},
		Short:   "Manage MCP servers",
	}
	cmd.AddCommand(newMCPsListCmd(v))
	cmd.AddCommand(newMCPsGetCmd(v))
	cmd.AddCommand(newMCPsCreateCmd(v))
	cmd.AddCommand(newMCPsDeleteCmd(v))
	cmd.AddCommand(newMCPsTestCmd(v))
	return cmd
}

func parseHeaders(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(values))
	for _, kv := range values {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, newUserError("invalid header %q (expected KEY=VALUE)", kv)
		}
		out[parts[0]] = parts[1]
	}
	return out, nil
}

func newMCPsListCmd(v *viper.Viper) *cobra.Command {
	var keyword string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			items, err := client.V1.MCPs.List(context.Background(), hibotv1.V1MCPListParams{Keyword: keyword})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, m := range items {
				rows = append(rows, []string{m.ID, m.Name, m.Transport, m.Endpoint, m.Status})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(items, []string{"ID", "NAME", "TRANSPORT", "ENDPOINT", "STATUS"}, rows)
		},
	}
	cmd.Flags().StringVar(&keyword, "keyword", "", "Filter by name keyword")
	return cmd
}

func newMCPsGetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			m, err := client.V1.MCPs.Get(context.Background(), hibotv1.V1MCPGetParams{ID: args[0]})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(m,
				[]string{"ID", "NAME", "TRANSPORT", "ENDPOINT", "STATUS"},
				[][]string{{m.ID, m.Name, m.Transport, m.Endpoint, m.Status}})
		},
	}
}

func newMCPsCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		name         string
		description  string
		transport    string
		endpoint     string
		headers      []string
		authType     string
		credName     string
		credKey      string
		credValue    string
		credProvider string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register an MCP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return newUserError("--name is required")
			}
			if endpoint == "" {
				return newUserError("--endpoint is required")
			}
			if transport == "" {
				transport = hibot.V1MCPTransportStreamableHTTP
			}
			hdrs, err := parseHeaders(headers)
			if err != nil {
				return err
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibot.V1MCPNewParams{
				Name:        name,
				Description: description,
				Transport:   transport,
				Endpoint:    endpoint,
				Headers:     hdrs,
				AuthType:    authType,
			}
			if credName != "" || credValue != "" {
				cfg := &hibot.V1MCPCredentialInputParams{Name: credName, ProviderType: credProvider}
				if credValue != "" {
					if credKey == "" {
						credKey = "token"
					}
					cfg.Secrets = append(cfg.Secrets, hibot.V1CredentialSecretInputParams{
						KeyName:     credKey,
						SecretType:  "string",
						SecretValue: credValue,
					})
				}
				params.CredentialConfig = cfg
			}
			m, err := client.V1.MCPs.New(context.Background(), params)
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(m,
				[]string{"ID", "NAME", "TRANSPORT", "ENDPOINT"},
				[][]string{{m.ID, m.Name, m.Transport, m.Endpoint}})
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "MCP name (required)")
	f.StringVar(&description, "description", "", "Description")
	f.StringVar(&transport, "transport", "", "Transport (default streamable-http)")
	f.StringVar(&endpoint, "endpoint", "", "Endpoint URL (required)")
	f.StringArrayVar(&headers, "header", nil, "Header KEY=VALUE (repeatable)")
	f.StringVar(&authType, "auth-type", "", "Auth type")
	f.StringVar(&credName, "credential-name", "", "Credential provider name")
	f.StringVar(&credProvider, "credential-provider-type", "", "Credential provider type (e.g. basic)")
	f.StringVar(&credKey, "credential-key", "", "Credential secret key name (default token)")
	f.StringVar(&credValue, "credential-value", "", "Credential secret value (sets SecretValue)")
	return cmd
}

func newMCPsDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.MCPs.Delete(context.Background(), hibotv1.V1MCPDeleteParams{ID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "mcp %s deleted\n", args[0])
			return nil
		},
	}
}

func newMCPsTestCmd(v *viper.Viper) *cobra.Command {
	var (
		transport    string
		endpoint     string
		headers      []string
		authType     string
		credName     string
		credKey      string
		credValue    string
		credProvider string
	)
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test connection to an MCP endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			if endpoint == "" {
				return newUserError("--endpoint is required")
			}
			if transport == "" {
				transport = hibot.V1MCPTransportStreamableHTTP
			}
			hdrs, err := parseHeaders(headers)
			if err != nil {
				return err
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibotv1.V1MCPTestConnectionParams{
				Transport: transport,
				Endpoint:  endpoint,
				Headers:   hdrs,
				AuthType:  authType,
			}
			if credName != "" || credValue != "" {
				cfg := &hibot.V1MCPCredentialInputParams{Name: credName, ProviderType: credProvider}
				if credValue != "" {
					if credKey == "" {
						credKey = "token"
					}
					cfg.Secrets = append(cfg.Secrets, hibot.V1CredentialSecretInputParams{
						KeyName:     credKey,
						SecretType:  "string",
						SecretValue: credValue,
					})
				}
				params.CredentialConfig = cfg
			}
			result, err := client.V1.MCPs.TestConnection(context.Background(), params)
			if err != nil {
				return err
			}
			rows := [][]string{{fmt.Sprintf("%t", result.Success), fmt.Sprintf("%d", result.ToolCount), result.Error}}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"SUCCESS", "TOOLS", "ERROR"}, rows)
		},
	}
	f := cmd.Flags()
	f.StringVar(&transport, "transport", "", "Transport (default streamable-http)")
	f.StringVar(&endpoint, "endpoint", "", "Endpoint URL (required)")
	f.StringArrayVar(&headers, "header", nil, "Header KEY=VALUE (repeatable)")
	f.StringVar(&authType, "auth-type", "", "Auth type")
	f.StringVar(&credName, "credential-name", "", "Credential provider name")
	f.StringVar(&credProvider, "credential-provider-type", "", "Credential provider type (e.g. basic)")
	f.StringVar(&credKey, "credential-key", "", "Credential secret key name (default token)")
	f.StringVar(&credValue, "credential-value", "", "Credential secret value (sets SecretValue)")
	return cmd
}
