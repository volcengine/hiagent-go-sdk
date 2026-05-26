package mocktop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type Server struct {
	t      testing.TB
	server *httptest.Server

	mu     sync.Mutex
	seen   map[string]int
	bodies map[string]map[string]any
}

func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{t: t, seen: map[string]int{}, bodies: map[string]map[string]any{}}
	s.server = httptest.NewServer(http.HandlerFunc(s.handle))
	return s
}

func (s *Server) URL() string { return s.server.URL }

func (s *Server) Close() { s.server.Close() }

func (s *Server) Body(action string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bodies[action]
}

func (s *Server) RequireActions(actions ...string) {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, action := range actions {
		if s.seen[action] == 0 {
			s.t.Fatalf("action %s was not called", action)
		}
	}
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("Action")
	s.recordSeen(action)

	var body map[string]any
	if action != "UploadBlob" {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.t.Fatalf("decode %s request: %v", action, err)
		}
		s.recordBody(action, body)
	}

	switch action {
	case "GetModel":
		writeResult(w, `{"Items":[{"ID":"doubao-seed-2.0-pro-260215"}]}`)
	case "ListEnv":
		writeResult(w, `{"Items":[{"ID":"env-1","Name":"default-env","ImageType":"hermes","CreatedAt":"2026-01-01T00:00:00Z"}]}`)
	case "CreateAgent":
		writeResult(w, `{"ID":"agent-1"}`)
	case "CreateSession":
		writeResult(w, `{"ID":"session-1"}`)
	case "Chat":
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "event: delta\ndata: {\"request_id\":\"req-test\",\"delta\":{\"text\":\"ok\"}}\n\n")
		_, _ = fmt.Fprint(w, "event: completed\ndata: {\"request_id\":\"req-test\",\"message\":{\"ID\":\"message-1\",\"Content\":\"ok\"}}\n\n")
	case "UploadBlob":
		if got := r.URL.Query().Get("Filename"); got == "" {
			s.t.Fatalf("UploadBlob Filename is empty")
		}
		writeResult(w, `{"BlobID":"blob-1"}`)
	case "CreateSkill":
		writeResult(w, `{"ID":"skill-version-1"}`)
	case "CreateMCP":
		writeResult(w, `{"ID":"mcp-1"}`)
	case "CreateResource":
		writeResult(w, `{"ID":"resource-1"}`)
	case "CreateAgentPromptTemplate":
		writeResult(w, `{"ID":"prompt-1"}`)
	default:
		s.t.Fatalf("unexpected action %q", action)
	}
}

func (s *Server) recordSeen(action string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[action]++
}

func (s *Server) recordBody(action string, body map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bodies[action] = body
}

func writeResult(w http.ResponseWriter, result string) {
	_, _ = fmt.Fprintf(w, `{"ResponseMetadata":{"RequestId":"req-test"},"Result":%s}`, result)
}
