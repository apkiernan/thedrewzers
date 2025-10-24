# Performance Optimization Research

**Feature**: Website Performance Optimization
**Branch**: 001-performance-optimization
**Date**: 2025-10-23
**Purpose**: Document technology decisions and best practices for achieving 100 Lighthouse scores

---

## Research Overview

This document consolidates research findings for three major performance optimization areas:

1. **Image Optimization** - Format conversion, responsive sizes, LQIP generation
2. **Critical CSS Extraction** - Evaluation and recommendation
3. **Font Optimization** - Subsetting, format conversion, loading strategy

All decisions prioritize constitutional alignment: simplicity, build-time optimization, and no runtime complexity.

---

## 1. Image Optimization

### Decision: Hybrid CLI + Go Approach

**Rationale**: Use CLI tools (`cwebp`, `avifenc`) for format conversion combined with Go's existing `disintegration/imaging` library for resizing and blur generation. This avoids complex C dependencies while maintaining build-time simplicity.

### 1.1 Image Format Conversion

**Technology Choice**: Command-line tools

| Tool              | Purpose       | Installation           | Quality Settings                     |
| ----------------- | ------------- | ---------------------- | ------------------------------------ |
| `cwebp`           | WebP encoding | `brew install webp`    | `-q 85 -preset photo -m 6`           |
| `avifenc`         | AVIF encoding | `brew install libavif` | `--cq-level 18 -a tune=ssim`         |
| `cjpeg` (mozjpeg) | JPEG fallback | `brew install mozjpeg` | `-quality 85 -optimize -progressive` |

**Alternatives Considered**:

- **h2non/bimg** (Go wrapper for libvips) - Rejected: Requires CGO, C dependencies, complicates Lambda builds
- **golang.org/x/image/webp** - Rejected: Decode-only, no encoding support
- **kolesa-team/go-webp** - Rejected: CGO bindings add complexity

**Expected File Size Reductions**:

- WebP: 25-34% smaller than JPEG
- AVIF: 50% smaller than JPEG
- Total page weight reduction: 60-85% (mobile users)

### 1.2 Responsive Image Sizes (srcset)

**Technology Choice**: Custom Go implementation using existing `disintegration/imaging v1.6.2`

**Recommended Breakpoints**:

```
640w  - Mobile portrait (2x DPR for 320px screens)
768w  - Mobile landscape / Small tablets
1024w - Tablets / Small laptops
1280w - Standard laptops
1920w - Full HD displays
2560w - Retina/4K displays (max for web)
```

**Resize Filter**: `imaging.Lanczos` (or `imaging.CatmullRom` for photos to avoid ringing artifacts)

**Alternatives Considered**:

- **Image processing servers** (imgproxy, picfit) - Rejected: Runtime processing, server overhead, violates simplicity
- **Hugo-style templates** - Rejected: Not using Hugo

### 1.3 LQIP (Low-Quality Image Placeholder)

**Technology Choice**: Simple Blur Approach (Tiny Blurred JPEG)

**Implementation**:

- Resize image to 20px wide using `imaging.Resize`
- Apply blur using `imaging.Blur(tiny, 2.0)`
- Save with JPEG quality 20
- Expected size: 2-5KB per placeholder

**HTML Pattern**:

```html
<img
  src="photo-lqip.jpg"
  data-src="photo-1920w.avif"
  srcset="photo-640w.avif 640w, ..."
  sizes="100vw"
  style="filter: blur(20px); transition: filter 0.3s;"
  onload="this.style.filter='none'; this.src=this.dataset.src"
/>
```

**Alternatives Considered**:

- **BlurHash/ThumbHash** - Rejected: Requires client-side JavaScript decoder (2KB lib), adds complexity
- **Dominant color** - Rejected: Too simple, poor user experience

### 1.4 Build Pipeline Integration

**Makefile Workflow**:

```makefile
optimize-images:
	@echo "Generating responsive image sizes..."
	@go run cmd/optimize-images/main.go
	@echo "Converting to WebP format..."
	@find $(DIST_IMAGES) -name "*-[0-9]*w.jpg" -exec sh -c \
		'cwebp -q 85 -preset photo "$$1" -o "$${1%.jpg}.webp"' sh {} \;
	@echo "Converting to AVIF format..."
	@find $(DIST_IMAGES) -name "*-[0-9]*w.jpg" -exec sh -c \
		'avifenc --cq-level 18 "$$1" "$${1%.jpg}.avif"' sh {} \;
	@echo "Generating LQIP placeholders..."
	@go run cmd/generate-lqip/main.go
```

**Go Programs to Create**:

1. `cmd/optimize-images/main.go` - Generate responsive sizes
2. `cmd/generate-lqip/main.go` - Generate blur placeholders
3. Update `cmd/gallery-metadata/main.go` - Include srcset metadata

