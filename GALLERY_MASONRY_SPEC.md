# Gallery Masonry Layout Migration Specification

## Document Overview

**Purpose**: Migrate the current fixed-grid gallery to a dynamic masonry layout with enhanced visual effects and interactions.

**Reference**: UI_DESIGN_RECOMMENDATIONS.md Section 1.2 & 4.2

**Estimated Effort**: 6-8 hours (Phase 1), 12+ hours (with lightbox)

**Priority**: High-Impact Feature (Wow Factor: 8-10/10)

---

## Current State Analysis

### Existing Implementation

**Template**: `internal/views/gallery.templ`
- Standard CSS Grid layout (`grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4`)
- 263 photos (molly_andrewENG.jpg + molly_andrewENG-2.jpg through molly_andrewENG-263.jpg)
- Forced aspect-square sizing (all images same height)
- Basic lazy loading with `loading="lazy"`
- Simple hover effects (scale, shadow, text overlay)

**Handler**: `internal/handlers/gallery.go`
- Minimal server-side logic
- Direct Templ rendering

### Limitations

1. **Visual Monotony**: All images forced to squares loses composition variety
2. **Wasted Space**: Variable aspect ratios not utilized
3. **Static Layout**: No dynamic repositioning or optimization
4. **Basic Interactions**: No lightbox, zoom, or advanced navigation
5. **Performance**: All 263 images rendered immediately (though lazy-loaded)

---

## Target Architecture

### Design Goals

1. **Variable Heights**: Allow images to display at natural aspect ratios
2. **Intelligent Positioning**: Minimize column height differences for balanced layout
3. **Enhanced Interactivity**: Smooth hover effects, animations, transitions
4. **Progressive Enhancement**: Blur-up placeholders, staggered animations
5. **Performance**: Intersection Observer for scroll animations, optimized rendering
6. **Responsive**: Adaptive column count (1/2/3/4 based on viewport)

### Technical Approach

**Option A: CSS Grid Masonry (Recommended for Phase 1)**
- Simpler implementation
- Native CSS performance
- Better for static site generation
- Works without JavaScript for basic layout

**Option B: JavaScript Column Distribution (Phase 2 Enhancement)**
- More control over image placement
- Can optimize for minimum height variance
- Enables advanced features (filtering, sorting)
- Requires client-side calculation

### Recommended Strategy: Hybrid Approach

**Phase 1 (6-8 hours)**: CSS Grid-based masonry with JavaScript enhancements
**Phase 2 (12+ hours)**: Full lightbox with advanced features

---

## Phase 1: Core Masonry Implementation

### 1.1 Image Metadata Preparation

**Goal**: Determine natural aspect ratios for all 263 images to calculate grid spans.

**Option A: Build-time Metadata Generation** (Recommended)
Create a Go script to analyze images and generate metadata:

```go
// cmd/gallery-metadata/main.go
package main

import (
    "encoding/json"
    "fmt"
    "image"
    _ "image/jpeg"
    "os"
    "path/filepath"
)

type ImageMetadata struct {
    Filename     string  `json:"filename"`
    Width        int     `json:"width"`
    Height       int     `json:"height"`
    AspectRatio  float64 `json:"aspectRatio"`
    GridRowSpan  int     `json:"gridRowSpan"`
}

func main() {
    staticDir := "static"
    images := []ImageMetadata{}

    // Process first image
    images = append(images, processImage(filepath.Join(staticDir, "molly_andrewENG.jpg")))

    // Process numbered images (2-263)
    for i := 2; i <= 263; i++ {
        filename := fmt.Sprintf("molly_andrewENG-%d.jpg", i)
        images = append(images, processImage(filepath.Join(staticDir, filename)))
    }

    // Write metadata JSON
    data, _ := json.MarshalIndent(images, "", "  ")
    os.WriteFile("static/gallery-metadata.json", data, 0644)
}

func processImage(path string) ImageMetadata {
    file, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    config, _, err := image.DecodeConfig(file)
    if err != nil {
        panic(err)
    }

    aspectRatio := float64(config.Width) / float64(config.Height)

    // Calculate grid row span (assuming 20px row height)
    // Normalize to a base width of 300px
    normalizedHeight := 300.0 / aspectRatio
    gridRowSpan := int(normalizedHeight / 20)

    return ImageMetadata{
        Filename:    filepath.Base(path),
        Width:       config.Width,
        Height:      config.Height,
        AspectRatio: aspectRatio,
        GridRowSpan: gridRowSpan,
    }
}
```

