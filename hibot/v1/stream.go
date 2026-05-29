package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/response"
	ssestream "github.com/volcengine/hiagent-go-sdk/hibot/internal/stream"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

const (
	// 兼容 gateway legacy delivery 路径 (Deliver / DeliverFailed) 发出的事件名；
	// 新 V4 事件路径 (HandleEvent) 已统一为 delta / completed / failed。
	v1SessionChatEventMessageChunkCompat     = "message" + ".chunk"
	v1SessionChatEventMessageCompletedCompat = "message" + ".completed"
	v1SessionChatEventMessageFailedCompat    = "message" + ".failed"
	// 私有化 Hermes runtime 经由 gateway 落地后发出的下划线分隔事件名
	// (message_started / message_delta / message_completed / message_failed
	// / run_completed)。SDK 也将它们归一化到统一三态：delta / completed / failed。
	v1SessionChatEventMessageDeltaUnderscore     = "message_delta"
	v1SessionChatEventMessageChunkUnderscore     = "message_chunk"
	v1SessionChatEventMessageCompletedUnderscore = "message_completed"
	v1SessionChatEventMessageFailedUnderscore    = "message_failed"
	v1SessionChatEventRunCompletedUnderscore     = "run_completed"
	v1SessionChatEventRunFailedUnderscore        = "run_failed"
)

func (s *SessionsService) Chat(ctx context.Context, sessionID string, params V1SessionChatParams) (*V1Message, error) {
	stream := s.chatStream(ctx, sessionID, params, true)
	defer stream.Close()
	for stream.Next() {
		event := stream.Current()
		if event.Type == V1SessionChatEventFailed {
			return nil, fmt.Errorf("hibot: chat failed: %s", event.Error.Message)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}
	return stream.FinalMessage()
}

func (s *SessionsService) ChatStreaming(ctx context.Context, sessionID string, params V1SessionChatParams) *V1SessionChatStream {
	return s.chatStream(ctx, sessionID, params, false)
}

// chatStream 是 Chat / ChatStreaming 的共享底座。
// autoApproveAll=true 时自动注入 Approve="all"：webchat 非流式聚合调用方
// 在收到批回复前不需要显式审批；流式订阅方仍可通过 SSE approval_request
// 事件参与人审。
func (s *SessionsService) chatStream(ctx context.Context, sessionID string, params V1SessionChatParams, autoApproveAll bool) *V1SessionChatStream {
	agentID := params.AgentID
	if agentID == "" {
		agentID = s.agentIDForSession(sessionID)
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"SessionID":   sessionID,
		"AgentID":     agentID,
		"Content":     params.Input,
	}
	if len(params.Files) > 0 {
		body["Files"] = params.Files
	}
	if params.ClientMessageID != "" {
		body["ClientMessageID"] = params.ClientMessageID
	}
	if autoApproveAll {
		body["Approve"] = "all"
	}
	stream := &V1SessionChatStream{}
	resp, err := s.client.requester.DoStream(ctx, request.Action{
		Service: s.client.services.Gateway,
		Version: version.Chat,
		Action:  "Chat",
		Body:    body,
	})
	if err != nil {
		stream.err = err
		return stream
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			stream.err = readErr
			return stream
		}
		stream.err = &response.APIError{StatusCode: resp.StatusCode, Message: string(b)}
		return stream
	}
	stream.resp = resp
	stream.decoder = ssestream.NewDecoder(resp.Body)
	return stream
}

type V1SessionChatStream struct {
	resp    *http.Response
	decoder *ssestream.Decoder
	current V1SessionChatEvent
	final   *V1Message
	err     error
	closed  bool
}

func (s *V1SessionChatStream) Next() bool {
	if s.err != nil || s.decoder == nil {
		return false
	}
	frame, err := s.decoder.Next()
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}
	event, err := decodeChatEvent(frame.Event, frame.Data)
	if err != nil {
		s.err = err
		return false
	}
	s.current = event
	if event.Message != nil {
		s.final = event.Message
	}
	return true
}

func (s *V1SessionChatStream) Current() V1SessionChatEvent {
	return s.current
}

func (s *V1SessionChatStream) Err() error {
	return s.err
}

func (s *V1SessionChatStream) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	if s.resp == nil || s.resp.Body == nil {
		return nil
	}
	return s.resp.Body.Close()
}

func (s *V1SessionChatStream) FinalMessage() (*V1Message, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.final != nil {
		return s.final, nil
	}
	if s.current.Message != nil {
		return s.current.Message, nil
	}
	return nil, fmt.Errorf("hibot: final message is not available")
}

