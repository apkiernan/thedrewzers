# Quick Start Guide: Performance Optimization

**Feature**: Website Performance Optimization
**Branch**: 001-performance-optimization
**Date**: 2025-10-23

---

## Overview

This guide provides step-by-step instructions for implementing the performance optimizations identified in the research phase. Follow these phases in order for best results.

**Estimated Total Time**: 3-5 weeks (can be done incrementally)

**Expected Outcome**:
- Lighthouse Performance: 90-95+ (mobile), 95-100 (desktop)
- Page weight reduction: 60-85%
- Load time: <2s on 4G, <1s on cable
- Core Web Vitals: All metrics in "Good" range

---

## Prerequisites

### System Requirements

**Operating System**: macOS, Linux, or WSL on Windows

**Tools to Install**:
```bash
# Image optimization tools
brew install webp libavif mozjpeg  # macOS
apt-get install webp libavif-bin libjpeg-turbo-progs  # Ubuntu/Debian

# Font optimization tools
npm install -g glyphhanger
pip install fonttools brotli

# Verification
cwebp -version
avifenc --version
glyphhanger --version
pyftsubset --help
```

### Development Environment

- Go 1.23.3+ (already installed)
- Node.js 18+ (already installed for Tailwind)
- Make (already installed)
- Git (already installed)

---

## Phase 1: Font Optimization (2-4 hours)

**Impact**: 70% font payload reduction, immediate Lighthouse improvements

### Step 1.1: Subset and Convert Fonts

```bash
# Navigate to fonts directory
cd /Users/drewzer/Projects/thedrewzers/static/fonts

# Subset Bodoni Moda Normal
pyftsubset BodoniModa-Variable.ttf \
  --unicodes="U+0020-007F,U+00A0-00FF" \
  --flavor=woff2 \
  --layout-features='*' \
  --output-file="BodoniModa-Variable.woff2"

# Subset Bodoni Moda Italic
pyftsubset BodoniModa-Italic-Variable.ttf \
  --unicodes="U+0020-007F,U+00A0-00FF" \
  --flavor=woff2 \
  --layout-features='*' \
  --output-file="BodoniModa-Italic-Variable.woff2"

# Subset BonheurRoyale
pyftsubset BonheurRoyale-Regular.ttf \
  --unicodes="U+0020-007F,U+00A0-00FF" \
  --flavor=woff2 \
  --layout-features='*' \
  --output-file="BonheurRoyale-Regular.woff2"

# Verify file sizes (should see 70%+ reduction)
ls -lh *.woff2
```

**Expected Output**:
```
BodoniModa-Variable.woff2         ~30-47KB (was 158KB TTF)
BodoniModa-Italic-Variable.woff2  ~32-51KB (was 171KB TTF)
BonheurRoyale-Regular.woff2       ~26-39KB (was 130KB TTF)
```

### Step 1.2: Update CSS with font-display

Edit `/Users/drewzer/Projects/thedrewzers/src/input.css`:

```css
@font-face {
  font-family: "BonheurRoyale-Regular";
  src: url("../fonts/BonheurRoyale-Regular.woff2") format("woff2");
  font-display: swap;
}

@font-face {
  font-family: "Bodoni Moda";
  src: url("../fonts/BodoniModa-Variable.woff2") format("woff2");
  font-weight: 100 900;
  font-style: normal;
  font-display: swap;
}

@font-face {
  font-family: "Bodoni Moda";
  src: url("../fonts/BodoniModa-Italic-Variable.woff2") format("woff2");
  font-weight: 100 900;
  font-style: italic;
  font-display: swap;
}
```

### Step 1.3: Add Font Preloading

Edit `/Users/drewzer/Projects/thedrewzers/internal/views/app.templ`:

Add before existing `<link rel="stylesheet">` tags:

```html
<!-- Preload critical font -->
<link rel="preload"
      href="/static/fonts/BodoniModa-Variable.woff2"
      as="font"
      type="font/woff2"
      crossorigin>
```