### 1.5 Common Gotchas

- **Processing Time**: 25 images × 6 sizes × 3 formats = 450 files (~2-3 minutes)
  - **Mitigation**: Track checksums, only process changed images
- **Disk Space**: ~100-150MB for all formats/sizes
  - **Mitigation**: Use .gitignore for source-images/
- **EXIF Orientation**: Some cameras use EXIF instead of rotating pixels
  - **Mitigation**: Use `imaging.AutoOriented(src)` before processing
- **Color Shift**: AVIF may shift colors if color profile not preserved
  - **Mitigation**: Add `-a color:enable-icc=true` to avifenc

### 1.6 Expected Performance Impact

**Before Optimization**:

- Mobile user (4G): Downloads full 1024×1466 images (~300KB each)
- Total gallery weight: ~7.5MB
- Load time: ~15-20 seconds on 4G

**After Optimization**:

- Mobile user: Downloads 640w AVIF (~40KB each)
- Total gallery weight: ~1MB
- Load time: ~2-3 seconds on 4G
- **~85% reduction in page weight**

---

## 2. Critical CSS Extraction

### Decision: DO NOT Implement Critical CSS

**Rationale**: Current CSS is already optimized at ~8KB gzipped total. Critical CSS extraction provides negligible benefits for files under 20KB and introduces complexity that violates constitutional constraints.

### 2.1 Current State Analysis

**CSS File Sizes**:

- `tailwind.css`: 18KB minified (~4.5KB gzipped)
- `styles.css`: 11KB unminified (custom styles)
- **Total gzipped**: ~7-8KB

**Existing Optimizations**:

- ✅ Tailwind purging active (content: `["./internal/views/*.templ"]`)
- ✅ Minification enabled via `--minify` flag
- ✅ JIT mode (tree-shaking)
- ✅ CloudFront CDN caching

### 2.2 Tools Evaluated

| Tool          | Type               | Pros                  | Cons                    | Complexity |
| ------------- | ------------------ | --------------------- | ----------------------- | ---------- |
| **Critters**  | DOM reconstruction | Fast, lightweight     | Not viewport-aware      | Medium     |
| **Critical**  | Headless browser   | Most accurate         | Puppeteer overhead      | High       |
| **Penthouse** | Headless browser   | Good parallelization  | Manual configuration    | High       |
| **Custom Go** | Parser-based       | No Node.js dependency | High development effort | Very High  |

### 2.3 Performance Analysis

**Caching Penalty for Multi-Page Sites**:

- **With external CSS**: 8KB downloaded once, cached for subsequent pages
- **With inlined critical CSS**: ~6KB inlined per page = 18KB total (non-cacheable)
- **Net difference**: 10KB MORE data transferred with critical CSS

**Expected Improvements**:

- First Contentful Paint (FCP): <50ms improvement (negligible)
- Largest Contentful Paint (LCP): Minimal impact
- **Conclusion**: Critical CSS makes performance WORSE for this use case

### 2.4 Recommended Alternative: CSS Preloading

**Simple Preload Hint** (5-minute implementation):

```html
<link rel="preload" href="/static/css/tailwind.css" as="style" />
<link rel="preload" href="/static/css/styles.css" as="style" />
<link rel="stylesheet" href="/static/css/tailwind.css" />
<link rel="stylesheet" href="/static/css/styles.css" />
```

**Expected Impact**: 20-50ms faster CSS load on cold cache with minimal complexity

### 2.5 Focus on Real Bottlenecks

**Higher Impact Optimizations**:

1. **Image optimization** (60-85% page weight reduction)
2. **Font optimization** (70% font payload reduction)
3. **JavaScript bundling/minification**
4. **CloudFront caching optimization**

**CSS is already a non-issue** - no further optimization needed.

---

## 3. Font Optimization

### Decision: Self-Hosted WOFF2 with Subsetting

**Rationale**: Self-hosting outperforms Google Fonts CDN in 2025 due to cache partitioning. Subsetting reduces font payload by 70-89% while WOFF2 format provides 30% better compression than TTF.

### 3.1 Current State Analysis

**Self-Hosted Fonts** (in `/static/fonts/`):

- `BodoniModa-Variable.ttf` - 158KB (weight 100-900)
- `BodoniModa-Italic-Variable.ttf` - 171KB (weight 100-900)
- `BonheurRoyale-Regular.ttf` - 130KB (decorative script)
- **Total**: ~460KB (uncompressed TTF)

**Issues**:

- ❌ Unoptimized format (TTF instead of WOFF2)
- ❌ No subsetting (includes thousands of unused glyphs)
- ❌ No `font-display` strategy (risk of FOIT)
- ❌ No preloading (critical fonts not prioritized)

### 3.2 Font Subsetting

**Technology Choice**: `glyphhanger` + `pyftsubset`

