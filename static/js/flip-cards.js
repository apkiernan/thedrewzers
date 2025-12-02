/**
 * 3D Flip Cards for Wedding Details Section
 * Handles card flipping animations with keyboard and mouse support
 */

(function () {
  "use strict";

  // Wait for DOM to be ready
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  function init() {
    const cards = document.querySelectorAll(".flip-card-container");

    if (!cards.length) {
      return; // Exit if no flip cards found
    }

    cards.forEach((card) => {
      setupFlipCard(card);
    });
  }

  function setupFlipCard(cardContainer) {
    let isFlipped = false;
    const card = cardContainer.querySelector(".flip-card");

    // Click/tap handler
    cardContainer.addEventListener("click", (e) => {
      // Don't flip if clicking on a link
      if (e.target.tagName === "A" || e.target.closest("a")) {
        return;
      }

      toggleFlip();
    });

    // Keyboard support
    cardContainer.setAttribute("tabindex", "0");
    cardContainer.setAttribute("role", "button");
    cardContainer.setAttribute("aria-label", "Flip card to see details");

    cardContainer.addEventListener("keydown", (e) => {
      // Enter or Space to flip
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        toggleFlip();
      }
    });

    function toggleFlip() {
      isFlipped = !isFlipped;
      cardContainer.classList.toggle("flipped", isFlipped);

      // Clear any inline transform styles to let CSS take over
      if (card) {
        card.style.transform = "";
      }

      // Update ARIA label
      cardContainer.setAttribute(
        "aria-label",
        isFlipped ? "Flip card back to front" : "Flip card to see details"
      );
    }
  }

  // Respect reduced motion preference
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    document.documentElement.style.setProperty("--flip-duration", "0s");
  }
})();
