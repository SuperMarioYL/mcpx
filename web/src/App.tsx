import { useCallback, useEffect, useRef, useState } from 'react';
import Pipeline from './components/Pipeline';
import { GITHUB_URL, STRINGS, TERMINAL_LINES, type Lang } from './i18n';

/* ---------- tiny hooks ---------- */

function useReveal<T extends HTMLElement>() {
  const ref = useRef<T>(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      el.classList.add('in');
      return;
    }
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            el.classList.add('in');
            io.disconnect();
          }
        }
      },
      { threshold: 0.15 }
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);
  return ref;
}

function initialLang(): Lang {
  try {
    const saved = localStorage.getItem('mcpx-lang');
    if (saved === 'zh' || saved === 'en') return saved;
  } catch {
    /* private mode */
  }
  // Default by visitor timezone: UTC+8 (Shanghai/Taipei/HK/…) → zh, else en.
  try {
    if (new Date().getTimezoneOffset() === -480) return 'zh';
  } catch { /* timezone unreadable */ }
  return 'en';
}

/* ---------- shared icons ---------- */

const GithubMark = () => (
  <svg viewBox="0 0 16 16" aria-hidden="true">
    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z" />
  </svg>
);

/* ---------- sections ---------- */

function Nav(props: { lang: Lang; onToggleLang: () => void }) {
  const t = STRINGS[props.lang];
  const [open, setOpen] = useState(false);

  useEffect(() => {
    document.body.style.overflow = open ? 'hidden' : '';
    return () => {
      document.body.style.overflow = '';
    };
  }, [open]);

  const close = () => setOpen(false);

  return (
    <nav className="nav">
      <a className="nav-logo" href="/" aria-label="mcpx">
        <svg width="22" height="22" viewBox="0 0 64 64" aria-hidden="true">
          <rect width="64" height="64" rx="14" fill="#16161f" />
          <line x1="12" y1="32" x2="52" y2="32" stroke="rgba(255,255,255,0.22)" strokeWidth="2" />
          <circle cx="14" cy="32" r="5" fill="#1a1a24" stroke="rgba(255,255,255,0.4)" strokeWidth="2" />
          <circle cx="50" cy="32" r="5" fill="#1a1a24" stroke="rgba(255,255,255,0.4)" strokeWidth="2" />
          <circle cx="32" cy="32" r="9.5" fill="#1e1e2c" stroke="#e6a144" strokeWidth="2.4" />
          <path d="M27.5 27.5 36.5 36.5 M36.5 27.5 27.5 36.5" stroke="#f4c87e" strokeWidth="3" strokeLinecap="round" />
        </svg>
        mcpx
      </a>
      <div className={`nav-menu${open ? ' active' : ''}`}>
        <ul className="nav-links">
          <li>
            <a href="#how" onClick={close}>
              {t.nav.how}
            </a>
          </li>
          <li>
            <a href="#features" onClick={close}>
              {t.nav.features}
            </a>
          </li>
          <li>
            <a href="#clients" onClick={close}>
              {t.nav.clients}
            </a>
          </li>
        </ul>
        <div className="nav-actions">
          <button
            className="btn-lang"
            onClick={() => {
              props.onToggleLang();
              close();
            }}
            aria-label={t.nav.langLabel}
          >
            {props.lang === 'zh' ? 'EN' : '中文'}
          </button>
          <a className="btn-github" href={GITHUB_URL} target="_blank" rel="noopener noreferrer">
            <svg width="15" height="15" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z" />
            </svg>
            {t.nav.github}
          </a>
        </div>
      </div>
      <button
        className={`menu-toggle${open ? ' active' : ''}`}
        onClick={() => setOpen(!open)}
        aria-label={open ? t.nav.menuClose : t.nav.menuOpen}
        aria-expanded={open}
      >
        <span />
        <span />
      </button>
    </nav>
  );
}

function InstallRow(props: { lang: Lang }) {
  const t = STRINGS[props.lang];
  const [copied, setCopied] = useState(false);
  const timer = useRef<number>();

  const copy = useCallback(async () => {
    const text = t.hero.installCmd;
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      ta.remove();
    }
    setCopied(true);
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setCopied(false), 1800);
  }, [t.hero.installCmd]);

  return (
    <div className="install-row">
      <code className="install-cmd">
        <span className="prompt">$</span>
        {t.hero.installCmd}
      </code>
      <button className={`btn-copy${copied ? ' copied' : ''}`} onClick={copy} aria-label={t.hero.copy}>
        {copied ? (
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        ) : (
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <rect x="9" y="9" width="12" height="12" rx="2" />
            <path d="M5 15V5a2 2 0 0 1 2-2h10" />
          </svg>
        )}
        {copied ? t.hero.copied : 'copy'}
      </button>
    </div>
  );
}