// Accumulate 消费整个流，累加所有 delta.Text，并在 completed 事件到来时
// 合并服务端返回的最终 Message。返回的 V1Message 的 Content 至少包含
// 累加后的 delta 文本；若服务端 completed 携带了完整 Content，则采用
// 服务端版本（视作权威值）。任意 failed 事件会被转化为 error。
func (s *V1SessionChatStream) Accumulate() (*V1Message, error) {
	var (
		buf      []byte
		finalMsg *V1Message
	)
	for s.Next() {
		event := s.current
		switch event.Type {
		case V1SessionChatEventFailed:
			msg := event.Error.Message
			if msg == "" {
				msg = event.Error.Code
			}
			if msg == "" {
				msg = "unknown error"
			}
			return nil, fmt.Errorf("hibot: chat failed: %s", msg)
		case V1SessionChatEventDelta:
			if event.Delta.Text != "" {
				buf = append(buf, event.Delta.Text...)
			}
		case V1SessionChatEventCompleted:
			if event.Message != nil {
				finalMsg = event.Message
			}
		}
	}
	if err := s.err; err != nil {
		return nil, err
	}
	if finalMsg == nil {
		finalMsg = s.final
	}
	if finalMsg == nil {
		if len(buf) == 0 {
			return nil, fmt.Errorf("hibot: final message is not available")
		}
		return &V1Message{Role: "assistant", Content: string(buf)}, nil
	}
	out := *finalMsg
	if out.Content == "" && len(buf) > 0 {
		out.Content = string(buf)
	}
	return &out, nil
}

func decodeChatEvent(eventName, data string) (V1SessionChatEvent, error) {
	event := V1SessionChatEvent{Type: normalizeChatEventName(eventName), rawData: data}
	if data == "" {
		return event, nil
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return event, fmt.Errorf("hibot: decode sse data: %w", err)
	}
	if rawType, ok := payload["type"]; ok && event.Type == "" {
		_ = json.Unmarshal(rawType, &event.Type)
		event.Type = normalizeChatEventName(event.Type)
	}
	if rawRequestID, ok := firstRaw(payload, "request_id", "RequestID", "RequestId"); ok {
		_ = json.Unmarshal(rawRequestID, &event.RequestID)
	}
	if rawDelta, ok := payload["delta"]; ok {
		_ = json.Unmarshal(rawDelta, &event.Delta)
		if event.Delta.Text == "" {
			var deltaText string
			if err := json.Unmarshal(rawDelta, &deltaText); err == nil {
				event.Delta.Text = deltaText
			}
		}
	}
	if rawText, ok := firstRaw(payload, "text", "Text", "content", "Content"); ok && event.Delta.Text == "" {
		_ = json.Unmarshal(rawText, &event.Delta.Text)
	}
	if rawErr, ok := firstRaw(payload, "error", "Error"); ok {
		if err := json.Unmarshal(rawErr, &event.Error); err != nil || event.Error.Message == "" {
			_ = json.Unmarshal(rawErr, &event.Error.Message)
		}
	}
	if rawCode, ok := firstRaw(payload, "code", "Code"); ok && event.Error.Code == "" {
		_ = json.Unmarshal(rawCode, &event.Error.Code)
	}
	if rawMessage, ok := firstRaw(payload, "message", "Message"); ok && event.Error.Message == "" {
		_ = json.Unmarshal(rawMessage, &event.Error.Message)
	}
	if rawMessage, ok := firstRaw(payload, "message", "Message"); ok {
		var msg V1Message
		if err := json.Unmarshal(rawMessage, &msg); err == nil {
			event.Message = &msg
		}
	}
	if event.Type == V1SessionChatEventCompleted && event.Message == nil {
		msg := V1Message{}
		if rawID, ok := firstRaw(payload, "message_id", "MessageID", "ID"); ok {
			_ = json.Unmarshal(rawID, &msg.ID)
		}
		if rawContent, ok := firstRaw(payload, "content", "Content"); ok {
			_ = json.Unmarshal(rawContent, &msg.Content)
		}
		if msg.ID != "" || msg.Content != "" {
			event.Message = &msg
		}
	}
	return event, nil
}

func normalizeChatEventName(name string) string {
	switch name {
	case v1SessionChatEventMessageChunkCompat,
		v1SessionChatEventMessageDeltaUnderscore,
		v1SessionChatEventMessageChunkUnderscore:
		return V1SessionChatEventDelta
	case v1SessionChatEventMessageCompletedCompat,
		v1SessionChatEventMessageCompletedUnderscore,
		v1SessionChatEventRunCompletedUnderscore:
		return V1SessionChatEventCompleted
	case v1SessionChatEventMessageFailedCompat,
		v1SessionChatEventMessageFailedUnderscore,
		v1SessionChatEventRunFailedUnderscore:
		return V1SessionChatEventFailed
	case "tool_started":
		// gateway 实际下发的事件名（见 internal/components/gateway/ssehub.go
		// case RuntimeEventToolStarted）；SDK 常量去掉 -ed 与之统一。
		return V1SessionChatEventToolStart
	case "tool_completed":
		// 同上，对齐 RuntimeEventToolCompleted 的下发名。
		return V1SessionChatEventToolComplete
	default:
		return name
	}
}

func firstRaw(payload map[string]json.RawMessage, keys ...string) (json.RawMessage, bool) {
	for _, key := range keys {
		if raw, ok := payload[key]; ok {
			return raw, true
		}
	}
	return nil, false
}
