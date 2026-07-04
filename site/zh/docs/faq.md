# 常见问题

## mcpx 会弄坏我现有的配置吗？

不会，这是整个工具的第一设计约束。三重保护：**写之前**先把原文件完整复制成时间戳备份；**写的时候**只合并 `mcpServers.<name>`（或 TOML 的 `[mcp_servers.<name>]`）这一个条目，其余键与其它 server 条目原样保留；**落盘时**采用原子写入（先写同目录临时文件再 rename 覆盖），即使中途崩溃也不会留下半截配置文件。

唯一的已知例外：Codex 的 `config.toml` 走「解析后重新序列化」，所有表和值都会保留，但 TOML 文件里的**注释和手工排版不保留**。

## 备份放在哪里？长什么样？

备份是配置文件的同目录兄弟文件，命名为：

```text
<配置文件>.mcpx.bak.<YYYYMMDD-HHMMSS>
```

例如 `~/.claude.json.mcpx.bak.20260704-112634`。`add` 和 `remove` 在每次真正修改文件前都会各留一份。mcpx 不会自动清理备份——要恢复，把备份文件复制回原名即可。

## 「幂等」具体是什么语义？

同一条 `mcpx add` 跑两遍，第二遍是稳定 no-op：mcpx 对写入前后的配置做语义化比较（键顺序、空白差异不算变化），内容相同就不写文件、不产生新备份，输出显示 `already up to date (idempotent no-op)`。但握手仍会重新跑——所以重复 `add` 也可以当作「重新验证连接」来用。

## 握手失败会怎样？会回滚我的配置吗？

不会自动回滚，这是有意的 fail-soft 设计。三种握手结果：

- `✓ OK`——`initialize` 与 `tools/list` 都在 8 秒内往返成功；
- `! WARN`——握手超时。条目**保留、不回滚**：server 首跑时 `npx`/`uvx` 拉包可能就超过 8 秒，超时不等于配置错误；
- `✗ FAIL`——进程无法启动或协议错误，逐客户端如实报告原因。

无论哪种结果，写入本身已经完成且有备份。想撤掉就 `mcpx remove <server>`，或从备份恢复。

## 支持哪些客户端？没检测到我的客户端怎么办？

当前四个适配器：Claude Code（`~/.claude.json`）、Claude Desktop（macOS / Windows / Linux 各自的 `claude_desktop_config.json` 路径）、Codex（`~/.codex/config.toml`）、Cursor（`~/.cursor/mcp.json`）。

探测是保守的：**配置文件存在于磁盘才算检测到**。文件不存在时 mcpx 会打印 `not detected (no config at ...)` 并跳过——它绝不会替一个可能根本没安装的客户端凭空创建配置。可以用 `--clients` 显式限定操作范围。

## 如何移除一个 server？

```bash
mcpx remove <server>    # 别名: mcpx rm
```

从每个客户端删除该名字的条目，删除前先备份；其它 server 与无关配置键不受影响；本来就没有这个条目的客户端显示 `not present (skipped)`。同样支持 `--dry-run` 预览。

## mcpx 会安装 server 的依赖包吗？

不会。mcpx 写入的是**启动配置**（命令 + 参数 + 环境变量），不下载、不安装任何 server 的二进制或包。内置目录里的 server 用 `npx` / `uvx` 启动，这两个运行器会在 server 首次运行时自行拉取对应的包。

## mcpx 需要账号吗？会联网吗？

不需要账号，mcpx 自身也不发起任何网络请求——探测、备份、写入都是纯本地文件操作，握手是在本机以 stdio 子进程方式与 server 通信。（注意：被测的 server 本身可能联网，例如 `npx` 首次运行时下载包，那是 server 运行器的行为，不是 mcpx 的。）

## 通过 --env 传的密钥安全吗？

`--env K=V` 的值会以**明文**写进各客户端的配置文件（这与你手工配置 MCP server 时完全一致，是各客户端配置格式本身的约定），并在握手时注入 server 进程。请像对待客户端配置文件一样对待这些值：不要把含密钥的配置文件提交进版本库。

## 我可以先看看会发生什么再决定吗？

可以，任何命令加 `--dry-run`：逐客户端显示「将会做什么」（`would backup → write → merge` / `would remove` / `already up to date`），不写文件、不备份、不跑握手，最后明确打印 `dry-run: no files were written.`。
