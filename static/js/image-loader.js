/**
 * LQIP (Low-Quality Image Placeholder) Image Loader
 *
 * Handles progressive image loading with smooth blur-to-sharp transitions.
 * Loads LQIP placeholder first, then swaps to full-resolution image when loaded.
 */

(function() {
  'use strict';

  /**
   * Initialize LQIP image loading for all responsive images
   */
  function initLQIPLoader() {
    // Find all images with data-src attribute (full-resolution source)
    const images = document.querySelectorAll('img[data-src]');

    if (images.length === 0) {
      return;
    }

    // Use IntersectionObserver for lazy loading
    const imageObserver = new IntersectionObserver((entries, observer) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const img = entry.target;
          loadImage(img);
          observer.unobserve(img);
        }
      });
    }, {
      rootMargin: '50px 0px', // Start loading slightly before image enters viewport
      threshold: 0.01
    });

    // Observe all images
    images.forEach(img => {
      // Add loading class for initial blur effect
      img.classList.add('lqip-loading');

      // If image is above the fold, load immediately
      const rect = img.getBoundingClientRect();
      if (rect.top < window.innerHeight) {
        loadImage(img);
      } else {
        imageObserver.observe(img);
      }
    });
  }

  /**
   * Load full-resolution image and handle transition
   */
  function loadImage(img) {
    const fullSrc = img.dataset.src;

    if (!fullSrc) {
      return;
    }

    // Create a new image to preload full-resolution version
    const tempImg = new Image();

    tempImg.onload = function() {
      // Swap to full-resolution image
      img.src = fullSrc;

      // Remove blur and apply loaded class
      img.classList.remove('lqip-loading');
      img.classList.add('lqip-loaded');

      // Clean up data attribute
      delete img.dataset.src;
    };

    tempImg.onerror = function() {
      console.error('Failed to load image:', fullSrc);
      // Still remove loading class even on error
      img.classList.remove('lqip-loading');
    };

    // Start loading
    tempImg.src = fullSrc;
  }

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initLQIPLoader);
  } else {
    initLQIPLoader();
  }
})();
