# gh-inline

[![CI](https://github.com/yuorei/gh-inline/actions/workflows/ci.yml/badge.svg)](https://github.com/yuorei/gh-inline/actions/workflows/ci.yml)
[![CodeQL](https://github.com/yuorei/gh-inline/actions/workflows/codeql.yml/badge.svg)](https://github.com/yuorei/gh-inline/actions/workflows/codeql.yml)
[![Latest Release](https://img.shields.io/github/v/release/yuorei/gh-inline)](https://github.com/yuorei/gh-inline/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/yuorei/gh-inline.svg)](https://pkg.go.dev/github.com/yuorei/gh-inline)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

Pull Request の**インラインレビューコメント（差分上のコメント）だけ**を、
スレッド単位（resolved / unresolved、outdated を含む）で一覧・フィルタ・JSON出力する
[`gh`](https://cli.github.com/) 拡張機能。

[English README is here](README.md)

## なぜ作ったか

`gh pr view` や `gh pr review` では、これはできない。`gh pr view --comments`
が表示するのは PR 本体の会話（issue コメント）であって、インラインのレビュー
コメントではない。「このPRの未解決インラインコメントスレッドが欲しい」に対応
する `gh pr` サブコマンドは存在しない。resolved 状態・outdated 状態・行位置・
スレッド構造といった情報は GitHub の GraphQL API（`reviewThreads`）にしか
存在せず、`gh` はそれを直接は取得できない。

`gh-inline` はその欠けているコマンド。

```sh
gh inline --unresolved
```

| インラインコメントについて欲しいもの | `gh pr view` | `gh-inline` |
| --- | :---: | :---: |
| コメント本文・投稿者・日時 | ✅（issueコメントと混在） | ✅ |
| 対象ファイル・行番号・複数行範囲 | ❌ | ✅ |
| diffの左右どちら側か | ❌ | ✅ |
| Resolved / Unresolved | ❌ | ✅ |
| Unresolvedだけに絞る | ❌ | ✅ `--unresolved` |
| outdated（古いdiff上）なコメント | ❌ | ✅ `--outdated` / `--no-outdated` |
| スレッド単位・返信も含めてまとめる | ❌ | ✅ |
| ファイル・行・投稿者で絞る | ❌ | ✅ |
| reactions | ❌ | ✅（`--json`） |
| 機械可読な出力 | ⚠️（`--json` はあるがスレッド情報なし） | ✅ `--json[=fields]` |

## インストール

```sh
gh extension install yuorei/gh-inline
```

[`gh`](https://cli.github.com/) v2.0 以上と、認証済みのセッション
（`gh auth login`）が必要。自分でビルドする場合は
[ソースからビルド](#ソースからビルド) を参照。

## 使い方

```sh
gh inline [<PR selector>] [flags]
```

`<PR selector>` は `gh pr view` と同じ規則。PR番号・URL・ブランチ名、または
省略（カレントブランチに紐づくPRを使う）。

```sh
# カレントブランチのPRの、未解決のインラインコメントスレッドだけ
gh inline --unresolved

# 別リポジトリの特定PR
gh inline 123 -R owner/repo

# 特定ファイルへのコメントだけ
gh inline 123 --file internal/api/client.go

# 特定レビュアーが関わったスレッドだけ（返信も含めて全部表示）
gh inline 123 --author octocat

# 差分が変わって古くなったコメントを除外
gh inline 123 --no-outdated

# スクリプト用にJSONで
gh inline 123 --json
gh inline 123 --json=path,line,author,body,isResolved | jq '...'
```

### フラグ一覧

| flag | 説明 |
| --- | --- |
| `-R, --repo owner/repo` | 対象リポジトリ（省略時はカレントディレクトリのリポジトリ） |
| `--unresolved` | unresolved なスレッドのみ |
| `--resolved` | resolved なスレッドのみ |
| `--outdated` | outdated（古いdiff上）なスレッドのみ |
| `--no-outdated` | outdated なスレッドを除外 |
| `--file <pattern>` | 対象ファイルで絞り込み（完全一致、または `*.go` のようなglob） |
| `--line <n>` | 対象行番号で絞り込み |
| `--author <login>` | コメント投稿者で絞り込み |
| `--json[=fields]` | JSON出力。値なしなら全フィールド、`--json=path,line,author` のように指定も可 |
| `-v, --version` | バージョン表示 |
| `-h, --help` | ヘルプ表示 |

スレッド単位のフラグ（`--unresolved` / `--resolved` / `--outdated` /
`--no-outdated` / `--file` / `--line`）は**スレッド全体**に対してマッチする。
スレッドが条件に一致すれば、返信を含むそのスレッドの全コメントが表示される
（返信単体だけが条件に一致してもスレッド自体は表示されない）。`--author` も
同様で、該当ユーザーの投稿が1件でもあるスレッドは会話の文脈ごと表示する
（そのユーザーの投稿だけを抜き出すわけではない）。

### `--json` で出力されるフィールド

```
threadId, isResolved, isOutdated, resolvedBy, path, line, startLine,
diffSide, commentId, databaseId, isReply, replyToId, author, body,
createdAt, updatedAt, url, commit, reactions
```

## 仕組み

`gh-inline` は **外部のGoライブラリに一切依存しない**（標準ライブラリのみ）。
GitHub の API と直接通信するのではなく、すでにインストール・認証済みの
`gh` コマンドをサブプロセスとして呼び出している。

- `gh repo view` / `gh pr view` で、対象リポジトリとPR番号を解決する。
  `gh pr view` 自体と全く同じ挙動（`-R`、ブランチ名指定、引数なし＝カレント
  ブランチのPR、のすべてに対応）。
- `gh api graphql` で `reviewThreads`（resolved / outdated / スレッド構造が
  取得できる唯一の場所）を取得する。スレッド数が多いPRのページネーションにも
  対応。

この方式により、gh-inline は自前で認証・トークン管理・HTTPクライアントを
持つ必要がなく、コード量も小さく監査しやすい状態を保てる。

## ソースからビルド

```sh
git clone https://github.com/yuorei/gh-inline.git
cd gh-inline
make install   # ./gh-inline をビルドして `gh inline` としてインストール
```

開発フロー全体は [CONTRIBUTING.md](CONTRIBUTING.md) を参照
（英語だが、コマンドはそのままコピペで使える）。

## 制限事項

- 1スレッドあたりのコメントは最大100件まで取得する（それ以上ある場合は
  stderr に警告を出す）。
- `--file` の glob マッチは Go の
  [`path.Match`](https://pkg.go.dev/path#Match) の仕様に従う
  （`*` は `/` をまたがない）。

## コミュニティ

- 🐛 バグ報告・要望は [Issueを作成](https://github.com/yuorei/gh-inline/issues/new/choose)。
- 🔧 コードで貢献したい場合は [CONTRIBUTING.md](CONTRIBUTING.md) を参照。
- 🔒 セキュリティ上の問題を見つけた場合は [SECURITY.md](SECURITY.md) を参照
  （公開Issueには書かないでください）。
- 📜 このプロジェクトは [Contributor Covenant](CODE_OF_CONDUCT.md) に従う。
- 📝 主な変更点は [CHANGELOG.md](CHANGELOG.md) に記録する。

## License

[MIT](LICENSE)
