package exampleutil

import (
	"context"
	"fmt"

	"github.com/volcengine/hiagent-go-sdk/hibot"
)

func PrintChatStream(ctx context.Context, client *hibot.Client, sessionID string, params hibot.V1SessionChatParams) error {
	stream := client.V1.Sessions.ChatStreaming(ctx, sessionID, params)
	defer stream.Close()

	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case hibot.V1SessionChatEventDelta:
			fmt.Print(event.Delta.Text)
		case hibot.V1SessionChatEventCompleted:
			fmt.Println("\ncompleted")
		case hibot.V1SessionChatEventFailed:
			return fmt.Errorf("chat failed: %s", event.Error.Message)
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("chat stream: %w", err)
	}
	final, err := stream.FinalMessage()
	if err != nil {
		return fmt.Errorf("final message: %w", err)
	}
	fmt.Printf("message_id=%s\n", final.ID)
	return nil
}
