package main

import (
	"path"
	"time"
)

// row is one inline review comment, flattened out of its thread with the
// thread's resolved/outdated/position state denormalized onto it. This is
// the unit both table rendering and --json operate on.
type row struct {
	ThreadID   string    `json:"threadId"`
	IsResolved bool      `json:"isResolved"`
	IsOutdated bool      `json:"isOutdated"`
	ResolvedBy string    `json:"resolvedBy,omitempty"`
	Path       string    `json:"path"`
	Line       int       `json:"line,omitempty"`
	StartLine  int       `json:"startLine,omitempty"`
	DiffSide   string    `json:"diffSide,omitempty"`
	CommentID  string    `json:"commentId"`
	DatabaseID string    `json:"databaseId,omitempty"`
	IsReply    bool      `json:"isReply"`
	ReplyToID  string    `json:"replyToId,omitempty"`
	Author     string    `json:"author"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	URL        string    `json:"url"`
	Commit     string    `json:"commit,omitempty"`
	Reactions  []string  `json:"reactions,omitempty"`
}

// flatten turns threads into rows. Filtering by resolved/outdated/file/line
// happens at thread level first (a thread either matches or it doesn't), so
// that when a thread matches, all of its comments -- including replies --
// stay together in the output.
func flatten(threads []reviewThread, f filters) []row {
	var rows []row

	for _, t := range threads {
		if !f.matchesThread(t) {
			continue
		}
		if f.author != "" && !threadHasAuthor(t, f.author) {
			continue
		}

		resolvedBy := ""
		if t.ResolvedBy != nil {
			resolvedBy = t.ResolvedBy.Login
		}

		for _, c := range t.Comments.Nodes {
			r := row{
				ThreadID:   t.ID,
				IsResolved: t.IsResolved,
				IsOutdated: t.IsOutdated,
				ResolvedBy: resolvedBy,
				Path:       coalesce(c.Path, t.Path),
				DiffSide:   t.DiffSide,
				CommentID:  c.ID,
				DatabaseID: c.FullDatabaseID,
				Author:     authorLogin(c.Author),
				Body:       c.Body,
				CreatedAt:  c.CreatedAt,
				UpdatedAt:  c.UpdatedAt,
				URL:        c.URL,
			}
			if c.Line != nil {
				r.Line = *c.Line
			} else if c.OriginalLine != nil {
				r.Line = *c.OriginalLine
			} else if t.Line != nil {
				r.Line = *t.Line
			}
			if c.StartLine != nil {
				r.StartLine = *c.StartLine
			} else if t.StartLine != nil {
				r.StartLine = *t.StartLine
			}
			if c.ReplyTo != nil {
				r.IsReply = true
				r.ReplyToID = c.ReplyTo.ID
			}
			if c.Commit != nil {
				r.Commit = c.Commit.OID
			}
			for _, rx := range c.Reactions.Nodes {
				r.Reactions = append(r.Reactions, rx.Content)
			}

			rows = append(rows, r)
		}
	}

	return rows
}

type filters struct {
	resolved   bool // only threads matching this...
	unresolved bool // ...pair is mutually exclusive, validated in main
	outdated   bool
	noOutdated bool
	file       string
	line       int
	author     string
}

func (f filters) matchesThread(t reviewThread) bool {
	if f.resolved && !t.IsResolved {
		return false
	}
	if f.unresolved && t.IsResolved {
		return false
	}
	if f.outdated && !t.IsOutdated {
		return false
	}
	if f.noOutdated && t.IsOutdated {
		return false
	}
	if f.file != "" {
		ok, err := path.Match(f.file, t.Path)
		if (err != nil || !ok) && t.Path != f.file {
			return false
		}
	}
	if f.line != 0 {
		if !threadHasLine(t, f.line) {
			return false
		}
	}
	return true
}

func threadHasLine(t reviewThread, line int) bool {
	if t.Line != nil && *t.Line == line {
		return true
	}
	if t.StartLine != nil && *t.StartLine == line {
		return true
	}
	for _, c := range t.Comments.Nodes {
		if c.Line != nil && *c.Line == line {
			return true
		}
		if c.OriginalLine != nil && *c.OriginalLine == line {
			return true
		}
	}
	return false
}

func threadHasAuthor(t reviewThread, login string) bool {
	for _, c := range t.Comments.Nodes {
		if authorLogin(c.Author) == login {
			return true
		}
	}
	return false
}

func authorLogin(a *actor) string {
	if a == nil {
		return "ghost"
	}
	return a.Login
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
