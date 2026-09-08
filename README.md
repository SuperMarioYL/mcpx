[English](./README.en.md) · [Website](https://mcpx.lei6393.com) · [GitHub](https://github.com/SuperMarioYL/mcpx)

<picture>
  <source media="(max-width: 600px) and (prefers-color-scheme: dark)" srcset="./assets/presentation/hero-mobile-dark.svg">
  <source media="(max-width: 600px)" srcset="./assets/presentation/hero-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="./assets/presentation/hero-dark.svg">
  <img src="./assets/presentation/hero-light.svg" width="960" alt="Hero diagram">
</picture>

# mcpx

**把同一 MCP 服务配置写入多个客户端。**

mcpx 检测支持的客户端配置文件，向选定客户端合并同一个 stdio 服务定义，并可启动服务执行 initialize/tools-list 冒烟检查。

## 为什么需要它

在多种配置格式重复维护命令、参数与环境变量容易产生差异。同一 ServerSpec 为这些配置提供共同输入，适配器负责各文件格式。

- **统一输入定义** — command、args 与 env 交给各适配器。
- **可重复更新** — 等价服务配置不会重复写入。
- **变更前备份** — 修改已有配置前生成同目录备份。

## 架构

<picture>
  <source media="(max-width: 600px) and (prefers-color-scheme: dark)" srcset="./assets/presentation/architecture-mobile-dark.svg">
  <source media="(max-width: 600px)" srcset="./assets/presentation/architecture-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="./assets/presentation/architecture-dark.svg">
  <img src="./assets/presentation/architecture-light.svg" width="960" alt="Architecture diagram">
</picture>

客户端适配器解析路径与结构；配置 writer 在变更前备份、合并指定条目，并将等价重复输入视为无操作。随后 add 启动指定 stdio 命令执行协议冒烟检查，不启动客户端应用本身。

| 组件 | 职责 |
| --- | --- |
| `ServerSpec / catalog` | internal/catalog |
| `Client adapters` | internal/clients |
| `Backup + merge` | internal/config |
| `Stdio handshake` | internal/handshake |

## 安装与快速上手

使用仓库清单声明的运行时版本。以下源码安装步骤可复现随仓示例。

```bash
git clone https://github.com/SuperMarioYL/mcpx.git
cd mcpx
go build ./cmd/mcpx
```

Go 示例写入 JSON 与 TOML fixture，重复合并、读取条目，并仅从自身临时文件移除。

```bash
go run ./examples/presentation
```

## 实际运行示例

<picture>
  <source media="(max-width: 600px) and (prefers-color-scheme: dark)" srcset="./assets/presentation/process-mobile-dark.svg">
  <source media="(max-width: 600px)" srcset="./assets/presentation/process-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="./assets/presentation/process-dark.svg">
  <img src="./assets/presentation/process-light.svg" width="960" alt="Process diagram">
</picture>

Both adapters write one entry with a backup; the second merge is unchanged, and removal reports a change.

```text
claude-code: first_changed=true backup=true repeat_changed=false servers=1
claude-code: remove_changed=true
codex: first_changed=true backup=true repeat_changed=false servers=1
codex: remove_changed=true
```

完整命令与输出保存在 [docs/demo-results.json](./docs/demo-results.json). 输入和复现代码均随仓提供。

![已有终端录制](./assets/demo.gif)

保留已有录制供参考；上方文字示例给出当前可复现的操作。

## 用法

安装后在仓库根目录运行以下命令；处理自己的数据时替换相应路径。

```bash
go run ./cmd/mcpx catalog
go run ./cmd/mcpx add filesystem --clients claude-code,codex --dry-run
# Apply to your detected clients after reviewing the preview:
go run ./cmd/mcpx add filesystem --clients claude-code,codex
go run ./cmd/mcpx list
go run ./cmd/mcpx remove filesystem --dry-run
```

## 配置

--clients 支持 claude-code、claude-desktop、codex 和 cursor；不存在的配置会跳过。--dry-run 不写文件、不握手；--yes 跳过写入询问。自定义服务使用 --command、带引号的 --args 与重复 --env K=V。环境条目会保存到客户端配置，应按部署选择合适的凭据机制。真实修改后重新加载客户端。

## 集成与职责分工

<picture>
  <source media="(max-width: 600px) and (prefers-color-scheme: dark)" srcset="./assets/presentation/integrations-mobile-dark.svg">
  <source media="(max-width: 600px)" srcset="./assets/presentation/integrations-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="./assets/presentation/integrations-dark.svg">
  <img src="./assets/presentation/integrations-light.svg" width="960" alt="Integrations diagram">
</picture>

根据工作流选择输入与输出路径。本文本地示例验证其中明确说明的子流程。

| 路径 | 已实现职责 |
| --- | --- |
| Claude Code / Desktop | mcpServers JSON entries |
| Cursor | Existing mcp.json adapter |
| Codex | mcp_servers TOML tables |
| Stdio JSON-RPC | initialize and tools/list |
| Local backups | Before changed config writes |

## 限制与后续方向

- 示例操作临时目录配置，不验证在线客户端或 MCP 握手。
- 服务握手不证明客户端已加载配置或所有工具可用。警告与失败可能保留已写条目；应检查逐客户端输出，不能只依赖退出码。
- remove 按名称删除条目，不追踪是否最初由 mcpx 创建。真实 add 启动的服务命令可能下载包或执行代码。

更多客户端适配器、doctor 检查与团队共享配置仍是后续方向；未包含托管 Teams 后端。

## 许可与贡献

许可见 [LICENSE](./LICENSE). 反馈问题时请提供最小输入、执行命令和实际输出。
