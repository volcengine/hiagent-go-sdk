package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// mockTOP is a tiny TOP-style mock for tests in this package; it records the
// last request body keyed by Action and returns canned Result envelopes.
type mockTOP struct {
	t        testing.TB
	server   *httptest.Server
	mu       sync.Mutex
	lastBody map[string]map[string]any
	handlers map[string]string // action -> Result JSON
}

func newMockTOP(t testing.TB, handlers map[string]string) *mockTOP {
	t.Helper()
	m := &mockTOP{
		t:        t,
		lastBody: map[string]map[string]any{},
		handlers: handlers,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handle))
	t.Cleanup(m.server.Close)
	return m
}

func (m *mockTOP) URL() string { return m.server.URL }

func (m *mockTOP) Body(action string) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastBody[action]
}

func (m *mockTOP) handle(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("Action")
	var body map[string]any
	if r.Header.Get("Content-Type") == "application/json" {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	m.mu.Lock()
	m.lastBody[action] = body
	m.mu.Unlock()

	result, ok := m.handlers[action]
	if !ok {
		m.t.Fatalf("unexpected action %q", action)
	}
	_, _ = fmt.Fprintf(w, `{"ResponseMetadata":{"RequestId":"req-test"},"Result":%s}`, result)
}

func TestAgentsCreate_HTTPMock(t *testing.T) {
	mock := newMockTOP(t, map[string]string{
		"CreateAgent": `{"ID":"agent-1"}`,
	})

	root := NewRootCmd()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{
		"--endpoint=" + mock.URL(),
		"--ak=AK", "--sk=SK", "--workspace-id=ws-1", "--region=cn",
		"--output=json",
		"agents", "create",
		"--name=demo", "--model-id=model-1", "--env-id=env-1",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v\noutput: %s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "agent-1") {
		t.Fatalf("expected agent-1 in output, got %q", stdout.String())
	}
	body := mock.Body("CreateAgent")
	if body["Name"] != "demo" {
		t.Fatalf("expected Name=demo in request, got %#v", body)
	}
	if body["WorkspaceID"] != "ws-1" {
		t.Fatalf("expected WorkspaceID injected, got %#v", body)
	}
}

func TestAgentsCreate_MissingFlags(t *testing.T) {
	root := NewRootCmd()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{
		"--endpoint=http://example", "--ak=AK", "--sk=SK", "--workspace-id=ws", "--region=cn",
		"agents", "create",
	})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected user error")
	}
	if ExitCodeFor(err) != 2 {
		t.Fatalf("expected exit code 2, got %d", ExitCodeFor(err))
	}
}