**Makefile Addition**:
```makefile
gallery-metadata:
	@echo "Generating gallery metadata..."
	go run cmd/gallery-metadata/main.go
```

**Option B: Client-side Calculation** (Fallback)
Calculate row spans dynamically in JavaScript after images load.

---

### 1.2 Template Updates

**File**: `internal/views/gallery.templ`

**Changes**:

1. **Load Metadata**: Pass image metadata to template
2. **Remove aspect-square**: Allow natural aspect ratios
3. **Add Grid Row Spans**: Use CSS custom properties for dynamic row spans
4. **Add Animation Classes**: Prepare for staggered fade-ins

```templ
package views

import (
    "fmt"
    "encoding/json"
    "os"
)

// ImageData represents gallery image metadata
type ImageData struct {
    Filename    string  `json:"filename"`
    Width       int     `json:"width"`
    Height      int     `json:"height"`
    AspectRatio float64 `json:"aspectRatio"`
    GridRowSpan int     `json:"gridRowSpan"`
}

// Load metadata at build time
func loadGalleryMetadata() []ImageData {
    data, err := os.ReadFile("static/gallery-metadata.json")
    if err != nil {
        // Fallback to basic rendering without metadata
        return nil
    }

    var images []ImageData
    json.Unmarshal(data, &images)
    return images
}

templ GalleryPage() {
    @galleryMetadata := loadGalleryMetadata()

    <div class="min-h-screen bg-white">
        // Header (unchanged)
        <header class="bg-pattern-subtle py-16">
            <div class="max-w-7xl mx-auto px-6">
                <div class="text-center">
                    <h1 class="script text-6xl md:text-8xl text-blue-300 font-extralight mb-4">Our Gallery</h1>
                    <p class="text-gray-600 text-lg">Engagement Photos</p>
                    <div class="mt-8">
                        <a href="/" class="text-blue-300 hover:text-blue-400 transition-colors underline">← Back to Home</a>
                    </div>
                </div>
            </div>
        </header>

        // Masonry Gallery Grid
        <div class="max-w-7xl mx-auto px-6 py-16">
            <div
                id="gallery-masonry"
                class="gallery-masonry grid gap-4"
                style="
                    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
                    grid-auto-rows: 20px;
                "
            >
                if galleryMetadata != nil {
                    for i, img := range galleryMetadata {
                        <div
                            class="gallery-item group relative overflow-hidden rounded-lg shadow-md hover:shadow-2xl transition-all duration-500 opacity-0 gallery-item-animate"
                            style={ fmt.Sprintf("grid-row: span %d;", img.GridRowSpan) }
                            data-index={ fmt.Sprintf("%d", i) }
                        >
                            <img
                                src={ fmt.Sprintf("/static/%s", img.Filename) }
                                alt={ fmt.Sprintf("Engagement photo %d", i+1) }
                                class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                loading="lazy"
                                data-aspect-ratio={ fmt.Sprintf("%.3f", img.AspectRatio) }
                            />
                            <div class="gallery-overlay absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500 flex items-end justify-center pb-6">
                                <div class="text-center transform translate-y-4 group-hover:translate-y-0 transition-transform duration-500">
                                    <span class="text-white text-sm font-medium">Photo { fmt.Sprintf("%d", i+1) }</span>
                                    <div class="mt-2 flex gap-3 justify-center">
                                        <button
                                            class="text-white hover:text-blue-300 transition-colors"
                                            aria-label="View full size"
                                            data-photo-index={ fmt.Sprintf("%d", i) }
                                        >
                                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"/>
                                            </svg>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    }
                } else {
                    // Fallback: render without metadata (client-side calculation)
                    <div class="gallery-item group relative overflow-hidden rounded-lg shadow-md hover:shadow-2xl transition-all duration-500">
                        <img
                            src="/static/molly_andrewENG.jpg"
                            alt="Engagement photo 1"
                            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                            loading="lazy"
                        />
                    </div>
                    for i := 2; i <= 263; i++ {
                        <div class="gallery-item group relative overflow-hidden rounded-lg shadow-md hover:shadow-2xl transition-all duration-500">
                            <img
                                src={ fmt.Sprintf("/static/molly_andrewENG-%d.jpg", i) }
                                alt={ fmt.Sprintf("Engagement photo %d", i) }
                                class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                loading="lazy"
                            />
                        </div>
                    }
                }
            </div>
        </div>

        // Back to top button (unchanged)
        <div class="text-center pb-16">
            <button
                onclick="window.scrollTo({top: 0, behavior: 'smooth'})"
                class="inline-flex items-center gap-2 px-6 py-3 bg-blue-300 text-white rounded-full hover:bg-blue-400 transition-colors shadow-lg hover:shadow-xl"
            >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" class="w-5 h-5">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18"/>
                </svg>
                Back to Top
            </button>
        </div>

        // Include gallery JavaScript
        <script src="/static/js/gallery.js"></script>
    </div>
}
```

