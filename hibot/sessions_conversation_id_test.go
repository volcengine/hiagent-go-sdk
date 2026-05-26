package hibot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

// ConversationID 必须满足 ^[A-Za-z0-9_-]{1,64}$。
var conversationIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func TestCreateSessionAutoInjectsConversationIDForWebChat(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("Action"); got != "CreateSession" {
			t.Fatalf("Action = %q, want CreateSession", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"ID":"session-1"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	sess, err := client.V1.Sessions.New(context.Background(), V1SessionNewParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if sess.ID != "session-1" {
		t.Fatalf("session.ID = %q, want session-1", sess.ID)
	}

	payload, ok := captured["Payload"].(map[string]any)
	if !ok {
		t.Fatalf("Payload missing or wrong type: %#v", captured["Payload"])
	}
	if got := payload["Channel"]; got != "webchat" {
		t.Fatalf("Channel = %v, want webchat", got)
	}
	cid, ok := payload["ConversationID"].(string)
	if !ok || cid == "" {
		t.Fatalf("ConversationID missing in webchat payload: %#v", payload)
	}
	if !conversationIDRegex.MatchString(cid) {
		t.Fatalf("ConversationID %q does not match %s", cid, conversationIDRegex)
	}
}

func TestCreateSessionSkipsConversationIDForIMChannel(t *testing.T) {
	t.Parallel()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"ID":"session-1"}}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.V1.Sessions.New(context.Background(), V1SessionNewParams{
		AgentID: "agent-1",
		Peer: &V1SessionPeerParams{
			Channel:  "feishu",
			PeerKind: "user",
			PeerID:   "open-id-1",
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	payload := captured["Payload"].(map[string]any)
	if got := payload["Channel"]; got != "feishu" {
		t.Fatalf("Channel = %v, want feishu", got)
	}
	// 非 webchat 渠道 SDK 不注入 ConversationID。
	if _, exists := payload["ConversationID"]; exists {
		t.Fatalf("ConversationID should not be injected for IM channel: %#v", payload)
	}
}

func TestCreateSessionConversationIDsAreUnique(t *testing.T) {
	t.Parallel()

	seen := map[string]struct{}{}
	for i := 0; i < 8; i++ {
		var body map[string]any
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&body)
			_, _ = fmt.Fprint(w, `{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"ID":"session-1"}}`)
		}))
		c := newTestClient(t, ts.URL)
		if _, err := c.V1.Sessions.New(context.Background(), V1SessionNewParams{AgentID: "agent-1"}); err != nil {
			ts.Close()
			t.Fatalf("create session: %v", err)
		}
		ts.Close()
		cid := body["Payload"].(map[string]any)["ConversationID"].(string)
		if _, dup := seen[cid]; dup {
			t.Fatalf("duplicate ConversationID generated: %s", cid)
		}
		seen[cid] = struct{}{}
	}
}
