import './styles.css';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { ScrollToPlugin } from 'gsap/ScrollToPlugin';
import { DICT } from './i18n.js';
import { buildDOM } from './template.js';

gsap.registerPlugin(ScrollTrigger, ScrollToPlugin);

// expose for headless verification / debugging (harmless in prod)
if (typeof window !== 'undefined') {
  window.__mcpx = { gsap, ScrollTrigger };
}

const REDUCED = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

// ---------- mount markup ----------
const app = document.getElementById('app');
app.innerHTML = buildDOM();

// ---------- i18n ----------
let lang = (navigator.language || 'zh').toLowerCase().startsWith('zh') ? 'zh' : 'zh'; // default zh per spec

function applyLang(l) {
  lang = l;
  const d = DICT[l];
  document.documentElement.lang = l === 'zh' ? 'zh-CN' : 'en';
  document.querySelectorAll('[data-i18n]').forEach((el) => {
    const key = el.getAttribute('data-i18n');
    if (d[key] != null) el.textContent = d[key];
  });
  document.querySelectorAll('.lang-toggle button').forEach((b) => {
    b.classList.toggle('active', b.dataset.lang === l);
  });
  ScrollTrigger.refresh();
}
applyLang(lang);

document.querySelectorAll('.lang-toggle button').forEach((b) => {
  b.addEventListener('click', () => applyLang(b.dataset.lang));
});

// ---------- copy command ----------
const copyBtn = document.getElementById('copyBtn');
copyBtn?.addEventListener('click', async () => {
  try {
    await navigator.clipboard.writeText('mcpx add filesystem');
    const orig = copyBtn.innerHTML;
    copyBtn.innerHTML =
      '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>';
    copyBtn.style.color = 'var(--cyan)';
    setTimeout(() => {
      copyBtn.innerHTML = orig;
      copyBtn.style.color = '';
    }, 1400);
  } catch {
    /* clipboard blocked — no-op */
  }
});

// ---------- topbar scrolled state ----------
const topbar = document.getElementById('topbar');
ScrollTrigger.create({
  start: 'top -40',
  end: 99999,
  onUpdate: (self) => topbar.classList.toggle('scrolled', self.scroll() > 40),
});

// ---------- smooth in-page anchors ----------
document.querySelectorAll('a[href^="#"]').forEach((a) => {
  a.addEventListener('click', (e) => {
    const id = a.getAttribute('href');
    if (id === '#' || id === '#top') {
      e.preventDefault();
      gsap.to(window, { duration: REDUCED ? 0 : 0.9, scrollTo: 0, ease: 'power2.inOut' });
      return;
    }
    const target = document.querySelector(id);
    if (target) {
      e.preventDefault();
      gsap.to(window, {
        duration: REDUCED ? 0 : 1.0,
        scrollTo: { y: target, offsetY: 0 },
        ease: 'power2.inOut',
      });
    }
  });
});

// ---------- reveal animations ----------
function setupReveals() {
  if (REDUCED) {
    gsap.set('.reveal', { opacity: 1, y: 0 });
    return;
  }
  // hero: staggered load-in
  const heroReveals = document.querySelectorAll('#hero .reveal');
  gsap.to(heroReveals, {
    opacity: 1,
    y: 0,
    duration: 1,
    ease: 'power3.out',
    stagger: 0.09,
    delay: 0.15,
  });

  // everything else on scroll
  document.querySelectorAll('.reveal:not(#hero .reveal)').forEach((el) => {
    gsap.to(el, {
      opacity: 1,
      y: 0,
      duration: 0.9,
      ease: 'power3.out',
      scrollTrigger: { trigger: el, start: 'top 82%' },
    });
  });
}

// ---------- 3D scene boot with graceful fallback ----------
let scene = null;

function showFallback() {
  const canvas = document.getElementById('bg-canvas');
  if (canvas) canvas.remove();
  if (!document.querySelector('.bg-fallback')) {
    const f = document.createElement('div');
    f.className = 'bg-fallback';
    document.body.prepend(f);
  }
}

async function bootScene() {
  const canvas = document.getElementById('bg-canvas');
  if (REDUCED || !canvas) {
    showFallback();
    return;
  }
  // quick WebGL capability probe
  try {
    const test = document.createElement('canvas');
    const gl = test.getContext('webgl2') || test.getContext('webgl');
    if (!gl) throw new Error('no-webgl');
  } catch {
    showFallback();
    return;
  }

  try {
    const { createScene } = await import('./scene.js');
    scene = createScene(canvas);
  } catch (err) {
    console.warn('[mcpx] 3D scene unavailable, using fallback:', err?.message || err);
    showFallback();
    return;
  }

  // drive scene.progress from overall scroll through the scrolly stretch
  const scrolly = document.querySelector('.scrolly');
  ScrollTrigger.create({
    trigger: scrolly,
    start: 'top bottom',
    end: 'bottom bottom',
    scrub: 0.6,
    onUpdate: (self) => scene.setProgress(self.progress),
  });
}

// ---------- go ----------
setupReveals();
bootScene();

window.addEventListener('beforeunload', () => scene?.dispose());
