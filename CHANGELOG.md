# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-09-07

### Fixed

- **Version-surface lockstep** — backfill the CHANGELOG with the missing
  [0.2.0]/[0.3.0]/[0.4.0] entries: three minor releases of fixes had never been
  documented, so the head still read `[0.1.0]` while the binary reported `0.4.0`.
  Add a `content_version` field to `web/site.json`, and add a single-source-of-
  truth lockstep test (`TestVersionLockstep`) asserting the VERSION file, the
  `mcpx --version` output, the CHANGELOG head, and the site `content_version` all
  agree. Bump the MCP `clientInfo` handshake version literal to stay in lockstep.

## [0.4.0] - 2026-08-28

### Fixed

- **Process-tree teardown** — reap the whole server process group at handshake
  teardown so the `npx`/`uvx` grandchild runtime is not orphaned to init when the
  wrapper is killed (Unix `Setpgid` + `kill(-pgid)`, Windows
  `CREATE_NEW_PROCESS_GROUP` + `taskkill /T`).
- **README backup-path format** — align the "actual output" backup-path examples
  in README.md / README.en.md with the real `backupName` format
  (`.mcpx.bak.<YYYYMMDD-HHMMSS.mmm>`), which the v0.2.0 backup fix had changed
  without updating the docs.

## [0.3.0] - 2026-08-22

### Fixed

- **`--args` comma mis-split** — stop treating a comma in a single token as a
  separator; route every `--args` string through the quote-aware splitter so a
  comma-bearing token (a URL query, minified JSON) stays one arg.
- **Web i18n license** — correct the landing-page license strings from MIT to
  Apache-2.0 to match the LICENSE file and README badges.

## [0.2.0] - 2026-08-02

### Fixed

- **Handshake parent-env inherit** — layer `--env` vars over the parent
  environment when spawning the server (Go's `os/exec` replaces the parent env
  when `cmd.Env` is non-nil), so `npx`/`uvx` catalog servers keep `PATH`/`HOME`
  and the handshake no longer false-fails on a documented `--env TOKEN=abc`.
- **Config-file permissions** — preserve the existing config-file mode on write
  (fall back to `0600` for new files) instead of widening every config to `0644`,
  so a hardened `~/.claude.json` holding MCP env secrets is not silently relaxed.
- **Backup sub-second uniqueness** — give backup files millisecond-resolution
  timestamps plus a `-N` collision guard so scripted back-to-back ops on the same
  config never clobber the prior backup.

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

[0.5.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.5.0
[0.4.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.4.0
[0.3.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.3.0
[0.2.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.2.0
[0.1.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.1.0
