# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release: `gh inline` lists a pull request's inline (diff) review
  comments, grouped by thread.
- Filters: `--unresolved`, `--resolved`, `--outdated`, `--no-outdated`,
  `--file`, `--line`, `--author`.
- `--json[=fields]` for machine-readable output, alongside the default
  human-readable table.
- `-R/--repo` and PR selector resolution matching `gh pr view`.
- `--version` / `-v`.

[Unreleased]: https://github.com/yuorei/gh-inline/compare/main...HEAD
