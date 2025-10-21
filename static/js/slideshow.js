(function () {
  "use strict";

  // Configuration
  const CONFIG = {
    slideDuration: 5000, // 5 seconds per slide
    fadeDuration: 1200, // 1.2 seconds fade transition
    initialDelay: 3000, // 3 seconds before starting slideshow
    preloadCount: 3, // Number of images to preload immediately
    resumeDelay: 2000, // Auto-resume after 2 seconds of no interaction
    totalSlides: 7, // Total number of slides
    totalDots: 7, // Number of progress dots
  };

  // State
  let currentSlide = 0;
  let slideshowTimer = null;
  let resumeTimer = null;
  let isPaused = false;
  let isInitialized = false;
  let prefersReducedMotion = false;

  // DOM Elements
  let container = null;
  let slides = [];
  let dots = [];
  let prevButton = null;
  let nextButton = null;
  let loadedImages = new Set();

  /**
   * Initialize the slideshow
   */
  function init() {
    if (isInitialized) return;

    // Check for reduced motion preference
    const motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    prefersReducedMotion = motionQuery.matches;

    // Get DOM elements
    container = document.querySelector(".slideshow-container");
    if (!container) {
      console.warn("Slideshow container not found");
      return;
    }

    slides = Array.from(container.querySelectorAll(".slideshow-image"));
    dots = Array.from(container.querySelectorAll(".slideshow-dot"));
    prevButton = container.querySelector(".slideshow-prev");
    nextButton = container.querySelector(".slideshow-next");

    if (slides.length === 0) {
      console.warn("No slideshow images found");
      return;
    }

    // Mark first image as loaded (it's already in the HTML)
    loadedImages.add(0);

    // Set up event listeners
    setupEventListeners();

    // Preload priority images
    preloadPriorityImages();

    // Start slideshow after initial delay (unless reduced motion)
    if (!prefersReducedMotion) {
      setTimeout(() => {
        startSlideshow();
      }, CONFIG.initialDelay);
    }

    // Background preload remaining images
    if ("requestIdleCallback" in window) {
      requestIdleCallback(() => preloadRemainingImages());
    } else {
      setTimeout(() => preloadRemainingImages(), 5000);
    }

    isInitialized = true;
  }

  /**
   * Set up event listeners
   */
  function setupEventListeners() {
    // Navigation buttons
    if (prevButton) {
      prevButton.addEventListener("click", () => {
        pauseSlideshow();
        previousSlide();
        resumeSlideshow();
      });
    }

    if (nextButton) {
      nextButton.addEventListener("click", () => {
        pauseSlideshow();
        nextSlide();
        resumeSlideshow();
      });
    }

    // Dot navigation
    dots.forEach((dot, index) => {
      dot.addEventListener("click", () => handleDotClick(index));
    });

    // Keyboard navigation
    container.addEventListener("keydown", handleKeyPress);

    // Hover pause (desktop)
    container.addEventListener("mouseenter", () => pauseSlideshow(true));
    container.addEventListener("mouseleave", () => resumeSlideshow());

    // Listen for motion preference changes
    const motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    motionQuery.addEventListener("change", (e) => {
      prefersReducedMotion = e.matches;
      if (prefersReducedMotion && !isPaused) {
        stopSlideshow();
      }
    });
  }

  /**
   * Preload priority images (next 3)
   */
  function preloadPriorityImages() {
    for (let i = 1; i <= CONFIG.preloadCount && i < slides.length; i++) {
      preloadImage(i);
    }
  }

  /**
   * Preload remaining images in background
   */
  function preloadRemainingImages() {
    for (let i = CONFIG.preloadCount + 1; i < slides.length; i++) {
      if (!loadedImages.has(i)) {
        preloadImage(i);
        // Small delay between each preload to avoid blocking
        if (i % 5 === 0 && "requestIdleCallback" in window) {
          requestIdleCallback(() => {});
        }
      }
    }
  }

  /**
   * Preload a specific image
   */
  function preloadImage(index) {
    if (loadedImages.has(index) || index >= slides.length) return;

    const img = slides[index];
    if (!img) return;

    const src = img.getAttribute("src");
    if (src && !img.complete) {
      const preloader = new Image();
      preloader.src = src;
      preloader.onload = () => {
        loadedImages.add(index);
      };
    } else {
      loadedImages.add(index);
    }
  }

  /**
   * Start the slideshow
   */
  function startSlideshow() {
    if (isPaused || prefersReducedMotion) return;

    slideshowTimer = setInterval(() => {
      nextSlide();
    }, CONFIG.slideDuration);
  }

  /**
   * Stop the slideshow
   */
  function stopSlideshow() {
    if (slideshowTimer) {
      clearInterval(slideshowTimer);
      slideshowTimer = null;
    }
    if (resumeTimer) {
      clearTimeout(resumeTimer);
      resumeTimer = null;
    }
  }

  /**
   * Pause the slideshow
   */
  function pauseSlideshow(clearResume = false) {
    if (isPaused) return;

    isPaused = true;
    stopSlideshow();

    if (clearResume && resumeTimer) {
      clearTimeout(resumeTimer);
      resumeTimer = null;
    }
  }

  /**
   * Resume the slideshow
   */
  function resumeSlideshow() {
    if (!isPaused || prefersReducedMotion) return;

    // Clear any existing resume timer
    if (resumeTimer) {
      clearTimeout(resumeTimer);
    }

    // Set up auto-resume after delay
    resumeTimer = setTimeout(() => {
      isPaused = false;
      startSlideshow();
    }, CONFIG.resumeDelay);
  }

  /**
   * Go to next slide
   */
  function nextSlide() {
    const nextIndex = (currentSlide + 1) % slides.length;
    goToSlide(nextIndex);
  }

  /**
   * Go to previous slide
   */
  function previousSlide() {
    const prevIndex = (currentSlide - 1 + slides.length) % slides.length;
    goToSlide(prevIndex);
  }

  /**
   * Go to specific slide
   */
  function goToSlide(index) {
    if (index === currentSlide || index < 0 || index >= slides.length) return;

    // Preload next few images if not already loaded
    for (let i = 1; i <= 3; i++) {
      const nextIndex = (index + i) % slides.length;
      preloadImage(nextIndex);
    }

    // Update slide classes
    slides[currentSlide].classList.remove("active", "fade-in");
    slides[index].classList.add("active", "fade-in");

    // Update dot indicators
    updateDots(index);

    // Update ARIA live region for screen readers
    updateAriaLiveRegion(index);

    currentSlide = index;
  }

  /**
   * Update dot indicators
   */
  function updateDots(slideIndex) {
    // Calculate which dot should be active (each dot represents ~5 slides)
    const slidesPerDot = Math.ceil(CONFIG.totalSlides / CONFIG.totalDots);
    const activeDotIndex = Math.floor(slideIndex / slidesPerDot);

    dots.forEach((dot, index) => {
      if (index === activeDotIndex) {
        dot.classList.add("active");
        dot.setAttribute("aria-selected", "true");
      } else {
        dot.classList.remove("active");
        dot.setAttribute("aria-selected", "false");
      }
    });
  }

  /**
   * Handle dot click
   */
  function handleDotClick(dotIndex) {
    // Each dot jumps to a set of slides
    const slidesPerDot = Math.ceil(CONFIG.totalSlides / CONFIG.totalDots);
    const targetSlide = dotIndex * slidesPerDot;

    pauseSlideshow();
    goToSlide(targetSlide);
    resumeSlideshow();
  }

  /**
   * Handle keyboard navigation
   */
  function handleKeyPress(event) {
    switch (event.key) {
      case " ":
      case "Enter":
        event.preventDefault();
        if (isPaused) {
          isPaused = false;
          startSlideshow();
        } else {
          pauseSlideshow(true);
        }
        break;

      case "ArrowLeft":
        event.preventDefault();
        pauseSlideshow();
        previousSlide();
        resumeSlideshow();
        break;

      case "ArrowRight":
        event.preventDefault();
        pauseSlideshow();
        nextSlide();
        resumeSlideshow();
        break;

      case "Escape":
        event.preventDefault();
        pauseSlideshow(true);
        break;
    }
  }

  /**
   * Update ARIA live region for screen readers
   */
  function updateAriaLiveRegion() {
    let liveRegion = document.getElementById("slideshow-live-region");

    if (!liveRegion) {
      liveRegion = document.createElement("div");
      liveRegion.id = "slideshow-live-region";
      liveRegion.className = "sr-only";
      liveRegion.setAttribute("role", "status");
      liveRegion.setAttribute("aria-live", "polite");
      liveRegion.setAttribute("aria-atomic", "true");
      document.body.appendChild(liveRegion);
    }
  }

  // Initialize when DOM is ready
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
