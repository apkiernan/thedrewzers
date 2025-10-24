/**
 * Gallery Lightbox Viewer
 * Full-featured lightbox with keyboard navigation, swipe gestures, zoom, and preloading
 */

export class GalleryLightbox {
  constructor(images) {
    this.images = images; // Array of image elements from gallery
    this.currentIndex = 0;
    this.isOpen = false;
    this.startX = 0;
    this.startY = 0;
    this.scale = 1;
    this.translateX = 0;
    this.translateY = 0;
    this.isDragging = false;
    this.lastTouchDistance = 0;

    this.createDOM();
    this.attachEventListeners();
  }

  /**
   * Create lightbox DOM structure
   */
  createDOM() {
    const lightbox = document.createElement("div");
    lightbox.id = "lightbox";
    lightbox.className = "lightbox";
    lightbox.innerHTML = `
      <div class="lightbox-backdrop"></div>

      <div class="lightbox-content">
        <img
          class="lightbox-image"
          alt="Gallery photo"
          draggable="false"
        />
      </div>

      <div class="lightbox-controls">
        <button
          class="lightbox-close"
          aria-label="Close lightbox"
          title="Close (ESC)"
        >
          <svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>

        <button
          class="lightbox-prev"
          aria-label="Previous photo"
          title="Previous (←)"
        >
          <svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
          </svg>
        </button>

        <button
          class="lightbox-next"
          aria-label="Next photo"
          title="Next (→)"
        >
          <svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
          </svg>
        </button>

        <div class="lightbox-counter">
          <span class="lightbox-counter-current">1</span>
          <span class="lightbox-counter-separator">/</span>
          <span class="lightbox-counter-total">${this.images.length}</span>
        </div>

        <button
          class="lightbox-zoom-in"
          aria-label="Zoom in"
          title="Zoom in (+)"
        >
          <svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v6m-3-3h6"/>
          </svg>
        </button>

        <button
          class="lightbox-zoom-out"
          aria-label="Zoom out"
          title="Zoom out (-)"
        >
          <svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM7 10h6"/>
          </svg>
        </button>

        <button
          class="lightbox-zoom-reset"
          aria-label="Reset zoom"
          title="Reset zoom (0)"
        >
          <svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
        </button>
      </div>
    `;

    document.body.appendChild(lightbox);

    // Cache DOM elements
    this.lightbox = lightbox;
    this.backdrop = lightbox.querySelector(".lightbox-backdrop");
    this.content = lightbox.querySelector(".lightbox-content");
    this.image = lightbox.querySelector(".lightbox-image");
    this.closeBtn = lightbox.querySelector(".lightbox-close");
    this.prevBtn = lightbox.querySelector(".lightbox-prev");
    this.nextBtn = lightbox.querySelector(".lightbox-next");
    this.counterCurrent = lightbox.querySelector(".lightbox-counter-current");
    this.zoomInBtn = lightbox.querySelector(".lightbox-zoom-in");
    this.zoomOutBtn = lightbox.querySelector(".lightbox-zoom-out");
    this.zoomResetBtn = lightbox.querySelector(".lightbox-zoom-reset");
  }

  /**
   * Attach all event listeners
   */
  attachEventListeners() {
    // Close lightbox
    this.closeBtn.addEventListener("click", () => this.close());
    this.backdrop.addEventListener("click", () => this.close());

    // Navigation
    this.prevBtn.addEventListener("click", () => this.prev());
    this.nextBtn.addEventListener("click", () => this.next());

    // Zoom controls
    this.zoomInBtn.addEventListener("click", () => this.zoomIn());
    this.zoomOutBtn.addEventListener("click", () => this.zoomOut());
    this.zoomResetBtn.addEventListener("click", () => this.resetZoom());

    // Keyboard navigation
    this.handleKeydown = this.handleKeydown.bind(this);

    // Touch/swipe gestures
    this.content.addEventListener(
      "touchstart",
      this.handleTouchStart.bind(this),
      {
        passive: false,
      }
    );
    this.content.addEventListener(
      "touchmove",
      this.handleTouchMove.bind(this),
      {
        passive: false,
      }
    );
    this.content.addEventListener("touchend", this.handleTouchEnd.bind(this));

    // Mouse drag for pan
    this.content.addEventListener("mousedown", this.handleMouseDown.bind(this));
    this.content.addEventListener("mousemove", this.handleMouseMove.bind(this));
    this.content.addEventListener("mouseup", this.handleMouseUp.bind(this));
    this.content.addEventListener("mouseleave", this.handleMouseUp.bind(this));

    // Mouse wheel zoom
    this.content.addEventListener("wheel", this.handleWheel.bind(this), {
      passive: false,
    });
  }

  /**
   * Open lightbox at specific index
   */
  open(index) {
    this.currentIndex = index;
    this.isOpen = true;
    this.resetZoom();

    this.lightbox.classList.add("lightbox-open");
    document.body.style.overflow = "hidden";

    this.loadImage(index);
    this.updateCounter();
    this.preloadAdjacent();

    // Add keyboard listener
    document.addEventListener("keydown", this.handleKeydown);

    // Focus management for accessibility
    this.closeBtn.focus();
  }

  /**
   * Close lightbox
   */
  close() {
    this.isOpen = false;
    this.lightbox.classList.remove("lightbox-open");
    document.body.style.overflow = "";

    // Remove keyboard listener
    document.removeEventListener("keydown", this.handleKeydown);

    // Reset zoom state
    this.resetZoom();
  }

  /**
   * Navigate to previous photo
   */
  prev() {
    this.currentIndex =
      (this.currentIndex - 1 + this.images.length) % this.images.length;
    this.loadImage(this.currentIndex);
    this.updateCounter();
    this.preloadAdjacent();
    this.resetZoom();
  }

