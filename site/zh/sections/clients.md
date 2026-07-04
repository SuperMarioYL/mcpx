一个 server 在 mcpx 内部是一份与客户端无关的规格（名称 + 启动命令 + 参数 + 环境变量），由内置的客户端适配器翻译成各家自己的 schema 写进去。当前四个适配器：

| 客户端 | 配置文件 | 格式 | 条目形状 | 探测语义 |
| --- | --- | --- | --- | --- |
| Claude Code | `~/.claude.json` | JSON | `mcpServers.<name>`，含 `"type": "stdio"` | 文件存在即写入 |
| Claude Desktop | macOS：`~/Library/Application Support/Claude/claude_desktop_config.json`<br>Windows：`%APPDATA%\Claude\claude_desktop_config.json`<br>Linux：`~/.config/Claude/claude_desktop_config.json` | JSON | `mcpServers.<name>`（无 `type` 字段） | 文件存在即写入；其它 OS 报告 skipped |
| Codex | `~/.codex/config.toml` | TOML | `[mcp_servers.<name>]`：`command` / `args` / `env` | 文件存在即写入 |
| Cursor | `~/.cursor/mcp.json` | JSON | `mcpServers.<name>`（无 `type` 字段） | fail-soft：无配置文件则跳过 |

### 四步各自的语义

- **探测是保守的**——逐个解析各客户端在本机的配置路径，只有配置文件真实存在才算 detected；文件不存在报 `not detected`，路径在当前 OS 上无法解析报 `skipped` 并给出原因，永远不会成为写入目标。`--clients` 可限定范围（如 `--clients claude-code,codex`）。
- **备份只在将要改动时发生**——原配置完整复制为同目录时间戳文件 `<配置文件>.mcpx.bak.<YYYYMMDD-HHMMSS>`；`remove` 删除前同样先备份；幂等 no-op 与 `--dry-run` 都不产生备份。
- **合并只动一条**——JSON 客户端按 key-path 只设置 `mcpServers.<name>` 这一个键，写入前做语义化比较（键序、空白差异不算变化），重复执行是稳定 no-op；TOML 客户端（Codex）解析整个 `config.toml`、只替换 `[mcp_servers.<name>]` 一张表再序列化写回，其余表的数据全部保留（注释与手工排版不保留）。落盘一律原子写入：先写同目录临时文件再 rename，中途崩溃也不会留下半截配置。
- **握手是第一等公民**——写完就把 server 当作 stdio 子进程拉起来，用裸 JSON-RPC 依次发 `initialize`、`notifications/initialized`、`tools/list`，8 秒超时。`✓ OK` 表示两个请求都往返成功；`! WARN` 表示超时——条目保留、不回滚（server 首跑时 `npx`/`uvx` 拉包可能就超过 8 秒，超时不等于配置错）；`✗ FAIL` 逐客户端如实报告原因。握手直接基于 `os/exec` + 标准库 JSON，不引入任何 MCP SDK。

### 同一条目在不同客户端里的形状

同一条 `mcpx add filesystem`，在 Claude Code（JSON）里是：

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

在 Codex（TOML）里是：

```toml
[mcp_servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "."]
```

（`env` 表仅在通过 `--env` 提供了环境变量时写入；Claude Desktop 与 Cursor 的 JSON 形状与 Claude Code 相同但不含 `type` 字段。）