### Step 1.4: Rebuild and Test

```bash
# Rebuild CSS
npm run build

# Test locally
make server

# Visit http://localhost:8080 and verify:
# - Fonts load correctly
# - Network tab shows WOFF2 files (not TTF)
# - Text appears immediately (no invisible text flash)
```

### Step 1.5: Deploy Font Optimizations

```bash
# Build static site
make static-build

# Deploy to S3
make upload-static

# Invalidate CloudFront cache
make invalidate-cache
```

**Verification**:
- Run Lighthouse audit (should see +5-15 points improvement)
- Check font file sizes in Network tab
- Verify no FOIT (Flash of Invisible Text)

---

## Phase 2: Image Optimization (2-4 weeks)

**Impact**: 60-85% page weight reduction, biggest UX improvement

### Step 2.1: Create Image Optimization Scripts

Create `/Users/drewzer/Projects/thedrewzers/cmd/optimize-images/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

var widths = []int{640, 768, 1024, 1280, 1920, 2560}

func main() {
	sourceDir := "static/images"
	distDir := "dist/images"

	// Create output directory
	os.MkdirAll(distDir, 0755)

	// Process all JPG images in source directory
	images, err := filepath.Glob(filepath.Join(sourceDir, "*.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	for _, imagePath := range images {
		if strings.Contains(imagePath, "-lqip.jpg") || strings.Contains(imagePath, "w.jpg") {
			continue // Skip already processed images
		}

		log.Printf("Processing: %s\n", imagePath)
		if err := generateResponsiveSizes(imagePath, distDir); err != nil {
			log.Printf("Error processing %s: %v\n", imagePath, err)
		}
	}
}

func generateResponsiveSizes(inputPath, distDir string) error {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	// Auto-orient based on EXIF
	src = imaging.AutoOriented(src)

	baseFilename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	originalWidth := src.Bounds().Dx()

	for _, width := range widths {
		// Skip if original is smaller
		if originalWidth < width {
			continue
		}

		// Resize using Lanczos filter for quality
		resized := imaging.Resize(src, width, 0, imaging.Lanczos)

		outputPath := filepath.Join(distDir, fmt.Sprintf("%s-%dw.jpg", baseFilename, width))
		if err := imaging.Save(resized, outputPath, imaging.JPEGQuality(85)); err != nil {
			return err
		}

		log.Printf("  Generated: %s\n", outputPath)
	}

	return nil
}
```

### Step 2.2: Create LQIP Generation Script

Create `/Users/drewzer/Projects/thedrewzers/cmd/generate-lqip/main.go`:

```go
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

func main() {
	sourceDir := "static/images"
	distDir := "dist/images"

	// Process all JPG images
	images, err := filepath.Glob(filepath.Join(sourceDir, "*.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	for _, imagePath := range images {
		if strings.Contains(imagePath, "-lqip.jpg") || strings.Contains(imagePath, "w.jpg") {
			continue
		}

		log.Printf("Generating LQIP for: %s\n", imagePath)
		if err := generateLQIP(imagePath, distDir); err != nil {
			log.Printf("Error: %v\n", err)
		}
	}
}

func generateLQIP(inputPath, distDir string) error {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	src = imaging.AutoOriented(src)

	// Resize to 20px wide (maintains aspect ratio)
	tiny := imaging.Resize(src, 20, 0, imaging.Box)

	// Apply blur
	blurred := imaging.Blur(tiny, 2.0)

	baseFilename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(distDir, baseFilename+"-lqip.jpg")

	// Save with very low quality
	if err := imaging.Save(blurred, outputPath, imaging.JPEGQuality(20)); err != nil {
		return err
	}

	log.Printf("  Generated LQIP: %s\n", outputPath)
	return nil
}
```

### Step 2.3: Update Makefile

Edit `/Users/drewzer/Projects/thedrewzers/Makefile`, add:

```makefile
# Image optimization target
.PHONY: optimize-images
SOURCE_IMAGES=static/images
DIST_IMAGES=dist/images

optimize-images:
	@echo "Creating output directory..."
	@mkdir -p $(DIST_IMAGES)
	@echo "Generating responsive image sizes..."
	@go run cmd/optimize-images/main.go
	@echo "Converting to WebP format..."
	@find $(DIST_IMAGES) -name "*-[0-9]*w.jpg" ! -name "*-lqip.jpg" -exec sh -c \
		'cwebp -q 85 -preset photo -m 6 "$$1" -o "$${1%.jpg}.webp"' sh {} \;
	@echo "Converting to AVIF format..."
	@find $(DIST_IMAGES) -name "*-[0-9]*w.jpg" ! -name "*-lqip.jpg" -exec sh -c \
		'avifenc --min 0 --max 63 -a end-usage=q -a cq-level=18 -a tune=ssim --jobs 8 "$$1" "$${1%.jpg}.avif"' sh {} \;
	@echo "Generating LQIP placeholders..."
	@go run cmd/generate-lqip/main.go
	@echo "Image optimization complete!"

# Update static-build to include image optimization
static-build: tpl optimize-images
	@echo "Building static site..."
	@go run ./cmd/build
	@echo "Copying static assets to dist..."
	@cp -r static/css dist/
	@cp -r static/js dist/
	@cp -r static/fonts dist/
	@echo "Static site ready in ./dist/"
```

### Step 2.4: Update Templates for Responsive Images

Edit image templates (e.g., `/Users/drewzer/Projects/thedrewzers/internal/views/gallery.templ`):

Replace:
```html
<img src="/static/images/photo.jpg" alt="Wedding photo">
```

With:
```html
<picture>
  <source type="image/avif"
          srcset="/static/images/photo-640w.avif 640w,
                  /static/images/photo-1280w.avif 1280w,
                  /static/images/photo-1920w.avif 1920w"
          sizes="(max-width: 768px) 100vw, 50vw">
  <source type="image/webp"
          srcset="/static/images/photo-640w.webp 640w,
                  /static/images/photo-1280w.webp 1280w,
                  /static/images/photo-1920w.webp 1920w"
          sizes="(max-width: 768px) 100vw, 50vw">
  <img src="/static/images/photo-lqip.jpg"
       srcset="/static/images/photo-640w.jpg 640w,
               /static/images/photo-1280w.jpg 1280w,
               /static/images/photo-1920w.jpg 1920w"
       sizes="(max-width: 768px) 100vw, 50vw"
       loading="lazy"
       alt="Wedding photo"
       style="filter: blur(20px); transition: filter 0.3s;"
       onload="this.style.filter='none'">
</picture>
```

### Step 2.5: Test Image Optimization

```bash
# Run image optimization
make optimize-images

# Verify output
ls -lh dist/images/

# Should see files like:
# photo-640w.jpg, photo-640w.webp, photo-640w.avif
# photo-1280w.jpg, photo-1280w.webp, photo-1280w.avif
# photo-lqip.jpg

# Build and test
make static-build
make server

# Verify in browser:
# - Images load progressively (blur → sharp)
# - Modern browsers load AVIF
# - Network tab shows smaller file sizes
```

### Step 2.6: Deploy Image Optimizations

```bash
make deploy  # Runs static-build + upload-static + invalidate-cache
```

**Verification**:
- Run Lighthouse audit (should see major improvements in LCP, page weight)
- Check Network tab for AVIF/WebP downloads
- Verify CLS (Cumulative Layout Shift) < 0.1

---

## Phase 3: CloudFront Caching Optimization (4-8 hours)

**Impact**: Better cache hit rates, reduced latency

### Step 3.1: Update Terraform Configuration

Edit `/Users/drewzer/Projects/thedrewzers/terraform/cloudfront.tf`:

```hcl
# Add cache behavior for static assets
ordered_cache_behavior {
  path_pattern     = "/static/*"
  target_origin_id = aws_s3_bucket.static_website.id

  allowed_methods  = ["GET", "HEAD", "OPTIONS"]
  cached_methods   = ["GET", "HEAD"]

  forwarded_values {
    query_string = false
    cookies {
      forward = "none"
    }
    headers = ["Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"]
  }

  min_ttl                = 0
  default_ttl            = 31536000  # 1 year
  max_ttl                = 31536000
  compress               = true
  viewer_protocol_policy = "redirect-to-https"
}

# HTML pages should have shorter cache with revalidation
ordered_cache_behavior {
  path_pattern     = "/*.html"
  target_origin_id = aws_s3_bucket.static_website.id

  allowed_methods  = ["GET", "HEAD"]
  cached_methods   = ["GET", "HEAD"]

  forwarded_values {
    query_string = false
    cookies {
      forward = "none"
    }
  }

  min_ttl                = 0
  default_ttl            = 86400  # 1 day
  max_ttl                = 604800  # 1 week
  compress               = true
  viewer_protocol_policy = "redirect-to-https"
}
```

### Step 3.2: Configure S3 Metadata

Update build script to set proper Content-Type headers:

```bash
# In Makefile or deployment script
aws s3 sync dist/ s3://thedrewzers-wedding-static/ \
  --exclude "*" \
  --include "*.avif" \
  --content-type "image/avif" \
  --cache-control "public, max-age=31536000, immutable"

aws s3 sync dist/ s3://thedrewzers-wedding-static/ \
  --exclude "*" \
  --include "*.webp" \
  --content-type "image/webp" \
  --cache-control "public, max-age=31536000, immutable"

aws s3 sync dist/ s3://thedrewzers-wedding-static/ \
  --exclude "*" \
  --include "*.woff2" \
  --content-type "font/woff2" \
  --cache-control "public, max-age=31536000, immutable"
```

### Step 3.3: Apply Terraform Changes

```bash
cd terraform/

# Review changes
make tf-plan

# Apply if looks good
make tf-apply

# Verify CloudFront distribution updated
aws cloudfront list-distributions
```

---

## Phase 4: Performance Monitoring (2-4 hours)

**Impact**: Ongoing visibility into performance

### Step 4.1: Set Up Lighthouse CI

Create `/.lighthouse/lighthouserc.json`:

```json
{
  "ci": {
    "collect": {
      "url": ["http://localhost:8080/", "http://localhost:8080/venue.html"],
      "numberOfRuns": 3
    },
    "assert": {
      "preset": "lighthouse:recommended",
      "assertions": {
        "categories:performance": ["error", {"minScore": 0.9}],
        "categories:accessibility": ["error", {"minScore": 1}],
        "categories:best-practices": ["error", {"minScore": 1}],
        "categories:seo": ["error", {"minScore": 1}],
        "first-contentful-paint": ["error", {"maxNumericValue": 2000}],
        "largest-contentful-paint": ["error", {"maxNumericValue": 2500}],
        "cumulative-layout-shift": ["error", {"maxNumericValue": 0.1}]
      }
    }
  }
}
```

### Step 4.2: Add Lighthouse CI to Build

```bash
# Install Lighthouse CI
npm install -g @lhci/cli

# Add to package.json scripts
"scripts": {
  "build": "tailwindcss -i ./src/input.css -o ./static/css/tailwind.css --minify",
  "watch": "tailwindcss -i ./src/input.css -o ./static/css/tailwind.css --watch",
  "lighthouse": "lhci autorun"
}
```

### Step 4.3: Run Performance Audit

```bash
# Start local server
make server &

# Run Lighthouse CI
npm run lighthouse

# Review results
# Should see all scores in green (90-100)
```

---

## Testing Checklist

### Manual Testing

- [ ] Fonts load correctly on all pages
- [ ] No invisible text flash (FOIT)
- [ ] Images load progressively (blur → sharp transition)
- [ ] Mobile devices load appropriately sized images
- [ ] Network tab shows AVIF/WebP for modern browsers
- [ ] Page loads fast on throttled 3G connection

