// ============================================================
// mcpx 3D hero — the hub->clients network as a luminous object.
// Central glowing core (mcpx) + 4 orbiting client nodes + luminous
// connection lines with particles flowing hub->node. Bloom for sheen.
// Exposes a `progress` setter (0..1) so ScrollTrigger can evolve it.
// Degrades: throws on no-WebGL so caller can show the CSS fallback.
// ============================================================

import * as THREE from 'three';
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js';
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass.js';

const COL = {
  violet: new THREE.Color('#8b7cf6'),
  blue: new THREE.Color('#4d9bff'),
  cyan: new THREE.Color('#34e5d0'),
};
const NODE_COLORS = [COL.violet, COL.blue, COL.cyan, new THREE.Color('#a78bfa')];
const NODE_LABELS = 4;

export function createScene(canvas) {
  // --- renderer (throws if WebGL unsupported -> caller falls back) ---
  const renderer = new THREE.WebGLRenderer({
    canvas,
    antialias: true,
    alpha: true,
    powerPreference: 'high-performance',
  });
  if (!renderer.getContext()) throw new Error('no-webgl');

  const DPR = Math.min(window.devicePixelRatio || 1, 2);
  renderer.setPixelRatio(DPR);
  renderer.setClearColor(0x06070c, 1);

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(46, 1, 0.1, 100);
  camera.position.set(0, 0, 14);

  // group we rotate for parallax + scroll.
  // `rig` offsets the whole object to the right so it sits beside the
  // left-aligned hero copy rather than behind it (kept legible on wide
  // screens; recentred on narrow ones via setLayout()).
  const rig = new THREE.Group();
  const world = new THREE.Group();
  rig.add(world);
  scene.add(rig);

  // -------------------------------------------------------
  // Central core: layered icosahedron wireframe + inner glow
  // -------------------------------------------------------
  const coreGeo = new THREE.IcosahedronGeometry(2.05, 1);
  const coreWire = new THREE.LineSegments(
    new THREE.WireframeGeometry(coreGeo),
    new THREE.LineBasicMaterial({ color: COL.blue, transparent: true, opacity: 0.55 })
  );
  world.add(coreWire);

  const coreInner = new THREE.Mesh(
    new THREE.IcosahedronGeometry(1.5, 1),
    new THREE.MeshBasicMaterial({ color: COL.violet, transparent: true, opacity: 0.14, wireframe: true })
  );
  world.add(coreInner);

  // soft glowing point at the very center (sprite)
  const glowTex = makeGlowTexture();
  const coreGlow = new THREE.Sprite(
    new THREE.SpriteMaterial({ map: glowTex, color: COL.blue, transparent: true, opacity: 0.72, blending: THREE.AdditiveBlending, depthWrite: false })
  );
  coreGlow.scale.set(5, 5, 1);
  world.add(coreGlow);

  // faint particle halo around the core
  const halo = makeHalo(700, 6.5);
  world.add(halo.points);

  // -------------------------------------------------------
  // 4 client nodes on an orbit + connection lines + flow particles
  // -------------------------------------------------------
  const nodeGroup = new THREE.Group();
  world.add(nodeGroup);

  const nodes = [];
  const radius = 6.2;
  for (let i = 0; i < NODE_LABELS; i++) {
    const ang = (i / NODE_LABELS) * Math.PI * 2 + Math.PI / 4;
    const tilt = (i % 2 === 0 ? 1 : -1) * 1.35;
    const pos = new THREE.Vector3(Math.cos(ang) * radius, tilt + Math.sin(ang) * 1.2, Math.sin(ang) * radius);
    const c = NODE_COLORS[i];

    const mesh = new THREE.Mesh(
      new THREE.OctahedronGeometry(0.52, 0),
      new THREE.MeshBasicMaterial({ color: c, wireframe: true, transparent: true, opacity: 0.9 })
    );
    mesh.position.copy(pos);
    nodeGroup.add(mesh);

    const nglow = new THREE.Sprite(
      new THREE.SpriteMaterial({ map: glowTex, color: c, transparent: true, opacity: 0, blending: THREE.AdditiveBlending, depthWrite: false })
    );
    nglow.scale.set(2.2, 2.2, 1);
    nglow.position.copy(pos);
    nodeGroup.add(nglow);

    // connection line hub -> node
    const lineGeo = new THREE.BufferGeometry().setFromPoints([new THREE.Vector3(0, 0, 0), pos]);
    const line = new THREE.Line(
      lineGeo,
      new THREE.LineBasicMaterial({ color: c, transparent: true, opacity: 0.0 })
    );
    nodeGroup.add(line);

    nodes.push({ pos, mesh, nglow, line, color: c, baseAng: ang, appearAt: 0.06 + i * 0.16 });
  }

  // flow particles that stream hub -> node along the lines
  const FLOW_PER = 14;
  const flowCount = NODE_LABELS * FLOW_PER;
  const flowGeo = new THREE.BufferGeometry();
  const flowPos = new Float32Array(flowCount * 3);
  const flowCol = new Float32Array(flowCount * 3);
  const flowT = new Float32Array(flowCount); // 0..1 along its line
  const flowNode = new Int32Array(flowCount);
  for (let i = 0; i < flowCount; i++) {
    const ni = Math.floor(i / FLOW_PER);
    flowNode[i] = ni;
    flowT[i] = (i % FLOW_PER) / FLOW_PER;
    const c = nodes[ni].color;
    flowCol[i * 3] = c.r; flowCol[i * 3 + 1] = c.g; flowCol[i * 3 + 2] = c.b;
  }
  flowGeo.setAttribute('position', new THREE.BufferAttribute(flowPos, 3));
  flowGeo.setAttribute('color', new THREE.BufferAttribute(flowCol, 3));
  const flowMat = new THREE.PointsMaterial({
    size: 0.16, vertexColors: true, transparent: true, opacity: 0.0,
    blending: THREE.AdditiveBlending, depthWrite: false, map: glowTex,
  });
  const flow = new THREE.Points(flowGeo, flowMat);
  nodeGroup.add(flow);

  // -------------------------------------------------------
  // Post-processing: bloom
  // -------------------------------------------------------
  const composer = new EffectComposer(renderer);
  composer.addPass(new RenderPass(scene, camera));
  const bloom = new UnrealBloomPass(new THREE.Vector2(1, 1), 0.85, 0.62, 0.2);
  composer.addPass(bloom);

  // -------------------------------------------------------
  // state
  // -------------------------------------------------------
  let progress = 0;          // scroll 0..1 (drives node reveal + dolly)
  const pointer = { x: 0, y: 0, tx: 0, ty: 0 };
  let running = true;
  let raf = 0;
  const clock = new THREE.Clock();

  function setLayout() {
    const w = window.innerWidth;
    // On wide screens push the object to the right of the text column and
    // scale it up; on narrow screens recentre + shrink so it reads as ambient.
    if (w >= 1280) {
      rig.position.set(4.6, -0.6, 0);
      rig.scale.setScalar(0.98);
    } else if (w >= 1024) {
      rig.position.set(4.0, -0.6, -0.5);
      rig.scale.setScalar(0.9);
    } else if (w >= 700) {
      rig.position.set(2.0, 0.2, -1.8);
      rig.scale.setScalar(0.82);
    } else {
      rig.position.set(0, 2.6, -2.6);
      rig.scale.setScalar(0.68);
    }
  }

  function resize() {
    const w = canvas.clientWidth || window.innerWidth;
    const h = canvas.clientHeight || window.innerHeight;
    renderer.setSize(w, h, false);
    composer.setSize(w, h);
    bloom.setSize(w, h);
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
    setLayout();
  }
  resize();

  function onPointer(e) {
    const nx = (e.clientX / window.innerWidth) * 2 - 1;
    const ny = (e.clientY / window.innerHeight) * 2 - 1;
    pointer.tx = nx;
    pointer.ty = ny;
  }
  window.addEventListener('pointermove', onPointer, { passive: true });
  window.addEventListener('resize', resize);

  // pause when tab hidden
  function onVisibility() {
    if (document.hidden) {
      running = false;
      cancelAnimationFrame(raf);
    } else if (!running) {
      running = true;
      clock.getDelta(); // drop the elapsed gap
      loop();
    }
  }
  document.addEventListener('visibilitychange', onVisibility);

  const tmp = new THREE.Vector3();

  function tick() {
    const dt = Math.min(clock.getDelta(), 0.05);
    const t = clock.elapsedTime;

    // ease pointer
    pointer.x += (pointer.tx - pointer.x) * 0.05;
    pointer.y += (pointer.ty - pointer.y) * 0.05;

    // base auto-rotation + scroll rotation + parallax
    world.rotation.y = t * 0.12 + progress * Math.PI * 0.9 + pointer.x * 0.35;
    world.rotation.x = Math.sin(t * 0.15) * 0.06 - pointer.y * 0.22 + progress * 0.12;

    // core life
    coreWire.rotation.y -= dt * 0.25;
    coreWire.rotation.x += dt * 0.12;
    coreInner.rotation.y += dt * 0.4;
    const pulse = 1 + Math.sin(t * 1.6) * 0.03;
    coreWire.scale.setScalar(pulse);
    coreGlow.material.opacity = 0.52 + Math.sin(t * 1.6) * 0.1;

    halo.points.rotation.y += dt * 0.03;

    // camera dolly on scroll (gentle push-in as advantages reveal)
    camera.position.z = 14 - progress * 3.2;
    camera.position.y = progress * 0.6;
    camera.lookAt(0, 0, 0);

    // node reveal keyed to scroll progress
    for (let i = 0; i < nodes.length; i++) {
      const n = nodes[i];
      const appear = smoothstep(n.appearAt, n.appearAt + 0.14, progress);
      n.mesh.material.opacity = 0.25 + appear * 0.7;
      n.mesh.rotation.x += dt * 0.9;
      n.mesh.rotation.y += dt * 1.2;
      n.mesh.scale.setScalar(0.6 + appear * 0.6 + Math.sin(t * 2 + i) * 0.04);
      n.nglow.material.opacity = appear * (0.55 + Math.sin(t * 2.4 + i) * 0.12);
      n.line.material.opacity = appear * 0.5;
    }

    // flow particles stream hub->node, brighten with reveal
    const speed = 0.28 + progress * 0.22;
    flowMat.opacity = 0.15 + smoothstep(0.05, 0.4, progress) * 0.75;
    for (let i = 0; i < flowCount; i++) {
      let tt = flowT[i] + dt * speed;
      if (tt > 1) tt -= 1;
      flowT[i] = tt;
      const n = nodes[flowNode[i]];
      const appear = smoothstep(n.appearAt, n.appearAt + 0.14, progress);
      // fade particle if node not yet revealed
      const eased = tt;
      tmp.copy(n.pos).multiplyScalar(eased * appear);
      flowPos[i * 3] = tmp.x;
      flowPos[i * 3 + 1] = tmp.y;
      flowPos[i * 3 + 2] = tmp.z;
    }
    flowGeo.attributes.position.needsUpdate = true;

    composer.render();
  }

  function loop() {
    if (!running) return;
    tick();
    raf = requestAnimationFrame(loop);
  }
  loop();

  return {
    setProgress(p) { progress = Math.max(0, Math.min(1, p)); },
    resize,
    dispose() {
      running = false;
      cancelAnimationFrame(raf);
      window.removeEventListener('pointermove', onPointer);
      window.removeEventListener('resize', resize);
      document.removeEventListener('visibilitychange', onVisibility);
      scene.traverse((o) => {
        if (o.geometry) o.geometry.dispose();
        if (o.material) {
          const mats = Array.isArray(o.material) ? o.material : [o.material];
          mats.forEach((m) => {
            if (m.map) m.map.dispose();
            m.dispose();
          });
        }
      });
      glowTex.dispose();
      composer.dispose?.();
      renderer.dispose();
    },
  };
}

