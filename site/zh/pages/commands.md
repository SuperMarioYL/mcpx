# 命令参考

```text
mcpx <command> [flags]
```

| 命令 | 作用 |
| --- | --- |
| `mcpx add <server>` | 把一个 MCP server 安装进每个探测到的客户端（备份 → 幂等写入 → 握手验证） |
| `mcpx list` | 显示探测到的客户端及各自已安装的 MCP server |
| `mcpx remove <server>` | 从每个客户端删除该 server 条目（别名 `rm`，删除前先备份） |
| `mcpx catalog` | 列出可直接按名字安装的内置 server |

## 全局 flag（所有子命令可用）

| Flag | 说明 |
| --- | --- |
| `--clients <id1,id2>` | 限制操作到指定客户端子集，可选 id：`claude-code`、`claude-desktop`、`codex`、`cursor` |
| `--dry-run` | 只显示将要发生的变化，不写任何文件、不备份、不跑握手 |
| `--yes` / `-y` | 跳过写入/删除前的确认提示（脚本与 CI 场景） |
| `--version` / `-v` | 打印版本号 |
| `--help` / `-h` | 帮助 |

## mcpx add

```text
mcpx add <server> [--command <bin>] [--args '<a b c>'] [--env K=V ...]
```

`<server>` 是内置目录里的名字，或你自定义的条目名。行为要点：

- 内置 server（见下方 catalog）直接按名字装：`mcpx add filesystem`；
- 不在目录里的 server 走 from-spec 路径：必须给 `--command`，`--args` 与 `--env` 可选；`--command` 同时会覆盖同名内置条目；
- 写入前打印确认提示 `Write "<server>" into N client(s)? [y/N]`（`--yes` / `--dry-run` 跳过）；
- 每个客户端独立报告：写入结果（`backup → write → merged` / `already up to date (idempotent no-op)`）+ 握手结果（OK / WARN / FAIL）；
- 重复执行同一条 `add` 是稳定 no-op：不改文件、不新增备份，但握手会重新验证。

| Flag | 说明 |
| --- | --- |
| `--command <bin>` | server 的可执行程序（安装目录外的 server 时必填），如 `npx`、`uvx`、绝对路径 |
| `--args '<a b c>'` | server 参数，空格分隔；支持引号包裹含空格的参数（`'a "b c" d'`）；不含空格时也接受逗号分隔（`'-y,pkg,.'`） |
| `--env K=V` | 传给 server 进程的环境变量，可重复多次；会写进客户端配置的 `env` 字段并在握手时生效 |

```bash
mcpx add filesystem
mcpx add mytool --command npx --args '-y my-mcp-server' --env TOKEN=abc
mcpx add filesystem --clients claude-code,codex --dry-run
```

## mcpx list

```text
mcpx list
```

先打印探测摘要（哪些客户端命中、哪些没检测到、哪些被跳过及原因），然后按客户端分组列出：配置文件路径 + 每个已安装 server 的名字与完整启动命令行。没有任何 server 的客户端显示 `(no MCP servers)`。

## mcpx remove

```text
mcpx remove <server>    # 别名: mcpx rm <server>
```

- 从每个探测到的客户端里删除名为 `<server>` 的条目，删除前对每个将被修改的文件先做时间戳备份；
- 只删这一个条目——其它 server、其它配置键原样保留；
- 客户端里本来就没有该条目时显示 `not present (skipped)`，不算错误；
- 同样支持 `--dry-run`（显示 `would remove`）与 `--yes`。

## mcpx catalog

```text
mcpx catalog
```

列出内置 server 目录。内置目录只记录**启动命令**——mcpx 不负责安装底层包，`npx` / `uvx` 会在 server 首次运行时自动拉取：

| 名字 | 说明 | 启动命令 |
| --- | --- | --- |
| `everything` | 覆盖全部 MCP 特性的参考 server（测试用） | `npx -y @modelcontextprotocol/server-everything` |
| `fetch` | 抓取网页并转为 Markdown | `uvx mcp-server-fetch` |
| `filesystem` | 本地文件系统读写（限定根目录） | `npx -y @modelcontextprotocol/server-filesystem .` |
| `git` | Git 仓库操作（log / diff / status / blame） | `uvx mcp-server-git` |
| `memory` | 知识图谱记忆存储 | `npx -y @modelcontextprotocol/server-memory` |
| `sequential-thinking` | 结构化分步推理草稿板 | `npx -y @modelcontextprotocol/server-sequential-thinking` |
| `sqlite` | 查询本地 SQLite 数据库 | `uvx mcp-server-sqlite --db-path ./data.db` |
| `time` | 时间与时区换算工具 | `uvx mcp-server-time` |

目录之外的任何 server：`mcpx add <name> --command <bin> --args '<a b c>' --env K=V`。

## 环境变量与退出码

- mcpx 自身只依赖**用户主目录**来解析各客户端配置路径；Windows 上额外读取 `APPDATA` 定位 Claude Desktop 配置。
- 通过 `--env` 传入的变量属于目标 server：既写进客户端配置的 `env` 字段，也在握手启动 server 进程时注入。
- 输出颜色在非终端环境（管道 / CI）自动关闭。
- 退出码：成功为 `0`；任何错误打印到 stderr 并以非零码退出。
