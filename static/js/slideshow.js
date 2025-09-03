(function () {
  'use strict';

  function prefersReducedMotion() {
    return window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  }

  function initSlideshow(container) {
    // Mark container as JS-enabled to disable CSS fallback
    container.classList.add('js-slideshow');
    const slides = Array.from(container.querySelectorAll('.wedding-slide'));
    if (!slides.length) return;

    const intervalMs = parseInt(container.dataset.interval, 10) || 5000;
    const fadeMs = 400; // keep in sync with CSS transition
    const reduce = prefersReducedMotion();

    // Initialize: first slide visible, others hidden via class
    let index = 0;
    slides.forEach((s, i) => {
      s.classList.remove('fade-in', 'fade-out', 'show', 'wp-visible');
      s.classList.toggle('wp-visible', i === 0);
    });
    let transitioning = false;

    function getOpacityTransitionMs(el) {
      const cs = window.getComputedStyle(el);
      const props = cs.transitionProperty.split(',').map(s => s.trim());
      const durations = cs.transitionDuration.split(',').map(s => s.trim());
      const delays = cs.transitionDelay.split(',').map(s => s.trim());
      // Normalize lengths
      const n = Math.max(props.length, durations.length, delays.length);
      function toMs(val) {
        if (!val) return 0;
        return val.endsWith('ms') ? parseFloat(val) : parseFloat(val) * 1000;
      }
      let total = 0;
      for (let i = 0; i < n; i++) {
        const p = props[i % props.length];
        const d = toMs(durations[i % durations.length]);
        const de = toMs(delays[i % delays.length]);
        if (p === 'opacity' || p === 'all') {
          total = Math.max(total, d + de);
        }
      }
      return total;
    }

    function waitForOpacityTransition(el) {
      return new Promise((resolve) => {
        const maxMs = Math.max(getOpacityTransitionMs(el), fadeMs);
        let done = false;
        const onEnd = (e) => {
          if (e.propertyName === 'opacity') {
            cleanup();
          }
        };
        const cleanup = () => {
          if (done) return;
          done = true;
          el.removeEventListener('transitionend', onEnd);
          clearTimeout(timer);
          resolve();
        };
        const timer = setTimeout(cleanup, maxMs + 50);
        el.addEventListener('transitionend', onEnd);
      });
    }

    async function show(idx) {
      if (transitioning) return; // guard against overlaps
      const current = slides[index];
      const next = slides[idx];
      if (current === next) return;

      transitioning = true;

      // Fade out current fully first
      if (current) {
        if (reduce) {
          current.classList.remove('wp-visible');
        } else {
          // Removing 'show' triggers opacity transition to 0
          const waitOut = waitForOpacityTransition(current);
          current.classList.remove('wp-visible');
          await waitOut;
        }
      }

      // Prepare and fade in next
      // Force reflow before adding classes (ensure prior style has applied)
      // eslint-disable-next-line no-unused-expressions
      next.offsetHeight;
      if (reduce) {
        next.classList.add('wp-visible');
      } else {
        const waitIn = waitForOpacityTransition(next);
        next.classList.add('wp-visible');
        await waitIn;
      }

      index = idx;
      transitioning = false;
    }

    // Cycle using setTimeout to avoid overlap with transitions
    let timer = null;
    function scheduleNext() {
      if (timer) clearTimeout(timer);
      timer = setTimeout(async () => {
        const nextIdx = (index + 1) % slides.length;
        await show(nextIdx);
        scheduleNext();
      }, intervalMs);
    }
    scheduleNext();

    // Restart timer on visibility change (optional resiliency)
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        if (timer) clearTimeout(timer);
      } else {
        scheduleNext();
      }
    });
  }

  function init() {
    const containers = document.querySelectorAll('.wedding-slides');
    containers.forEach(initSlideshow);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