---

### 1.3 JavaScript Implementation

**File**: `static/js/gallery.js` (new file)

```javascript
/**
 * Gallery Masonry Layout & Animations
 * Handles staggered animations, responsive layout, and scroll effects
 */

class GalleryMasonry {
    constructor() {
        this.gallery = document.getElementById('gallery-masonry');
        this.items = Array.from(document.querySelectorAll('.gallery-item'));
        this.observer = null;

        this.init();
    }

    init() {
        this.setupIntersectionObserver();
        this.handleResize();

        // Adjust layout after images load
        this.items.forEach(item => {
            const img = item.querySelector('img');
            if (img.complete) {
                this.handleImageLoad(item);
            } else {
                img.addEventListener('load', () => this.handleImageLoad(item));
            }
        });

        // Responsive handling
        window.addEventListener('resize', this.debounce(() => this.handleResize(), 200));
    }

    /**
     * Setup Intersection Observer for staggered fade-in animations
     */
    setupIntersectionObserver() {
        const options = {
            root: null,
            rootMargin: '50px',
            threshold: 0.1
        };

        this.observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const item = entry.target;
                    const index = parseInt(item.dataset.index || '0');

                    // Staggered animation delay
                    const delay = (index % 12) * 50; // 50ms between items in viewport

                    setTimeout(() => {
                        item.style.opacity = '1';
                        item.style.transform = 'translateY(0)';
                    }, delay);

                    // Stop observing once animated
                    this.observer.unobserve(item);
                }
            });
        }, options);

        // Observe all items
        this.items.forEach(item => {
            item.style.transform = 'translateY(20px)'; // Initial state
            this.observer.observe(item);
        });
    }

    /**
     * Calculate grid row span if metadata not available
     */
    handleImageLoad(item) {
        const img = item.querySelector('img');

        // If row span already set (from metadata), skip
        if (item.style.gridRow) return;

        // Calculate row span based on aspect ratio
        const aspectRatio = img.naturalWidth / img.naturalHeight;
        const containerWidth = item.offsetWidth;
        const naturalHeight = containerWidth / aspectRatio;
        const rowHeight = 20; // Must match grid-auto-rows
        const rowSpan = Math.ceil(naturalHeight / rowHeight);

        item.style.gridRow = `span ${rowSpan}`;
    }

    /**
     * Adjust column count based on viewport width
     */
    handleResize() {
        const width = window.innerWidth;
        let columns = 4;

        if (width < 640) columns = 1;
        else if (width < 768) columns = 2;
        else if (width < 1024) columns = 3;

        this.gallery.style.gridTemplateColumns = `repeat(${columns}, 1fr)`;
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

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new GalleryMasonry());
} else {
    new GalleryMasonry();
}
```

