/**
 * Gallery Masonry Layout & Animations
 * Handles staggered animations, responsive layout, and scroll effects
 * Uses column-based masonry algorithm to eliminate gaps
 */

import { GalleryLightbox } from "./lightbox.js";

class GalleryMasonry {
  constructor() {
    this.gallery = document.getElementById("gallery-masonry");
    if (!this.gallery) {
      console.log("Gallery masonry element not found - using fallback grid");
      return;
    }

    this.items = Array.from(document.querySelectorAll(".gallery-item"));
    this.observer = null;
    this.columnHeights = [];
    this.columnCount = 4;
    this.gap = 12; // Must match CSS gap
    this.imagesLoaded = 0;

    this.init();
  }

  init() {
    // Set gallery to relative positioning for absolute item placement
    this.gallery.style.position = "relative";

    // Hide items initially to prevent layout shift
    this.items.forEach((item) => {
      item.style.opacity = "0";
      item.style.visibility = "hidden";
    });

    this.calculateColumns();

    // Position all items immediately using declared dimensions (width/height attributes)
    // This prevents layout shifts when LQIP placeholders load
    this.items.forEach((item) => {
      this.positionItem(item);
      this.imagesLoaded++;
    });

    // Update gallery height after initial positioning
    this.updateGalleryHeight();

    // Setup intersection observer after initial positioning
    setTimeout(() => this.setupIntersectionObserver(), 100);

    // Responsive handling
    window.addEventListener(
      "resize",
      this.debounce(() => this.handleResize(), 200)
    );

    // Add keyboard navigation
    this.setupKeyboardNavigation();

    // Initialize lightbox
    this.initializeLightbox();
  }

  /**
   * Initialize lightbox and attach click handlers
   */
  initializeLightbox() {
    // Create lightbox instance using ES module import
    this.lightbox = new GalleryLightbox(this.items);

    // Attach click handlers to gallery items
    this.items.forEach((item, index) => {
      item.addEventListener("click", () => {
        this.lightbox.open(index);
      });

      // Update cursor to indicate clickable
      item.style.cursor = "pointer";
    });
  }

