# Command reference

```text
mcpx <command> [flags]
```

| Command | Purpose |
| --- | --- |
| `mcpx add <server>` | Install an MCP server into every detected client (backup → idempotent write → handshake) |
| `mcpx list` | Show detected clients and the MCP servers each has installed |
| `mcpx remove <server>` | Remove the server entry from every client (alias `rm`; backup taken first) |
| `mcpx catalog` | List the built-in servers installable by name |

## Global flags (available on every subcommand)

| Flag | Description |
| --- | --- |
| `--clients <id1,id2>` | Restrict the operation to a subset of clients; valid ids: `claude-code`, `claude-desktop`, `codex`, `cursor` |
| `--dry-run` | Show what would change; writes nothing, takes no backups, runs no handshake |
| `--yes` / `-y` | Skip the confirmation prompt before writing/removing (for scripts and CI) |
| `--version` / `-v` | Print the version |
| `--help` / `-h` | Help |

## mcpx add

```text
mcpx add <server> [--command <bin>] [--args '<a b c>'] [--env K=V ...]
```

`<server>` is a name from the built-in catalog, or any entry name of your choosing. Behavior:

- Built-in servers (see the catalog below) install by name alone: `mcpx add filesystem`.
- Anything else takes the from-spec path: `--command` is required, `--args` and `--env` are optional; a non-empty `--command` also overrides a catalog entry of the same name.
- Before writing, mcpx prompts `Write "<server>" into N client(s)? [y/N]` (skipped by `--yes` and `--dry-run`).
- Each client reports independently: the write outcome (`backup → write → merged` / `already up to date (idempotent no-op)`) plus the handshake verdict (OK / WARN / FAIL).
- Repeating the same `add` is a stable no-op: no file change, no new backup — but the handshake re-runs, re-verifying the server.

| Flag | Description |
| --- | --- |
| `--command <bin>` | Executable that launches the server (required for non-catalog servers), e.g. `npx`, `uvx`, or an absolute path |
| `--args '<a b c>'` | Server arguments, space-separated; quoting is honored for arguments containing spaces (`'a "b c" d'`); a comma-separated form (`'-y,pkg,.'`) is also accepted when the string has no spaces |
| `--env K=V` | Environment variable for the server process, repeatable; written into the client config's `env` field and injected during the handshake |

```bash
mcpx add filesystem
mcpx add mytool --command npx --args '-y my-mcp-server' --env TOKEN=abc
mcpx add filesystem --clients claude-code,codex --dry-run
```

## mcpx list

```text
mcpx list
```

Prints the detection summary first (which clients were found, which weren't, which were skipped and why), then one block per client: its config file path plus the name and full command line of every installed MCP server. A client with none shows `(no MCP servers)`.

## mcpx remove

```text
mcpx remove <server>    # alias: mcpx rm <server>
```

- Deletes the entry named `<server>` from every detected client, taking a timestamped backup of each file it is about to modify.
- Removes only that one entry — other servers and unrelated config keys are preserved.
- A client that never had the entry shows `not present (skipped)`; that is not an error.
- Supports `--dry-run` (shows `would remove`) and `--yes` like everything else.

## mcpx catalog

```text
mcpx catalog
```

Lists the built-in server catalog. Catalog entries are **launch commands only** — mcpx does not install the underlying packages; `npx` / `uvx` fetch them on the server's first run:

| Name | Description | Launch command |
| --- | --- | --- |
| `everything` | Reference server exercising all MCP features (for testing) | `npx -y @modelcontextprotocol/server-everything` |
| `fetch` | Fetch and convert web pages to markdown | `uvx mcp-server-fetch` |
| `filesystem` | Local filesystem access (read/write under given roots) | `npx -y @modelcontextprotocol/server-filesystem .` |
| `git` | Git repository operations (log / diff / status / blame) | `uvx mcp-server-git` |
| `memory` | Knowledge-graph memory store | `npx -y @modelcontextprotocol/server-memory` |
| `sequential-thinking` | Structured step-by-step reasoning scratchpad | `npx -y @modelcontextprotocol/server-sequential-thinking` |
| `sqlite` | Query a local SQLite database | `uvx mcp-server-sqlite --db-path ./data.db` |
| `time` | Time and timezone conversion utilities | `uvx mcp-server-time` |

For anything not listed: `mcpx add <name> --command <bin> --args '<a b c>' --env K=V`.

## Environment variables and exit codes

- mcpx itself relies only on the **user home directory** to resolve client config paths, plus `APPDATA` on Windows to locate the Claude Desktop config.
- Variables passed via `--env` belong to the target server: they are written into the client config's `env` field and injected when the handshake launches the server process.
- Colored output switches itself off in non-terminal environments (pipes / CI).
- Exit codes: `0` on success; any error is printed to stderr and exits non-zero.
