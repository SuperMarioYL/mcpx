Inside mcpx, a server is a client-agnostic spec (name + command + args + environment). Built-in client adapters translate that spec into each client's own schema. Four adapters today:

| Client | Config file | Format | Entry shape | Detection semantics |
| --- | --- | --- | --- | --- |
| Claude Code | `~/.claude.json` | JSON | `mcpServers.<name>` with `"type": "stdio"` | written iff the file exists |
| Claude Desktop | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`<br>Windows: `%APPDATA%\Claude\claude_desktop_config.json`<br>Linux: `~/.config/Claude/claude_desktop_config.json` | JSON | `mcpServers.<name>` (no `type` field) | written iff the file exists; other OSes report `skipped` |
| Codex | `~/.codex/config.toml` | TOML | `[mcp_servers.<name>]` with `command` / `args` / `env` | written iff the file exists |
| Cursor | `~/.cursor/mcp.json` | JSON | `mcpServers.<name>` (no `type` field) | fail-soft: skipped when no config file |

### What each stage means

- **Detection is conservative** — mcpx resolves each client's config path on this machine; a client counts as detected only when its config file actually exists. A missing file is reported as `not detected`; a path unresolvable on the current OS is reported as `skipped` with the reason, and never becomes a write target. `--clients` restricts the scope (e.g. `--clients claude-code,codex`).
- **Backups happen only when something will change** — the original config is copied in full to a timestamped sibling, `<config>.mcpx.bak.<YYYYMMDD-HHMMSS>`; `remove` backs up before deleting too; idempotent no-ops and `--dry-run` create none.
- **The merge touches one entry** — JSON clients get a key-path write that sets only `mcpServers.<name>`, with a semantic comparison first (key order and whitespace don't count as change), so repeats are stable no-ops; the TOML client (Codex) has its whole `config.toml` parsed, only the `[mcp_servers.<name>]` table replaced, and the file re-serialized — all other tables keep their data (comments and hand-formatting are not preserved). Every write is atomic: temp file in the same directory, then rename — a crash mid-write can never leave a truncated config.
- **Verification is a first-class step** — after writing, mcpx launches the server as a stdio subprocess and drives raw JSON-RPC through `initialize`, `notifications/initialized`, and `tools/list`, capped at 8 seconds. `✓ OK` means both requests round-tripped; `! WARN` means a timeout — the entry is kept, never rolled back (a first run may spend more than 8 seconds just having `npx`/`uvx` fetch its package; a timeout is not proof of a broken config); `✗ FAIL` reports the reason per client. The handshake is built directly on `os/exec` plus standard-library JSON — no MCP SDK dependency.

### The same entry, per client

The same `mcpx add filesystem`, in Claude Code (JSON):

```json
{
  "mcpServers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "env": {}
    }
  }
}
```

and in Codex (TOML):

```toml
[mcp_servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "."]
```

(The `env` table is written only when environment variables were supplied via `--env`; Claude Desktop and Cursor use the same JSON shape as Claude Code minus the `type` field.)
