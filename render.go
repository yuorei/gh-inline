package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
)

// renderJSON prints rows as a JSON array. When fields is non-empty, each
// element is reduced to just those keys (matching gh's own --json <fields>
// convention); otherwise the full row is emitted.
func renderJSON(w io.Writer, rows []row, fields []string) error {
	if len(fields) == 0 {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		full, err := toMap(r)
		if err != nil {
			return err
		}
		picked := make(map[string]any, len(fields))
		for _, f := range fields {
			picked[f] = full[f]
		}
		out = append(out, picked)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func toMap(r row) (map[string]any, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// renderTable prints rows grouped by thread: one header line per thread
// (resolved/outdated state, file:line, resolver), then one indented line per
// comment in that thread, replies marked with an arrow.
func renderTable(w io.Writer, rows []row) {
	if len(rows) == 0 {
		fmt.Fprintln(w, "No inline comments found.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	defer tw.Flush()

	currentThread := ""
	for _, r := range rows {
		if r.ThreadID != currentThread {
			currentThread = r.ThreadID
			status := "UNRESOLVED"
			if r.IsResolved {
				status = "RESOLVED"
				if r.ResolvedBy != "" {
					status += " by " + r.ResolvedBy
				}
			}
			outdated := ""
			if r.IsOutdated {
				outdated = " [outdated]"
			}
			loc := r.Path
			switch {
			case r.StartLine != 0 && r.StartLine != r.Line && r.Line != 0:
				loc += ":" + strconv.Itoa(r.StartLine) + "-" + strconv.Itoa(r.Line)
			case r.Line != 0:
				loc += ":" + strconv.Itoa(r.Line)
			}
			fmt.Fprintf(tw, "\n%s%s\t%s\n", status, outdated, loc)
		}

		marker := "-"
		if r.IsReply {
			marker = "  ↳"
		}
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n",
			marker, r.Author, r.UpdatedAt.Format("2006-01-02 15:04"), truncate(oneLine(r.Body), 80))
	}
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
