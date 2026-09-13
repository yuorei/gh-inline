# Security Policy

## Supported Versions

Only the latest released version of `gh-inline` is supported with security
fixes. Since this is a small CLI tool with no persistent state, users are
expected to upgrade with:

```sh
gh extension upgrade inline
```

## Reporting a Vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, use GitHub's
[private vulnerability reporting](https://github.com/yuorei/gh-inline/security/advisories/new)
for this repository. If that's not available, contact the maintainer
directly via the email address on their
[GitHub profile](https://github.com/yuorei).

Please include:

- A description of the vulnerability and its potential impact
- Steps to reproduce it (a minimal example is very helpful)
- The version of `gh-inline` and `gh` you tested against

You should expect an initial response within a few days. This project is
maintained on a best-effort basis, so please be patient.

## Scope

`gh-inline` is a thin, read-only client: it only ever calls `gh repo view`,
`gh pr view`, and `gh api graphql` on your behalf, using your existing `gh`
authentication, and never writes to GitHub or stores credentials of its own.
Reports involving `gh` itself (authentication, token storage, etc.) should go
to [cli/cli](https://github.com/cli/cli/security) instead.