  /**
   * Navigate to next photo
   */
  next() {
    this.currentIndex = (this.currentIndex + 1) % this.images.length;
    this.loadImage(this.currentIndex);
    this.updateCounter();
    this.preloadAdjacent();
    this.resetZoom();
  }

  /**
   * Load image at specific index
   */
  loadImage(index) {
    const img = this.images[index].querySelector("img");

    // Use full-resolution image if available (data-src), otherwise use current src
    // This ensures we show the high-quality image in the lightbox, not the LQIP
    const imageSrc = img.dataset.src || img.src;

    this.image.src = imageSrc;
    this.image.alt = img.alt;

    // Add loading state
    this.image.classList.add("loading");

    // Remove loading state when image loads
    this.image.onload = () => {
      this.image.classList.remove("loading");
    };
  }

  /**
   * Update photo counter
   */
  updateCounter() {
    this.counterCurrent.textContent = this.currentIndex + 1;
  }

  /**
   * Preload adjacent images for smooth navigation
   */
  preloadAdjacent() {
    const prevIndex =
      (this.currentIndex - 1 + this.images.length) % this.images.length;
    const nextIndex = (this.currentIndex + 1) % this.images.length;

    [prevIndex, nextIndex].forEach((index) => {
      const galleryImg = this.images[index].querySelector("img");
      // Preload full-resolution image if available
      const imageSrc = galleryImg.dataset.src || galleryImg.src;

      const img = new Image();
      img.src = imageSrc;
    });
  }

  /**
   * Handle keyboard navigation
   */
  handleKeydown(e) {
    if (!this.isOpen) return;

    switch (e.key) {
      case "Escape":
        this.close();
        break;
      case "ArrowLeft":
        this.prev();
        break;
      case "ArrowRight":
        this.next();
        break;
      case "+":
      case "=":
        this.zoomIn();
        break;
      case "-":
      case "_":
        this.zoomOut();
        break;
      case "0":
        this.resetZoom();
        break;
    }
  }

  /**
   * Handle touch start for swipe gestures
   */
  handleTouchStart(e) {
    if (e.touches.length === 1) {
      // Single touch - swipe to navigate
      this.startX = e.touches[0].clientX;
      this.startY = e.touches[0].clientY;
    } else if (e.touches.length === 2) {
      // Pinch to zoom
      e.preventDefault();
      const touch1 = e.touches[0];
      const touch2 = e.touches[1];
      this.lastTouchDistance = Math.hypot(
        touch2.clientX - touch1.clientX,
        touch2.clientY - touch1.clientY
      );
    }
  }

  /**
   * Handle touch move for swipe and pinch
   */
  handleTouchMove(e) {
    if (e.touches.length === 2) {
      // Pinch zoom
      e.preventDefault();
      const touch1 = e.touches[0];
      const touch2 = e.touches[1];
      const distance = Math.hypot(
        touch2.clientX - touch1.clientX,
        touch2.clientY - touch1.clientY
      );

      if (this.lastTouchDistance > 0) {
        const delta = distance - this.lastTouchDistance;
        this.scale += delta * 0.01;
        this.scale = Math.max(1, Math.min(this.scale, 4));
        this.updateTransform();
      }

      this.lastTouchDistance = distance;
    }
  }

  /**
   * Handle touch end for swipe detection
   */
  handleTouchEnd(e) {
    const endX = e.changedTouches[0].clientX;
    const endY = e.changedTouches[0].clientY;
    const deltaX = endX - this.startX;
    const deltaY = endY - this.startY;

    // Only trigger swipe if horizontal movement is dominant
    if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > 50) {
      if (deltaX > 0) {
        this.prev();
      } else {
        this.next();
      }
    }

    this.lastTouchDistance = 0;
  }

  /**
   * Handle mouse down for pan
   */
  handleMouseDown(e) {
    if (this.scale > 1) {
      this.isDragging = true;
      this.startX = e.clientX - this.translateX;
      this.startY = e.clientY - this.translateY;
      this.content.style.cursor = "grabbing";
    }
  }

  /**
   * Handle mouse move for pan
   */
  handleMouseMove(e) {
    if (this.isDragging && this.scale > 1) {
      this.translateX = e.clientX - this.startX;
      this.translateY = e.clientY - this.startY;
      this.updateTransform();
    }
  }

  /**
   * Handle mouse up to end pan
   */
  handleMouseUp() {
    this.isDragging = false;
    this.content.style.cursor = "";
  }

  /**
   * Handle mouse wheel for zoom
   */
  handleWheel(e) {
    if (e.ctrlKey || e.metaKey) {
      e.preventDefault();
      const delta = e.deltaY > 0 ? -0.1 : 0.1;
      this.scale += delta;
      this.scale = Math.max(1, Math.min(this.scale, 4));
      this.updateTransform();
    }
  }

  /**
   * Zoom in
   */
  zoomIn() {
    this.scale = Math.min(this.scale + 0.25, 4);
    this.updateTransform();
  }

  /**
   * Zoom out
   */
  zoomOut() {
    this.scale = Math.max(this.scale - 0.25, 1);
    if (this.scale === 1) {
      this.translateX = 0;
      this.translateY = 0;
    }
    this.updateTransform();
  }

  /**
   * Reset zoom to 100%
   */
  resetZoom() {
    this.scale = 1;
    this.translateX = 0;
    this.translateY = 0;
    this.updateTransform();
  }

  /**
   * Update image transform (zoom/pan)
   */
  updateTransform() {
    this.image.style.transform = `scale(${this.scale}) translate(${
      this.translateX / this.scale
    }px, ${this.translateY / this.scale}px)`;
  }
}

// ES module export - GalleryLightbox is exported at class definition
