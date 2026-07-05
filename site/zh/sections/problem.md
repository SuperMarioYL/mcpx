MCP server 的供给侧已经爆炸——光 `awesome-mcp-servers` 一个目录就有 90k+ star。找到一个好用的 server 越来越容易，但把它装进客户端这一步，还停留在手工时代。

### N 个客户端 × 5-6 步手工活

查这个客户端的配置文件在哪、查它要求的 schema、手改 JSON 或 TOML、重启客户端、失败了回去再改——整套流程还要**乘以你用的客户端数量**。同一个 `filesystem` server，在 Claude Code（`~/.claude.json`，带 `"type": "stdio"`）、Codex（`~/.codex/config.toml` 的 TOML 表）、Cursor（`~/.cursor/mcp.json`，不带 `type`）里是三种路径、三种形状。而客户端自带的安装命令（如 `claude mcp add`）只写它自己——每家厂商都不会替竞品维护配置写入。

### 塌缩成一条本地命令

`mcpx add <server>` 一条命令：探测你机器上的每个客户端，按各家自己的 schema 写入，逐一握手验证，最后告诉你 `3/3 clients ready`。单二进制、全程本地、无账号、无云；核心开源、MIT 许可、永不收费。

### 安全契约：备份先行 · 幂等 · 写完即验证

写之前先把原配置复制成时间戳备份；合并的最小单位是单个 server 条目，你手工维护的其它配置键纹丝不动，重复执行是稳定 no-op；写完当场启动 server 跑一次 `initialize` + `tools/list` 握手，连不上立刻知道。撤回只需一条 `mcpx remove`。
