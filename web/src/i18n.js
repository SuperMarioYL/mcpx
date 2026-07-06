// Lightweight bilingual dictionary. Both languages are truthful + parallel.
// Copy is grounded strictly in the mcpx README + source packages.

export const DICT = {
  zh: {
    'nav.docs': 'GitHub',
    'hero.eyebrow': '本地 · 无账号 · 单二进制',
    'hero.l1': '给 agent 加 MCP server',
    'hero.l2': '一条命令装进所有客户端',
    'hero.sub': '探测你已装的客户端，写前备份，幂等合并，再逐一 stdio 握手验证——全程单进程，不上云。',
    'hero.cta': '开始使用',
    'hero.github': '在 GitHub 查看',
    'hero.scroll': '下拉看产品优势',

    'adv.1.idx': '优势',
    'adv.1.h': '一条命令，装进所有客户端',
    'adv.1.p': '一次 mcpx add，自动探测本机已装的 Claude Code、Claude Desktop、Codex、Cursor，按各家 schema 一次性写好——不用为每个客户端手改一份不同的 JSON。',
    'adv.1.c1': '探测 4 类客户端',
    'adv.1.c2': 'JSON key-path / TOML 表',

    'adv.2.idx': '优势',
    'adv.2.h': '备份先行 · 幂等写入',
    'adv.2.p': '落盘前先备份原配置，只合并你要装的那一条 server 条目，绝不碰你的其它键；原子写入。再跑一次是安全的空操作。',
    'adv.2.c1': '写前自动备份',
    'adv.2.c2': '只动一条 · 原子写入',

    'adv.3.idx': '优势',
    'adv.3.h': '逐一握手验证',
    'adv.3.p': '写完不算完——mcpx 起子进程走 stdio JSON-RPC，跑一遍 initialize + tools/list，8 秒内确认每个客户端真的连得上、拿得到工具列表。',
    'adv.3.c1': 'stdio JSON-RPC',
    'adv.3.c2': '8s 握手超时',

    'adv.4.idx': '优势',
    'adv.4.h': '本地 · 无账号 · 单二进制',
    'adv.4.p': '没有常驻服务、没有云、不需要注册账号，也不引入 MCP SDK 依赖。一个跨平台静态二进制，装进 PATH 即用。',
    'adv.4.c1': '无常驻服务',
    'adv.4.c2': '跨平台单二进制',

    'proof.h': '装完当场验证，看得见的结果',
    'proof.p': '每个客户端都打印备份路径、写入结果和握手状态。撤回只删自己写的那一条，备份始终保留。',
    'proof.title': 'mcpx add filesystem',

    'clients.lead': 'mcpx 只对确证路径的客户端动手，未探测到的会明确标注跳过——绝不猜测写入。',
    'clients.cc.status': '实测',
    'clients.cd.status': '实测',
    'clients.cx.status': '实测',
    'clients.cu.status': 'fail-soft：无配置则跳过',

    'cta.h': '把 5-6 步 × N 客户端，收敛成一条命令',
    'cta.p': '开源、MIT、无付费墙。本地写配这一核心动作，永不设墙。',
    'cta.btn': '在 GitHub 上开始',

    'footer.tag': '一条命令的跨客户端 MCP 本地安装器',
  },

  en: {
    'nav.docs': 'GitHub',
    'hero.eyebrow': 'Local · No account · Single binary',
    'hero.l1': 'Give your agent an MCP server',
    'hero.l2': 'installed everywhere, in one command',
    'hero.sub': 'Detects the clients you already have, backs up before writing, merges idempotently, then handshake-verifies each over stdio — one process, no cloud.',
    'hero.cta': 'Get started',
    'hero.github': 'View on GitHub',
    'hero.scroll': 'Scroll for the advantages',

    'adv.1.idx': 'Advantage',
    'adv.1.h': 'One command, into every client',
    'adv.1.p': 'A single mcpx add auto-detects the Claude Code, Claude Desktop, Codex and Cursor installs on your machine and writes each one to its own schema at once — no more hand-editing a different JSON per client.',
    'adv.1.c1': 'Detects 4 client types',
    'adv.1.c2': 'JSON key-path / TOML table',

    'adv.2.idx': 'Advantage',
    'adv.2.h': 'Backup first · idempotent write',
    'adv.2.p': 'Backs up the original config before touching disk, merges only the one server entry you asked for, never touches your other keys, and writes atomically. Re-running is a safe no-op.',
    'adv.2.c1': 'Auto-backup before write',
    'adv.2.c2': 'One entry · atomic write',

    'adv.3.idx': 'Advantage',
    'adv.3.h': 'Handshake-verified, one by one',
    'adv.3.p': 'Writing isn’t enough — mcpx spawns the server over stdio JSON-RPC, runs initialize + tools/list, and confirms within 8 seconds that each client really connects and returns its tool list.',
    'adv.3.c1': 'stdio JSON-RPC',
    'adv.3.c2': '8s handshake cap',

    'adv.4.idx': 'Advantage',
    'adv.4.h': 'Local · no account · one binary',
    'adv.4.p': 'No daemon, no cloud, no account to create, and no MCP SDK dependency pulled in. One cross-platform static binary — drop it on your PATH and go.',
    'adv.4.c1': 'No daemon',
    'adv.4.c2': 'Cross-platform binary',

    'proof.h': 'Verified on the spot — results you can see',
    'proof.p': 'Every client prints its backup path, the write result and the handshake status. Removing deletes only the entry mcpx wrote; your backup always stays.',
    'proof.title': 'mcpx add filesystem',

    'clients.lead': 'mcpx only writes to clients whose config path it confirms; anything undetected is explicitly marked as skipped — it never guesses.',
    'clients.cc.status': 'verified',
    'clients.cd.status': 'verified',
    'clients.cx.status': 'verified',
    'clients.cu.status': 'fail-soft: skipped if no config',

    'cta.h': 'Collapse 5-6 steps × N clients into one command',
    'cta.p': 'Open source, MIT, no paywall. The core act — writing config locally — is never gated.',
    'cta.btn': 'Get started on GitHub',

    'footer.tag': 'The one-command cross-client MCP installer',
  },
};

export const GITHUB_URL = 'https://github.com/SuperMarioYL/mcpx';