function Terminal(props: { lang: Lang }) {
  const t = STRINGS[props.lang];
  const ref = useReveal<HTMLDivElement>();
  return (
    <div className="terminal-wrap" id="how">
      <div className="terminal reveal" ref={ref}>
        <div className="terminal-bar">
          <div className="dots" aria-hidden="true">
            <span />
            <span />
            <span />
          </div>
          <span className="t-title">{t.terminal.title}</span>
        </div>
        <div className="terminal-body">
          <pre aria-label="mcpx add filesystem — example output">
            {TERMINAL_LINES.map((l, i) => (
              <span key={i} className={`t-line ${l.kind}`} style={{ transitionDelay: `${i * 70}ms` }}>
                {l.text}
              </span>
            ))}
          </pre>
        </div>
      </div>
      <p className="terminal-caption-row">
        <span className="terminal-caption">{t.terminal.caption}</span>
      </p>
    </div>
  );
}

const CLIENT_GLYPHS = [
  /* Claude Code — terminal prompt */
  <svg viewBox="0 0 24 24" key="cc" aria-hidden="true">
    <polyline points="4 17 10 11 4 5" />
    <line x1="12" y1="19" x2="20" y2="19" />
  </svg>,
  /* Claude Desktop — app window */
  <svg viewBox="0 0 24 24" key="cd" aria-hidden="true">
    <rect x="2" y="4" width="20" height="16" rx="2" />
    <line x1="2" y1="9" x2="22" y2="9" />
    <circle cx="5.5" cy="6.5" r="0.5" />
  </svg>,
  /* Codex — braces */
  <svg viewBox="0 0 24 24" key="cx" aria-hidden="true">
    <path d="M8 3H7a2 2 0 0 0-2 2v5a2 2 0 0 1-2 2 2 2 0 0 1 2 2v5c0 1.1.9 2 2 2h1" />
    <path d="M16 21h1a2 2 0 0 0 2-2v-5c0-1.1.9-2 2-2a2 2 0 0 1-2-2V5a2 2 0 0 0-2-2h-1" />
  </svg>,
  /* Cursor — pointer */
  <svg viewBox="0 0 24 24" key="cu" aria-hidden="true">
    <path d="M4 4l7.5 16 2.2-6.3L20 11.5z" />
  </svg>
];

function Clients(props: { lang: Lang }) {
  const t = STRINGS[props.lang];
  const ref = useReveal<HTMLDivElement>();
  return (
    <section className="clients reveal" id="clients" ref={ref}>
      <p className="clients-caption">{t.clients.caption}</p>
      <div className="clients-row">
        {t.clients.items.map((c, i) => (
          <div className="client-item" key={c.name}>
            <span className="glyph">{CLIENT_GLYPHS[i]}</span>
            <span className="c-name">{c.name}</span>
            <span className="c-detail">{c.detail}</span>
          </div>
        ))}
      </div>
      <p className="clients-note">{t.clients.note}</p>
    </section>
  );
}

const FEATURE_ICONS = [
  /* one command every client — layers */
  <svg viewBox="0 0 24 24" key="f1" aria-hidden="true">
    <polygon points="12 2 2 7 12 12 22 7 12 2" />
    <polyline points="2 17 12 22 22 17" />
    <polyline points="2 12 12 17 22 12" />
  </svg>,
  /* backup first — copy/archive */
  <svg viewBox="0 0 24 24" key="f2" aria-hidden="true">
    <rect x="9" y="9" width="12" height="12" rx="2" />
    <path d="M5 15V5a2 2 0 0 1 2-2h10" />
  </svg>,
  /* handshake verify — shield check */
  <svg viewBox="0 0 24 24" key="f3" aria-hidden="true">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    <polyline points="9 12 11 14 15 10" />
  </svg>,
  /* single binary — box */
  <svg viewBox="0 0 24 24" key="f4" aria-hidden="true">
    <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
    <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
    <line x1="12" y1="22.08" x2="12" y2="12" />
  </svg>
];

