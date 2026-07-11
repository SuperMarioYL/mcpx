#!/usr/bin/env node
// ============================================================
// mcpx site · prebuild fonts
// Subsets Noto Sans SC → public/fonts/cjk.woff2 covering exactly
// the non-ASCII glyphs used by this site (src/i18n.ts + index.html
// + public/404.html). No Google Fonts at runtime, ever.
//
// Fail-soft: if pyftsubset or the source TTF is unavailable but a
// previously generated cjk.woff2 is already committed, the build
// proceeds with a warning (CI app-mode has no fonttools).
// ============================================================
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const OUT = path.join(ROOT, 'public', 'fonts', 'cjk.woff2');
const SOURCES = [
  path.join(ROOT, 'src', 'i18n.ts'),
  path.join(ROOT, 'index.html'),
  path.join(ROOT, 'public', '404.html')
];
const NOTO_SRC = process.env.NOTO_SC_TTF || path.join(os.tmpdir(), 'NotoSansSC-VF.ttf');

const log = (...a) => console.log('[mcpx:prebuild-fonts]', ...a);

// 1. collect every non-ASCII character actually used
const chars = new Set();
for (const f of SOURCES) {
  if (!fs.existsSync(f)) continue;
  for (const ch of fs.readFileSync(f, 'utf8')) {
    if (ch.codePointAt(0) > 0x7f) chars.add(ch);
  }
}
// keep ASCII digits/latin out (Space Grotesk covers them); keep CJK + punctuation
const text = [...chars].sort().join('');
if (!text) {
  log('no non-ASCII text found; nothing to subset.');
  process.exit(0);
}
log(`collected ${chars.size} unique non-ASCII glyphs`);

// 2. check tooling
const hasTool = (() => {
  try {
    execFileSync('pyftsubset', ['--help'], { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
})();

if (!hasTool || !fs.existsSync(NOTO_SRC)) {
  if (fs.existsSync(OUT) && fs.statSync(OUT).size > 1000) {
    log(
      `pyftsubset or Noto source unavailable — keeping committed ${path.relative(ROOT, OUT)} ` +
        `(${fs.statSync(OUT).size} bytes). Re-run locally after copy changes.`
    );
    process.exit(0);
  }
  console.error(
    '[mcpx:prebuild-fonts] ERROR: no cjk.woff2 and cannot generate one ' +
      '(need pyftsubset + NOTO_SC_TTF or a cached NotoSansSC-VF.ttf).'
  );
  process.exit(1);
}

// 3. subset
const tmpTxt = path.join(os.tmpdir(), 'mcpx-cjk-text.txt');
fs.writeFileSync(tmpTxt, text, 'utf8');
fs.mkdirSync(path.dirname(OUT), { recursive: true });
execFileSync('pyftsubset', [
  NOTO_SRC,
  `--text-file=${tmpTxt}`,
  '--flavor=woff2',
  `--output-file=${OUT}`,
  '--layout-features=*',
  '--no-hinting',
  '--desubroutinize'
]);
log(`wrote ${path.relative(ROOT, OUT)} (${fs.statSync(OUT).size} bytes)`);
