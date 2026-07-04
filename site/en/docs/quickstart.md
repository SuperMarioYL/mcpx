# Quickstart

From zero to "every client has the same MCP server, handshake-verified" in about two minutes.

## Install

Pick any of the three paths below — each ends with a single statically compiled `mcpx` binary on your PATH.

### Option 1 - go install (requires Go 1.24+)

```bash
go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest
```

### Option 2 - download a prebuilt binary

Every [GitHub Release](https://github.com/SuperMarioYL/mcpx/releases) ships archives for Linux / macOS / Windows × amd64 / arm64 (`.zip` on Windows, `.tar.gz` elsewhere), named after the `mcpx_<version>_<os>_<arch>` template, with a SHA-256 `checksums.txt` alongside.

```bash
# Example: macOS Apple Silicon, v0.1.0 — pick your platform on the Releases page
curl -LO https://github.com/SuperMarioYL/mcpx/releases/download/v0.1.0/mcpx_0.1.0_darwin_arm64.tar.gz
tar -xzf mcpx_0.1.0_darwin_arm64.tar.gz
sudo mv mcpx /usr/local/bin/
mcpx --version
```

### Option 3 - build from source

```bash
git clone https://github.com/SuperMarioYL/mcpx.git
cd mcpx
go build ./cmd/mcpx
./mcpx --version
```

## Your first install: mcpx add filesystem

`filesystem` is the local-filesystem server from the built-in catalog (`mcpx catalog` lists all 8). Run one command:

```bash
mcpx add filesystem
```

On a machine with Claude Code, Claude Desktop, and Codex installed, the output looks like this:

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at /Users/you/.cursor/mcp.json)

Installing filesystem (npx -y @modelcontextprotocol/server-filesystem .) into 3 client(s)
Write "filesystem" into 3 client(s)? [y/N] y
  Claude Code      backup → write → merged
                   backup: /Users/you/.claude.json.mcpx.bak.20260704-112634
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Claude Desktop   backup → write → merged
                   backup: /Users/you/Library/Application Support/Claude/claude_desktop_config.json.mcpx.bak.20260704-112634
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Codex            backup → write → merged
                   backup: /Users/you/.codex/config.toml.mcpx.bak.20260704-112635
                   ✓ handshake OK (initialize + tools/list, 12 tools)

✓ 3/3 clients ready
Restart (or reload) the affected clients to pick up the new server.
```

Line by line:

1. **Detection** — mcpx resolves each known client's config path and counts a client as "detected" only when its config file actually exists. There is no `~/.cursor/mcp.json` here, so Cursor is explicitly skipped — mcpx never invents a config file to guess-write into.
2. **Plan and confirmation** — it prints the exact server launch command and the number of target clients, then asks for confirmation (skip with `--yes` / `-y`; `--dry-run` never prompts).
3. **Backup → write → merge** — for each client, the original config is first copied to a timestamped sibling backup (`<config>.mcpx.bak.<timestamp>`), then only the `filesystem` entry is merged into that client's own schema; every other key is left untouched.
4. **Handshake** — mcpx spawns the server as a stdio subprocess and runs a JSON-RPC `initialize` plus `tools/list`; it reports OK only when both round-trip, along with the number of tools the server exposes.
5. **Summary** — `3/3 clients ready`. Restart (or reload) the clients and the server is live.

## Confirm the result: mcpx list

```bash
mcpx list
```

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at /Users/you/.cursor/mcp.json)

Claude Code (/Users/you/.claude.json)
  filesystem           npx -y @modelcontextprotocol/server-filesystem .
Claude Desktop (/Users/you/Library/Application Support/Claude/claude_desktop_config.json)
  filesystem           npx -y @modelcontextprotocol/server-filesystem .
Codex (/Users/you/.codex/config.toml)
  filesystem           npx -y @modelcontextprotocol/server-filesystem .
```

One block per client: its config path plus every MCP server it currently has, with the full launch command line.

## What happens if you run it again (idempotency)

Re-running `mcpx add filesystem` shows `already up to date (idempotent no-op)` for each client: the file content doesn't change and no new backup is created — but the handshake still runs, so a repeated `add` doubles as a re-verify.

## Preview before touching anything

```bash
mcpx add fetch --dry-run
```

`--dry-run` prints what would happen per client (`would backup → write → merge` or `already up to date`) and writes nothing — no files, no backups, no handshake.

## Undo

```bash
mcpx remove filesystem
```

Removes only the `filesystem` entry (with a fresh backup taken first); other servers and unrelated config keys are untouched.

## Next steps

- Every command and flag: [Command reference](../commands/)
- What happens inside each step: [How it works](../architecture/)
- Safety, backups, failure semantics: [FAQ](../faq/)