---

### 1.4 CSS Enhancements

**File**: `static/css/styles.css` (additions)

```css
/* Gallery Masonry Layout */
.gallery-masonry {
    /* Grid defined inline in template for dynamic configuration */
}

.gallery-item {
    cursor: pointer;
    will-change: transform, opacity, box-shadow;
    transition: opacity 0.6s cubic-bezier(0.16, 1, 0.3, 1),
                transform 0.6s cubic-bezier(0.16, 1, 0.3, 1),
                box-shadow 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

.gallery-item:hover {
    z-index: 10;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

/* Enhanced hover effect with subtle 3D tilt */
@media (hover: hover) and (pointer: fine) {
    .gallery-item {
        transform-style: preserve-3d;
    }

    .gallery-item:hover {
        transform: translateY(-8px) scale(1.02);
    }
}

/* Overlay enhancements */
.gallery-overlay {
    backdrop-filter: blur(8px);
}

/* Smooth image transitions */
.gallery-item img {
    will-change: transform;
}

/* Loading skeleton (optional for progressive enhancement) */
.gallery-item.loading {
    background: linear-gradient(
        90deg,
        #f0f0f0 25%,
        #e0e0e0 50%,
        #f0f0f0 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
    0% { background-position: -200% 0; }
    100% { background-position: 200% 0; }
}

/* Respect reduced motion preferences */
@media (prefers-reduced-motion: reduce) {
    .gallery-item,
    .gallery-item img,
    .gallery-overlay {
        transition: none !important;
        animation: none !important;
    }

    .gallery-item-animate {
        opacity: 1 !important;
        transform: none !important;
    }
}
```

---

### 1.5 Build Integration

**Update**: `internal/views/app.templ`

Ensure gallery.js is loaded on the gallery page:

```templ
// In the <head> or before </body>
if isGalleryPage {
    <script src="/static/js/gallery.js" defer></script>
}
```

**Update**: `Makefile`

Add metadata generation to build process:

```makefile
# Add to existing build targets
all: tpl styles gallery-metadata server

gallery-metadata:
	@echo "Generating gallery image metadata..."
	@go run cmd/gallery-metadata/main.go

static-build: tpl styles gallery-metadata
	@echo "Building static site..."
	# ... existing static build commands
```

---

## Phase 1 Testing Checklist

- [ ] Gallery metadata generates correctly for all 263 images
- [ ] Images display at natural aspect ratios (no forced squares)
- [ ] Layout balances across columns (no significant height differences)
- [ ] Staggered fade-in animations work on scroll
- [ ] Hover effects are smooth and performant
- [ ] Responsive behavior works (1/2/3/4 columns)
- [ ] Lazy loading still functional
- [ ] No layout shift after images load
- [ ] Reduced motion preferences respected
- [ ] Works without JavaScript (graceful degradation)
- [ ] Static build generates properly
- [ ] Performance: no jank, smooth 60fps scrolling

---

## Phase 2: Advanced Features (Future Enhancement)

### 2.1 Full-Featured Lightbox

**Reference**: UI_DESIGN_RECOMMENDATIONS.md Section 4.1

**Features**:
- Modal overlay with backdrop blur
- Keyboard navigation (arrow keys, ESC)
- Swipe gestures (mobile)
- Zoom/pan functionality
- Thumbnail strip navigation
- Photo counter (47 / 263)
- Share/download buttons
- Preload adjacent images

**Estimated Effort**: 12+ hours

**New File**: `static/js/lightbox.js`

---

## Performance Considerations

### Image Optimization

**Current**: JPEG files at original dimensions