**Installation**:

```bash
npm install -g glyphhanger
pip install fonttools brotli
```

**Unicode Range for Wedding Website**:

```
U+0020-007F  # Basic Latin (A-Z, a-z, 0-9, punctuation)
U+00A0-00FF  # Latin-1 Supplement (accented characters)
```

**Subsetting Command**:

```bash
pyftsubset BodoniModa-Variable.ttf \
  --unicodes="U+0020-007F,U+00A0-00FF" \
  --flavor=woff2 \
  --layout-features='*' \
  --output-file="BodoniModa-Variable.woff2"
```

**Expected File Size Reductions**:

- BodoniModa-Variable: 158KB → ~30-47KB (70-81% reduction)
- BodoniModa-Italic-Variable: 171KB → ~32-51KB (70-81% reduction)
- BonheurRoyale-Regular: 130KB → ~26-39KB (70-80% reduction)
- **Total**: 460KB → ~90-140KB (WOFF2, subsetted)

**Alternatives Considered**:

- **Online tools** (Font Squirrel, CloudConvert) - Rejected: May break variable font axes
- **Manual WOFF2 conversion only** (no subsetting) - Rejected: Leaves 70% optimization on table

### 3.3 Font-Display Strategy

**Technology Choice**: `font-display: swap`

**Rationale**:

- Ensures text is always visible (no FOIT - Flash of Invisible Text)
- Simple implementation, no complex fallback strategies
- Prevents high Cumulative Layout Shift (CLS) when combined with preloading

**CSS Implementation**:

```css
@font-face {
  font-family: "Bodoni Moda";
  src: url("../fonts/BodoniModa-Variable.woff2") format("woff2");
  font-weight: 100 900;
  font-style: normal;
  font-display: swap;
}
```

**Alternative: `font-display: optional`** - Rejected: Too risky for brand-critical fonts

### 3.4 Font Preloading

**Technology Choice**: Preload critical font with `crossorigin` attribute

**Implementation**:

```html
<link
  rel="preload"
  href="/static/fonts/BodoniModa-Variable.woff2"
  as="font"
  type="font/woff2"
  crossorigin
/>
```

**Critical Rule**: `crossorigin` is **mandatory** even for same-origin fonts (due to CORS requirements)

**Best Practices**:

- ✅ Limit to 1-2 critical fonts maximum
- ✅ Only preload fonts used above-the-fold
- ❌ Don't preload all font variations

### 3.5 Self-Hosted vs Google Fonts CDN

**Decision: Self-Host All Fonts**

**Why Self-Hosting Wins in 2025**:

1. **Cache partitioning eliminated CDN advantage** - Browsers now partition cache by site origin
2. **Fewer network connections** - Single HTTP/2 connection reused
3. **Google's short cache lifetime** - 24-hour CSS cache forces daily re-downloads
4. **Full control over optimization** - Can subset, customize, control `font-display`
5. **Privacy & GDPR compliance** - No third-party requests

**Performance Data**:

- Self-hosting improved Mobile PageSpeed: 66 → 94 (+28 points) in case studies
- Eliminated 2-3 DNS lookups
- Reduced load time by 500-800ms on 4G

### 3.6 Expected Performance Impact

**Before Optimization**:

- Total font size: ~460KB (TTF, unsubsetted)
- Format: TTF (suboptimal for web)
- Loading strategy: Default (risk of FOIT)
- Preloading: None

**After Optimization**:

- Total font size: ~90-140KB (WOFF2, subsetted) - **70% reduction**
- Format: WOFF2 (optimal compression)
- Loading strategy: `font-display: swap` (text always visible)
- Preloading: Critical font preloaded
- External dependencies: None

**Lighthouse Score Impact**:

- Performance: +5-15 points
- Best Practices: +5 points (proper font-display)
- CLS: Improved (preload + swap reduces font swap shift)
- LCP: Improved (critical font loads earlier)

### 3.7 Font Download Time Comparison

**Current State (3G connection at 400kbps)**:

- 460KB @ 400kbps = ~9 seconds

**After Optimization**:

- 120KB @ 400kbps = ~2.4 seconds

**Time Savings**: 6.6 seconds on slow connections

---

## Constitutional Alignment Check

### Principle I: Static-First Architecture ✅

**Image Optimization**: All processing at build time (make optimize-images)
**Critical CSS**: Not implemented (simplicity maintained)
**Font Optimization**: Subsetting and conversion at build time

### Principle II: Build-Time Optimization ✅

**Image Optimization**: CLI tools run during `make static-build`
**Font Optimization**: Subsetting happens once during setup, fonts committed to repo

### Principle III: Simplicity & Pragmatism ✅

