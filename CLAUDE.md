# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Wedding website built with Go, Templ templating, and Tailwind CSS. Deployable to AWS Lambda with static assets served from S3.

## Key Commands

### Development
- `make server` - Run local dev server with hot reload on port 8080
- `make all` - Build templates, compile CSS, and start server
- `make tpl` - Generate Go code from templ templates
- `make styles` - Watch and compile Tailwind CSS

### Static Site Generation
- `make static-build` - Generate static HTML files in ./dist/
- `make static-deploy` - Upload static site to S3 and invalidate CloudFront cache
- `make invalidate-cache` - Manually invalidate CloudFront cache

### Deployment
- `make deploy` - Build Lambda, generate static site, and deploy everything
- `make lambda-build` - Build Lambda binary (linux/amd64)
- `make upload-static` - Upload only static assets to S3
- `make tf-plan` - Preview Terraform changes
- `make tf-apply` - Apply Terraform changes

### Performance Optimization
- `make optimize-images` - Generate responsive image sizes and modern formats (AVIF/WebP/JPEG)
- `npm run lighthouse` - Run Lighthouse CI performance audit
- `npm run build` - Build production CSS with Tailwind minification

### Other Commands
- `make clean` - Remove build artifacts
- `npm run watch` - Watch CSS changes

## Architecture

### Entry Points
- **Local Dev**: `cmd/main/main.go` - Runs on port 8080 with static file serving
- **Lambda**: `cmd/lambda/main.go` - AWS Lambda handler for API routes only (`/api/*`)
- **Static Build**: `cmd/build/main.go` - Generates static HTML files from templates

### Core Structure
- **Handlers**: `internal/handlers/` - HTTP request handlers (homepage.go, venue.go)
- **Templates**: `internal/views/*.templ` - Templ components compiled to Go
- **Static Assets**: `static/` - CSS, fonts, images served locally or from S3
- **Generated Site**: `dist/` - Static HTML output from build process

### Key Patterns
1. **Static-First Architecture**: 
   - HTML pages served directly from S3 via CloudFront
   - Lambda only handles API requests (`/api/*`)
   - No Lambda cold starts for page loads
2. **Templ Templates**: Type-safe HTML generation at build time
3. **Client-Side State**: First-time visitor overlay managed via localStorage
4. **API Configuration**: `static/js/config.js` defines API endpoint

### AWS Infrastructure (Terraform)
- **S3 Bucket**: Static website hosting for HTML/CSS/JS
- **CloudFront**: CDN serving static site, routes `/api/*` to Lambda
- **Lambda Function**: Handles API requests only (future RSVP functionality)
- **API Gateway**: HTTP API for Lambda integration

### Development Flow
1. `air` watches .go, .templ, .css files
2. Auto-runs `templ generate` and rebuilds on changes
3. Templates compile .templ → _templ.go files
4. Tailwind processes CSS via npm scripts

## Important Notes
- Lambda binary must be named `bootstrap` for custom runtime
- Static paths in Lambda redirect to S3 bucket
- Default S3 bucket: `thedrewzers-wedding-static`
- Default region: `us-east-1`
- CloudFront invalidation: Set `CLOUDFRONT_DISTRIBUTION_ID` env var or let it auto-detect from Terraform outputs

## Active Technologies
- Go 1.23.3, Node.js (Tailwind CSS 3.4.14) (001-performance-optimization)
- AWS S3 (static assets), CloudFront CDN (distribution), no database (001-performance-optimization)
- Go 1.23.3 + Templ v0.2.793 (templating), Tailwind CSS 3.4.14 (styling) (002-wedding-party-page)
- Static files (Go slice/array for wedding party data in template) (002-wedding-party-page)

## Recent Changes
- 001-performance-optimization: Added Go 1.23.3, Node.js (Tailwind CSS 3.4.14)

## Performance Optimization

### Overview
The site is heavily optimized for performance with automated build-time optimization:
- **Font optimization**: WOFF2 subsets (70% reduction: 460KB → 90KB)
- **Image optimization**: Responsive sizes + modern formats (AVIF/WebP/JPEG)
- **CloudFront caching**: Aggressive caching (1-year for assets, 1-day for HTML)
- **Resource loading**: Preloading critical resources, deferred JavaScript
- **Mobile performance**: 86% Lighthouse score (production median)

### Running Lighthouse Audits
```bash
# Run Lighthouse CI audit (uses .lighthouse/lighthouserc.json config)
npm run lighthouse

# Results include:
# - Performance score (mobile + desktop)
# - Core Web Vitals (LCP, CLS, TBT, FCP, Speed Index)
# - Accessibility, Best Practices, SEO scores
```

### Image Optimization Workflow
```bash
# Optimize all images in static/images/
make optimize-images

# This generates:
# - Responsive JPEG sizes: 640w, 768w, 1024w, 1280w, 1920w, 2560w
# - AVIF versions: Best compression for modern browsers
# - WebP versions: Good compression for older browsers
# - LQIP placeholders: Tiny blurred images for instant loading
# Output: dist/images/
```

**Affected files**: Templates use `<picture>` elements with `srcset` and `sizes` for automatic format/size selection.

### Font Optimization
Fonts are pre-optimized as WOFF2 subsets with Latin charset:
- `static/fonts/optimized/BodoniModa-Variable.woff2` (25KB, -84%)
- `static/fonts/optimized/BodoniModa-Italic-Variable.woff2` (31KB, -82%)
- `static/fonts/optimized/BonheurRoyale-Regular.woff2` (20KB, -85%)

Fonts are preloaded in `internal/views/app.templ` for instant rendering.

### Performance Targets
- **Mobile**: Performance 90+, LCP ≤2.5s, CLS ≤0.1, TBT <200ms
- **Desktop**: Performance 95+, LCP ≤2.0s, CLS ≤0.05, TBT <100ms
- **Page Weight**: <1MB total (mobile), fonts <150KB, CSS ~28KB minified
- **Caching**: 95%+ cache hit rate for return visitors

See `specs/001-performance-optimization/` for detailed documentation.
