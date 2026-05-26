package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	obj := map[string]any{"id": "a", "n": 1}
	if err := renderJSON(&buf, obj); err != nil {
		t.Fatalf("renderJSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["id"] != "a" {
		t.Fatalf("got %v", got)
	}
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	if err := renderTable(&buf, []string{"ID", "NAME"}, [][]string{{"1", "alpha"}, {"22", "beta"}}); err != nil {
		t.Fatalf("renderTable: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") {
		t.Fatalf("missing header: %q", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("missing rows: %q", out)
	}
}

func TestEmitterMessage_SuppressedForJSON(t *testing.T) {
	var buf bytes.Buffer
	e := newEmitter("json", &buf)
	e.emitMessage("hello")
	if buf.Len() != 0 {
		t.Fatalf("emitMessage should be suppressed for json, got %q", buf.String())
	}
	var buf2 bytes.Buffer
	e2 := newEmitter("table", &buf2)
	e2.emitMessage("hi %d", 1)
	if !strings.Contains(buf2.String(), "hi 1") {
		t.Fatalf("expected hi 1, got %q", buf2.String())
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hello", 0) != "hello" {
		t.Fatalf("n=0 should not truncate")
	}
	if got := truncate("hello", 3); got != "he…" {
		t.Fatalf("got %q", got)
	}
	if got := truncate("hi", 5); got != "hi" {
		t.Fatalf("short input should pass through, got %q", got)
	}
}
