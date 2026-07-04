![mcpx four-stage pipeline (light)](/assets/atlas-light.svg)
![mcpx four-stage pipeline (dark)](/assets/atlas-dark.svg)

*A single `mcpx add` runs a four-stage pipeline inside one process: **detect** (only config files that actually exist on disk count — never guess-write) → **backup** (a timestamped sibling file, taken only when content is actually about to change) → **idempotent merge** (touch a single server entry, write atomically) → **handshake** (stdio subprocess + raw JSON-RPC, capped at 8 seconds). One binary, no daemon, no MCP SDK dependency.*