function FeatureCard(props: {
  span: 7 | 5;
  delay: 0 | 1 | 2 | 3;
  icon: JSX.Element;
  title: string;
  body: string;
  chips: string[];
}) {
  const ref = useReveal<HTMLDivElement>();
  return (
    <article
      className={`feature-card span-${props.span} reveal${props.delay ? ` d${props.delay}` : ''}`}
      ref={ref}
    >
      <div className="feature-icon">{props.icon}</div>
      <h3>{props.title}</h3>
      <p>{props.body}</p>
      <div className="chip-row">
        {props.chips.map((c) => (
          <span className="chip" key={c}>
            {c}
          </span>
        ))}
      </div>
    </article>
  );
}

function Features(props: { lang: Lang }) {
  const t = STRINGS[props.lang];
  const headRef = useReveal<HTMLDivElement>();
  const spans: (7 | 5)[] = [7, 5, 5, 7];
  const delays: (0 | 1 | 2 | 3)[] = [0, 1, 0, 1];
  return (
    <section className="section" id="features">
      <div className="section-head reveal" ref={headRef}>
        <span className="eyebrow">{t.features.eyebrow}</span>
        <h2>
          {t.features.h2a}
          <br />
          <strong>{t.features.h2b}</strong>
        </h2>
      </div>
      <div className="feature-grid">
        {t.features.cards.map((card, i) => (
          <FeatureCard
            key={card.title}
            span={spans[i]}
            delay={delays[i]}
            icon={FEATURE_ICONS[i]}
            title={card.title}
            body={card.body}
            chips={card.chips}
          />
        ))}
      </div>
    </section>
  );
}

function Cta(props: { lang: Lang }) {
  const t = STRINGS[props.lang];
  const ref = useReveal<HTMLDivElement>();
  return (
    <section className="cta-card">
      <div className="cta-grid" aria-hidden="true" />
      <div className="cta-inner reveal" ref={ref}>
        <h2>{t.cta.h2}</h2>
        <p className="cta-sub">{t.cta.sub}</p>
        <a className="btn-cta" href={GITHUB_URL} target="_blank" rel="noopener noreferrer">
          {t.cta.button}
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
        </a>
        <p className="cta-note">{t.cta.note}</p>
      </div>
    </section>
  );
}

function Footer(props: { lang: Lang }) {
  const t = STRINGS[props.lang];
  return (
    <footer className="footer">
      <div className="f-left">
        <span className="f-name">mcpx</span>
        <span>— {t.footer.tag}</span>
      </div>
      <div className="f-right">
        <span>{t.footer.license}</span>
        <a href={GITHUB_URL} target="_blank" rel="noopener noreferrer" aria-label={t.footer.githubAria}>
          <GithubMark />
        </a>
      </div>
    </footer>
  );
}

/* ---------- app ---------- */

export default function App() {
  const [lang, setLang] = useState<Lang>(initialLang);
  const t = STRINGS[lang];

  useEffect(() => {
    document.documentElement.lang = t.htmlLang;
    document.title = t.docTitle;
    try {
      localStorage.setItem('mcpx-lang', lang);
    } catch {
      /* private mode */
    }
  }, [lang, t]);

  return (
    <>
      <Nav lang={lang} onToggleLang={() => setLang(lang === 'zh' ? 'en' : 'zh')} />

      <section className="hero-card">
        <div className="hero-grid" aria-hidden="true" />

        <Pipeline
          labelServer={t.hero.nodeServer}
          labelMcpx={t.hero.nodeMcpx}
          labelVerify={t.hero.nodeVerify}
        />

        <div className="hero-content">
          <span className="hero-eyebrow">
            <span className="dot" aria-hidden="true" />
            {t.hero.eyebrow}
          </span>
          <h1 className="hero-heading">
            {t.hero.h1a}
            <strong>{t.hero.h1b}</strong>
          </h1>
          <p className="hero-sub">
            {t.hero.sub1}
            <br />
            {t.hero.sub2}
          </p>
          <InstallRow lang={lang} />
          <div className="hero-ctas">
            <a className="btn-cta" href={GITHUB_URL} target="_blank" rel="noopener noreferrer">
              {t.hero.ctaPrimary}
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </a>
            <a className="btn-ghost" href={GITHUB_URL} target="_blank" rel="noopener noreferrer">
              <GithubMark />
              {t.hero.ctaGithub}
            </a>
          </div>
        </div>

        <Terminal lang={lang} />
      </section>

      <Clients lang={lang} />
      <Features lang={lang} />
      <Cta lang={lang} />
      <Footer lang={lang} />
    </>
  );
}
