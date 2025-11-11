# Research: Masonry Gallery Layout

**Feature**: 001-masonry-gallery
**Date**: 2025-11-10
**Status**: Complete

## Executive Summary

The gallery page currently implements a JavaScript-based masonry layout using absolute positioning and column-balancing algorithm. Analysis of the existing code (`static/js/gallery.js`) reveals the implementation is functionally complete with proper handling of:

- Dynamic column calculation based on viewport width
- Image positioning using declared width/height attributes
- Aspect ratio preservation
- Intersection Observer for staggered animations
- Responsive resize handling with smooth transitions

**Finding**: The current implementation should already eliminate whitespace within image cards through the `object-cover` CSS class and explicit width/height styling. The research focuses on identifying any potential gaps and documenting best practices for maintaining zero-whitespace masonry layouts.

## Research Areas

### 1. Current Implementation Analysis

**Question**: How does the existing masonry implementation prevent whitespace?

**Findings**:

The current `static/js/gallery.js` implementation (lines 167-229) uses the following approach:

1. **Column-based positioning**: Images are placed in the shortest column using absolute positioning
2. **Explicit dimensions**: Each item gets explicit `width` and `height` via inline styles
3. **Aspect ratio calculation**: Uses declared `width`/`height` attributes from image metadata
4. **Object-fit coverage**: Template applies `object-cover` class to ensure images fill containers

**Key Code Sections**:
```javascript
// From positionItem() method (gallery.js:167-229)
const columnWidth = (containerWidth - this.gap * (this.columnCount - 1)) / this.columnCount;
let itemHeight = columnWidth / aspectRatio;

item.style.width = `${columnWidth}px`;
item.style.height = `${itemHeight}px`;
```

**Assessment**: ✅ The implementation correctly calculates dimensions to eliminate container whitespace. Image cards are sized precisely to their container dimensions.

### 2. CSS object-fit Best Practices

**Question**: What CSS techniques ensure images fill their containers without whitespace?

**Findings**:

**Current Approach** (from `gallery.templ:74`):
```html
<img class="w-full h-full object-cover ..." />
```

**CSS Properties**:
- `w-full` (width: 100%) - Image fills container width
- `h-full` (height: 100%) - Image fills container height
- `object-cover` - Scales image to fill container, cropping if needed to maintain aspect ratio
- `object-position: center` (default) - Centers the image within container

**Alternatives Considered**:

| Property | Pros | Cons | Decision |
|----------|------|------|----------|
| `object-fit: cover` | Fills container completely, no gaps | May crop image edges | ✅ SELECTED - Best for masonry |
| `object-fit: contain` | Shows full image | Creates letterboxing/pillarboxing gaps | ❌ REJECTED - Violates no-whitespace requirement |
| `object-fit: fill` | Fills container | Distorts aspect ratio | ❌ REJECTED - Degrades image quality |
| `background-size: cover` | Same as object-fit | Requires img→div refactor | ❌ REJECTED - Unnecessary complexity |

**Rationale**: `object-cover` is industry standard for masonry layouts (Pinterest, Unsplash, Google Photos all use this approach).

### 3. Absolute Positioning vs CSS Grid Masonry

**Question**: Should we use CSS Grid masonry or continue with JavaScript absolute positioning?

**Findings**:

**CSS Grid Masonry** (`grid-template-rows: masonry`):
- **Browser Support**: Firefox 87+ only (with flag), Chrome/Safari experimental
- **Spec Status**: CSS Grid Level 3 draft, not finalized
- **Production Ready**: ❌ Not yet (as of 2025-11-10)
- **Fallback Complexity**: High - requires JavaScript polyfill anyway

**JavaScript Absolute Positioning** (current approach):
- **Browser Support**: Universal (IE11+)
- **Performance**: Excellent with proper implementation
- **Control**: Fine-grained positioning control
- **Production Ready**: ✅ Battle-tested (Masonry.js, Pinterest, etc.)

**Decision**: ✅ CONTINUE with JavaScript absolute positioning

**Rationale**: CSS Grid masonry is not production-ready. The existing JavaScript implementation provides superior browser compatibility and control. Switching would add complexity without user benefit.

### 4. Layout Shift Prevention (CLS = 0)

**Question**: How do we achieve zero Cumulative Layout Shift during image loading?

**Findings**:

**Current Strategy** (from `gallery.js:39-44`):
```javascript
// Position all items immediately using declared dimensions (width/height attributes)
// This prevents layout shifts when LQIP placeholders load
this.items.forEach((item) => {
  this.positionItem(item);
  this.imagesLoaded++;
});
```

**Key Techniques**:
1. **Declared dimensions**: Images have `width` and `height` attributes from metadata
2. **Immediate positioning**: Items positioned before images load using declared dimensions
3. **Hidden initially**: Items start with `opacity: 0; visibility: hidden` to prevent flash
4. **LQIP placeholders**: Low-quality image placeholders (`*-lqip.jpg`) provide instant visual feedback

