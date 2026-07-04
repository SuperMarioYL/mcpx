# Why mcpx

## The pain: N clients × 5-6 manual steps each

The supply side of MCP servers has exploded — the `awesome-mcp-servers` directory alone has 90k+ stars, and `chrome-devtools-mcp`, shipped by the Chrome DevTools team itself, has 45k+. Finding a good server keeps getting easier. **Installing it into your clients is still stuck in the manual era:**

1. Look up where this client keeps its config (every client uses a different path);
2. Look up the schema it expects (every client uses a different shape);
3. Hand-edit the JSON or TOML and hope you didn't break a neighboring key;
4. Restart the client;
5. Try it, fail, edit again, restart again.

And that whole loop **multiplies by the number of clients you use**. The same `filesystem` server is written three different ways across three clients:

- Claude Code: `mcpServers.<name>` inside `~/.claude.json`, with `"type": "stdio"`;
- Codex: a `[mcp_servers.<name>]` TOML table inside `~/.codex/config.toml`;
- Cursor: `mcpServers.<name>` inside `~/.cursor/mcp.json`, without `type`.

mcpx collapses `5-6 steps × N clients` into one command: `mcpx add <server>` — detect, back up, write each client's own schema, handshake-verify each one, and finish with `3/3 clients ready`.

## Versus a client's own command (e.g. `claude mcp add`)

Claude Code's built-in `claude mcp add` is genuinely good — **but it only writes Claude Code's own config**. That's not an oversight; it's structural: no client vendor is going to maintain config writers for its competitors. However many clients you run, that's how many commands you learn and how many times you repeat yourself.

mcpx stands outside all of them: one command writes into **every** client it detects on your machine. If you only ever use one client, you may genuinely not need mcpx — its value concentrates on multi-client workflows, and running multiple coding agents side by side is fast becoming the norm.

## Versus hand-editing config files

Hand-editing always works, and it's the most error-prone path there is: wrong schema, a mangled neighboring key, one client forgotten, no verification at the end. Every mcpx step targets one of those failure modes:

- **Backup first** — the original file is copied to a timestamped backup before anything is written;
- **Touch one entry only** — the unit of merge is a single server entry; everything you maintain by hand stays exactly as it was;
- **Verify immediately** — the server is launched on the spot for an `initialize` + `tools/list` handshake, so a broken install surfaces now, not after you've restarted the client.

## Versus directories and registry-style installers

Directories like `awesome-mcp-servers` answer "what exists" — but after you've found a server, none of the clone / configure / restart work goes away. A directory is not an installer. Registry-style tools such as Smithery and mcp-get lean toward single-point installs or cloud registries, with some capabilities behind accounts.

mcpx's difference is three things stacked together: **(1) it writes into every client detected on your machine**, not just one place; **(2) it runs a handshake smoke test after installing**, proving the server actually connects; **(3) it is entirely local — no account, no cloud.**

| Capability | mcpx | `claude mcp add` | Directory (awesome-mcp-servers) | Smithery / mcp-get |
| --- | :-: | :-: | :-: | :-: |
| One command writes into **every** detected client | ✓ | — (Claude Code only) | — (it's a directory) | partial (single-point / registry) |
| Handshake smoke test after install | ✓ | — | — | — |
| Backup + idempotent merge (leaves your other keys alone) | ✓ | ✓ | n/a | partial |
| Local, no account, no cloud | ✓ | ✓ | ✓ | partial (cloud / registry-leaning) |
| Server discovery / catalog size | 8 built-in | — | ✓ 90k★, unmatched | ✓ |

In one line: directories tell you *what exists*, each client's own `add` writes *itself*, and mcpx *writes across clients and proves the server actually connects*.

## Honest boundaries

What mcpx deliberately does not do is worth stating just as plainly:

- **It doesn't host servers** — it writes local launch configs, it is not a cloud service;
- **It doesn't install dependencies** — catalog entries are launch commands; `npx` / `uvx` fetch the packages on the server's first run;
- **It only manages the entries it writes** — `remove` deletes the entry matching the name you give, nothing else;
- **It never guesses paths** — clients whose config path isn't confirmed on the current OS are skipped, and say so.

The local core — detect, back up, cross-client write, handshake — is **open source, MIT-licensed, account-free, and will never sit behind a paywall**. A team-oriented config-sync layer is on the roadmap, but it is not implemented today and takes nothing away from the local core.
