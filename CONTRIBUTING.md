# Contributing to gh-inline

Thanks for taking the time to contribute! Bug reports, feature requests, and
pull requests are all welcome.

## Development setup

```sh
git clone https://github.com/yuorei/gh-inline.git
cd gh-inline
make build          # builds ./gh-inline
make install        # builds and installs it as `gh inline`
```

The extension has no runtime dependencies beyond the Go standard library and
the `gh` CLI itself (all GitHub API calls are made by shelling out to `gh`).

## Workflow

1. Open an issue first for anything non-trivial, so we can agree on the
   approach before you invest time in it.
2. Create a branch off `main`.
3. Make your change. Keep pull requests focused on a single concern.
4. Run the checks locally before opening a PR:

   ```sh
   make fmt   # gofmt -l .   (should print nothing)
   make vet   # go vet ./...
   make test  # go test ./...
   ```

5. Open a pull request describing what changed and why.

## Testing against a real pull request

Most of the logic (`flatten`, `renderTable`, `renderJSON`, argument parsing)
is covered by unit tests and needs no network access:

```sh
go test ./...
```

To exercise the full path end-to-end (talking to the GitHub API through
`gh`), point the built binary at any public PR:

```sh
go build -o gh-inline .
./gh-inline -R cli/cli 14337
```

## Code style

- `gofmt` is the source of truth for formatting.
- Prefer the Go standard library. This project intentionally avoids external
  Go dependencies (see [README.md](README.md#why-no-dependencies)) -- please
  raise it in an issue first if you think an exception is warranted.
- No `any`/`interface{}` where a concrete type will do.

## Reporting bugs

Please include:

- The exact command you ran (flags included)
- What you expected vs. what happened
- `gh --version` and `gh inline --version` output
- Whether the PR/repo is public or private (no need to share the URL if
  private, but it helps to know)
