#!/usr/bin/env bash
# What using mcpx looks like — three commands, no account, no cloud.
set -euo pipefail

# 1) Install a built-in server into every detected client (backup → merge → handshake).
mcpx add filesystem

# 2) See which clients now have which MCP servers.
mcpx list

# 3) Install a server that isn't in the catalog, from an explicit spec.
mcpx add my-server \
  --command npx \
  --args '-y @my-scope/my-mcp-server' \
  --env API_TOKEN=secret

# Undo — removes only the entries mcpx wrote, backup preserved.
mcpx remove filesystem

# Preview without touching any file.
mcpx add fetch --dry-run

# Restrict to a subset of clients.
mcpx add git --clients claude-code,codex --yes
