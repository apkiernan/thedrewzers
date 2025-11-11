# Quickstart: Masonry Gallery Layout

**Feature**: 001-masonry-gallery
**Target Audience**: Developers implementing the masonry gallery fixes
**Time to Complete**: 15 minutes setup + implementation

## Overview

This feature fixes empty whitespace issues in the gallery masonry layout. The current implementation already has a functional masonry system using JavaScript absolute positioning. This quickstart guides you through verifying the implementation, identifying whitespace issues, and applying fixes.

## Prerequisites

- Go 1.23.3+ installed
- Node.js 16+ (for Tailwind CSS)
- `templ` CLI installed (`go install github.com/a-h/templ/cmd/templ@latest`)
- Project cloned and dependencies installed

## Quick Commands

```bash
# Development: Run local server with hot reload
make server

# Build static site for deployment
make static-build

# Deploy to S3 + invalidate CloudFront
make deploy

# Regenerate image metadata (if needed)
make optimize-images

# Run Lighthouse performance audit
npm run lighthouse
```

## Project Structure (Key Files)

```text
internal/views/
  └── gallery.templ           # Template with masonry markup

static/
  ├── gallery-metadata.json   # Image dimensions
  └── js/
      ├── gallery.js          # Masonry positioning logic
      └── lightbox.js         # Image lightbox

dist/
  ├── gallery.html            # Generated static HTML
  └── js/
      └── gallery.*.min.js    # Minified JavaScript
```

## Step 1: Verify Current Implementation

### 1.1 Start Development Server

```bash
# Terminal 1: Start server with hot reload
make server

# Server runs at http://localhost:8080
# Navigate to http://localhost:8080/gallery
```

### 1.2 Inspect Gallery Layout

**Visual Inspection**:
1. Open gallery page in browser
2. Open DevTools (F12) → Elements tab
3. Inspect a `.gallery-item` element
4. Check computed styles for:
   - `width` and `height` (should be explicit pixel values)
   - `position: absolute` (set by JavaScript)
   - Image: `width: 100%`, `height: 100%`, `object-fit: cover`

**Expected Layout**:
- Gallery container has `position: relative`
- Each `.gallery-item` has absolute positioning
- Images fill their containers edge-to-edge
- No visible gaps within image cards

### 1.3 Check Console for Errors

**Open Console** (F12 → Console):
- No JavaScript errors
- Should see: "Gallery masonry element found" (if logging enabled)
- Intersection Observer working (items fade in on scroll)

**Common Issues**:
```javascript
// If you see: "Gallery masonry element not found"
// Check: <div id="gallery-masonry"> exists in template

// If images don't load:
// Check: gallery-metadata.json exists in static/
// Check: Image files exist (run `make optimize-images`)
```

## Step 2: Identify Whitespace Issues

### 2.1 Visual Whitespace Detection

**Use DevTools Ruler**:
1. Open DevTools → Elements
2. Hover over `.gallery-item`
3. Check highlighted box dimensions:
   - Container box should match image content exactly
   - No padding/margin inside container
   - Image should fill entire container (blue highlight)

**Measurement Tool**:
```javascript
// Run in console to measure all items
document.querySelectorAll('.gallery-item').forEach((item, i) => {
  const img = item.querySelector('img');
  const itemBox = item.getBoundingClientRect();
  const imgBox = img.getBoundingClientRect();

  console.log(`Item ${i}:`,
    `Container: ${itemBox.width}x${itemBox.height}`,
    `Image: ${imgBox.width}x${imgBox.height}`,
    `Match: ${Math.abs(itemBox.width - imgBox.width) < 1}`
  );
});
```

### 2.2 Common Whitespace Causes

| Issue | Symptom | Fix Location |
|-------|---------|--------------|
| Missing `object-cover` | Letterboxing/pillarboxing | `gallery.templ:74` (add class) |
| Padding on container | Gaps around image | `gallery.templ:36` (remove padding) |
| Incorrect dimensions | Stretched/squashed | `gallery.js:167-229` (check calculations) |
| Border on container | Thin gaps | CSS (remove border) |

