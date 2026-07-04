# FAQ

## Will mcpx break my existing config?

No — that is the tool's first design constraint. Three layers of protection: **before writing**, the original file is copied in full to a timestamped backup; **while writing**, only the `mcpServers.<name>` entry (or the `[mcp_servers.<name>]` TOML table) is merged — every other key and every other server entry is preserved; **on disk**, writes are atomic (temp file in the same directory, then rename), so even a crash mid-write cannot leave a truncated config file.

The one known exception: Codex's `config.toml` goes through parse-and-re-serialize. All tables and values survive, but **TOML comments and hand-formatting are not preserved**.

## Where do backups live, and what do they look like?

A backup is a sibling of the config file, named:

```text
<config>.mcpx.bak.<YYYYMMDD-HHMMSS>
```

e.g. `~/.claude.json.mcpx.bak.20260704-112634`. Both `add` and `remove` take one before each real modification. mcpx never cleans backups up automatically — to restore, copy the backup back over the original.

## What exactly does "idempotent" mean here?

Run the same `mcpx add` twice and the second run is a stable no-op: mcpx compares the config semantically before and after (key order and whitespace don't count as change), and when the content is identical it writes nothing, creates no new backup, and prints `already up to date (idempotent no-op)`. The handshake still re-runs, though — so a repeated `add` doubles as a "re-verify the connection" command.

## What happens when the handshake fails? Does it roll back my config?

No automatic rollback — that's a deliberate fail-soft design. The three verdicts:

- `✓ OK` — `initialize` and `tools/list` both round-tripped within 8 seconds;
- `! WARN` — the handshake timed out. The entry is **kept, never rolled back**: a server's first run may spend more than 8 seconds just having `npx`/`uvx` fetch its package, and a timeout is not proof of a broken config;
- `✗ FAIL` — the process couldn't start or a protocol error occurred; the reason is reported per client.

In every case the write itself already happened and is backed up. To undo, run `mcpx remove <server>` or restore the backup.

## Which clients are supported? What if mine isn't detected?

Four adapters today: Claude Code (`~/.claude.json`), Claude Desktop (the platform-specific `claude_desktop_config.json` path on macOS / Windows / Linux), Codex (`~/.codex/config.toml`), and Cursor (`~/.cursor/mcp.json`).

Detection is conservative: **a client counts as detected only when its config file exists on disk.** When the file is missing, mcpx prints `not detected (no config at ...)` and skips it — it will never fabricate a config for a client that may not even be installed. Use `--clients` to restrict the operation explicitly.

## How do I remove a server?

```bash
mcpx remove <server>    # alias: mcpx rm
```

Deletes that entry from every client, taking a backup first; other servers and unrelated keys are untouched; clients that never had the entry show `not present (skipped)`. `--dry-run` previews the removal.

## Does mcpx install the server's packages?

No. mcpx writes **launch configuration** (command + args + environment) — it downloads and installs no server binaries or packages. The built-in catalog launches servers through `npx` / `uvx`, and those runners fetch the actual package on the server's first run.

## Does mcpx need an account? Does it talk to the network?

No account, and mcpx itself makes no network requests — detection, backup, and writing are pure local file operations, and the handshake talks to the server over stdio on your machine. (The server under test may use the network itself — e.g. `npx` downloading its package on first run — but that's the server runner's behavior, not mcpx's.)

## Are secrets passed via --env safe?

Values given as `--env K=V` are written **in plain text** into each client's config file — exactly as they would be if you configured the MCP server by hand; that's how the clients' own config formats work — and injected into the server process during the handshake. Treat them like you treat the client config itself: don't commit config files containing secrets to version control.

## Can I see what would happen before committing to it?

Yes — add `--dry-run` to any command: it shows the plan per client (`would backup → write → merge` / `would remove` / `already up to date`), writes no files, takes no backups, runs no handshake, and ends with an explicit `dry-run: no files were written.`
