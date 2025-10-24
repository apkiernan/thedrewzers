# Image Optimization Workflow

**Feature**: 001-performance-optimization
**Date**: October 2024
**Status**: Complete

## Overview

Automated image optimization pipeline generating **210 optimized files** from 21 source images:
- 63 responsive JPEG sizes (640w, 768w, 1024w per image)
- 63 AVIF versions (best compression for modern browsers)
- 63 WebP versions (good compression for older browsers)
- 21 LQIP placeholders (Low Quality Image Placeholders)

## Build Pipeline

### Command
```bash
make optimize-images
```

### Process Flow
1. **Go program** (`cmd/optimize-images/main.go`): Generate responsive JPEG sizes
2. **Go program** (`cmd/generate-lqip/main.go`): Generate tiny blurred JPEG placeholders
3. **cwebp**: Convert JPEG → WebP format
4. **avifenc**: Convert JPEG → AVIF format

## Responsive Image Generation

### Tool: `cmd/optimize-images/main.go`

**Algorithm**: Lanczos resampling (highest quality)

```go
// Example implementation
imaging.Resize(img, targetWidth, 0, imaging.Lanczos)
```

### Output Sizes
- `640w`: Mobile portrait, small screens
- `768w`: Mobile landscape, tablets
- `1024w`: Desktop, default size for most views
- `1280w`: Large desktop displays (optional)
- `1920w`: Full HD displays (optional)
- `2560w`: 4K/retina displays (optional)

**Current implementation**: 640w, 768w, 1024w for optimal balance of quality and file count.

### File Naming Convention
```
{original-name}-{width}w.{extension}

Examples:
hero-01-640w.jpg
hero-01-768w.jpg
hero-01-1024w.jpg
molly_andrewENG-1-640w.jpg
```

### Source Directory
```
static/images/
├── slideshow/
│   ├── hero-01.jpg
│   ├── hero-02.jpg
│   └── ...
├── molly_andrewENG-1.jpg
├── molly_andrewENG-2.jpg
└── ...
```

### Output Directory
```
dist/images/
├── slideshow/
│   ├── hero-01-640w.jpg (148KB)
│   ├── hero-01-640w.webp (118KB)
│   ├── hero-01-640w.avif (103KB)
│   ├── hero-01-768w.jpg (204KB)
│   ├── hero-01-768w.webp (163KB)
│   ├── hero-01-768w.avif (139KB)
│   ├── hero-01-1024w.jpg (314KB)
│   ├── hero-01-1024w.webp (242KB)
│   ├── hero-01-1024w.avif (204KB)
│   └── hero-01-lqip.jpg (651 bytes)
├── molly_andrewENG-1-640w.jpg
├── molly_andrewENG-1-640w.webp
├── molly_andrewENG-1-640w.avif
└── ...
```

## Format Conversion

### WebP Generation

**Tool**: `cwebp` (libwebp 1.6.0)

```bash
cwebp -q 85 -preset photo -m 6 input.jpg -o output.webp
```

**Parameters**:
- `-q 85`: Quality setting (0-100, higher = better quality, larger file)
- `-preset photo`: Optimization preset for photographic images
- `-m 6`: Compression method (0-6, higher = slower but smaller)

**Typical Results**:
- 25-35% smaller than JPEG at equivalent quality
- Supported by Chrome, Edge, Firefox, Safari 14+

### AVIF Generation

**Tool**: `avifenc` (libavif 1.3.0)

```bash
avifenc --min 0 --max 63 -a end-usage=q -a cq-level=18 -a tune=ssim --jobs 8 input.jpg output.avif
```

**Parameters**:
- `--min 0 --max 63`: Quantizer range (lower = higher quality)
- `-a end-usage=q`: Constant quality mode
- `-a cq-level=18`: Quality level (0-63, lower = higher quality)
- `-a tune=ssim`: Tune for SSIM (structural similarity) metric
- `--jobs 8`: Parallel encoding threads

**Typical Results**:
- 40-50% smaller than JPEG at equivalent quality
- Supported by Chrome 85+, Firefox 93+, Safari 16.4+

### Format Comparison

Example (hero-01-1024w):

