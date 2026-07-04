# 快速开始

从零到「所有客户端都装上同一个 MCP server 并通过握手验证」，大约两分钟。

## 安装

三种方式任选其一，最终都是把一个静态编译的 `mcpx` 单二进制放进 PATH。

### 方式一 - go install（需要 Go 1.24+）

```bash
go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest
```

### 方式二 - 下载预编译二进制

[GitHub Releases](https://github.com/SuperMarioYL/mcpx/releases) 为每个版本提供 Linux / macOS / Windows × amd64 / arm64 六个平台的压缩包（Windows 为 `.zip`，其余为 `.tar.gz`），文件名遵循 `mcpx_<版本>_<系统>_<架构>` 模板，并附带 `checksums.txt`（SHA-256）供校验。

```bash
# 以 macOS Apple Silicon + v0.1.0 为例；请到 Releases 页选择你的平台
curl -LO https://github.com/SuperMarioYL/mcpx/releases/download/v0.1.0/mcpx_0.1.0_darwin_arm64.tar.gz
tar -xzf mcpx_0.1.0_darwin_arm64.tar.gz
sudo mv mcpx /usr/local/bin/
mcpx --version
```

### 方式三 - 源码构建

```bash
git clone https://github.com/SuperMarioYL/mcpx.git
cd mcpx
go build ./cmd/mcpx
./mcpx --version
```

## 第一次安装：mcpx add filesystem

`filesystem` 是内置目录里的本地文件系统 server（`mcpx catalog` 可查看全部 8 个）。跑一条命令：

```bash
mcpx add filesystem
```

在一台装了 Claude Code、Claude Desktop 和 Codex 的机器上，输出形如：

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at /Users/you/.cursor/mcp.json)

Installing filesystem (npx -y @modelcontextprotocol/server-filesystem .) into 3 client(s)
Write "filesystem" into 3 client(s)? [y/N] y
  Claude Code      backup → write → merged
                   backup: /Users/you/.claude.json.mcpx.bak.20260704-112634
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Claude Desktop   backup → write → merged
                   backup: /Users/you/Library/Application Support/Claude/claude_desktop_config.json.mcpx.bak.20260704-112634
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Codex            backup → write → merged
                   backup: /Users/you/.codex/config.toml.mcpx.bak.20260704-112635
                   ✓ handshake OK (initialize + tools/list, 12 tools)

✓ 3/3 clients ready
Restart (or reload) the affected clients to pick up the new server.
```

逐行解读：

1. **探测**：mcpx 解析每个已知客户端的配置路径，只有配置文件真实存在才算「detected」。Cursor 没有 `~/.cursor/mcp.json`，所以被明确标注跳过——mcpx 绝不凭空新建配置去猜写。
2. **安装计划与确认**：打印将要写入的 server 启动命令与目标客户端数量，然后要求确认（加 `--yes` / `-y` 可跳过；`--dry-run` 下不会提问）。
3. **备份 → 写入 → 合并**：对每个客户端，先把原配置复制成带时间戳的同目录备份（`<配置文件>.mcpx.bak.<时间戳>`），再只把 `filesystem` 这一个条目合并进该客户端自己的 schema，其余键原样保留。
4. **握手验证**：mcpx 以 stdio 子进程方式启动这个 server，走 JSON-RPC 发送 `initialize` 和 `tools/list`，两个请求都往返成功才判 OK，并报告 server 暴露的工具数。
5. **汇总**：`3/3 clients ready`。重启（或重载）客户端即可使用。

## 确认结果：mcpx list

```bash
mcpx list
```

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at /Users/you/.cursor/mcp.json)

Claude Code (/Users/you/.claude.json)
  filesystem           npx -y @modelcontextprotocol/server-filesystem .
Claude Desktop (/Users/you/Library/Application Support/Claude/claude_desktop_config.json)
  filesystem           npx -y @modelcontextprotocol/server-filesystem .
Codex (/Users/you/.codex/config.toml)
  filesystem           npx -y @modelcontextprotocol/server-filesystem .
```

每个客户端一段：配置文件路径 + 它当前安装的全部 MCP server 及启动命令。

## 再跑一遍会怎样（幂等）

再次执行 `mcpx add filesystem`，每个客户端会显示 `already up to date (idempotent no-op)`：文件内容不变、不产生新备份，但握手仍会重新跑一遍——`add` 同时也是一个「重新验证」命令。

## 先预览，再动手

```bash
mcpx add fetch --dry-run
```

`--dry-run` 打印每个客户端「将会发生什么」（`would backup → write → merge` 或 `already up to date`），不写任何文件、不做备份、不跑握手。

## 撤回

```bash
mcpx remove filesystem
```

只删除 `filesystem` 这一个条目（删除前同样先备份），其它 server 与无关配置键不受影响。

## 下一步

- 完整命令与 flag：[命令参考](../commands/)
- 每一步内部发生了什么：[工作原理](../architecture/)
- 安全性、备份、失败语义：[常见问题](../faq/)
