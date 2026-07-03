# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-07-04

First public release. One command installs an MCP server into every coding-agent
client on the machine — no account, no cloud.

### Added

- **`mcpx add <server>`** (m1) — detect installed clients, back up each config,
  idempotently merge one MCP server entry into every client, then run a stdio
  JSON-RPC handshake (`initialize` + `tools/list`) smoke test and print per-client
  OK / WARN / FAIL. A handshake timeout is reported as WARN and the entry is kept,
  never rolled back.
- **`mcpx list`** (m2) — show every detected client, its config path, and the MCP
  servers it currently has installed.
- **`mcpx remove <server>`** (m2) — remove the entry from every client (a fresh
  backup is written first; unrelated entries are preserved).
- **`mcpx catalog`** (m3) — list the 8 built-in servers (`everything`, `fetch`,
  `filesystem`, `git`, `memory`, `sequential-thinking`, `sqlite`, `time`) plus the
  generic from-spec path `--command <bin> --args '<a b c>' --env K=V`.
- **Client adapters** — Claude Code (`~/.claude.json`), Claude Desktop (macOS
  Application Support JSON), Codex (`~/.codex/config.toml`, TOML), and Cursor
  (`~/.cursor/mcp.json`, fail-soft when absent). JSON clients use `sjson`
  key-path merge; the TOML client uses a surgical `go-toml/v2` merge so other
  tables (`[projects.*]`, `[features]`, …) are preserved.
- **Global flags** — `--clients c1,c2`, `--dry-run`, `--yes/-y`, `--version/-v`,
  `--help/-h`.
- Cross-platform single binary, GitHub Actions CI (gofmt + vet + test + build
  matrix), goreleaser release workflow, and a vhs demo tape.

[0.1.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.1.0
