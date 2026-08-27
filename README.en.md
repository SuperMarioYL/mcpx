<div align="right"><sub><b>English</b>&nbsp;&nbsp;⇄&nbsp;&nbsp;<a href="./README.md">简体中文</a></sub></div>

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/hero-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="./assets/hero-light.svg">
    <img src="./assets/hero-light.svg" width="880" alt="mcpx — one command installs an MCP server into every client you have">
  </picture>
</p>

<p><sub>mcpx is a one-command, no-account, local <b>MCP</b> cross-client installer: detect clients → write config → handshake-verify.</sub></p>

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-black" alt="license"></a>
  <img src="https://img.shields.io/github/v/release/SuperMarioYL/mcpx" alt="latest release">
  <a href="https://github.com/SuperMarioYL/mcpx/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/SuperMarioYL/mcpx/ci.yml?branch=main&label=ci" alt="ci"></a>
  <img src="https://img.shields.io/badge/go-1.24-00ADD8?logo=go&logoColor=white" alt="go 1.24">
  <img src="https://img.shields.io/badge/MCP-ready-5E5CE6" alt="MCP ready">
  <img src="https://img.shields.io/badge/coding%20agent-native-0071E3" alt="coding agent native">
</p>

**Adding an MCP server to your agent still means hand-editing a different JSON per client — `mcpx add <server>` collapses that into one command that writes into every client you have and handshake-verifies each.**

<h2><img src="https://api.iconify.design/tabler:topology-star-3.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Architecture</h2>

One `mcpx add`: detect the coding-agent clients installed on this machine (Claude Code / Claude Desktop / Codex / Cursor), **back up → idempotently merge** into each one in its own schema (JSON key-path or a TOML table), then spawn the server over stdio JSON-RPC and run a handshake. Single process, single binary, no daemon, no MCP SDK.

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/atlas-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="./assets/atlas-light.svg">
    <img src="./assets/atlas-light.svg" width="880" alt="Architecture: mcpx add → detect clients → per-client adapter (Claude Code JSON / Codex TOML / Cursor JSON) → backup + idempotent merge → handshake">
  </picture>
</p>

<h2><img src="https://api.iconify.design/tabler:info-circle.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Why this exists</h2>

The supply side of MCP servers is already huge (`awesome-mcp-servers` 90k★, `chrome-devtools-mcp` 45k★), but wiring one into a client is still manual: every client keeps its config at a **different path with a different schema**. You write it once for each Coding Agent — Claude Code, Cursor, Codex — restart each, and hope the handshake works. `claude mcp add` only solves Claude Code itself; the value concentrates on people who run **more than one client**. mcpx collapses `5-6 steps × N clients` into one local command — backup first, touch only the entry it wrote, verify on the spot.

<h2><img src="https://api.iconify.design/tabler:rocket.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Quickstart</h2>

```bash
go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest   # single binary onto PATH (< 30s)
mcpx add filesystem                                        # detect → backup → write → handshake
mcpx list                                                  # confirm every client got it
```

<details>
<summary>Real output (dev machine with 3 clients)</summary>

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at ~/.cursor/mcp.json)

Installing filesystem (npx -y @modelcontextprotocol/server-filesystem .) into 3 client(s)
  Claude Code      backup → write → merged
                   backup: ~/.claude.json.mcpx.bak.20260704-112615.123
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Claude Desktop   backup → write → merged
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Codex            backup → write → merged
                   ✓ handshake OK (initialize + tools/list, 12 tools)

✓ 3/3 clients ready
Restart (or reload) the affected clients to pick up the new server.
```
</details>

<h2><img src="https://api.iconify.design/tabler:terminal-2.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Usage</h2>

```bash
# Install a built-in server by name (backup → idempotent write → handshake)
mcpx add filesystem

# Not in the catalog? Pass an explicit command / args / env
mcpx add my-server --command npx --args '-y @my-scope/my-mcp-server' --env API_TOKEN=secret

# Restrict to specific clients, skip the prompt
mcpx add git --clients claude-code,codex --yes

# Preview without touching any file
mcpx add fetch --dry-run

# Undo — removes only the entry mcpx wrote, backup preserved
mcpx remove filesystem

