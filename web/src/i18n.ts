export type Lang = 'zh' | 'en';

export interface Strings {
  htmlLang: string;
  docTitle: string;
  nav: {
    how: string;
    features: string;
    clients: string;
    github: string;
    langLabel: string;
    menuOpen: string;
    menuClose: string;
  };
  hero: {
    eyebrow: string;
    h1a: string;
    h1b: string;
    sub1: string;
    sub2: string;
    installLabel: string;
    installCmd: string;
    copy: string;
    copied: string;
    ctaPrimary: string;
    ctaGithub: string;
    nodeServer: string;
    nodeMcpx: string;
    nodeVerify: string;
  };
  terminal: {
    title: string;
    caption: string;
  };
  clients: {
    caption: string;
    note: string;
    items: { name: string; detail: string }[];
  };
  features: {
    eyebrow: string;
    h2a: string;
    h2b: string;
    cards: { title: string; body: string; chips: string[] }[];
  };
  cta: {
    h2: string;
    sub: string;
    button: string;
    note: string;
  };
  footer: {
    tag: string;
    license: string;
    githubAria: string;
  };
}

export const STRINGS: Record<Lang, Strings> = {
  zh: {
    htmlLang: 'zh-CN',
    docTitle: 'mcpx — 一条命令，把 MCP server 装进所有客户端',
    nav: {
      how: '工作方式',
      features: '为什么',
      clients: '客户端',
      github: 'GitHub',
      langLabel: '切换到英文',
      menuOpen: '打开菜单',
      menuClose: '关闭菜单'
    },
    hero: {
      eyebrow: '本地运行 · 无账号 · 单一二进制',
      h1a: '给 agent 加 MCP server',
      h1b: '一条命令，装进所有客户端',
      sub1: '探测你已装的客户端，写前备份，幂等合并，',
      sub2: '再逐一 stdio 握手验证。全程单进程，不上云。',
      installLabel: '安装',
      installCmd: 'go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest',
      copy: '复制安装命令',
      copied: '已复制',
      ctaPrimary: '开始使用',
      ctaGithub: '在 GitHub 查看',
      nodeServer: 'MCP server',
      nodeMcpx: 'mcpx add',
      nodeVerify: '握手验证'
    },
    terminal: {
      title: '~ mcpx add filesystem',
      caption: '真实输出 —— 探测 → 备份 → 写入 → 握手，每个客户端逐一确认。'
    },
    clients: {
      caption: '开箱支持的客户端',
      note: '只写入探测到的客户端 —— 探测不到就明确跳过，绝不盲写。',
      items: [
        { name: 'Claude Code', detail: '~/.claude.json · JSON' },
        { name: 'Claude Desktop', detail: 'claude_desktop_config.json · JSON' },
        { name: 'Codex', detail: '~/.codex/config.toml · TOML' },
        { name: 'Cursor', detail: '~/.cursor/mcp.json · JSON' }
      ]
    },
    features: {
      eyebrow: '为什么是 mcpx',
      h2a: '不只写进每个客户端，',
      h2b: '还要验证真的连得上',
      cards: [
        {
          title: '一条命令，装进所有客户端',
          body: '一次 mcpx add，自动探测本机已装的 Claude Code、Claude Desktop、Codex、Cursor，按各家 schema 一次性写好 —— 不用再为每个客户端手改一份不同的 JSON。',
          chips: ['探测 4 类客户端', 'JSON key-path / TOML 表']
        },
        {
          title: '备份先行 · 幂等写入',
          body: '落盘前先备份原配置，只合并你要装的那一条 server 条目，绝不碰其它键，原子写入。再跑一次是安全的空操作，mcpx remove 也只删它写过的那条。',
          chips: ['写前自动备份', '只动一条 · 原子写入']
        },
        {
          title: '逐一握手验证',
          body: '写完不算完：mcpx 起子进程走 stdio JSON-RPC，跑一遍 initialize + tools/list，8 秒超时内确认每个客户端真的连得上、拿得到工具列表。',
          chips: ['stdio JSON-RPC', 'initialize + tools/list', '8s 超时']
        },
        {
          title: '本地 · 无账号 · 单一二进制',
          body: '没有常驻服务、没有云、不需要注册账号，也不引入 MCP SDK 依赖。一个跨平台静态 Go 二进制，装进 PATH 即用。',
          chips: ['无常驻服务', '跨平台 Go 二进制', '零 SDK 依赖']
        }
      ]
    },
    cta: {
      h2: '把 5-6 步 × N 个客户端，收敛成一条命令',
      sub: '开源、MIT、无付费墙 —— 本地写配这一核心动作，永不设墙。',
      button: '在 GitHub 上开始',
      note: 'MIT 许可 · 开源核心永久免费'
    },
    footer: {
      tag: '一条命令的跨客户端 MCP 本地安装器',
      license: 'MIT © 2026 SuperMarioYL',
      githubAria: '在 GitHub 上查看 mcpx'
    }
  },
  en: {
    htmlLang: 'en',
    docTitle: 'mcpx — one command installs an MCP server into every client',
    nav: {
      how: 'How it works',
      features: 'Why mcpx',
      clients: 'Clients',
      github: 'GitHub',
      langLabel: 'Switch to Chinese',
      menuOpen: 'Open menu',
      menuClose: 'Close menu'
    },
    hero: {
      eyebrow: 'Local · no account · one binary',
      h1a: 'Give your agent an MCP server',
      h1b: 'installed everywhere, in one command',
      sub1: 'Detects the clients you already have, backs up before writing,',
      sub2: 'merges idempotently, then handshake-verifies each over stdio.',
      installLabel: 'Install',
      installCmd: 'go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest',
      copy: 'Copy install command',
      copied: 'Copied',
      ctaPrimary: 'Get started',
      ctaGithub: 'View on GitHub',
      nodeServer: 'MCP server',
      nodeMcpx: 'mcpx add',
      nodeVerify: 'handshake'
    },
    terminal: {
      title: '~ mcpx add filesystem',
      caption: 'Real output — detect → backup → write → handshake, confirmed per client.'
    },
    clients: {
      caption: 'Supported out of the box',
      note: 'Writes only to detected clients — undetected ones are skipped loudly, never guess-written.',
      items: [
        { name: 'Claude Code', detail: '~/.claude.json · JSON' },
        { name: 'Claude Desktop', detail: 'claude_desktop_config.json · JSON' },
        { name: 'Codex', detail: '~/.codex/config.toml · TOML' },
        { name: 'Cursor', detail: '~/.cursor/mcp.json · JSON' }
      ]
    },
    features: {
      eyebrow: 'Why mcpx',
      h2a: 'Not just written to every client —',
      h2b: 'verified to actually connect',
      cards: [
        {
          title: 'One command, every client',
          body: 'A single mcpx add auto-detects the Claude Code, Claude Desktop, Codex and Cursor installs on your machine and writes each one in its own schema at once — no more hand-editing a different JSON per client.',
          chips: ['detects 4 client kinds', 'JSON key-path / TOML table']
        },
        {
          title: 'Backup first, idempotent write',
          body: 'Backs up the original config before touching disk, merges only the one server entry you asked for, never touches your other keys, and writes atomically. Re-running is a safe no-op; mcpx remove deletes only what it wrote.',
          chips: ['auto-backup before write', 'one entry · atomic write']
        },
        {
          title: 'Handshake-verified, one by one',
          body: "Writing isn't enough: mcpx spawns the server over stdio JSON-RPC, runs initialize + tools/list, and confirms within an 8-second cap that each client really connects and returns its tool list.",
          chips: ['stdio JSON-RPC', 'initialize + tools/list', '8s cap']
        },
        {
          title: 'Local, no account, one binary',
          body: 'No daemon, no cloud, no account to create, and no MCP SDK dependency pulled in. One cross-platform static Go binary: drop it on your PATH and go.',
          chips: ['no daemon', 'cross-platform Go binary', 'zero SDK deps']
        }
      ]
    },
    cta: {
      h2: 'Collapse 5-6 steps × N clients into one command',
      sub: 'Open source, MIT, no paywall — the core act of writing config locally is never gated.',
      button: 'Get started on GitHub',
      note: 'MIT licensed · the open-source core is free forever'
    },
    footer: {
      tag: 'The one-command cross-client MCP installer',
      license: 'MIT © 2026 SuperMarioYL',
      githubAria: 'View mcpx on GitHub'
    }
  }
};