## Step 3: Apply Fixes

### 3.1 Ensure object-fit Coverage

**File**: `internal/views/gallery.templ:74`

```templ
<img
  src={...}
  alt={...}
  class="w-full h-full object-cover ..."  // ← Verify these classes
  loading="lazy"
/>
```

**Verify**:
- `w-full` → `width: 100%`
- `h-full` → `height: 100%`
- `object-cover` → Fill container, crop if needed

### 3.2 Add Performance Optimization

**File**: `static/css/styles.css` or `tailwind.config.js`

Add `will-change` hint for animations:

```css
/* If using custom CSS */
.gallery-item-animate {
  will-change: transform, opacity;
}
```

Or in Tailwind (via tailwind.config.js):

```javascript
// Add to theme.extend
theme: {
  extend: {
    // ... existing config
  }
}

// Or add to HTML
<div class="gallery-item ... will-change-transform will-change-opacity">
```

### 3.3 Add Focus Indicator

**File**: `static/css/styles.css`

```css
.gallery-item:focus-visible {
  outline: 3px solid #93c5fd; /* blue-300 */
  outline-offset: 4px;
  z-index: 20;
}
```

## Step 4: Test the Implementation

### 4.1 Visual Testing Checklist

Open `http://localhost:8080/gallery` and verify:

- [ ] No whitespace gaps inside image cards
- [ ] Images fill containers edge-to-edge
- [ ] Aspect ratios preserved (no distortion)
- [ ] Columns balanced (heights similar)
- [ ] Smooth scroll animations (staggered fade-in)
- [ ] Hover effects work (scale, shadow, overlay)

### 4.2 Responsive Testing

Test breakpoints using DevTools Device Toolbar:

```text
Mobile (375px):    1 column
Tablet (768px):    2-3 columns
Desktop (1280px):  4 columns
Wide (1920px):     4 columns
```

**Resize Test**:
1. Drag browser window smaller/larger
2. Verify layout smoothly transitions
3. No items overlap or disappear
4. Gallery height adjusts correctly

### 4.3 Accessibility Testing

**Keyboard Navigation**:
- [ ] Press `Tab` to focus gallery items
- [ ] Focus indicator visible (outline)
- [ ] Press `Enter` or `Space` to open lightbox
- [ ] All items reachable via keyboard

**Screen Reader**:
```html
<!-- Each item should announce -->
"View photo 1 in full size, button"
```

### 4.4 Performance Testing

**Run Lighthouse Audit**:

```bash
# Terminal 2 (while dev server running)
npm run lighthouse

# Check scores:
# Performance: 90+
# CLS: 0 (zero layout shift)
# LCP: <2.5s
```

**Manual FPS Check**:
1. Open DevTools → Performance
2. Start recording
3. Scroll through gallery
4. Stop recording
5. Check: 60fps maintained, no long tasks

## Step 5: Build and Deploy

### 5.1 Build Static Site

```bash
# Generate static HTML + minified assets
make static-build

# Verify output
ls -lh dist/gallery.html
ls -lh dist/js/gallery.*.min.js

# Check file sizes
# gallery.html: ~50-100KB
# gallery.js: ~3-5KB (minified)
```

### 5.2 Local Testing of Static Build

```bash
# Serve dist/ directory (using Python)
cd dist
python3 -m http.server 8000

# Or using Node.js
npx serve dist -p 8000

# Open: http://localhost:8000/gallery.html
# Test: All functionality works from static files
```

### 5.3 Deploy to Production

```bash
# Full deployment (build + upload + cache invalidate)
make deploy

# Or step-by-step:
make static-build       # 1. Build
make upload-static      # 2. Upload to S3
make invalidate-cache   # 3. Clear CloudFront

# Verify deployment
# Visit: https://your-domain.com/gallery
```