**CLS Prevention Checklist**:
- ✅ Width/height attributes set on all images
- ✅ Aspect ratio calculated from metadata (not naturalWidth/naturalHeight)
- ✅ Items positioned before opacity transition
- ✅ Container height set explicitly (`this.gallery.style.height`)
- ✅ No reflow on image load (dimensions pre-calculated)

**Assessment**: Current implementation follows best practices for zero CLS.

### 5. Responsive Breakpoint Strategy

**Question**: What column counts should be used at different viewport widths?

**Findings**:

**Current Breakpoints** (from `gallery.js:130-145`):
```javascript
if (width < 640) {
  this.columnCount = 1;
} else if (width < 768) {
  this.columnCount = 2;
} else if (width < 1024) {
  this.columnCount = 3;
} else {
  this.columnCount = 4;
}
```

**Industry Benchmarks**:

| Service | Mobile (<640px) | Tablet (768-1024px) | Desktop (1024px+) |
|---------|-----------------|---------------------|-------------------|
| Pinterest | 1-2 cols | 3-4 cols | 5-7 cols |
| Unsplash | 1 col | 2-3 cols | 3-4 cols |
| Google Photos | 2-3 cols | 4-5 cols | 6-8 cols |
| **Current Implementation** | 1 col | 2-3 cols | 4 cols |

**Assessment**: ✅ Current breakpoints align with Unsplash (photo-focused gallery). Appropriate for wedding website use case.

**Recommendation**: Keep existing breakpoints. They balance visual density with image appreciation.

### 6. Performance Optimization Patterns

**Question**: What optimizations ensure 60fps scrolling with staggered animations?

**Findings**:

**Current Optimizations**:

1. **Intersection Observer** (gallery.js:86-125):
   - Lazy animation triggering (only visible items animate)
   - Viewport margin: `100px` (preload slightly offscreen)
   - Threshold: `0.01` (trigger when 1% visible)
   - Staggered delay: `(index % 8) * 40ms` (smooth cascade)

2. **Debounced Resize** (gallery.js:287-298):
   - 200ms debounce on window resize
   - Prevents excessive recalculations
   - Smooth transitions during resize

3. **GPU-Accelerated Properties**:
   - Uses `transform: translateY()` for animations (not `top`)
   - Uses `opacity` transitions (GPU-composited)
   - Avoids layout-triggering properties during animations

4. **will-change Hints** (needed?):
   - ❌ NOT currently used
   - Could add `will-change: transform, opacity` for animation hints
   - **Trade-off**: Memory overhead vs. animation smoothness

**Performance Checklist**:
- ✅ Intersection Observer (better than scroll events)
- ✅ Debounced resize handler
- ✅ Transform-based animations
- ⚠️ Missing `will-change` hints (minor optimization opportunity)
- ✅ Reduced motion support (lines 302-321)

**Recommendation**: Add `will-change: transform, opacity` to `.gallery-item-animate` class for marginal performance improvement.

### 7. Accessibility Considerations

**Question**: How do we ensure keyboard navigation and screen reader support?

**Findings**:

**Current Implementation**:

1. **Keyboard Navigation** (gallery.js:272-283):
   ```javascript
   item.addEventListener("keydown", (e) => {
     if (e.key === "Enter" || e.key === " ") {
       e.preventDefault();
       item.click();
     }
   });
   ```

2. **ARIA Attributes** (gallery.templ:36):
   ```html
   <div ... role="button" tabindex="0"
        aria-label="View photo X in full size">
   ```

3. **Focus Management**:
   - ✅ `tabindex="0"` allows keyboard focus
   - ✅ `role="button"` announces interactive element
   - ✅ Enter/Space trigger click (standard button behavior)

**Missing Enhancements**:
- ⚠️ No visible focus indicator (`:focus-visible` styling)
- ⚠️ No announcement of masonry layout structure
- ⚠️ No skip link for long galleries

**Recommendation**: Add focus-visible styles. Screen reader announcements are adequate for static image gallery.

## Decisions Summary

### Decision 1: Masonry Implementation Approach

**Decision**: Continue with JavaScript absolute positioning (do not migrate to CSS Grid masonry)

**Rationale**:
- CSS Grid masonry not production-ready (experimental in Chrome/Safari, Firefox-only with flag)
- Current JavaScript implementation is battle-tested and performant
- Switching would add complexity without user benefit
- Industry standard (Pinterest, Masonry.js use same approach)

