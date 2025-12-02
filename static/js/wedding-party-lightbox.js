/**
 * Wedding Party Lightbox Integration
 * Makes wedding party member photos clickable to open in lightbox viewer
 */

import { GalleryLightbox } from "./lightbox.js";

class WeddingPartyLightbox {
  constructor() {
    // Get all wedding party member cards that have real photos (not default avatar)
    this.photoCards = Array.from(
      document.querySelectorAll(".wedding-party-photo")
    ).filter((card) => {
      const img = card.querySelector("img");
      return img && !img.src.includes("default-avatar");
    });

    if (this.photoCards.length === 0) {
      return;
    }

    this.init();
  }

  init() {
    // Initialize lightbox with photo cards
    this.lightbox = new GalleryLightbox(this.photoCards);

    // Make photos clickable
    this.attachClickHandlers();
  }

  /**
   * Attach click handlers to photo cards
   */
  attachClickHandlers() {
    this.photoCards.forEach((card, index) => {
      const img = card.querySelector("img");
      if (img) {
        img.style.cursor = "pointer";
        img.addEventListener("click", (e) => {
          e.preventDefault();
          e.stopPropagation();
          this.lightbox.open(index);
        });

        // Add keyboard support for accessibility
        img.setAttribute("tabindex", "0");
        img.setAttribute("role", "button");
        img.setAttribute("aria-label", `View ${img.alt} photo in full size`);

        img.addEventListener("keydown", (e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            e.stopPropagation();
            this.lightbox.open(index);
          }
        });
      }
    });
  }
}

// Initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener(
    "DOMContentLoaded",
    () => new WeddingPartyLightbox()
  );
} else {
  new WeddingPartyLightbox();
}
