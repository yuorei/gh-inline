package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func sampleRows() []row {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return []row{
		{ThreadID: "T1", IsResolved: false, Path: "main.go", Line: 10, Author: "alice", Body: "root comment", UpdatedAt: now},
		{ThreadID: "T1", IsResolved: false, Path: "main.go", Line: 10, Author: "bob", Body: "a reply", IsReply: true, ReplyToID: "C1", UpdatedAt: now},
		{ThreadID: "T2", IsResolved: true, ResolvedBy: "carol", Path: "util.go", Line: 42, Author: "carol", Body: "resolved thread", UpdatedAt: now},
	}
}

func TestRenderTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	renderTable(&buf, nil)
	if !strings.Contains(buf.String(), "No inline comments found") {
		t.Fatalf("expected an empty-state message, got %q", buf.String())
	}
}

func TestRenderTableGroupsByThread(t *testing.T) {
	var buf bytes.Buffer
	renderTable(&buf, sampleRows())
	out := buf.String()

	if strings.Count(out, "UNRESOLVED") != 1 {
		t.Fatalf("expected exactly one UNRESOLVED header for thread T1, got:\n%s", out)
	}
	if !strings.Contains(out, "RESOLVED by carol") {
		t.Fatalf("expected a 'RESOLVED by carol' header, got:\n%s", out)
	}
	if !strings.Contains(out, "↳") {
		t.Fatalf("expected the reply marker for bob's comment, got:\n%s", out)
	}
}

func TestRenderTableLineRange(t *testing.T) {
	var buf bytes.Buffer
	renderTable(&buf, []row{
		{ThreadID: "T1", Path: "main.go", StartLine: 10, Line: 15, Author: "alice", Body: "multi-line comment"},
	})
	if !strings.Contains(buf.String(), "main.go:10-15") {
		t.Fatalf("expected a line range in the header, got:\n%s", buf.String())
	}
}

func TestRenderJSONAllFields(t *testing.T) {
	var buf bytes.Buffer
	if err := renderJSON(&buf, sampleRows(), nil); err != nil {
		t.Fatalf("renderJSON: %v", err)
	}

	var out []row
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON matching row: %v\n%s", err, buf.String())
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(out))
	}
}

func TestRenderJSONFieldSelection(t *testing.T) {
	var buf bytes.Buffer
	if err := renderJSON(&buf, sampleRows(), []string{"path", "author"}); err != nil {
		t.Fatalf("renderJSON: %v", err)
	}

	var out []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	for _, o := range out {
		if len(o) != 2 {
			t.Fatalf("expected exactly 2 selected fields, got %v", o)
		}
		if _, ok := o["path"]; !ok {
			t.Fatalf("expected 'path' field, got %v", o)
		}
		if _, ok := o["author"]; !ok {
			t.Fatalf("expected 'author' field, got %v", o)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Fatalf("expected short strings to pass through unchanged, got %q", got)
	}
	got := truncate("hello world", 6)
	if []rune(got)[5] != '…' {
		t.Fatalf("expected truncated string to end with an ellipsis, got %q", got)
	}
}

func TestOneLine(t *testing.T) {
	if got := oneLine("a\nb\r\nc"); got != "a b c" {
		t.Fatalf("expected newlines collapsed to spaces, got %q", got)
	}
}
