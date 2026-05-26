package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/volcengine/hiagent-go-sdk/hibot"
	hibotv1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

// fakeStream lets us drive runStreamingChat without HTTP. Since
// V1SessionChatStream lives in the SDK and Next/Current require the embedded
// reader, we test runStreamingChat-equivalent logic by reusing the same
// switch via a small replica that mirrors the production switch on
// V1SessionChatEvent.Type. To keep testing close to production, we still call
// the real types from the SDK.

type fakeChatEvent = hibot.V1SessionChatEvent

// runStreamingChatEvents mirrors runStreamingChat but consumes a fixed slice
// of events instead of iterating an SDK stream. This keeps the assertion
// surface stable while exercising the dispatch logic.
func runStreamingChatEvents(events []fakeChatEvent, w *bytes.Buffer, verbose bool) error {
	for _, event := range events {
		switch event.Type {
		case hibot.V1SessionChatEventDelta:
			if event.Delta.Text != "" {
				_, _ = w.WriteString(event.Delta.Text)
			}
		case hibot.V1SessionChatEventCompleted:
			id := ""
			if event.Message != nil {
				id = event.Message.ID
			}
			_, _ = w.WriteString("\n[completed message_id=" + id + "]\n")
		case hibot.V1SessionChatEventFailed:
			msg := event.Error.Message
			if msg == "" {
				msg = event.Error.Code
			}
			return &chatFailedError{msg: msg}
		default:
			if verbose && event.Type != "" {
				_, _ = w.WriteString("\n[event:" + event.Type + "]\n")
			}
		}
	}
	return nil
}

type chatFailedError struct{ msg string }

func (e *chatFailedError) Error() string { return "chat failed: " + e.msg }

func TestRunStreamingChat_DeltaCompleted(t *testing.T) {
	events := []fakeChatEvent{
		{Type: hibot.V1SessionChatEventDelta, Delta: hibot.V1SessionTextDelta{Text: "hel"}},
		{Type: hibot.V1SessionChatEventDelta, Delta: hibot.V1SessionTextDelta{Text: "lo"}},
		{Type: hibot.V1SessionChatEventCompleted, Message: &hibot.V1Message{ID: "m1"}},
	}
	var buf bytes.Buffer
	if err := runStreamingChatEvents(events, &buf, false); err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "hello\n[completed message_id=m1]") {
		t.Fatalf("got %q", out)
	}
}

func TestRunStreamingChat_Failed(t *testing.T) {
	events := []fakeChatEvent{
		{Type: hibot.V1SessionChatEventFailed, Error: hibotv1.V1SessionChatError{Message: "boom"}},
	}
	var buf bytes.Buffer
	if err := runStreamingChatEvents(events, &buf, false); err == nil {
		t.Fatalf("expected error")
	} else if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("got %v", err)
	}
}

func TestRunStreamingChat_VerboseOtherEvents(t *testing.T) {
	events := []fakeChatEvent{
		{Type: hibot.V1SessionChatEventToolStart},
		{Type: hibot.V1SessionChatEventCompleted, Message: &hibot.V1Message{ID: "m"}},
	}
	var buf bytes.Buffer
	if err := runStreamingChatEvents(events, &buf, true); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(buf.String(), "[event:") {
		t.Fatalf("expected verbose event line, got %q", buf.String())
	}

	var buf2 bytes.Buffer
	if err := runStreamingChatEvents(events, &buf2, false); err != nil {
		t.Fatalf("err: %v", err)
	}
	if strings.Contains(buf2.String(), "[event:") {
		t.Fatalf("verbose=false should suppress, got %q", buf2.String())
	}
}