| Format | Size | Reduction | Browser Support |
|--------|------|-----------|-----------------|
| JPEG | 314KB | baseline | Universal |
| WebP | 242KB | -23% | Modern browsers (Safari 14+) |
| AVIF | 204KB | -35% | Latest browsers (Safari 16.4+) |

## LQIP (Low Quality Image Placeholder)

### Tool: `cmd/generate-lqip/main.go`

**Purpose**: Instant perceived loading with tiny blurred placeholder

```go
// Resize to 20px width, maintain aspect ratio
thumbnail := imaging.Resize(img, 20, 0, imaging.Lanczos)

// Apply Gaussian blur (sigma: 8.0)
blurred := imaging.Blur(thumbnail, 8.0)

// Encode as JPEG quality 60
imaging.Save(blurred, outputPath, imaging.JPEGQuality(60))
```

**Results**:
- File size: ~650 bytes (0.65KB)
- Dimensions: 20px wide (aspect ratio preserved)
- Blur: Heavy (sigma 8.0)
- Loading time: <10ms on any connection

**Usage**:
1. LQIP loads instantly (base64 encoded or separate request)
2. Placeholder displayed with blur effect
3. Full-resolution image loads in background
4. Smooth transition replaces LQIP with sharp image

**Note**: Current implementation generates LQIP files but templates don't actively use blur-up technique yet. Files are available for future enhancement.

## Template Integration

### Picture Element with Format Negotiation

```html
<picture>
  <!-- AVIF: Best compression, latest browsers -->
  <source type="image/avif"
          srcset="/static/images/hero-01-640w.avif 640w,
                  /static/images/hero-01-768w.avif 768w,
                  /static/images/hero-01-1024w.avif 1024w"
          sizes="(max-width: 768px) 100vw, (max-width: 1024px) 90vw, 1024px"/>

  <!-- WebP: Good compression, wider support -->
  <source type="image/webp"
          srcset="/static/images/hero-01-640w.webp 640w,
                  /static/images/hero-01-768w.webp 768w,
                  /static/images/hero-01-1024w.webp 1024w"
          sizes="(max-width: 768px) 100vw, (max-width: 1024px) 90vw, 1024px"/>

  <!-- JPEG: Universal fallback -->
  <img src="/static/images/hero-01-1024w.jpg"
       srcset="/static/images/hero-01-640w.jpg 640w,
               /static/images/hero-01-768w.jpg 768w,
               /static/images/hero-01-1024w.jpg 1024w"
       sizes="(max-width: 768px) 100vw, (max-width: 1024px) 90vw, 1024px"
       alt="Hero image"
       width="1024"
       height="1535"
       loading="lazy"/>
</picture>
```

### Browser Selection Logic

Browser automatically selects the best format and size:

1. **Format**: Checks `<source>` elements top-to-bottom
   - Modern browser (AVIF support) → Uses AVIF source
   - Older modern browser (WebP support) → Uses WebP source
   - Fallback → Uses JPEG from `<img>` tag

2. **Size**: Based on `srcset` and `sizes` attributes
   - Viewport width ≤768px → Downloads 640w image
   - Viewport width ≤1024px → Downloads 768w image
   - Viewport width >1024px → Downloads 1024w image
   - DPR (Device Pixel Ratio) also considered

### Preventing Layout Shift

**Always specify dimensions**:

```html
<img width="1024" height="1535" ... />
```

**CSS aspect ratio** (Tailwind):

```css
.responsive-image {
  aspect-ratio: 2 / 3;
  object-fit: cover;
}
```

**Result**: Zero layout shift (CLS = 0)

## Lazy Loading

### Above-the-fold images
```html
<img loading="eager" ... />  <!-- First hero image -->
```

### Below-the-fold images
```html
<img loading="lazy" ... />  <!-- Gallery images, lower slides -->
```

**Behavior**:
- `eager`: Load immediately (default)
- `lazy`: Load when near viewport (~1000px threshold)

**Implementation**: `internal/views/gallery.templ` uses lazy loading for all gallery images.

## Cache Configuration

### S3 Upload (Makefile)
```bash
aws s3 sync dist/images s3://$(S3_BUCKET)/static/images/ \
  --acl public-read \
  --cache-control "public, max-age=31536000, immutable"
```

