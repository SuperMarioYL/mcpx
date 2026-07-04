The supply side of MCP servers has exploded — the `awesome-mcp-servers` directory alone has 90k+ stars. Finding a good server keeps getting easier; installing it into your clients is still stuck in the manual era.

### N clients × 5-6 manual steps

Look up where this client keeps its config, look up the schema it expects, hand-edit the JSON or TOML, restart the client, fail, edit again — and that whole loop **multiplies by the number of clients you use**. The same `filesystem` server is written three different ways: Claude Code (`~/.claude.json`, with `"type": "stdio"`), Codex (a TOML table in `~/.codex/config.toml`), Cursor (`~/.cursor/mcp.json`, without `type`). And a client's own installer (e.g. `claude mcp add`) only writes itself — no vendor maintains config writers for its competitors.

### Collapsed into one local command

`mcpx add <server>`, once: it detects every client on your machine, writes each client's own schema, handshake-verifies each one, and finishes with `3/3 clients ready`. A single binary, entirely local — no account, no cloud; the core is open source, MIT-licensed, and will never sit behind a paywall.

### The safety contract: backup-first · idempotent · verified

Before writing, the original config is copied to a timestamped backup; the unit of merge is a single server entry, so everything you maintain by hand stays exactly as it was and re-running is a stable no-op; after writing, the server is launched on the spot for an `initialize` + `tools/list` handshake, so a broken install surfaces immediately. Undo is one `mcpx remove` away.
