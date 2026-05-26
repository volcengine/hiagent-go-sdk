package v1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

// generateConversationID 生成符合 `^[A-Za-z0-9_-]{1,64}$` 的 ConversationID：
// 使用 16 字节随机数转 32 位 hex，长度与字符集均严格满足。
//
// rand.Read 失败概率极低（仅当系统熵源不可用），失败时返回空串让上层退化
// 为不携带 ConversationID 的旧行为，保证主流程不被熵源故障拖垮。
var generateConversationID = func() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buf[:])
}

type SessionsService struct {
	client        *Client
	mu            sync.RWMutex
	sessionAgents map[string]string
}

func newSessionsService(c *Client) *SessionsService {
	return &SessionsService{
		client:        c,
		sessionAgents: make(map[string]string),
	}
}

func (s *SessionsService) New(ctx context.Context, params V1SessionNewParams) (*V1Session, error) {
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"AgentID":     params.AgentID,
	}
	// Peer 仅在调用方需要显式指定渠道（飞书 / 企微 等 IM 渠道）或按 user 隔离
	// 会话（SaaS 多租户嵌入）时才传入。WebChat 主流程没有外部 Channel/Peer
	// 概念，由 SDK 注入 webchat/system 默认值（PeerID 用 AgentID 兜底）。
	payload := map[string]any{
		"Channel":  "webchat",
		"PeerKind": "system",
		"PeerID":   params.AgentID,
	}
	if params.Peer != nil {
		if params.Peer.Channel != "" {
			payload["Channel"] = params.Peer.Channel
		}
		if params.Peer.PeerKind != "" {
			payload["PeerKind"] = params.Peer.PeerKind
		}
		if params.Peer.PeerID != "" {
			payload["PeerID"] = params.Peer.PeerID
		}
	}
	// ConversationID 仅在 webchat 渠道由 SDK 自动生成并透传；其它渠道留空，
	// 这里直接跳过以避免污染请求体。
	if payload["Channel"] == "webchat" {
		if cid := generateConversationID(); cid != "" {
			payload["ConversationID"] = cid
		}
	}
	body["Payload"] = payload
	var result V1Session
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateSession",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create session response missing ID")
	}
	result.AgentID = params.AgentID
	if params.AgentID != "" {
		s.mu.Lock()
		s.sessionAgents[result.ID] = params.AgentID
		s.mu.Unlock()
	}
	return &result, nil
}

func (s *SessionsService) List(ctx context.Context, params V1SessionListParams) (*V1SessionList, error) {
	var result V1SessionList
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListSessions",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *SessionsService) Get(ctx context.Context, params V1SessionGetParams) (*V1Session, error) {
	if params.SessionID == "" {
		return nil, errors.New("hibot: session id is required")
	}
	var result V1Session
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetSession",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get session response missing ID")
	}
	return &result, nil
}

func (s *SessionsService) GetByKey(ctx context.Context, params V1SessionGetByKeyParams) (*V1Session, error) {
	if params.SessionKey == "" {
		return nil, errors.New("hibot: session key is required")
	}
	var result V1Session
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetSessionByKey",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get session by key response missing ID")
	}
	return &result, nil
}

func (s *SessionsService) Archive(ctx context.Context, params V1SessionArchiveParams) error {
	if params.SessionID == "" {
		return errors.New("hibot: session id is required")
	}
	payload := map[string]any{}
	if params.Summary != "" {
		payload["Summary"] = params.Summary
	}
	if params.Consolidate != nil {
		payload["Consolidate"] = *params.Consolidate
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"SessionID":   params.SessionID,
	}
	if len(payload) > 0 {
		body["Payload"] = payload
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ArchiveSession",
		Body:    body,
	}, nil)
}

func (s *SessionsService) Delete(ctx context.Context, params V1SessionDeleteParams) error {
	if params.SessionID == "" {
		return errors.New("hibot: session id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteSession",
		Body:    params,
	}, nil)
}

func (s *SessionsService) ListMessages(ctx context.Context, params V1MessageListParams) (*V1MessageList, error) {
	if params.SessionID == "" {
		return nil, errors.New("hibot: session id is required")
	}
	var result V1MessageList
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListMessages",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *SessionsService) GetMessage(ctx context.Context, params V1MessageGetParams) (*V1Message, error) {
	if params.SessionID == "" || params.MessageID == "" {
		return nil, errors.New("hibot: session id and message id are required")
	}
	var result V1Message
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetMessage",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get message response missing ID")
	}
	return &result, nil
}

func (s *SessionsService) InjectMessage(ctx context.Context, params V1MessageInjectParams) (*V1Message, error) {
	if params.SessionID == "" {
		return nil, errors.New("hibot: session id is required")
	}
	payload := map[string]any{}
	if params.Role != "" {
		payload["Role"] = params.Role
	}
	if params.Content != "" {
		payload["Content"] = params.Content
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"SessionID":   params.SessionID,
		"Payload":     payload,
	}
	var result V1Message
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "InjectMessage",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: inject message response missing ID")
	}
	return &result, nil
}

func (s *SessionsService) agentIDForSession(sessionID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionAgents[sessionID]
}