  /**
   * Setup Intersection Observer for staggered fade-in animations
   */
  setupIntersectionObserver() {
    const options = {
      root: null,
      rootMargin: "100px",
      threshold: 0.01,
    };

    this.observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          const item = entry.target;

          // Only animate if item is positioned (visible)
          if (item.style.visibility !== "visible") {
            return;
          }

          const index = parseInt(item.dataset.index || "0");

          // Staggered animation delay
          // Only stagger within visible group (8 items max for smoother effect)
          const delay = (index % 8) * 40; // 40ms between items in viewport

          setTimeout(() => {
            item.style.opacity = "1";
            item.style.transform = "translateY(0)";
          }, delay);

          // Stop observing once animated
          this.observer.unobserve(item);
        }
      });
    }, options);

    // Observe all items with initial transform state
    this.items.forEach((item) => {
      item.style.transform = "translateY(20px)"; // Initial state for animation
      this.observer.observe(item);
    });
  }

  /**
   * Calculate number of columns based on viewport width
   */
  calculateColumns() {
    const width = window.innerWidth;

    if (width < 640) {
      this.columnCount = 1;
    } else if (width < 768) {
      this.columnCount = 2;
    } else if (width < 1024) {
      this.columnCount = 3;
    } else {
      this.columnCount = 4;
    }

    // Initialize column heights array
    this.columnHeights = new Array(this.columnCount).fill(0);
  }

  /**
   * Position items using column-based masonry algorithm
   */
  handleImageLoad(item) {
    const img = item.querySelector("img");

    this.imagesLoaded++;

    // Position item absolutely using masonry algorithm
    this.positionItem(item);

    // Recalculate gallery height after all images load
    if (this.imagesLoaded === this.items.length) {
      this.updateGalleryHeight();
    }
  }

  /**
   * Position a single item in the shortest column
   */
  positionItem(item) {
    const img = item.querySelector("img");
    const index = parseInt(item.dataset.index || "0");

    // Calculate item dimensions
    const containerWidth = this.gallery.offsetWidth;
    const columnWidth =
      (containerWidth - this.gap * (this.columnCount - 1)) / this.columnCount;

    // Use declared width/height attributes if available (more reliable than naturalWidth/Height with LQIP)
    // Otherwise fall back to natural dimensions
    let aspectRatio;
    if (img.getAttribute('width') && img.getAttribute('height')) {
      const declaredWidth = parseInt(img.getAttribute('width'));
      const declaredHeight = parseInt(img.getAttribute('height'));
      aspectRatio = declaredWidth / declaredHeight;
    } else if (img.naturalWidth && img.naturalHeight) {
      aspectRatio = img.naturalWidth / img.naturalHeight;
    } else {
      // Fallback if neither available
      aspectRatio = 1;
    }

    let itemHeight = columnWidth / aspectRatio;

    // Add variety using pseudo-random pattern
    const pattern = (index * 7 + 3) % 11;

    if (pattern === 0) {
      itemHeight *= 1.15;
    } else if (pattern % 3 === 0) {
      itemHeight *= 1.08;
    } else if (pattern % 5 === 0) {
      itemHeight *= 0.9;
    }

    // Find shortest column
    let shortestColumn = 0;
    let shortestHeight = this.columnHeights[0];

    for (let i = 1; i < this.columnCount; i++) {
      if (this.columnHeights[i] < shortestHeight) {
        shortestHeight = this.columnHeights[i];
        shortestColumn = i;
      }
    }

    // Position item
    const left = shortestColumn * (columnWidth + this.gap);
    const top = this.columnHeights[shortestColumn];

    item.style.position = "absolute";
    item.style.left = `${left}px`;
    item.style.top = `${top}px`;
    item.style.width = `${columnWidth}px`;
    item.style.height = `${itemHeight}px`;

    // Make item visible after positioning (ready for intersection observer)
    item.style.visibility = "visible";

    // Update column height
    this.columnHeights[shortestColumn] += itemHeight + this.gap;
  }

  /**
   * Update gallery container height to fit all items
   */
  updateGalleryHeight() {
    const maxHeight = Math.max(...this.columnHeights);
    this.gallery.style.height = `${maxHeight}px`;
  }

  /**
   * Handle window resize - recalculate columns and reposition all items
   */
  handleResize() {
    const previousColumnCount = this.columnCount;
    this.calculateColumns();

    // Only reposition if column count changed
    if (previousColumnCount !== this.columnCount) {
      // Add smooth transitions during resize
      this.items.forEach((item) => {
        item.style.transition =
          "left 0.3s ease, top 0.3s ease, width 0.3s ease, height 0.3s ease";
      });

      this.imagesLoaded = 0;
      this.items.forEach((item) => {
        this.positionItem(item);
        this.imagesLoaded++;
      });
      this.updateGalleryHeight();

      // Remove transitions after resize completes
      setTimeout(() => {
        this.items.forEach((item) => {
          item.style.transition = "";
        });
      }, 300);
    }
  }

  /**
   * Setup keyboard navigation for accessibility
   */
  setupKeyboardNavigation() {
    this.items.forEach((item) => {
      item.addEventListener("keydown", (e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          // Trigger click on the item (for future lightbox integration)
          item.click();
        }
      });
    });
  }

  /**
   * Debounce helper for resize events
   */
  debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }
}

// Check for reduced motion preference
const prefersReducedMotion = window.matchMedia(
  "(prefers-reduced-motion: reduce)"
).matches;

if (prefersReducedMotion) {
  // Disable animations for users who prefer reduced motion
  const style = document.createElement("style");
  style.textContent = `
        .gallery-item-animate {
            opacity: 1 !important;
            transform: none !important;
        }
        .gallery-item,
        .gallery-item img,
        .gallery-overlay {
            transition: none !important;
        }
    `;
  document.head.appendChild(style);
}

// Initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => new GalleryMasonry());
} else {
  new GalleryMasonry();
}
