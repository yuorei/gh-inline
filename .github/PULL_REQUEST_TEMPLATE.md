<!-- Thanks for sending a pull request! -->

### Description

<!-- What does this change do, and why? -->

### How was this tested?

<!--
`go test ./...` covers the pure logic (filtering, rendering, arg parsing).
If you touched anything that talks to `gh` (resolve.go, graphql.go), please
also show the command actually run against a real PR, e.g.:

  ./gh-inline -R owner/repo 123 --unresolved
-->

### Checklist

- [ ] `make check` passes locally (`gofmt`, `go vet`, `go test`)
- [ ] I updated `README.md` **and** `README.ja.md` if user-facing behavior changed
- [ ] I added an entry to `CHANGELOG.md` under `[Unreleased]` if this is a user-facing change
