(function () {
  "use strict";

  // Configuration
  const CONFIG = {
    autoplayDuration: 5000, // 5 seconds per slide
    transitionDuration: 700, // 0.7 seconds slide transition
    initialDelay: 3000, // 3 seconds before starting autoplay
    resumeDelay: 2000, // Auto-resume after 2 seconds of no interaction
    totalSlides: 6, // Total number of original slides (hero-02 through hero-07)
    cloneCount: 3, // Number of clones on each side for infinite scroll
  };

  // State
  let currentIndex = CONFIG.cloneCount + 2; // Start at hero-04.jpg (her favorite!)
  let realCurrentSlide = 2; // The actual slide index (0-5, hero-04.jpg is index 2)
  let autoplayTimer = null;
  let resumeTimer = null;
  let isPaused = false;
  let isInitialized = false;
  let prefersReducedMotion = false;
  let isTransitioning = false;

  // DOM Elements
  let container = null;
  let track = null;
  let originalSlides = [];
  let allSlides = []; // Including clones
  let dots = [];
  let prevButton = null;
  let nextButton = null;

  /**
   * Initialize the carousel with infinite scroll
   */
  function init() {
    if (isInitialized) return;

    // Check for reduced motion preference
    const motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    prefersReducedMotion = motionQuery.matches;

    // Get DOM elements
    container = document.querySelector(".carousel-container");
    if (!container) {
      return;
    }

    track = container.querySelector(".carousel-track");
    originalSlides = Array.from(container.querySelectorAll(".carousel-slide"));
    dots = Array.from(container.querySelectorAll(".carousel-dot"));
    prevButton = container.querySelector(".carousel-prev");
    nextButton = container.querySelector(".carousel-next");

    if (originalSlides.length === 0) {
      console.warn("No carousel slides found");
      return;
    }

    // Create clones for infinite scroll
    createClones();

    // Set up event listeners
    setupEventListeners();

    // Initialize first slide
    goToSlide(currentIndex, true);

    // Start autoplay after initial delay (unless reduced motion)
    if (!prefersReducedMotion) {
      setTimeout(() => {
        startAutoplay();
      }, CONFIG.initialDelay);
    }

    isInitialized = true;
  }

  /**
   * Create clones at the beginning and end for infinite scroll
   */
  function createClones() {
    const fragment = document.createDocumentFragment();

    // Clone last few slides and prepend to beginning
    for (
      let i = CONFIG.totalSlides - CONFIG.cloneCount;
      i < CONFIG.totalSlides;
      i++
    ) {
      const clone = originalSlides[i].cloneNode(true);
      clone.classList.add("clone");
      clone.setAttribute("aria-hidden", "true");
      makeFocusableElementsNonFocusable(clone);
      fragment.appendChild(clone);
    }

    // Insert clones at the beginning
    track.insertBefore(fragment, track.firstChild);

    // Clone first few slides and append to end
    const endFragment = document.createDocumentFragment();
    for (let i = 0; i < CONFIG.cloneCount; i++) {
      const clone = originalSlides[i].cloneNode(true);
      clone.classList.add("clone");
      clone.setAttribute("aria-hidden", "true");
      makeFocusableElementsNonFocusable(clone);
      endFragment.appendChild(clone);
    }
    track.appendChild(endFragment);

    // Update allSlides to include clones
    allSlides = Array.from(track.querySelectorAll(".carousel-slide"));
  }

  /**
   * Make all focusable elements within a container non-focusable
   * This prevents aria-hidden elements from containing focusable descendants
   */
  function makeFocusableElementsNonFocusable(container) {
    const focusableSelectors =
      'a, button, input, select, textarea, [tabindex]:not([tabindex="-1"]), img, area';
    const focusableElements = container.querySelectorAll(focusableSelectors);
    focusableElements.forEach((el) => {
      el.setAttribute("tabindex", "-1");
    });
  }

  /**
   * Set up event listeners
   */
  function setupEventListeners() {
    // Navigation buttons
    if (prevButton) {
      prevButton.addEventListener("click", () => {
        if (isTransitioning) return;
        pauseAutoplay();
        previousSlide();
        resumeAutoplay();
      });
    }

    if (nextButton) {
      nextButton.addEventListener("click", () => {
        if (isTransitioning) return;
        pauseAutoplay();
        nextSlide();
        resumeAutoplay();
      });
    }

    // Dot navigation
    dots.forEach((dot, index) => {
      dot.addEventListener("click", () => {
        if (isTransitioning) return;
        pauseAutoplay();
        // Jump to the real slide (accounting for clones at the beginning)
        goToSlide(CONFIG.cloneCount + index);
        resumeAutoplay();
      });
    });

    // Keyboard navigation
    container.addEventListener("keydown", handleKeyPress);

    // Hover pause (desktop)
    container.addEventListener("mouseenter", () => pauseAutoplay(true));
    container.addEventListener("mouseleave", () => resumeAutoplay());

    // Listen for motion preference changes
    const motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    motionQuery.addEventListener("change", (e) => {
      prefersReducedMotion = e.matches;
      if (prefersReducedMotion && !isPaused) {
        stopAutoplay();
      }
    });

    // Resize handler to recalculate positioning
    let resizeTimeout;
    window.addEventListener("resize", () => {
      clearTimeout(resizeTimeout);
      resizeTimeout = setTimeout(() => {
        goToSlide(currentIndex, true);
      }, 150);
    });

    // Listen for transition end to handle clone jumps
    track.addEventListener("transitionend", handleTransitionEnd);
  }

  /**
   * Handle transition end for infinite scroll jump
   */
  function handleTransitionEnd(e) {
    if (e.target !== track) return;

    isTransitioning = false;

    // Check if we're at a clone and need to jump to real slide
    if (currentIndex < CONFIG.cloneCount) {
      // We're in the cloned beginning, jump to the real end
      const realIndex = CONFIG.totalSlides + currentIndex;
      currentIndex = realIndex;
      goToSlide(currentIndex, true);
    } else if (currentIndex >= CONFIG.cloneCount + CONFIG.totalSlides) {
      // We're in the cloned end, jump to the real beginning
      const realIndex = currentIndex - CONFIG.totalSlides;
      currentIndex = realIndex;
      goToSlide(currentIndex, true);
    }
  }

  /**
   * Start autoplay
   */
  function startAutoplay() {
    if (isPaused || prefersReducedMotion) return;

    autoplayTimer = setInterval(() => {
      nextSlide();
    }, CONFIG.autoplayDuration);
  }

  /**
   * Stop autoplay
   */
  function stopAutoplay() {
    if (autoplayTimer) {
      clearInterval(autoplayTimer);
      autoplayTimer = null;
    }
    if (resumeTimer) {
      clearTimeout(resumeTimer);
      resumeTimer = null;
    }
  }

  /**
   * Pause autoplay
   */
  function pauseAutoplay(clearResume = false) {
    if (isPaused) return;

    isPaused = true;
    stopAutoplay();

    if (clearResume && resumeTimer) {
      clearTimeout(resumeTimer);
      resumeTimer = null;
    }
  }

  /**
   * Resume autoplay
   */
  function resumeAutoplay() {
    if (!isPaused || prefersReducedMotion) return;

    // Clear any existing resume timer
    if (resumeTimer) {
      clearTimeout(resumeTimer);
    }

    // Set up auto-resume after delay
    resumeTimer = setTimeout(() => {
      isPaused = false;
      startAutoplay();
    }, CONFIG.resumeDelay);
  }

  /**
   * Go to next slide
   */
  function nextSlide() {
    if (isTransitioning) return;
    goToSlide(currentIndex + 1);
  }

  /**
   * Go to previous slide
   */
  function previousSlide() {
    if (isTransitioning) return;
    goToSlide(currentIndex - 1);
  }

  /**
   * Go to specific slide
   */
  function goToSlide(slideIndex, skipTransition = false) {
    if (slideIndex === currentIndex && !skipTransition) return;

    currentIndex = slideIndex;

    // Calculate the real slide index (0-6) for dot indicators
    let tempIndex = currentIndex - CONFIG.cloneCount;
    if (tempIndex < 0) {
      tempIndex = CONFIG.totalSlides + tempIndex;
    } else if (tempIndex >= CONFIG.totalSlides) {
      tempIndex = tempIndex - CONFIG.totalSlides;
    }
    realCurrentSlide = tempIndex;

    // Remove active class from all slides
    allSlides.forEach((slide) => slide.classList.remove("active"));

    // Add active class to current slide
    if (allSlides[currentIndex]) {
      allSlides[currentIndex].classList.add("active");
    }

    // Calculate the offset to center the current slide
    // Account for container padding and calculate true center

    // Get container's total width
    const containerWidth = container.offsetWidth;

    // Get container padding (80px on each side from CSS)
    const containerStyle = window.getComputedStyle(container);
    const paddingLeft = parseFloat(containerStyle.paddingLeft);
    const paddingRight = parseFloat(containerStyle.paddingRight);

    // The actual content area where slides live
    const contentWidth = containerWidth - paddingLeft - paddingRight;

    // Get the actual rendered width of a slide (50% of content area)
    const slideElement = allSlides[currentIndex];
    const slideWidth = slideElement.offsetWidth;

    // Calculate the natural left edge of this slide within the content area
    const slideNaturalLeft = currentIndex * slideWidth;

    // Center point of the slide within the track
    const slideNaturalCenter = slideNaturalLeft + slideWidth / 2;

    // Center of the content area (where we want the slide centered)
    const contentCenter = contentWidth / 2;

    // How much to translate to center the slide
    const offset = contentCenter - slideNaturalCenter;

    // Apply transform
    if (skipTransition) {
      track.style.transition = "none";
      track.style.transform = `translateX(${offset}px)`;
      // Force reflow
      track.offsetHeight;
      track.style.transition = "";
      isTransitioning = false;
    } else {
      isTransitioning = true;
      track.style.transform = `translateX(${offset}px)`;
    }

    // Update dot indicators (based on real slide index)
    updateDots(realCurrentSlide);

    // Update ARIA
    updateAriaLiveRegion(realCurrentSlide);
  }

  /**
   * Update dot indicators
   */
  function updateDots(slideIndex) {
    dots.forEach((dot, index) => {
      if (index === slideIndex) {
        dot.classList.add("active");
        dot.setAttribute("aria-selected", "true");
      } else {
        dot.classList.remove("active");
        dot.setAttribute("aria-selected", "false");
      }
    });
  }

  /**
   * Handle keyboard navigation
   */
  function handleKeyPress(event) {
    if (isTransitioning) return;

    switch (event.key) {
      case " ":
      case "Enter":
        event.preventDefault();
        if (isPaused) {
          isPaused = false;
          startAutoplay();
        } else {
          pauseAutoplay(true);
        }
        break;

      case "ArrowLeft":
        event.preventDefault();
        pauseAutoplay();
        previousSlide();
        resumeAutoplay();
        break;

      case "ArrowRight":
        event.preventDefault();
        pauseAutoplay();
        nextSlide();
        resumeAutoplay();
        break;

      case "Escape":
        event.preventDefault();
        pauseAutoplay(true);
        break;
    }
  }

  /**
   * Update ARIA live region for screen readers
   */
  function updateAriaLiveRegion(slideIndex) {
    let liveRegion = document.getElementById("carousel-live-region");

    if (!liveRegion) {
      liveRegion = document.createElement("div");
      liveRegion.id = "carousel-live-region";
      liveRegion.className = "sr-only";
      liveRegion.setAttribute("role", "status");
      liveRegion.setAttribute("aria-live", "polite");
      liveRegion.setAttribute("aria-atomic", "true");
      document.body.appendChild(liveRegion);
    }

    liveRegion.textContent = `Photo ${slideIndex + 1} of ${CONFIG.totalSlides}`;
  }

  // Initialize when DOM is ready
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
