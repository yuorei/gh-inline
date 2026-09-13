# gh-inline

[![CI](https://github.com/yuorei/gh-inline/actions/workflows/ci.yml/badge.svg)](https://github.com/yuorei/gh-inline/actions/workflows/ci.yml)
[![CodeQL](https://github.com/yuorei/gh-inline/actions/workflows/codeql.yml/badge.svg)](https://github.com/yuorei/gh-inline/actions/workflows/codeql.yml)
[![Latest Release](https://img.shields.io/github/v/release/yuorei/gh-inline)](https://github.com/yuorei/gh-inline/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/yuorei/gh-inline.svg)](https://pkg.go.dev/github.com/yuorei/gh-inline)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

A [`gh`](https://cli.github.com/) extension that lists a pull request's
**inline (diff) review comments** — and only those — grouped by thread, with
resolved/unresolved and outdated filtering built in.

[日本語版 README はこちら](README.ja.md)

## Why

`gh pr view` and `gh pr review` don't do this. `gh pr view --comments` prints
the PR's top-level conversation, not inline review comments. There is no
`gh pr` subcommand for "give me the unresolved inline comment threads on this
PR" — the data (resolved state, outdated state, per-line position, threading)
only exists in GitHub's GraphQL API, one `reviewThreads` query away, but
nothing in `gh` surfaces it directly.

`gh-inline` is that missing command:

```sh
gh inline --unresolved
```

| What you want from inline comments | `gh pr view` | `gh-inline` |
| --- | :---: | :---: |
| Comment body, author, timestamps | ✅ (mixed with issue comments) | ✅ |
| File path / line / line range | ❌ | ✅ |
| Diff side (added/removed) | ❌ | ✅ |
| Resolved vs. unresolved | ❌ | ✅ |
| Filter to unresolved only | ❌ | ✅ `--unresolved` |
| Outdated (stale diff) comments | ❌ | ✅ `--outdated` / `--no-outdated` |
| Grouped by thread, with replies | ❌ | ✅ |
| Filter by file / line / author | ❌ | ✅ |
| Reactions | ❌ | ✅ (`--json`) |
| Machine-readable output | ⚠️ (`--json`, no thread info) | ✅ `--json[=fields]` |

## Installation

```sh
gh extension install yuorei/gh-inline
```

Requires [`gh`](https://cli.github.com/) v2.0+ and an authenticated session
(`gh auth login`). See [Building from source](#building-from-source) to build
it yourself instead.

## Usage

```sh
gh inline [<PR selector>] [flags]
```

`<PR selector>` follows the same rules as `gh pr view`: a PR number, a URL, a
branch name, or nothing at all (uses the PR associated with the current
branch).

```sh
# Unresolved inline comment threads on the current branch's PR
gh inline --unresolved

# A specific PR in a specific repo
gh inline 123 -R owner/repo

# Only comments left on a specific file
gh inline 123 --file internal/api/client.go

# Only what a specific reviewer said (whole thread, replies included)
gh inline 123 --author octocat

# Hide stale comments left on since-changed diff lines
gh inline 123 --no-outdated

# Machine-readable output for scripting
gh inline 123 --json
gh inline 123 --json=path,line,author,body,isResolved | jq '...'
```

### Flags

| Flag | Description |
| --- | --- |
| `-R, --repo owner/repo` | Target repository (default: the current directory's) |
| `--unresolved` | Only unresolved threads |
| `--resolved` | Only resolved threads |
| `--outdated` | Only threads left on an outdated diff |
| `--no-outdated` | Hide threads left on an outdated diff |
| `--file <pattern>` | Filter by file path (exact match, or a glob like `*.go`) |
| `--line <n>` | Filter by line number |
| `--author <login>` | Filter by comment author |
| `--json[=fields]` | JSON output. No value = every field; or a comma-separated list, e.g. `--json=path,line,author` |
| `-v, --version` | Print the version |
| `-h, --help` | Show help |

Thread-level filters (`--unresolved`/`--resolved`/`--outdated`/`--no-outdated`/`--file`/`--line`)
match whole threads: if a thread matches, every comment in it — including
replies — is shown, even if an individual reply wouldn't match on its own.
`--author` works the same way: it surfaces any thread containing a comment
from that author, not just that author's own comments, so you keep the
context of the conversation.

### `--json` fields

```
threadId, isResolved, isOutdated, resolvedBy, path, line, startLine,
diffSide, commentId, databaseId, isReply, replyToId, author, body,
createdAt, updatedAt, url, commit, reactions
```

## How it works

`gh-inline` has **zero external Go dependencies** — only the standard
library. It doesn't talk to GitHub's API directly; instead it shells out to
the `gh` CLI you already have installed and authenticated:

- `gh repo view` / `gh pr view` to resolve the target repository and PR
  number, exactly the way `gh pr view` itself would (so `-R`, branch
  selectors, and "no argument = current branch's PR" all behave the same).
- `gh api graphql` to fetch `reviewThreads` (the only place GitHub exposes
  resolved/outdated/threading state), with pagination handled for PRs with
  many threads.

This means gh-inline never has to manage its own auth, token storage, or HTTP
client — and it stays a small, auditable amount of code.

## Building from source

```sh
git clone https://github.com/yuorei/gh-inline.git
cd gh-inline
make install   # builds ./gh-inline and installs it as `gh inline`
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development workflow.

## Limitations

- Threads with more than 100 comments only fetch the first 100 (a warning is
  printed to stderr when this happens).
- `--file` glob matching uses Go's [`path.Match`](https://pkg.go.dev/path#Match)
  semantics (`*` doesn't cross `/`).

## Community

- 🐛 Found a bug or have an idea? [Open an issue](https://github.com/yuorei/gh-inline/issues/new/choose).
- 🔧 Want to contribute code? See [CONTRIBUTING.md](CONTRIBUTING.md).
- 🔒 Found a security issue? See [SECURITY.md](SECURITY.md) — please don't file a public issue.
- 📜 This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
- 📝 Notable changes are tracked in [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE)
