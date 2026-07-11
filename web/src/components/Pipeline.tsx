import { useEffect, useRef } from 'react';

/**
 * Neumorphic icon pipeline with a traveling light beam:
 * MCP server (layers) ──beam──▶ mcpx (×) ──beam──▶ handshake (shield-check)
 *
 * Beam = two SVG paths (glow + core) stroked with a userSpaceOnUse linear
 * gradient whose x1/x2 window slides along the path via rAF, in four phases:
 * p1 (800ms) → splash (800ms) → p2 (800ms) → idle (1000ms), looped.
 * With prefers-reduced-motion the loop never starts and a static beam shows.
 */
export default function Pipeline(props: {
  labelServer: string;
  labelMcpx: string;
  labelVerify: string;
}) {
  const pipelineRef = useRef<HTMLDivElement>(null);
  const nodeStackRef = useRef<HTMLDivElement>(null);
  const nodeXRef = useRef<HTMLDivElement>(null);
  const nodeShieldRef = useRef<HTMLDivElement>(null);
  const glowPathRef = useRef<SVGPathElement>(null);
  const corePathRef = useRef<SVGPathElement>(null);
  const gradientRef = useRef<SVGLinearGradientElement>(null);
  const splashRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const pipeline = pipelineRef.current;
    const nodeStack = nodeStackRef.current;
    const nodeX = nodeXRef.current;
    const nodeShield = nodeShieldRef.current;
    const glowPath = glowPathRef.current;
    const corePath = corePathRef.current;
    const gradient = gradientRef.current;
    const splash = splashRef.current;
    if (!pipeline || !nodeStack || !nodeX || !nodeShield || !glowPath || !corePath || !gradient || !splash) {
      return;
    }

    const setPath = () => {
      const pRect = pipeline.getBoundingClientRect();
      const sRect = nodeStack.getBoundingClientRect();
      const xRect = nodeX.getBoundingClientRect();
      const shRect = nodeShield.getBoundingClientRect();
      const startX = sRect.left + sRect.width / 2 - pRect.left;
      const startY = sRect.top + sRect.height / 2 - pRect.top;
      const midX = xRect.left + xRect.width / 2 - pRect.left;
      const midY = xRect.top + xRect.height / 2 - pRect.top;
      const endX = shRect.left + shRect.width / 2 - pRect.left;
      const endY = shRect.top + shRect.height / 2 - pRect.top;
      const d = `M ${startX},${startY} L ${midX},${midY} L ${endX},${endY}`;
      glowPath.setAttribute('d', d);
      corePath.setAttribute('d', d);
    };

    setPath();
    window.addEventListener('resize', setPath);

    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
    const setGradientWindow = (percentage: number) => {
      const center = percentage * 100;
      gradient.setAttribute('x1', `${center - 5}%`);
      gradient.setAttribute('x2', `${center + 5}%`);
      gradient.setAttribute('y1', '0%');
      gradient.setAttribute('y2', '0%');
    };

    if (reduced.matches) {
      // Static, fully visible beam: spread the gradient across the whole run.
      gradient.setAttribute('x1', '-40%');
      gradient.setAttribute('x2', '140%');
      gradient.setAttribute('y1', '0%');
      gradient.setAttribute('y2', '0%');
      return () => window.removeEventListener('resize', setPath);
    }

    type Phase = 'p1' | 'splash' | 'p2' | 'idle';
    let state: Phase = 'p1';
    let lastStateChange = performance.now();
    let raf = 0;

    const showBeam = (on: boolean) => {
      glowPath.style.opacity = on ? '0.6' : '0';
      corePath.style.opacity = on ? '1' : '0';
    };

    const tick = (now: number) => {
      const elapsed = now - lastStateChange;
      if (state === 'p1') {
        const p = Math.min(elapsed / 800, 1);
        setGradientWindow(p * 0.5);
        nodeStack.classList.toggle('active', p < 0.4);
        if (p >= 1) {
          nodeStack.classList.remove('active');
          state = 'splash';
          lastStateChange = now;
          showBeam(false);
          splash.classList.add('animate');
        }
      } else if (state === 'splash') {
        if (elapsed >= 800) {
          state = 'p2';
          lastStateChange = now;
          splash.classList.remove('animate');
          showBeam(true);
        }
      } else if (state === 'p2') {
        const p = Math.min(elapsed / 800, 1);
        setGradientWindow(0.5 + p * 0.5);
        nodeShield.classList.toggle('active', p > 0.6);
        if (p >= 1) {
          nodeShield.classList.remove('active');
          state = 'idle';
          lastStateChange = now;
        }
      } else {
        if (elapsed >= 1000) {
          state = 'p1';
          lastStateChange = now;
        }
      }
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);

    return () => {
      window.removeEventListener('resize', setPath);
      cancelAnimationFrame(raf);
    };
  }, []);

  return (
    <div className="icon-pipeline" ref={pipelineRef} aria-hidden="true">
      <svg className="beam-svg">
        <defs>
          <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="2" result="blur" />
            <feComposite in="blur" in2="SourceGraphic" operator="over" />
          </filter>
          <linearGradient id="beam-gradient" gradientUnits="userSpaceOnUse" ref={gradientRef}>
            <stop offset="0%" stopColor="#b06a24" stopOpacity="0" />
            <stop offset="20%" stopColor="#b06a24" stopOpacity="0.8" />
            <stop offset="50%" stopColor="#ffffff" stopOpacity="1" />
            <stop offset="80%" stopColor="#f4c87e" stopOpacity="0.8" />
            <stop offset="100%" stopColor="#f4c87e" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path
          ref={glowPathRef}
          className="beam-path"
          stroke="url(#beam-gradient)"
          strokeWidth="2"
          fill="none"
          filter="url(#glow)"
          style={{ opacity: 0.6 }}
        />
        <path
          ref={corePathRef}
          className="beam-path"
          stroke="url(#beam-gradient)"
          strokeWidth="0.8"
          fill="none"
        />
      </svg>

      <div className="node-anchor">
        <div className="icon-node node-light-right" ref={nodeStackRef}>
          <svg viewBox="0 0 24 24">
            <polygon points="12 2 2 7 12 12 22 7 12 2" />
            <polyline points="2 17 12 22 22 17" />
            <polyline points="2 12 12 17 22 12" />
          </svg>
        </div>
        <span className="node-label">{props.labelServer}</span>
      </div>

      <div className="pipeline-line" />

      <div className="center-wrap node-anchor">
        <div className="splash" ref={splashRef} />
        <div className="icon-node-center" ref={nodeXRef}>
          {/* mcpx mark: crossing strokes meeting in a hub */}
          <svg viewBox="0 0 40 40">
            <path
              d="M11 11 L29 29 M29 11 L11 29"
              stroke="#ffffff"
              strokeWidth="3.4"
              strokeLinecap="round"
              fill="none"
            />
            <circle cx="20" cy="20" r="5.2" fill="#1e1e2c" stroke="#f4c87e" strokeWidth="2" />
          </svg>
        </div>
        <span className="node-label">{props.labelMcpx}</span>
      </div>

      <div className="pipeline-line right" />

      <div className="node-anchor">
        <div className="icon-node node-light-left" ref={nodeShieldRef}>
          <svg viewBox="0 0 24 24">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            <polyline points="9 12 11 14 15 10" />
          </svg>
        </div>
        <span className="node-label">{props.labelVerify}</span>
      </div>
    </div>
  );
}