## Troubleshooting

### Issue: Images Not Loading

**Symptoms**: Broken image icons, 404 errors

**Diagnosis**:
```bash
# Check metadata exists
cat static/gallery-metadata.json

# Check image files exist
ls static/images/*.jpg
ls static/images/*-640w.avif
ls static/images/*-768w.webp
```

**Fix**:
```bash
# Regenerate optimized images
make optimize-images

# Rebuild static site
make static-build
```

### Issue: Layout Shifts on Load

**Symptoms**: Images jump around when loading, poor CLS score

**Diagnosis**:
```javascript
// Check if width/height attributes present
document.querySelectorAll('.gallery-item img').forEach(img => {
  console.log(img.getAttribute('width'), img.getAttribute('height'));
});
```

**Fix**:
- Ensure `gallery-metadata.json` has correct dimensions
- Verify template includes `width` and `height` attributes (gallery.templ:72-73)
- Check JavaScript uses declared dimensions, not `naturalWidth` (gallery.js:179-182)

### Issue: Whitespace Still Visible

**Symptoms**: Gaps around images within cards

**Diagnosis**:
```javascript
// Check actual vs expected dimensions
document.querySelectorAll('.gallery-item').forEach(item => {
  const img = item.querySelector('img');
  const computedStyle = getComputedStyle(img);

  console.log('object-fit:', computedStyle.objectFit); // Should be "cover"
  console.log('width:', computedStyle.width);          // Should be "100%"
  console.log('height:', computedStyle.height);        // Should be "100%"
});
```

**Fix**:
```templ
<!-- Ensure these classes in gallery.templ -->
<img class="w-full h-full object-cover" ... />
```

### Issue: Animations Not Working

**Symptoms**: No fade-in effects, items appear instantly

**Diagnosis**:
```javascript
// Check Intersection Observer support
console.log('IntersectionObserver' in window); // Should be true

// Check items have animation classes
document.querySelectorAll('.gallery-item-animate').length; // Should match image count
```

**Fix**:
- Verify `.gallery-item-animate` class present (gallery.templ:36)
- Check JavaScript initializes (console should have no errors)
- Test in modern browser (Chrome 90+, Firefox 88+, Safari 14+)

## Development Workflow

### Iterative Development

```bash
# 1. Make changes to template or JavaScript
vim internal/views/gallery.templ
vim static/js/gallery.js

# 2. Save (air auto-rebuilds)
# Watch terminal for rebuild messages

# 3. Refresh browser (http://localhost:8080/gallery)
# Changes reflected immediately

# 4. Iterate until satisfied
```

### Pre-Commit Checklist

Before committing changes:

- [ ] `make static-build` succeeds without errors
- [ ] No console errors in browser
- [ ] Lighthouse performance score 90+
- [ ] Visual QA passed (all checklist items)
- [ ] Responsive testing complete
- [ ] Accessibility testing complete

## Next Steps

After completing this quickstart:

1. **Generate Tasks**: Run `/speckit.tasks` to create implementation checklist
2. **Implement Fixes**: Follow task list to apply identified fixes
3. **Test Thoroughly**: Use testing checklists from research.md
4. **Deploy**: Follow deployment steps above

## Reference Documentation

- **Specification**: [spec.md](spec.md) - User requirements
- **Research**: [research.md](research.md) - Technical decisions
- **Data Model**: [data-model.md](data-model.md) - Data structures
- **Plan**: [plan.md](plan.md) - Implementation plan

## Support

**Common Commands**:
```bash
make help              # Show all available commands
make server            # Local development
make static-build      # Build for deployment
make deploy            # Full deployment
make optimize-images   # Regenerate image metadata
npm run lighthouse     # Performance audit
```

**Debugging**:
- Check browser console for JavaScript errors
- Inspect elements with DevTools
- Run Lighthouse for performance insights
- Review `research.md` for design decisions
