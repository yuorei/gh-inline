package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestReorderMovesPositionalArgToEnd(t *testing.T) {
	got := reorder([]string{"123", "--unresolved"})
	want := []string{"--unresolved", "123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reorder() = %v, want %v", got, want)
	}
}

func TestReorderKeepsValueFlagPairsTogether(t *testing.T) {
	got := reorder([]string{"-R", "owner/repo", "123", "--file", "main.go"})
	want := []string{"-R", "owner/repo", "--file", "main.go", "123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reorder() = %v, want %v", got, want)
	}
}

func TestReorderHandlesEqualsForm(t *testing.T) {
	got := reorder([]string{"123", "--file=main.go"})
	want := []string{"--file=main.go", "123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reorder() = %v, want %v", got, want)
	}
}

func TestRunRejectsConflictingResolvedFlags(t *testing.T) {
	err := run([]string{"--resolved", "--unresolved"})
	if err == nil {
		t.Fatal("expected an error for --resolved combined with --unresolved")
	}
}

func TestRunRejectsConflictingOutdatedFlags(t *testing.T) {
	err := run([]string{"--outdated", "--no-outdated"})
	if err == nil {
		t.Fatal("expected an error for --outdated combined with --no-outdated")
	}
}

func TestRunRejectsTooManyArguments(t *testing.T) {
	err := run([]string{"123", "456"})
	if err == nil {
		t.Fatal("expected an error for more than one positional argument")
	}
}

func TestRunVersionExitsBeforeAnyGhCall(t *testing.T) {
	// --version must not touch $PATH's `gh`, so this must succeed even with
	// no working directory / repo / auth context at all.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := run([]string{"--version"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("run(--version) returned an error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("gh-inline version")) {
		t.Fatalf("expected version output, got %q", buf.String())
	}
}