// ---------- helpers ----------
function smoothstep(a, b, x) {
  const t = Math.max(0, Math.min(1, (x - a) / (b - a)));
  return t * t * (3 - 2 * t);
}

function makeGlowTexture() {
  const s = 128;
  const c = document.createElement('canvas');
  c.width = c.height = s;
  const ctx = c.getContext('2d');
  const g = ctx.createRadialGradient(s / 2, s / 2, 0, s / 2, s / 2, s / 2);
  g.addColorStop(0, 'rgba(255,255,255,1)');
  g.addColorStop(0.25, 'rgba(255,255,255,0.8)');
  g.addColorStop(0.55, 'rgba(255,255,255,0.25)');
  g.addColorStop(1, 'rgba(255,255,255,0)');
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, s, s);
  const tex = new THREE.CanvasTexture(c);
  tex.needsUpdate = true;
  return tex;
}

function makeHalo(count, r) {
  const geo = new THREE.BufferGeometry();
  const pos = new Float32Array(count * 3);
  const col = new Float32Array(count * 3);
  const palette = [COL.violet, COL.blue, COL.cyan];
  for (let i = 0; i < count; i++) {
    // random point on a fuzzy sphere shell
    const u = Math.random();
    const v = Math.random();
    const theta = 2 * Math.PI * u;
    const phi = Math.acos(2 * v - 1);
    const rr = r * (0.6 + Math.random() * 0.55);
    pos[i * 3] = rr * Math.sin(phi) * Math.cos(theta);
    pos[i * 3 + 1] = rr * Math.cos(phi) * 0.7;
    pos[i * 3 + 2] = rr * Math.sin(phi) * Math.sin(theta);
    const c = palette[i % 3];
    col[i * 3] = c.r; col[i * 3 + 1] = c.g; col[i * 3 + 2] = c.b;
  }
  geo.setAttribute('position', new THREE.BufferAttribute(pos, 3));
  geo.setAttribute('color', new THREE.BufferAttribute(col, 3));
  const mat = new THREE.PointsMaterial({
    size: 0.05, vertexColors: true, transparent: true, opacity: 0.5,
    blending: THREE.AdditiveBlending, depthWrite: false,
  });
  return { points: new THREE.Points(geo, mat) };
}
