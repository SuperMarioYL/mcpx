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
      eyebrow: '本地运行 · 不上云 · 不注册账号',
      h1a: '同一个 MCP server，给每个客户端改一遍配置',
      h1b: '一条命令写进全部，装完当场验证连得上',
      sub1: 'mcpx 找到你本机已装的所有 coding agent，把要加的 server 写进每一个——',
      sub2: '写前先备份，只动这一条，装完当场试连，确认每个都真的拿得到工具。',
      installLabel: '安装',
      installCmd: 'go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest',
      copy: '复制安装命令',
      copied: '已复制',
      ctaPrimary: '开始使用',
      ctaGithub: '在 GitHub 查看',
      nodeServer: 'MCP server',
      nodeMcpx: 'mcpx add',
      nodeVerify: '当场验证'
    },
    terminal: {
      title: '~ mcpx add filesystem',
      caption: '真实输出 —— 写进每一个已装的客户端，并当场确认连得上。'
    },
    clients: {
      caption: '开箱支持的客户端',
      note: '只动确证已装的客户端 —— 没装的明确跳过，绝不乱写。',
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
          body: '一条 mcpx add，自动找到你本机已装的 Claude Code、Claude Desktop、Codex、Cursor，把 server 写进每一个 —— 不用再为每个客户端手改一份不同的配置。',
          chips: ['一个 server，写进每个客户端', '不用再逐个手改配置']
        },
        {
          title: '写前备份 · 只动这一条',
          body: '写进任何配置之前，先备份原来的；只加你要装的那一个 server，绝不碰你的其它设置。再跑一次不会重复写入，mcpx remove 也只删它自己加的那条。',
          chips: ['写前先备份', '只动这一条，不碰其它']
        },
        {
          title: '装完当场验证连得上',
          body: '写完不算完 —— mcpx 会真的把每个 server 启动一次，确认它连得上、你的 agent 拿得到它的工具列表。装完当场就知道哪个客户端真的通了，不用重启后才发现连不上。',
          chips: ['真的启动一次', '确认拿得到工具列表', '装完当场出结果']
        },
        {
          title: '本地运行 · 不上云 · 不注册',
          body: '不上云、不注册账号，你的配置也不外发。装上就能用，拔掉即走，不留下常驻进程。',
          chips: ['不上云 · 不注册', '装上即用，拔掉即走', '配置不外发']
        }
      ]
    },
    cta: {
      h2: '把 5-6 步 × N 个客户端，收敛成一条命令',
      sub: '开源、Apache-2.0、无付费墙 —— 本地安装这一核心能力，永远免费。',
      button: '在 GitHub 上开始',
      note: 'Apache-2.0 许可 · 开源核心永久免费'
    },
    footer: {
      tag: '一条命令，把 MCP server 装进你所有的 coding agent',
      license: 'Apache-2.0 © 2026 SuperMarioYL',
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
      eyebrow: 'Local · no cloud · no account',
      h1a: 'The same MCP server, hand-configured into every client',
      h1b: 'one command writes them all — and proves each one connects',
      sub1: 'mcpx finds every coding agent installed on your machine and writes the server into each one —',
      sub2: 'backs up first, touches only this one entry, and starts each server to confirm it really connects and exposes its tools.',
      installLabel: 'Install',
      installCmd: 'go install github.com/SuperMarioYL/mcpx/cmd/mcpx@latest',
      copy: 'Copy install command',
      copied: 'Copied',
      ctaPrimary: 'Get started',
      ctaGithub: 'View on GitHub',
      nodeServer: 'MCP server',
      nodeMcpx: 'mcpx add',
      nodeVerify: 'verified'
    },
    terminal: {
      title: '~ mcpx add filesystem',
      caption: 'Real output — written into every installed client, and confirmed connecting on the spot.'
    },
    clients: {
      caption: 'Supported out of the box',
      note: 'Only touches clients it confirms are installed — undetected ones are skipped loudly, never guess-written.',
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
          title: 'One command, into every client',
          body: 'A single mcpx add finds the Claude Code, Claude Desktop, Codex and Cursor installs on your machine and writes the server into each one — no more hand-editing a different config file per client.',
          chips: ['one server, into every client', 'no per-client hand-editing']
        },
        {
          title: 'Backs up first, touches only this one',
          body: 'Before touching any config, it backs up the original; it adds only the one server you asked for and never touches your other settings. Re-running will not write duplicates, and mcpx remove deletes only what it added.',
          chips: ['backs up first', 'only this one entry']
        },
        {
          title: 'Verified connecting, right after install',
          body: 'Writing is not enough — mcpx actually starts each server and confirms it connects and that your agent can see its tools. Right after install you know which clients really work, instead of finding out it is broken after a restart.',
          chips: ['actually starts it', 'confirms tools are visible', 'answers right after install']
        },
        {
          title: 'Local · no cloud · no account',
          body: 'No cloud, no account, and your config never leaves your machine. Install and run; remove it and nothing keeps running in the background.',
          chips: ['no cloud · no account', 'install and run', 'config stays local']
        }
      ]
    },
    cta: {
      h2: 'Collapse 5-6 steps × N clients into one command',
      sub: 'Open source, Apache-2.0, no paywall — the core local install is free forever.',
      button: 'Get started on GitHub',
      note: 'Apache-2.0 licensed · the open-source core is free forever'
    },
    footer: {
      tag: 'One command, an MCP server into every coding agent you have',
      license: 'Apache-2.0 © 2026 SuperMarioYL',
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
