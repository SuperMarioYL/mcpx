从零到「所有客户端都装上同一个 MCP server 并通过握手验证」，大约两分钟。三种安装方式任选其一，最终都是把一个静态编译的 `mcpx` 单二进制放进 PATH。

### 安装方式一 - go install（需要 Go 1.24+）

```bash
go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest
```

### 安装方式二 - 下载预编译二进制

[GitHub Releases](https://github.com/SuperMarioYL/mcpx/releases) 为每个版本提供 Linux / macOS / Windows × amd64 / arm64 六个平台的压缩包，并附带 `checksums.txt`（SHA-256）供校验。

```bash
# 以 macOS Apple Silicon + v0.1.0 为例；请到 Releases 页选择你的平台
curl -LO https://github.com/SuperMarioYL/mcpx/releases/download/v0.1.0/mcpx_0.1.0_darwin_arm64.tar.gz
tar -xzf mcpx_0.1.0_darwin_arm64.tar.gz
sudo mv mcpx /usr/local/bin/
mcpx --version
```

### 安装方式三 - 源码构建

```bash
git clone https://github.com/SuperMarioYL/mcpx.git
cd mcpx
go build ./cmd/mcpx
./mcpx --version
```

### 跑起来 - mcpx add filesystem

`filesystem` 是内置目录里的本地文件系统 server（`mcpx catalog` 可查看全部 8 个）。在一台装了 Claude Code、Claude Desktop 和 Codex 的机器上：

```bash
mcpx add filesystem
```

```text
Detected 3 client(s): Claude Code, Claude Desktop, Codex
  · Cursor not detected (no config at /Users/you/.cursor/mcp.json)

Installing filesystem (npx -y @modelcontextprotocol/server-filesystem .) into 3 client(s)
Write "filesystem" into 3 client(s)? [y/N] y
  Claude Code      backup → write → merged
                   backup: /Users/you/.claude.json.mcpx.bak.20260704-112634
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Claude Desktop   backup → write → merged
                   ✓ handshake OK (initialize + tools/list, 12 tools)
  Codex            backup → write → merged
                   ✓ handshake OK (initialize + tools/list, 12 tools)

✓ 3/3 clients ready
Restart (or reload) the affected clients to pick up the new server.
```

### 读懂输出

- **探测**——只有配置文件真实存在于磁盘才算 detected；Cursor 没有 `~/.cursor/mcp.json`，所以被明确标注跳过，mcpx 绝不凭空新建配置去猜写。
- **备份 → 写入**——每个客户端先留一份时间戳备份（`<配置文件>.mcpx.bak.<时间戳>`），再只把这一个条目合并进该客户端自己的 schema，其余键原样保留。
- **握手**——以 stdio 子进程启动 server，`initialize` 与 `tools/list` 都往返成功才判 OK，并报告工具数。
- **幂等**——再跑一遍显示 `already up to date (idempotent no-op)`：不改文件、不新增备份，但握手重新验证。
- **预览与撤回**——任何命令加 `--dry-run` 只看不写；`mcpx remove filesystem` 撤回（删除前同样先备份）。
