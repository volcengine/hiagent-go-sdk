package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func newChatCmd(v *viper.Viper) *cobra.Command {
	var (
		input           string
		stream          bool
		clientMessageID string
	)
	cmd := &cobra.Command{
		Use:   "chat <session-id>",
		Short: "Send a chat message and stream the response",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			text, err := resolveChatInput(cmd, input)
			if err != nil {
				return err
			}
			client, err := buildClient(v)
			if err != nil {
				return err
			}
			params := hibot.V1SessionChatParams{
				Input:           text,
				ClientMessageID: clientMessageID,
			}
			verbose, _ := cmd.Flags().GetBool(flagVerbose)
			ctx := context.Background()
			out := cmd.OutOrStdout()

			if !stream {
				msg, err := client.V1.Sessions.Chat(ctx, sessionID, params)
				if err != nil {
					return err
				}
				format := resolveOutputFormat(cmd)
				e := newEmitter(format, out)
				return e.emitObject(msg,
					[]string{"ID", "ROLE", "CONTENT"},
					[][]string{{msg.ID, msg.Role, msg.Content}})
			}
			// Streaming path: write deltas directly, end with [completed].
			s := client.V1.Sessions.ChatStreaming(ctx, sessionID, params)
			defer s.Close()
			return runStreamingChat(s, out, verbose)
		},
	}
	cmd.Flags().StringVar(&input, "input", "", "Chat input text (default: read from stdin)")
	cmd.Flags().BoolVar(&stream, "stream", false, "Stream deltas to stdout")
	cmd.Flags().StringVar(&clientMessageID, "client-message-id", "", "Idempotency key for the user message")
	return cmd
}

// resolveChatInput resolves --input or, when missing, reads from stdin until EOF.
func resolveChatInput(cmd *cobra.Command, flagValue string) (string, error) {
	if flagValue != "" {
		return readContentArg(flagValue)
	}
	stdin := cmd.InOrStdin()
	// Avoid blocking forever when stdin is a TTY with no input.
	if f, ok := stdin.(*os.File); ok {
		if info, err := f.Stat(); err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			return "", newUserError("--input is required (or pipe data into stdin)")
		}
	}
	data, err := io.ReadAll(bufio.NewReader(stdin))
	if err != nil {
		return "", err
	}
	text := strings.TrimRight(string(data), "\r\n")
	if text == "" {
		return "", newUserError("--input is required (or pipe data into stdin)")
	}
	return text, nil
}

// runStreamingChat consumes the stream and writes delta text to w. completed
// triggers `\n[completed message_id=...]`; failed becomes an error. Other
// events are silenced unless verbose=true.
func runStreamingChat(s *hibot.V1SessionChatStream, w io.Writer, verbose bool) error {
	for s.Next() {
		event := s.Current()
		switch event.Type {
		case hibot.V1SessionChatEventDelta:
			if event.Delta.Text != "" {
				_, _ = io.WriteString(w, event.Delta.Text)
			}
		case hibot.V1SessionChatEventCompleted:
			id := ""
			if event.Message != nil {
				id = event.Message.ID
			}
			fmt.Fprintf(w, "\n[completed message_id=%s]\n", id)
		case hibot.V1SessionChatEventFailed:
			msg := event.Error.Message
			if msg == "" {
				msg = event.Error.Code
			}
			return fmt.Errorf("chat failed: %s", msg)
		default:
			if verbose && event.Type != "" {
				fmt.Fprintf(w, "\n[event:%s]\n", event.Type)
			}
		}
	}
	return s.Err()
}