**Recommended**:
1. Generate multiple sizes (400w, 800w, 1200w, 1600w)
2. Convert to WebP with JPEG fallback
3. Use `<picture>` element with srcset
4. Implement blur-up placeholders (LQIP)

**Script**: Update `cmd/photo/photo.go` to generate responsive image sets

### Loading Strategy

1. **Above the fold**: Load first 12 images immediately
2. **Below the fold**: Intersection Observer triggers lazy load
3. **Blur-up**: Show tiny blurred version while full image loads
4. **Preloading**: Preload images in next viewport section

---

## Accessibility Enhancements

### Required ARIA Attributes

```html
<div
    class="gallery-item"
    role="button"
    tabindex="0"
    aria-label="View photo 47 in lightbox"
>
```

### Keyboard Navigation

- **Tab**: Navigate between gallery items
- **Enter/Space**: Open lightbox
- **Arrow Keys**: Navigate in lightbox (Phase 2)
- **ESC**: Close lightbox (Phase 2)

### Focus Management

```css
.gallery-item:focus-visible {
    outline: 3px solid var(--color-primary);
    outline-offset: 4px;
}
```

---

## Migration Steps

### Step 1: Preparation (30 min)
1. Create `cmd/gallery-metadata/main.go`
2. Run metadata generation
3. Verify `static/gallery-metadata.json` created correctly

### Step 2: Template Update (1 hour)
1. Update `internal/views/gallery.templ`
2. Add metadata loading function
3. Update grid structure
4. Remove aspect-square classes
5. Add animation classes

### Step 3: JavaScript Implementation (2 hours)
1. Create `static/js/gallery.js`
2. Implement GalleryMasonry class
3. Add Intersection Observer
4. Add responsive handlers
5. Test staggered animations

### Step 4: CSS Enhancements (1 hour)
1. Update `static/css/styles.css`
2. Add masonry-specific styles
3. Enhance hover effects
4. Add loading states
5. Add reduced-motion support

### Step 5: Build Integration (30 min)
1. Update Makefile
2. Update app.templ to load gallery.js
3. Test static build generation
4. Verify deployment to S3

### Step 6: Testing & Refinement (2 hours)
1. Cross-browser testing
2. Mobile responsive testing
3. Performance profiling
4. Accessibility audit
5. Visual QA
6. Bug fixes and polish

---

## Success Metrics

### Visual Quality
- ✅ Images display at natural aspect ratios
- ✅ Balanced column heights (variance < 200px)
- ✅ Smooth, professional animations
- ✅ Enhanced hover effects

### Performance
- ✅ First Contentful Paint < 1.5s
- ✅ Largest Contentful Paint < 2.5s
- ✅ No Cumulative Layout Shift
- ✅ Smooth 60fps scrolling

### User Experience
- ✅ Staggered animations create visual interest
- ✅ Responsive across all device sizes
- ✅ Keyboard accessible
- ✅ Reduced motion supported

---

## Rollback Plan

If issues arise during deployment:

1. **Revert template**: Restore `internal/views/gallery.templ` from git
2. **Remove JS**: Comment out `<script src="/static/js/gallery.js">`
3. **Rebuild**: Run `make static-build`
4. **Redeploy**: Run `make static-deploy`

**Estimated rollback time**: < 5 minutes

---

## References

- **Design Doc**: UI_DESIGN_RECOMMENDATIONS.md
- **Current Implementation**: internal/views/gallery.templ:21
- **CSS Grid Masonry**: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Grid_Layout/Masonry_Layout
- **Intersection Observer**: https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
- **Web Vitals**: https://web.dev/vitals/

---

## Next Steps

After completing Phase 1, evaluate for Phase 2 enhancements:
- [ ] Full lightbox implementation
- [ ] Advanced image optimization (WebP, srcset)
- [ ] Blur-up placeholders

**Estimated Total Time**: 6-8 hours (Phase 1) + 12-20 hours (Phase 2)
