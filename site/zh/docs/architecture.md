# 工作原理

一次 `mcpx add` 在单个进程里跑完四步流水线：**探测 → 备份 → 幂等合并 → 握手验证**。单二进制、无常驻服务、不依赖 MCP SDK。

<div class="home-hero mx-auto">
  <img class="hero-img hero-img-light img-fluid" src="/assets/atlas-light.svg" alt="mcpx 架构 - mcpx add 经过客户端探测、各客户端适配器、备份加幂等合并、stdio 握手验证四个阶段">
  <img class="hero-img hero-img-dark img-fluid" src="/assets/atlas-dark.svg" alt="mcpx 架构 - mcpx add 经过客户端探测、各客户端适配器、备份加幂等合并、stdio 握手验证四个阶段">
</div>

## 第一步 - 探测（detect）

mcpx 内置一张客户端适配器注册表。探测时逐个解析各客户端在本机的配置路径，**只有配置文件真实存在于磁盘上，才算「detected」**：

- 配置文件不存在 → 报告 `not detected`，绝不新建文件去猜写；
- 当前操作系统上路径无法解析（例如 Claude Desktop 在不支持的 OS 上）→ 报告 `skipped` 并给出原因，永远不会成为写入目标。

`--clients` 可以把探测范围限制到指定客户端（如 `--clients claude-code,codex`）。

## 第二步 - 备份（backup）

只有在文件内容**确实将要改变**时才备份：把原配置完整复制为同目录的时间戳文件

```text
<配置文件>.mcpx.bak.<YYYYMMDD-HHMMSS>
```

例如 `~/.claude.json.mcpx.bak.20260704-112634`。`remove` 在删除条目前同样先备份。幂等 no-op（内容无变化）不产生备份，`--dry-run` 也不产生。

## 第三步 - 幂等合并（merge）

一个 server 在 mcpx 内部是一份与客户端无关的规格（名称 + 启动命令 + 参数 + 环境变量），由各客户端适配器翻译成该客户端自己的 schema 写进去：

- **JSON 客户端**（Claude Code / Claude Desktop / Cursor）：按 key-path 只设置 `mcpServers.<name>` 这一个键，文档其余部分原样保留。写入前做语义化比较（键序、空白差异不算变化），因此重复执行是稳定 no-op。
- **TOML 客户端**（Codex）：解析整个 `config.toml`，只替换 `[mcp_servers.<name>]` 这一张表再序列化写回——其余表（如 `[projects.*]`）的数据全部保留。注意：TOML 路径是「解析后重排」，**注释与手工排版不保留**（值不受影响）。
- **原子写入**：先写同目录临时文件再 rename 覆盖，中途崩溃也不会留下半截配置。

## 第四步 - 握手验证（handshake）

写完配置后，mcpx 直接把这个 server 当作 stdio 子进程拉起来，用裸 JSON-RPC 走一遍最小 MCP 会话：

1. 发送 `initialize`，等待响应；
2. 发送 `notifications/initialized` 通知；
3. 发送 `tools/list`，等待响应并统计工具数。

整个握手有 8 秒超时上限，三种结果：

| 结果 | 含义 | 对配置的影响 |
| --- | --- | --- |
| `✓ OK` | `initialize` 与 `tools/list` 均往返成功 | 条目已就绪 |
| `! WARN` | 握手超时（server 可能启动慢） | **条目保留，不回滚** |
| `✗ FAIL` | 进程启动失败 / 协议错误 / server 提前关闭 | 条目保留，逐客户端报告原因 |

握手实现直接基于 `os/exec` + 标准库 JSON，不引入任何 MCP SDK——mcpx 不会被 SDK 版本绑架。

## 客户端适配器一览

| 客户端 | 配置文件 | 格式 | 条目形状 | 探测语义 |
| --- | --- | --- | --- | --- |
| Claude Code | `~/.claude.json` | JSON | `mcpServers.<name>`，含 `"type": "stdio"` | 文件存在即写入 |
| Claude Desktop | macOS：`~/Library/Application Support/Claude/claude_desktop_config.json`<br>Windows：`%APPDATA%\Claude\claude_desktop_config.json`<br>Linux：`~/.config/Claude/claude_desktop_config.json` | JSON | `mcpServers.<name>`（无 `type` 字段） | 文件存在即写入；其它 OS 报告 skipped |
| Codex | `~/.codex/config.toml` | TOML | `[mcp_servers.<name>]`：`command` / `args` / `env` | 文件存在即写入 |
| Cursor | `~/.cursor/mcp.json` | JSON | `mcpServers.<name>`（无 `type` 字段） | fail-soft：无配置文件则跳过 |

## 写入的条目长什么样

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

## 设计取舍

- **保守探测优先**：宁可漏报也不误写——路径未确证的客户端一律 skip 并明说，这是「写别人的配置文件」应有的姿态。
- **只动自己那一条**：合并的最小单位是单个 server 条目，用户手工维护的其它配置键永远不被触碰。
- **验证是第一等公民**：写完就地证明「真的连得上」，而不是让用户重启后自己猜。
- **fail-soft 而非回滚**：握手超时或失败时保留条目并如实报告——server 启动慢不等于配置错，回滚反而可能销毁一次正确的写入。备份始终在，撤回只需一条 `mcpx remove`。