### Automated Testing

- [ ] Lighthouse Performance: 90+ (mobile), 95+ (desktop)
- [ ] Lighthouse Best Practices: 100
- [ ] Lighthouse Accessibility: 100
- [ ] Lighthouse SEO: 100
- [ ] LCP < 2.5s
- [ ] FID < 100ms
- [ ] CLS < 0.1
- [ ] TTI < 3s

### Cross-Browser Testing

- [ ] Chrome (latest)
- [ ] Safari (latest)
- [ ] Firefox (latest)
- [ ] Edge (latest)
- [ ] Mobile Safari (iOS)
- [ ] Mobile Chrome (Android)

---

## Troubleshooting

### Issue: Fonts not loading

**Symptoms**: Fallback fonts displayed, WOFF2 files show 404
**Solution**: Verify font paths in CSS match actual file locations

```bash
# Check fonts exist
ls -la static/fonts/*.woff2

# Verify CSS references match
grep "url(" src/input.css
```

### Issue: Images still large

**Symptoms**: Network tab shows large file sizes, AVIF not loading
**Solution**: Verify CLI tools ran successfully, check browser support

```bash
# Re-run image optimization with verbose output
make optimize-images

# Check if AVIF files were created
ls -la dist/images/*.avif

# Test AVIF support in browser
# Open DevTools → Network → Check image responses
```

### Issue: CloudFront not caching

**Symptoms**: Every request hits origin, cache-control headers missing
**Solution**: Check CloudFront behaviors, verify S3 metadata

```bash
# Check CloudFront cache statistics
aws cloudfront get-distribution --id YOUR_DISTRIBUTION_ID

# Verify S3 object metadata
aws s3api head-object --bucket thedrewzers-wedding-static --key static/fonts/BodoniModa-Variable.woff2
```

### Issue: Layout shifts during loading

**Symptoms**: CLS score high, content jumps as images load
**Solution**: Ensure image dimensions specified, LQIP properly sized

```html
<!-- Add explicit width/height attributes -->
<img src="..." width="1280" height="1920" loading="lazy">
```

---

## Performance Targets Summary

| Metric | Target | How to Measure |
|--------|--------|----------------|
| **Lighthouse Performance (Mobile)** | 90+ | Chrome DevTools → Lighthouse |
| **Lighthouse Performance (Desktop)** | 95+ | Chrome DevTools → Lighthouse |
| **LCP (Largest Contentful Paint)** | ≤2.5s | Lighthouse → Core Web Vitals |
| **FID (First Input Delay)** | ≤100ms | Real User Monitoring |
| **CLS (Cumulative Layout Shift)** | ≤0.1 | Lighthouse → Core Web Vitals |
| **TTI (Time to Interactive)** | <3s | Lighthouse → Performance |
| **Page Weight (Homepage)** | <1MB | Network tab → Total size |
| **Font Payload** | <150KB | Network tab → Fonts |
| **Image Payload (Mobile)** | <500KB | Network tab → Images (throttled to 4G) |

---

## Next Steps

After completing all phases:

1. **Monitor Performance**: Set up weekly Lighthouse CI runs
2. **Establish Baseline**: Document initial vs final performance metrics
3. **Performance Budget**: Create alerts if metrics regress
4. **Documentation**: Update CLAUDE.md with new build commands
5. **Team Training**: Document optimization workflow for future developers

---

## Additional Resources

- **Research Document**: `specs/001-performance-optimization/research.md`
- **Implementation Plan**: `specs/001-performance-optimization/plan.md`
- **Specification**: `specs/001-performance-optimization/spec.md`
- **Project Constitution**: `.specify/memory/constitution.md`

---

**Document Version**: 1.0
**Last Updated**: 2025-10-23
**Estimated Completion**: 3-5 weeks (incremental implementation)