export const GITHUB_URL = 'https://github.com/SuperMarioYL/mcpx';

/* Terminal transcript — kept in English in both locales: it is the CLI's real output. */
export const TERMINAL_LINES: { text: string; kind: 'cmd' | 'plain' | 'dim' | 'ok' }[] = [
  { text: '$ mcpx add filesystem', kind: 'cmd' },
  { text: 'Detected 3 client(s): Claude Code, Claude Desktop, Codex', kind: 'plain' },
  { text: '  · Cursor not detected (no config at ~/.cursor/mcp.json)', kind: 'dim' },
  { text: 'Installing filesystem into 3 client(s)', kind: 'plain' },
  { text: '  Claude Code      backup → write → merged', kind: 'plain' },
  { text: '                   ✓ handshake OK (initialize + tools/list, 12 tools)', kind: 'ok' },
  { text: '  Claude Desktop   backup → write → merged', kind: 'plain' },
  { text: '                   ✓ handshake OK (initialize + tools/list, 12 tools)', kind: 'ok' },
  { text: '  Codex            backup → write → merged', kind: 'plain' },
  { text: '                   ✓ handshake OK (initialize + tools/list, 12 tools)', kind: 'ok' },
  { text: '✓ 3/3 clients ready', kind: 'ok' }
];