**Image Optimization**: No CGO, no C dependencies, clear Makefile integration
**Critical CSS**: Rejected due to complexity without meaningful benefit
**Font Optimization**: Standard tools (glyphhanger, pyftsubset), no abstractions

**New Dependencies (justified)**:

- CLI tools: `cwebp`, `avifenc` (system-level, not Go dependencies)
- Font tools: `glyphhanger`, `pyftsubset` (one-time setup, not runtime)
- All tools are industry-standard, well-documented, actively maintained

### Principle IV: Infrastructure as Code ✅

**CloudFront Configuration**: Cache headers for images/fonts → `terraform/`
**S3 Bucket**: CORS headers for fonts → `terraform/`

### Principle V: Performance & Cost Efficiency ✅

**Performance Impact**:

- Image optimization: 60-85% page weight reduction
- Font optimization: 70% font payload reduction
- Expected Lighthouse Performance: 90-95+ (mobile), 95-100 (desktop)

**Cost Impact**:

- CloudFront bandwidth savings (smaller files)
- Reduced S3 requests (better caching)
- No additional AWS services required

---

## Implementation Priorities

### Phase 1: Font Optimization (Fastest Wins)

**Effort**: 2-4 hours
**Impact**: 70% font payload reduction, immediate Lighthouse improvements

**Steps**:

1. Install tools: `npm install -g glyphhanger`, `pip install fonttools brotli`
2. Subset and convert fonts to WOFF2
3. Update CSS with `font-display: swap`
4. Add preload to template
5. Test and deploy

### Phase 2: Image Optimization (Highest Impact)

**Effort**: 2-4 weeks
**Impact**: 60-85% page weight reduction, biggest user experience improvement

**Steps**:

1. Install CLI tools: `brew install webp libavif`
2. Create `cmd/optimize-images/main.go` for responsive sizes
3. Create `cmd/generate-lqip/main.go` for blur placeholders
4. Update Makefile with `optimize-images` target
5. Update templates to use `<picture>` element
6. Test and deploy

### Phase 3: CloudFront Caching Optimization

**Effort**: 4-8 hours
**Impact**: Better cache hit rates, reduced latency

**Steps**:

1. Update Terraform for optimal cache headers
2. Configure compression (gzip/brotli)
3. Set up cache behaviors for different asset types
4. Test cache performance

### Phase 4: Performance Monitoring

**Effort**: 2-4 hours
**Impact**: Ongoing performance visibility

**Steps**:

1. Set up Lighthouse CI
2. Create performance budgets
3. Monitor Core Web Vitals
4. Document performance baseline and targets

---

## Tools and Dependencies Summary

### CLI Tools (System-Level)

- `cwebp` - WebP encoding (brew install webp)
- `avifenc` - AVIF encoding (brew install libavif)
- `cjpeg` (mozjpeg) - JPEG optimization (brew install mozjpeg)
- `glyphhanger` - Font subsetting (npm install -g glyphhanger)
- `pyftsubset` - Font subsetting (pip install fonttools brotli)

### Go Dependencies (Existing)

- `github.com/disintegration/imaging v1.6.2` - Image resizing, blur

### Node Dependencies (Existing)

- `tailwindcss ^3.4.14` - CSS framework

### No Additional Runtime Dependencies

All tools are build-time only, zero runtime impact on Lambda or static site.

---

## Performance Targets (Success Criteria)

### Lighthouse Scores

- ✅ Desktop Performance: 95+
- ✅ Mobile Performance: 90+
- ✅ Best Practices: 100
- ✅ Accessibility: 100
- ✅ SEO: 100

### Core Web Vitals

- ✅ LCP: ≤2.5s
- ✅ FID: ≤100ms
- ✅ CLS: ≤0.1
- ✅ TTI: <3s

### Load Metrics

- ✅ Homepage weight: <1MB
- ✅ Initial load (4G): <2s
- ✅ Cached load: <0.5s

---

## References

### Image Optimization

- [cwebp Documentation](https://developers.google.com/speed/webp/docs/cwebp)
- [libavif GitHub](https://github.com/AOMediaCodec/libavif)
- [disintegration/imaging](https://github.com/disintegration/imaging)

### CSS Optimization

- [CSS Wizardry: Critical CSS? Not So Fast!](https://csswizardry.com/2022/09/critical-css-not-so-fast/)
- [web.dev: Extract Critical CSS](https://web.dev/articles/extract-critical-css)

### Font Optimization

- [glyphhanger](https://www.npmjs.com/package/glyphhanger)
- [fonttools](https://github.com/fonttools/fonttools)
- [web.dev: Font Best Practices](https://web.dev/articles/font-best-practices)
- [MDN: font-display](https://developer.mozilla.org/en-US/docs/Web/CSS/@font-face/font-display)

---

**Document Version**: 1.0
**Last Updated**: 2025-10-23
**Status**: Research Complete, Ready for Implementation Planning
