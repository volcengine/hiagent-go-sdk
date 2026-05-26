package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

func newSessionsCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "Manage Hibot Sessions",
	}
	cmd.AddCommand(newSessionsCreateCmd(v))
	cmd.AddCommand(newSessionsListCmd(v))
	cmd.AddCommand(newSessionsGetCmd(v))
	cmd.AddCommand(newSessionsDeleteCmd(v))
	cmd.AddCommand(newSessionsArchiveCmd(v))
	cmd.AddCommand(newSessionsMessagesCmd(v))
	return cmd
}

func newSessionsCreateCmd(v *viper.Viper) *cobra.Command {
	var (
		agentID  string
		channel  string
		peerKind string
		peerID   string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				return newUserError("--agent-id is required")
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibot.V1SessionNewParams{AgentID: agentID}
			// 仅当调用方需要把会话挂到 IM 渠道（飞书 / 企微 等）时才传 Peer；
			// 默认 webchat 主流程不需要任何 Peer 字段，由 SDK 注入兜底值。
			if channel != "" || peerKind != "" || peerID != "" {
				params.Peer = &hibot.V1SessionPeerParams{
					Channel:  channel,
					PeerKind: peerKind,
					PeerID:   peerID,
				}
			}
			sess, err := client.V1.Sessions.New(context.Background(), params)
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(sess,
				[]string{"ID", "AGENT_ID", "PEER_KIND", "PEER_ID"},
				[][]string{{sess.ID, sess.AgentID, sess.PeerKind, sess.PeerID}})
		},
	}
	f := cmd.Flags()
	f.StringVar(&agentID, "agent-id", "", "Agent ID (required)")
	f.StringVar(&channel, "channel", "", "IM channel (e.g. feishu/wecom). Leave empty for webchat default")
	f.StringVar(&peerKind, "peer-kind", "", "Peer kind for IM channel (e.g. user/group)")
	f.StringVar(&peerID, "peer-id", "", "Peer ID for IM channel (e.g. open_id / chat_id)")
	return cmd
}

func newSessionsListCmd(v *viper.Viper) *cobra.Command {
	var agentID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			result, err := client.V1.Sessions.List(context.Background(), hibotv1.V1SessionListParams{AgentID: agentID})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(result.Items))
			for _, s := range result.Items {
				rows = append(rows, []string{s.ID, s.AgentID, s.Status, s.CreatedAt})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"ID", "AGENT_ID", "STATUS", "CREATED_AT"}, rows)
		},
	}
	cmd.Flags().StringVar(&agentID, "agent-id", "", "Filter sessions by Agent ID")
	return cmd
}

func newSessionsGetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "get <session-id>",
		Short: "Get a Session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			sess, err := client.V1.Sessions.Get(context.Background(), hibotv1.V1SessionGetParams{SessionID: args[0]})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(sess,
				[]string{"ID", "AGENT_ID", "STATUS", "PEER_KIND", "PEER_ID"},
				[][]string{{sess.ID, sess.AgentID, sess.Status, sess.PeerKind, sess.PeerID}})
		},
	}
}

func newSessionsDeleteCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <session-id>",
		Short: "Delete a Session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Sessions.Delete(context.Background(), hibotv1.V1SessionDeleteParams{SessionID: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "session %s deleted\n", args[0])
			return nil
		},
	}
}

func newSessionsArchiveCmd(v *viper.Viper) *cobra.Command {
	var summary string
	cmd := &cobra.Command{
		Use:   "archive <session-id>",
		Short: "Archive a Session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			if err := client.V1.Sessions.Archive(context.Background(), hibotv1.V1SessionArchiveParams{
				SessionID: args[0],
				Summary:   summary,
			}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "session %s archived\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&summary, "summary", "", "Optional archival summary")
	return cmd
}

func newSessionsMessagesCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Manage messages within a Session",
	}
	cmd.AddCommand(newSessionsMessagesListCmd(v))
	cmd.AddCommand(newSessionsMessagesGetCmd(v))
	cmd.AddCommand(newSessionsMessagesInjectCmd(v))
	return cmd
}

func newSessionsMessagesListCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "list <session-id>",
		Short: "List messages in a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			result, err := client.V1.Sessions.ListMessages(context.Background(), hibotv1.V1MessageListParams{SessionID: args[0]})
			if err != nil {
				return err
			}
			rows := make([][]string, 0, len(result.Items))
			for _, m := range result.Items {
				rows = append(rows, []string{m.ID, m.Role, truncate(m.Content, 60), m.CreatedAt})
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(result, []string{"ID", "ROLE", "CONTENT", "CREATED_AT"}, rows)
		},
	}
}

func newSessionsMessagesGetCmd(v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "get <session-id> <message-id>",
		Short: "Get a message",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			msg, err := client.V1.Sessions.GetMessage(context.Background(), hibotv1.V1MessageGetParams{
				SessionID: args[0],
				MessageID: args[1],
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(msg,
				[]string{"ID", "ROLE", "CONTENT", "CREATED_AT"},
				[][]string{{msg.ID, msg.Role, truncate(msg.Content, 60), msg.CreatedAt}})
		},
	}
}

func newSessionsMessagesInjectCmd(v *viper.Viper) *cobra.Command {
	var (
		role    string
		content string
	)
	cmd := &cobra.Command{
		Use:   "inject <session-id>",
		Short: "Inject a message into a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if role == "" {
				return newUserError("--role is required")
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
			msg, err := client.V1.Sessions.InjectMessage(context.Background(), hibotv1.V1MessageInjectParams{
				SessionID: args[0],
				Role:      role,
				Content:   resolved,
			})
			if err != nil {
				return err
			}
			format := resolveOutputFormat(cmd)
			e := newEmitter(format, cmd.OutOrStdout())
			return e.emitObject(msg,
				[]string{"ID", "ROLE", "CONTENT"},
				[][]string{{msg.ID, msg.Role, truncate(msg.Content, 60)}})
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "Message role (user/assistant/system)")
	cmd.Flags().StringVar(&content, "content", "", "Message content (use @path/to/file)")
	return cmd
}