# List the built-in catalog
mcpx catalog
```

More in [`examples/quickstart.sh`](./examples/quickstart.sh).

**Subcommands**: `add <server>` · `list` · `remove <server>` (alias `rm`) · `catalog`
**Global flags**: `--clients c1,c2` · `--dry-run` · `--yes/-y` · `--version/-v` · `--help/-h`

<h2><img src="https://api.iconify.design/tabler:photo.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Demo</h2>

![demo](assets/demo.gif)

<h2><img src="https://api.iconify.design/tabler:plug-connected.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Supported clients</h2>

| Client | Config file | Format | Status |
|---|---|---|---|
| Claude Code | `~/.claude.json` | JSON (`mcpServers.<name>`, with `type:"stdio"`) | ✓ verified |
| Claude Desktop | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) | JSON (`mcpServers`) | ✓ verified |
| Codex | `~/.codex/config.toml` | TOML (`[mcp_servers.<name>]`) | ✓ verified |
| Cursor | `~/.cursor/mcp.json` | JSON (`mcpServers`) | fail-soft: skipped if no config |

Undetected clients are never guess-written — mcpx only touches clients whose path is confirmed, and prints a clear "skipped" note for the rest.

<h2><img src="https://api.iconify.design/tabler:git-compare.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Comparison</h2>

Honest positioning — each alternative is better at something:

| Capability | `mcpx` | `claude mcp add` | [awesome-mcp-servers](https://github.com/punkpeye/awesome-mcp-servers) | Smithery / mcp-get |
|---|:--:|:--:|:--:|:--:|
| One command writes into **every** detected client | ✓ | — (Claude Code only) | — (a directory) | partial (single-point / registry) |
| Handshake smoke-test after install | ✓ | — | — | — |
| Backup + idempotent merge (leaves your other keys alone) | ✓ | ✓ | n/a | partial |
| Local, no account, no cloud | ✓ | ✓ | ✓ | partial (cloud / registry-leaning) |
| Server discovery / catalog size | 8 built-in | — | ✓ **90k★, unmatched** | ✓ |

> In one line: catalogs tell you *what exists*, each client's own `add` writes *itself*, and mcpx writes *across clients and proves the server actually connects*.

<h2><img src="https://api.iconify.design/tabler:credit-card.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Pricing</h2>

The local core — detect, backup, cross-client write, handshake — is **open source with no paywall, forever**. Payment sits only at the "team sync / audit" layer:

| Tier | For | Price | What you get |
|---|---|---|---|
| **OSS Core** | Individual developers | Open source, free | Everything above this section |
| **mcpx Teams** | 5–30-person engineering teams | **$5 / seat / month** (team cap from $49/mo, ≤10 seats) | Hosted team manifest (a shared profile of "which servers + versions to install"), `mcpx add --profile <team>` to align the whole team in one command, CI checks that every member's config matches the team profile, audit log |

> Teams is a v0.3+ wedge — v0.1 only stubs the roadmap and ships no Teams code. The local write action never sits behind a paywall.

<h2><img src="https://api.iconify.design/tabler:map-2.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Roadmap</h2>

- [x] **m1** — `mcpx add`: detect → backup → idempotent write → handshake, per-client OK/WARN/FAIL
- [x] **m2** — `mcpx list` / `mcpx remove` (re-backup before delete, keep other entries)
- [x] **m3** — 8-server built-in catalog + `--command/--args/--env` from-spec; cross-platform single binary, bilingual README, demo, CI
- [ ] More client adapters (implemented once each client's config path is confirmed)
- [ ] `mcpx doctor`: health-check the MCP servers each client has installed
- [ ] **mcpx Teams**: shared profile + CI gate + audit (the paid wedge)

<h2><img src="https://api.iconify.design/tabler:license.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> License & contributing</h2>

Apache-2.0. Issues and PRs welcome — please mention which clients you use and which one you'd like supported next.

## Share this

```
mcpx — the one-command MCP installer that writes into every coding agent you have (Claude Code, Codex, Cursor) and handshake-verifies each. No account, no cloud. https://github.com/SuperMarioYL/mcpx
```

<p align="center"><sub><a href="./LICENSE">Apache-2.0</a> © 2026 SuperMarioYL</sub></p>
