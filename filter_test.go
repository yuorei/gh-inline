package main

import (
	"testing"
	"time"
)

func sampleThreads() []reviewThread {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t1 := reviewThread{
		ID:         "T1",
		IsResolved: false,
		IsOutdated: false,
		Path:       "main.go",
		Line:       new(10),
		DiffSide:   "RIGHT",
	}
	t1.Comments.Nodes = []reviewComment{
		{ID: "C1", Path: "main.go", Line: new(10), Body: "root comment", Author: &actor{Login: "alice"}, CreatedAt: now, UpdatedAt: now},
		{ID: "C2", Path: "main.go", Line: new(10), Body: "a reply", Author: &actor{Login: "bob"}, ReplyTo: &struct {
			ID string `json:"id"`
		}{ID: "C1"}, CreatedAt: now, UpdatedAt: now},
	}

	t2 := reviewThread{
		ID:         "T2",
		IsResolved: true,
		IsOutdated: true,
		Path:       "util.go",
		Line:       new(42),
		DiffSide:   "LEFT",
		ResolvedBy: &actor{Login: "carol"},
	}
	t2.Comments.Nodes = []reviewComment{
		{ID: "C3", Path: "util.go", Line: new(42), Body: "resolved thread", Author: &actor{Login: "carol"}, CreatedAt: now, UpdatedAt: now},
	}

	return []reviewThread{t1, t2}
}

func TestFlattenNoFilter(t *testing.T) {
	rows := flatten(sampleThreads(), filters{})
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0].ThreadID != "T1" || rows[1].ThreadID != "T1" || rows[2].ThreadID != "T2" {
		t.Fatalf("rows not grouped by thread: %+v", rows)
	}
	if !rows[1].IsReply || rows[1].ReplyToID != "C1" {
		t.Fatalf("expected row[1] to be a reply to C1, got %+v", rows[1])
	}
}

func TestFlattenUnresolvedKeepsWholeThread(t *testing.T) {
	rows := flatten(sampleThreads(), filters{unresolved: true})
	if len(rows) != 2 {
		t.Fatalf("expected the unresolved thread's 2 comments (root+reply), got %d", len(rows))
	}
	for _, r := range rows {
		if r.ThreadID != "T1" {
			t.Fatalf("expected only T1 rows, got %+v", r)
		}
	}
}

func TestFlattenResolved(t *testing.T) {
	rows := flatten(sampleThreads(), filters{resolved: true})
	if len(rows) != 1 || rows[0].ThreadID != "T2" {
		t.Fatalf("expected only T2's 1 comment, got %+v", rows)
	}
	if rows[0].ResolvedBy != "carol" {
		t.Fatalf("expected resolvedBy carol, got %q", rows[0].ResolvedBy)
	}
}

func TestFlattenOutdated(t *testing.T) {
	rows := flatten(sampleThreads(), filters{outdated: true})
	if len(rows) != 1 || rows[0].ThreadID != "T2" {
		t.Fatalf("expected only the outdated thread T2, got %+v", rows)
	}

	rows = flatten(sampleThreads(), filters{noOutdated: true})
	if len(rows) != 2 || rows[0].ThreadID != "T1" {
		t.Fatalf("expected only the non-outdated thread T1, got %+v", rows)
	}
}

func TestFlattenFile(t *testing.T) {
	rows := flatten(sampleThreads(), filters{file: "util.go"})
	if len(rows) != 1 || rows[0].Path != "util.go" {
		t.Fatalf("expected only util.go rows, got %+v", rows)
	}

	rows = flatten(sampleThreads(), filters{file: "*.go"})
	if len(rows) != 3 {
		t.Fatalf("expected glob '*.go' to match both threads, got %d rows", len(rows))
	}
}

func TestFlattenLine(t *testing.T) {
	rows := flatten(sampleThreads(), filters{line: 42})
	if len(rows) != 1 || rows[0].ThreadID != "T2" {
		t.Fatalf("expected only line-42 thread, got %+v", rows)
	}
}

func TestFlattenAuthorKeepsWholeThread(t *testing.T) {
	// bob only wrote the reply in T1, but the whole thread should surface.
	rows := flatten(sampleThreads(), filters{author: "bob"})
	if len(rows) != 2 {
		t.Fatalf("expected both of T1's comments when filtering by reply author bob, got %+v", rows)
	}

	rows = flatten(sampleThreads(), filters{author: "nobody"})
	if len(rows) != 0 {
		t.Fatalf("expected no rows for unknown author, got %+v", rows)
	}
}

func TestAuthorLoginHandlesDeletedUser(t *testing.T) {
	if got := authorLogin(nil); got != "ghost" {
		t.Fatalf("expected 'ghost' for a nil author, got %q", got)
	}
}