**Alternatives Considered**:
1. CSS Grid masonry → Rejected (browser support)
2. Column-count CSS → Rejected (limited control, can't eliminate gaps)
3. Flexbox masonry → Rejected (row-based, not true masonry)

### Decision 2: Whitespace Elimination Strategy

**Decision**: Use `object-fit: cover` with explicit container dimensions

**Rationale**:
- `object-cover` fills containers completely with no gaps
- Explicit width/height from JavaScript ensures precise container sizing
- Industry standard approach (all major photo galleries use this)
- Maintains aspect ratios while eliminating whitespace

**Alternatives Considered**:
1. `object-fit: contain` → Rejected (creates letterbox gaps)
2. `object-fit: fill` → Rejected (distorts images)
3. Background images → Rejected (unnecessary complexity)

### Decision 3: Layout Shift Prevention

**Decision**: Pre-position items using declared image dimensions before opacity fade-in

**Rationale**:
- Metadata provides width/height for all images
- Positioning before visibility eliminates reflow
- LQIP placeholders provide instant visual feedback
- Achieves CLS = 0 target

**Alternatives Considered**:
1. Wait for naturalWidth/naturalHeight → Rejected (causes layout shift)
2. Fixed aspect ratios → Rejected (distorts varied images)
3. Server-side rendering → N/A (static site)

### Decision 4: Responsive Breakpoints

**Decision**: Keep existing breakpoints (1/2/3/4 columns at 640/768/1024px)

**Rationale**:
- Aligns with Tailwind defaults and industry standards
- Balances visual density with image appreciation
- Works well for 21-image gallery size
- Matches user expectations from other photo galleries

**Alternatives Considered**:
1. More aggressive columns (Pinterest-style) → Rejected (too dense for wedding photos)
2. Fewer columns → Rejected (wastes screen space on desktop)
3. Dynamic columns by image count → Rejected (unnecessary complexity)

### Decision 5: Performance Optimizations

**Decision**: Add `will-change: transform, opacity` to animating items

**Rationale**:
- Minimal memory overhead for 21 images
- Hints browser to GPU-composite layers
- Improves animation smoothness on mid-range devices
- Easy to implement (1 line of CSS)

**Alternatives Considered**:
1. No will-change (current) → Replaced (marginal improvement available)
2. will-change on all items → Rejected (memory overhead)
3. JavaScript-based layer hints → Rejected (overkill)

### Decision 6: Accessibility Enhancements

**Decision**: Add focus-visible styling, keep current keyboard navigation

**Rationale**:
- Current keyboard navigation is functional (Enter/Space work)
- Focus indicators improve usability
- Screen reader support is adequate (role/aria-label)
- Minimal implementation effort

**Alternatives Considered**:
1. Complex focus management → Rejected (overkill for image gallery)
2. Skip links → Rejected (21 images not long enough)
3. Roving tabindex → Rejected (unnecessary for static gallery)

## Technical Recommendations

### Immediate Fixes (Zero Whitespace Goal)

1. **Verify object-cover is working**:
   - Inspect image cards for any overflow or gap issues
   - Ensure `w-full h-full` classes are applied
   - Verify no padding/margin on image elements

2. **Add will-change optimization**:
   ```css
   .gallery-item-animate {
     will-change: transform, opacity;
   }
   ```

3. **Add focus-visible styling**:
   ```css
   .gallery-item:focus-visible {
     outline: 3px solid #93c5fd; /* blue-300 */
     outline-offset: 4px;
     z-index: 20;
   }
   ```

### Validation Tests

Before considering this feature complete, verify:

- [ ] No visible gaps within image cards on all screen sizes
- [ ] Images fill containers edge-to-edge (inspect with DevTools)
- [ ] Aspect ratios preserved (no distortion)
- [ ] Layout stable during image loading (CLS = 0)
- [ ] Smooth animations on scroll (60fps)
- [ ] Responsive breakpoints work (1/2/3/4 cols)
- [ ] Keyboard navigation functional (Tab, Enter, Space)
- [ ] Focus indicators visible
- [ ] Reduced motion respected

## References

- [MDN: object-fit](https://developer.mozilla.org/en-US/docs/Web/CSS/object-fit)
- [CSS Grid Level 3: Masonry Layout](https://drafts.csswg.org/css-grid-3/#masonry-layout)
- [Web.dev: Cumulative Layout Shift](https://web.dev/cls/)
- [Pinterest Masonry Algorithm](https://medium.com/pinterest-engineering/building-pinterest-masonry-f2c3f6d6e7db)
- [Intersection Observer API](https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API)

## Open Questions

**Q1**: Should we add column count customization for users?
**A**: No - adds complexity without clear user benefit. Wedding guests don't need layout controls.

**Q2**: Should we lazy-load images below the fold?
**A**: Already implemented - images have `loading="lazy"` attribute (gallery.templ:75).

**Q3**: Do we need to support IE11?
**A**: No - project targets modern browsers (Chrome 90+, per Technical Context). Intersection Observer not supported in IE11.
