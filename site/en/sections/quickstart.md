From zero to "every client has the same MCP server, handshake-verified" in about two minutes. Pick any of the three install paths — each ends with a single statically compiled `mcpx` binary on your PATH.

### Install option 1 - go install (requires Go 1.24+)

```bash
go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest
```

### Install option 2 - download a prebuilt binary

Every [GitHub Release](https://github.com/SuperMarioYL/mcpx/releases) ships archives for Linux / macOS / Windows × amd64 / arm64, with a SHA-256 `checksums.txt` alongside.

```bash
# Example: macOS Apple Silicon, v0.1.0 — pick your platform on the Releases page
curl -LO https://github.com/SuperMarioYL/mcpx/releases/download/v0.1.0/mcpx_0.1.0_darwin_arm64.tar.gz
tar -xzf mcpx_0.1.0_darwin_arm64.tar.gz
sudo mv mcpx /usr/local/bin/
mcpx --version
```

### Install option 3 - build from source

```bash
git clone https://github.com/SuperMarioYL/mcpx.git
cd mcpx
go build ./cmd/mcpx
./mcpx --version
```

### Run it - mcpx add filesystem

`filesystem` is the local-filesystem server from the built-in catalog (`mcpx catalog` lists all 8). On a machine with Claude Code, Claude Desktop, and Codex installed:

```bash
mcpx add filesystem
```

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at /Users/you/.cursor/mcp.json)

Installing filesystem (npx -y @modelcontextprotocol/server-filesystem .) into 3 client(s)
Write "filesystem" into 3 client(s)? [y/N] y
  Claude Code      backup → write → merged
                   backup: /Users/you/.claude.json.mcpx.bak.20260704-112634
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Claude Desktop   backup → write → merged
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Codex            backup → write → merged
                   ✓ handshake OK (initialize + tools/list, 12 tools)

✓ 3/3 clients ready
Restart (or reload) the affected clients to pick up the new server.
```

### Reading the output

- **Detection** — a client counts as detected only when its config file actually exists on disk; there is no `~/.cursor/mcp.json` here, so Cursor is explicitly skipped — mcpx never invents a config file to guess-write into.
- **Backup → write** — each client gets a timestamped backup first (`<config>.mcpx.bak.<timestamp>`), then only this one entry is merged into that client's own schema; every other key is left untouched.
- **Handshake** — the server is spawned as a stdio subprocess; OK requires both `initialize` and `tools/list` to round-trip, and the tool count is reported.
- **Idempotency** — re-running shows `already up to date (idempotent no-op)`: no file change, no new backup, but the handshake re-runs as a re-verify.
- **Preview and undo** — add `--dry-run` to any command to look without touching; `mcpx remove filesystem` undoes the install (with a fresh backup taken first).