### CloudFront (terraform/cloudfront.tf)
- Path: `/static/*`
- Default TTL: 1 week
- Max TTL: 1 year
- Compression: Enabled

**Result**: Images cached for 1 year, return visitors load from browser cache instantly.

## Performance Impact

### File Size Reduction

Example (21 gallery images at 1024w):

| Metric | Before | After | Reduction |
|--------|--------|-------|-----------|
| Format | JPEG only | AVIF/WebP/JPEG | - |
| Average size (JPEG) | ~350KB | ~314KB | -10% (resize) |
| Average size (WebP) | - | ~242KB | -31% vs JPEG |
| Average size (AVIF) | - | ~204KB | -42% vs JPEG |
| **Total (21 images)** | **7.35MB** | **4.28MB (AVIF)** | **-42%** |

### Lighthouse Impact

- **LCP (Largest Contentful Paint)**: Improved from 4.5s → 1.2s (local) / 3.3s (production)
- **Total page weight**: Reduced from 7.5MB → <1MB (mobile, with AVIF)
- **CLS (Cumulative Layout Shift)**: 0 (zero layout shift due to width/height attributes)
- **Mobile Performance**: 86% (production median)

## Deployment

Images are deployed as part of static site deployment:

```bash
make static-deploy
```

This command:
1. Runs `make optimize-images` (generates all variants)
2. Builds static HTML files
3. Syncs images to S3 with immutable cache headers
4. Invalidates CloudFront cache

## Validation

### Check Image Loading
```bash
# Test format negotiation
curl -I -H "Accept: image/avif" \
  https://thekiernanwedding.com/static/images/molly_andrewENG-1-1024w.avif

# Expected headers:
# content-type: image/avif
# cache-control: public, max-age=31536000, immutable
# x-cache: Hit from cloudfront (after first request)
```

### Chrome DevTools
1. Open Network tab → Filter by "Img"
2. Refresh page
3. Verify:
   - Modern browsers load AVIF format
   - Correct size loaded based on viewport (e.g., 640w on mobile)
   - Cache status: "from disk cache" (return visits)

### Lighthouse
```bash
npm run lighthouse
```

Check "Serve images in next-gen formats" audit should pass.

## Troubleshooting

### Images not loading
- Verify optimized images exist in `dist/images/`
- Check templates reference correct paths (`/static/images/...`)
- Ensure `make optimize-images` ran successfully

### Wrong format loading
- Check browser support (AVIF requires Chrome 85+, Safari 16.4+)
- Verify `<source>` elements have correct `type` attribute
- Test with browser DevTools → Network tab → "Type" column

### Layout shift (CLS > 0)
- Add `width` and `height` attributes to all `<img>` tags
- Ensure dimensions match actual image aspect ratio
- Use CSS `aspect-ratio` property for responsive sizing

### Images too large (slow loading)
- Reduce quality settings: `cwebp -q 80` (instead of 85)
- Reduce AVIF quality: `-a cq-level=20` (instead of 18)
- Generate smaller sizes: Add 480w or reduce 1024w → 800w

## Future Enhancements

### LQIP Blur-Up Implementation
Currently LQIP files are generated but not actively used. To implement:

1. **Base64 encode LQIP** in template data:
```go
lqipBase64 := base64.StdEncoding.EncodeToString(lqipBytes)
```

2. **Add blur transition CSS**:
```css
img.blur-load {
  filter: blur(20px);
  transition: filter 0.3s;
}
img.blur-load.loaded {
  filter: blur(0);
}
```

3. **JavaScript to swap**:
```javascript
img.onload = function() {
  img.classList.add('loaded');
};
```

### Additional Optimizations
- **Art direction**: Different crops for mobile vs desktop using media queries
- **Additional sizes**: Add 480w for small mobiles, 1280w for large desktops
- **Progressive JPEG**: Use for baseline JPEG fallback
- **Smart crop**: Detect faces/subjects and crop intelligently

## References

- [Responsive images guide](https://developer.mozilla.org/en-US/docs/Learn/HTML/Multimedia_and_embedding/Responsive_images)
- [WebP documentation](https://developers.google.com/speed/webp)
- [AVIF specification](https://aomediacodec.github.io/av1-avif/)
- [Image optimization best practices](https://web.dev/fast/#optimize-your-images)
