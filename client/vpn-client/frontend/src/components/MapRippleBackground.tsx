import { useRef, useEffect } from 'react';
// @ts-ignore — Vite ?raw import, no type definition needed
import worldSvg from '../assets/images/world.svg?raw';

const BASE_FILL    = '#1e1b4b';
const RIPPLE_FILL  = '#818cf8';

export default function MapRippleBackground() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // Pre-cache centroids once — never touch getBBox again on click
    const pathCentroids = new Map<SVGPathElement, { x: number; y: number }>();
    container.querySelectorAll<SVGPathElement>('path').forEach((p) => {
      const b = p.getBBox();
      pathCentroids.set(p, { x: b.x + b.width / 2, y: b.y + b.height / 2 });
    });

    const allPaths = Array.from(pathCentroids.keys());

    const handleClick = (e: MouseEvent) => {
      const target = e.target as SVGPathElement;
      if (target.tagName.toLowerCase() !== 'path') return;

      const origin = pathCentroids.get(target);
      if (!origin) return;

      allPaths.forEach((p) => {
        const c = pathCentroids.get(p)!;
        const dist = Math.hypot(origin.x - c.x, origin.y - c.y);

        // Cancel previous animation so clicks don't stack
        p.getAnimations().forEach((a) => a.cancel());

        p.animate(
          [
            { fill: BASE_FILL },
            { fill: RIPPLE_FILL, offset: 0.4 },
            { fill: BASE_FILL },
          ],
          {
            duration: 800,
            delay: dist * 0.5,
            easing: 'ease-out',
            fill: 'none',
          }
        );
      });
    };

    container.addEventListener('click', handleClick);
    return () => container.removeEventListener('click', handleClick);
  }, []);

  return (
    <div
      ref={containerRef}
      className="world-map absolute inset-0"
      dangerouslySetInnerHTML={{ __html: worldSvg }}
    />
  );
}
