# How it works

A single `mcpx add` runs a four-stage pipeline inside one process: **detect → backup → idempotent merge → handshake**. One binary, no daemon, no MCP SDK dependency.

<div class="home-hero mx-auto">
  <img class="hero-img hero-img-light img-fluid" src="/assets/atlas-light.svg" alt="mcpx architecture - mcpx add flows through client detection, per-client adapters, backup plus idempotent merge, and a stdio handshake">
  <img class="hero-img hero-img-dark img-fluid" src="/assets/atlas-dark.svg" alt="mcpx architecture - mcpx add flows through client detection, per-client adapters, backup plus idempotent merge, and a stdio handshake">
</div>

## Stage 1 - Detect

mcpx ships a registry of client adapters. Detection resolves each client's config path on this machine, and a client counts as "detected" **only when its config file actually exists on disk**:

- Config file missing → reported as `not detected`; mcpx never creates a file to guess-write into.
- Path unresolvable on the current OS (e.g. Claude Desktop on an unsupported platform) → reported as `skipped` with the reason, and never becomes a write target.

`--clients` restricts detection to a subset (e.g. `--clients claude-code,codex`).

## Stage 2 - Backup

A backup is taken only when the file content is **actually about to change**: the original config is copied in full to a timestamped sibling

```text
<config>.mcpx.bak.<YYYYMMDD-HHMMSS>
```

e.g. `~/.claude.json.mcpx.bak.20260704-112634`. `remove` backs up before deleting, too. Idempotent no-ops (no content change) create no backup, and neither does `--dry-run`.

## Stage 3 - Idempotent merge

Inside mcpx, a server is a client-agnostic spec (name + command + args + environment). Each client adapter translates that spec into the client's own schema:

- **JSON clients** (Claude Code / Claude Desktop / Cursor): a key-path write sets only the `mcpServers.<name>` key; the rest of the document is preserved as-is. Before writing, mcpx compares the documents semantically (key order and whitespace don't count as change), so repeating the same install is a stable no-op.
- **TOML client** (Codex): the whole `config.toml` is parsed, only the `[mcp_servers.<name>]` table is replaced, and the file is re-serialized — all other tables (e.g. `[projects.*]`) keep their data. Note: since the TOML path is parse-then-re-serialize, **comments and hand-formatting are not preserved** (values are).
- **Atomic writes**: content goes to a temp file in the same directory, then a rename replaces the original — a crash mid-write can never leave a truncated config.

## Stage 4 - Handshake

After writing the config, mcpx launches the server itself as a stdio subprocess and drives a minimal MCP session over raw JSON-RPC:

1. send `initialize` and wait for the response;
2. send the `notifications/initialized` notification;
3. send `tools/list`, wait for the response, and count the tools.

The whole handshake is capped at 8 seconds, with three possible verdicts:

| Verdict | Meaning | Effect on the config |
| --- | --- | --- |
| `✓ OK` | Both `initialize` and `tools/list` round-tripped | Entry is ready |
| `! WARN` | Handshake timed out (server may just start slowly) | **Entry kept, never rolled back** |
| `✗ FAIL` | Spawn failure / protocol error / server closed early | Entry kept; the reason is reported per client |

The handshake is implemented directly on `os/exec` plus standard-library JSON — no MCP SDK dependency — so mcpx is never held hostage by an SDK version.

## Client adapters

| Client | Config file | Format | Entry shape | Detection semantics |
| --- | --- | --- | --- | --- |
| Claude Code | `~/.claude.json` | JSON | `mcpServers.<name>` with `"type": "stdio"` | written iff the file exists |
| Claude Desktop | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`<br>Windows: `%APPDATA%\Claude\claude_desktop_config.json`<br>Linux: `~/.config/Claude/claude_desktop_config.json` | JSON | `mcpServers.<name>` (no `type` field) | written iff the file exists; other OSes report `skipped` |
| Codex | `~/.codex/config.toml` | TOML | `[mcp_servers.<name>]` with `command` / `args` / `env` | written iff the file exists |
| Cursor | `~/.cursor/mcp.json` | JSON | `mcpServers.<name>` (no `type` field) | fail-soft: skipped when no config file |

## What the written entry looks like

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

## Design trade-offs

- **Conservative detection first** — better to under-report than to mis-write: any client whose path isn't confirmed is skipped and says so. That's the right posture for a tool that edits other programs' config files.
- **Touch only your own entry** — the unit of merge is a single server entry; keys the user maintains by hand are never rewritten.
- **Verification as a first-class step** — prove on the spot that the server actually connects, instead of letting the user restart and guess.
- **Fail-soft over rollback** — on a timeout or failure the entry stays in place and the truth is reported: a slow first start is not a broken config, and rolling back could destroy a perfectly correct write. The backup is always there, and undo is one `mcpx remove` away.
