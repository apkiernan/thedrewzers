/**
 * Hero Carousel Lightbox Integration
 * Makes carousel images clickable to open in lightbox viewer
 */

import { GalleryLightbox } from "./lightbox.js";

class HeroCarouselLightbox {
  constructor() {
    this.carousel = document.querySelector(".carousel-container");
    if (!this.carousel) {
      return;
    }

    // Get only the original slides (not clones)
    this.slides = Array.from(
      this.carousel.querySelectorAll(".carousel-slide:not(.clone)")
    );

    if (this.slides.length === 0) {
      return;
    }

    this.init();
  }

  init() {
    // Initialize lightbox with carousel slides
    this.lightbox = new GalleryLightbox(this.slides);

    // Make slides clickable
    this.attachClickHandlers();

    // Also handle clicks on cloned slides
    this.attachCloneHandlers();
  }

  /**
   * Attach click handlers to original slides
   */
  attachClickHandlers() {
    this.slides.forEach((slide, index) => {
      // Make the slide image clickable
      const img = slide.querySelector("img");
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
        img.setAttribute(
          "aria-label",
          `View engagement photo ${index + 1} in full size`
        );

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

  /**
   * Attach click handlers to cloned slides
   * Map clones back to their original slide indices
   */
  attachCloneHandlers() {
    const allSlides = this.carousel.querySelectorAll(".carousel-slide");
    const clones = Array.from(allSlides).filter((slide) =>
      slide.classList.contains("clone")
    );

    clones.forEach((clone) => {
      const img = clone.querySelector("img");
      if (img) {
        img.style.cursor = "pointer";

        // Find the corresponding original slide index by matching the image src
        const cloneSrc = img.getAttribute("src");
        const originalIndex = this.slides.findIndex((slide) => {
          const originalImg = slide.querySelector("img");
          return originalImg && originalImg.getAttribute("src") === cloneSrc;
        });

        if (originalIndex !== -1) {
          img.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            this.lightbox.open(originalIndex);
          });

          // Add keyboard support
          img.setAttribute("tabindex", "0");
          img.setAttribute("role", "button");
          img.setAttribute(
            "aria-label",
            `View engagement photo ${originalIndex + 1} in full size`
          );

          img.addEventListener("keydown", (e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              e.stopPropagation();
              this.lightbox.open(originalIndex);
            }
          });
        }
      }
    });
  }
}

// Initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => new HeroCarouselLightbox());
} else {
  new HeroCarouselLightbox();
}
