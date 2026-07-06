import { GITHUB_URL } from './i18n.js';

const ghIcon = `<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 .5C5.37.5 0 5.87 0 12.5c0 5.3 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58 0-.29-.01-1.04-.02-2.05-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.33-1.76-1.33-1.76-1.09-.75.08-.73.08-.73 1.2.09 1.84 1.24 1.84 1.24 1.07 1.83 2.8 1.3 3.49.99.11-.78.42-1.3.76-1.6-2.67-.3-5.47-1.34-5.47-5.95 0-1.31.47-2.39 1.24-3.23-.12-.31-.54-1.53.12-3.18 0 0 1.01-.32 3.3 1.23a11.5 11.5 0 0 1 6.01 0c2.29-1.55 3.3-1.23 3.3-1.23.66 1.65.24 2.87.12 3.18.77.84 1.24 1.92 1.24 3.23 0 4.62-2.81 5.64-5.49 5.94.43.37.81 1.1.81 2.22 0 1.6-.01 2.9-.01 3.29 0 .32.22.7.83.58C20.56 22.29 24 17.8 24 12.5 24 5.87 18.63.5 12 .5Z"/></svg>`;

const copyIcon = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/></svg>`;

const chev = `<svg class="chev" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 9l6 6 6-6"/></svg>`;

const check = `<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" style="vertical-align:-1px"><path d="M20 6L9 17l-5-5"/></svg>`;

function advantage(n, extraClass = '') {
  return `
  <section class="advantage ${extraClass}" data-adv="${n}">
    <div class="wrap">
      <div class="advantage-inner reveal">
        <div class="adv-index"><span data-i18n="adv.${n}.idx"></span> · <b>0${n}</b></div>
        <h2 data-i18n="adv.${n}.h"></h2>
        <p data-i18n="adv.${n}.p"></p>
        <div class="adv-meta">
          <span class="chip"><span class="k">›</span> <span data-i18n="adv.${n}.c1"></span></span>
          <span class="chip"><span class="k">›</span> <span data-i18n="adv.${n}.c2"></span></span>
        </div>
      </div>
    </div>
  </section>`;
}

export function buildDOM() {
  return `
  <canvas id="bg-canvas"></canvas>

  <header class="topbar" id="topbar">
    <a class="wordmark" href="#top" aria-label="mcpx home">mcpx<span class="dot"></span></a>
    <nav class="nav-right">
      <div class="lang-toggle" role="group" aria-label="language">
        <button data-lang="zh" class="active">中</button>
        <button data-lang="en">EN</button>
      </div>
      <a class="icon-link" href="${GITHUB_URL}" target="_blank" rel="noopener" aria-label="GitHub repository">${ghIcon}</a>
    </nav>
  </header>

  <main id="top">
    <!-- HERO -->
    <section class="hero" id="hero">
      <div class="wrap hero-inner">
        <div class="eyebrow reveal" data-i18n="hero.eyebrow"></div>
        <h1>
          <span class="l1 reveal" data-i18n="hero.l1"></span>
          <span class="l2 grad-text reveal" data-i18n="hero.l2"></span>
        </h1>
        <p class="hero-sub reveal" data-i18n="hero.sub"></p>

        <div class="hero-cmd reveal">
          <span><span class="prompt">$</span> mcpx add filesystem</span>
          <button class="cmd-copy" id="copyBtn" aria-label="copy command">${copyIcon}</button>
        </div>

        <div class="hero-actions reveal">
          <a class="btn btn-primary" href="${GITHUB_URL}" target="_blank" rel="noopener">
            <span data-i18n="hero.cta"></span>
          </a>
          <a class="btn btn-ghost" href="${GITHUB_URL}" target="_blank" rel="noopener">
            ${ghIcon}<span data-i18n="hero.github"></span>
          </a>
        </div>
      </div>

      <a class="scroll-cue" href="#adv-1" aria-label="scroll down">
        <span data-i18n="hero.scroll"></span>
        ${chev}
      </a>
    </section>

    <!-- SCROLLYTELLING ADVANTAGES -->
    <div class="scrolly ground">
      <span id="adv-1"></span>
      ${advantage(1)}
      ${advantage(2)}
      ${advantage(3)}
      ${advantage(4)}
    </div>

    <!-- TERMINAL PROOF -->
    <section class="proof ground">
      <div class="wrap">
        <div class="proof-head reveal">
          <h2 data-i18n="proof.h"></h2>
          <p data-i18n="proof.p"></p>
        </div>
        <div class="terminal reveal">
          <div class="terminal-bar">
            <span class="tdot"></span><span class="tdot"></span><span class="tdot"></span>
            <span class="ttitle" data-i18n="proof.title"></span>
          </div>
<pre><span class="dim">$ mcpx add filesystem</span>

Detected <span class="hl">3</span> client(s): Claude Code, Claude Desktop, Codex
  <span class="dim">· Cursor not detected (no config at ~/.cursor/mcp.json)</span>

Installing <span class="kw">filesystem</span> into 3 client(s)
  Claude Code      backup → write → merged
                   <span class="dim">backup: ~/.claude.json.mcpx-bak-20260704-1126</span>
                   <span class="ok">${check} handshake OK (initialize + tools/list, 12 tools)</span>
  Claude Desktop   backup → write → merged
                   <span class="ok">${check} handshake OK (initialize + tools/list, 12 tools)</span>
  Codex            backup → write → merged
                   <span class="ok">${check} handshake OK (initialize + tools/list, 12 tools)</span>

<span class="ok">${check} 3/3 clients ready</span>
</pre>
        </div>
      </div>
    </section>

    <!-- CLIENT STRIP -->
    <section class="clients ground">
      <div class="wrap">
        <p class="lead reveal" data-i18n="clients.lead"></p>
        <div class="client-grid reveal">
          <div class="client-cell">
            <span class="cname">Claude Code</span>
            <span class="cpath">~/.claude.json</span>
            <span class="cstatus">${check} <span data-i18n="clients.cc.status"></span></span>
          </div>
          <div class="client-cell">
            <span class="cname">Claude Desktop</span>
            <span class="cpath">…/Claude/claude_desktop_config.json</span>
            <span class="cstatus">${check} <span data-i18n="clients.cd.status"></span></span>
          </div>
          <div class="client-cell">
            <span class="cname">Codex</span>
            <span class="cpath">~/.codex/config.toml</span>
            <span class="cstatus">${check} <span data-i18n="clients.cx.status"></span></span>
          </div>
          <div class="client-cell">
            <span class="cname">Cursor</span>
            <span class="cpath">~/.cursor/mcp.json</span>
            <span class="cstatus soft">○ <span data-i18n="clients.cu.status"></span></span>
          </div>
        </div>
      </div>
    </section>

    <!-- FINAL CTA -->
    <section class="cta ground">
      <div class="cta-inner reveal">
        <h2 data-i18n="cta.h"></h2>
        <p data-i18n="cta.p"></p>
        <div class="hero-actions">
          <a class="btn btn-primary" href="${GITHUB_URL}" target="_blank" rel="noopener">
            ${ghIcon}<span data-i18n="cta.btn"></span>
          </a>
        </div>
      </div>
    </section>
  </main>

  <footer>
    <span class="wordmark">mcpx<span class="dot"></span></span>
    <span class="fmono" data-i18n="footer.tag"></span>
    <a href="${GITHUB_URL}" target="_blank" rel="noopener" class="fmono">github.com/SuperMarioYL/mcpx →</a>
  </footer>
  `;
}
