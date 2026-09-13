package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// resolveRepo mirrors `gh repo view`'s own repository resolution (current
// directory's remote, or -R/--repo when given) so gh-inline behaves like any
// other gh subcommand.
func resolveRepo(repoOverride string) (owner, name string, err error) {
	args := []string{"repo", "view", "--json", "owner,name"}
	if repoOverride != "" {
		args = append(args, repoOverride)
	}

	out, err := runGh(args...)
	if err != nil {
		return "", "", err
	}

	var parsed struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return "", "", fmt.Errorf("failed to parse gh repo view output: %w", err)
	}
	return parsed.Owner.Login, parsed.Name, nil
}

// resolvePR mirrors `gh pr view`'s own selector resolution: a PR number, a
// URL, a branch name, or (if empty) the current branch's PR.
func resolvePR(selector, repoOverride string) (int, error) {
	args := []string{"pr", "view", "--json", "number"}
	if selector != "" {
		args = append(args, selector)
	}
	if repoOverride != "" {
		args = append(args, "-R", repoOverride)
	}

	out, err := runGh(args...)
	if err != nil {
		return 0, err
	}

	var parsed struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return 0, fmt.Errorf("failed to parse gh pr view output: %w", err)
	}
	return parsed.Number, nil
}

func runGh(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return stdout.Bytes(), nil
}
