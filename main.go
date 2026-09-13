package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// version is set at build time via -ldflags "-X main.version=...";
// GoReleaser fills it in for tagged releases (see .goreleaser.yml).
var version = "dev"

const usage = `gh-inline: list inline (review) comments on a pull request, grouped by thread.

USAGE
  gh inline [<PR selector>] [flags]

  <PR selector> is anything "gh pr view" accepts: a number, a URL, a branch
  name, or nothing (uses the PR associated with the current branch).

FLAGS
  -R, --repo owner/repo   Select another repository (default: current directory's)
      --unresolved        Show only unresolved threads
      --resolved          Show only resolved threads
      --outdated          Show only threads left on an outdated diff
      --no-outdated       Hide threads left on an outdated diff
      --file <pattern>    Filter by file path (exact match, or a glob like '*.go')
      --line <n>          Filter by line number
      --author <login>    Filter by comment author
      --json[=fields]     Output JSON. With no value, all fields are included;
                           otherwise a comma-separated list of fields, e.g.
                           --json=path,line,author,body,isResolved
  -h, --help               Show this help
  -v, --version            Show version
`

// jsonValue implements flag.Value + the (unexported) boolFlag interface Go's
// flag package checks for, so both "--json" and "--json=a,b,c" parse.
type jsonValue struct {
	set    bool
	fields []string
}

func (j *jsonValue) String() string   { return strings.Join(j.fields, ",") }
func (j *jsonValue) IsBoolFlag() bool { return true }
func (j *jsonValue) Set(s string) error {
	j.set = true
	if s == "" || s == "true" {
		j.fields = nil
		return nil
	}
	j.fields = strings.Split(s, ",")
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gh-inline:", err)
		os.Exit(1)
	}
}

// reorder moves any positional (non-flag) argument to the end, so a flag
// package's normal "stop at the first non-flag arg" parsing doesn't force
// users to type flags before the PR selector, e.g. `gh-inline 123 --json`.
// gh-inline only ever takes a single optional positional argument, so the
// first token that isn't "-..." and isn't the value of a preceding
// value-taking flag is assumed to be it.
func reorder(args []string) []string {
	valueFlags := map[string]bool{
		"-repo": true, "--repo": true, "-R": true,
		"-file": true, "--file": true,
		"-author": true, "--author": true,
		"-line": true, "--line": true,
	}

	var flags []string
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			if valueFlags[a] && !strings.Contains(a, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positional = append(positional, a)
	}
	return append(flags, positional...)
}

func run(rawArgs []string) error {
	args := reorder(rawArgs)

	fs := flag.NewFlagSet("gh-inline", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	var repo string
	fs.StringVar(&repo, "repo", "", "")
	fs.StringVar(&repo, "R", "", "")

	var unresolved, resolved, outdated, noOutdated bool
	fs.BoolVar(&unresolved, "unresolved", false, "")
	fs.BoolVar(&resolved, "resolved", false, "")
	fs.BoolVar(&outdated, "outdated", false, "")
	fs.BoolVar(&noOutdated, "no-outdated", false, "")

	var file, author string
	var line int
	fs.StringVar(&file, "file", "", "")
	fs.StringVar(&author, "author", "", "")
	fs.IntVar(&line, "line", 0, "")

	var jv jsonValue
	fs.Var(&jv, "json", "")

	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "")
	fs.BoolVar(&showVersion, "v", false, "")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if showVersion {
		fmt.Fprintln(os.Stdout, "gh-inline version", version)
		return nil
	}

	if unresolved && resolved {
		return fmt.Errorf("--unresolved and --resolved cannot be used together")
	}
	if outdated && noOutdated {
		return fmt.Errorf("--outdated and --no-outdated cannot be used together")
	}

	selector := ""
	if fs.NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}
	if fs.NArg() == 1 {
		selector = fs.Arg(0)
	}

	owner, repoName, err := resolveRepo(repo)
	if err != nil {
		return fmt.Errorf("could not resolve repository: %w", err)
	}
	prNumber, err := resolvePR(selector, repo)
	if err != nil {
		return fmt.Errorf("could not resolve pull request: %w", err)
	}

	threads, err := fetchReviewThreads(owner, repoName, prNumber)
	if err != nil {
		return fmt.Errorf("could not fetch review threads: %w", err)
	}

	f := filters{
		resolved:   resolved,
		unresolved: unresolved,
		outdated:   outdated,
		noOutdated: noOutdated,
		file:       file,
		line:       line,
		author:     author,
	}
	rows := flatten(threads, f)

	if jv.set {
		return renderJSON(os.Stdout, rows, jv.fields)
	}
	renderTable(os.Stdout, rows)
	return nil
}
