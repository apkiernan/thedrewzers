/**
 * 3D Flip Cards for Wedding Details Section
 * Handles card flipping animations with keyboard and mouse support
 */

(function() {
  'use strict';

  // Wait for DOM to be ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  function init() {
    const cards = document.querySelectorAll('.flip-card-container');

    if (!cards.length) {
      return; // Exit if no flip cards found
    }

    cards.forEach(card => {
      setupFlipCard(card);
    });
  }

  function setupFlipCard(cardContainer) {
    let isFlipped = false;
    const card = cardContainer.querySelector('.flip-card');

    // Click/tap handler
    cardContainer.addEventListener('click', (e) => {
      // Don't flip if clicking on a link
      if (e.target.tagName === 'A' || e.target.closest('a')) {
        return;
      }

      toggleFlip();
    });

    // Keyboard support
    cardContainer.setAttribute('tabindex', '0');
    cardContainer.setAttribute('role', 'button');
    cardContainer.setAttribute('aria-label', 'Flip card to see details');

    cardContainer.addEventListener('keydown', (e) => {
      // Enter or Space to flip
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        toggleFlip();
      }
    });

    // Optional: Mouse tilt effect on desktop
    // Disabled for now to avoid conflicts with flip animation
    // if (window.matchMedia('(hover: hover) and (pointer: fine)').matches) {
    //   setupTiltEffect(cardContainer, card);
    // }

    function toggleFlip() {
      isFlipped = !isFlipped;
      cardContainer.classList.toggle('flipped', isFlipped);

      // Clear any inline transform styles to let CSS take over
      if (card) {
        card.style.transform = '';
      }

      // Update ARIA label
      cardContainer.setAttribute(
        'aria-label',
        isFlipped ? 'Flip card back to front' : 'Flip card to see details'
      );
    }
  }

  function setupTiltEffect(cardContainer, card) {
    let bounds;

    cardContainer.addEventListener('mouseenter', () => {
      // Don't apply tilt if card is flipped
      if (cardContainer.classList.contains('flipped')) return;
      bounds = cardContainer.getBoundingClientRect();
    });

    cardContainer.addEventListener('mousemove', (e) => {
      // Don't apply tilt if card is flipped or no bounds
      if (!bounds || cardContainer.classList.contains('flipped')) return;

      // Calculate mouse position relative to card center
      const mouseX = e.clientX - bounds.left;
      const mouseY = e.clientY - bounds.top;

      const centerX = bounds.width / 2;
      const centerY = bounds.height / 2;

      // Calculate rotation (max 3 degrees for subtle effect)
      const rotateX = ((mouseY - centerY) / centerY) * -3;
      const rotateY = ((mouseX - centerX) / centerX) * 3;

      // Apply tilt using CSS variables instead of inline transform
      // This way it won't conflict with the flip animation
      cardContainer.style.setProperty('--tilt-x', `${rotateX}deg`);
      cardContainer.style.setProperty('--tilt-y', `${rotateY}deg`);
    });

    cardContainer.addEventListener('mouseleave', () => {
      // Reset tilt
      cardContainer.style.removeProperty('--tilt-x');
      cardContainer.style.removeProperty('--tilt-y');
      bounds = null;
    });
  }

  // Respect reduced motion preference
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    document.documentElement.style.setProperty('--flip-duration', '0s');
  }
})();
