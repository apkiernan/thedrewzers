# Implementation Plan: Website Performance Optimization

**Branch**: `001-performance-optimization` | **Date**: 2025-10-23 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-performance-optimization/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Optimize the wedding website to achieve 100 Lighthouse and PageSpeed scores through comprehensive performance improvements including image optimization, aggressive caching, resource optimization, and Core Web Vitals compliance. The approach focuses on build-time optimization, maintaining the static-first architecture, and implementing modern web performance best practices without introducing runtime complexity or third-party dependencies.

## Technical Context

**Language/Version**: Go 1.23.3, Node.js (Tailwind CSS 3.4.14)
**Primary Dependencies**:

- Templ 0.2.793 (template engine)
- Tailwind CSS 3.4.14 (styling)
- AWS Lambda Go API Proxy 0.16.1 (Lambda runtime)
- Disintegration Imaging 1.6.2 (current image processing)
- Additional image optimization tools (to be determined in research)

**Storage**: AWS S3 (static assets), CloudFront CDN (distribution), no database
**Testing**: Manual visual testing, Lighthouse CI (to be added), build verification
**Target Platform**: Static website (S3 + CloudFront), AWS Lambda for API routes
**Project Type**: Web application (static-first architecture)
**Performance Goals**:

- Lighthouse Performance: 95+ desktop, 90+ mobile
- LCP ≤2.5s, FID ≤100ms, CLS ≤0.1, TTI <3s
- Initial page load <2s on 4G, <1s on cable
- Total homepage weight <1MB

**Constraints**:

- Must maintain static-first architecture (no SSR)
- Build-time optimization only (no runtime processing)
- AWS-only infrastructure (S3, CloudFront, Lambda)
- No new paid services without justification
- Compatible with existing build process (make commands)

**Scale/Scope**:

- Small wedding website (~10 pages)
- Expected traffic: <1000 visitors/month
- Image-heavy content (wedding photos)
- Mobile-first audience (60%+ mobile traffic)

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

### Principle I: Static-First Architecture ✅ PASS

**Evaluation**: This feature maintains and enhances static-first architecture. All optimizations occur at build time; no runtime HTML generation is introduced.

**Compliance**:

- Image optimization happens during `make static-build`
- CSS optimization happens during `npm run build`
- HTML remains pre-generated in `dist/`
- No changes to Lambda runtime behavior

### Principle II: Build-Time Optimization ✅ PASS

**Evaluation**: Performance optimization is inherently a build-time concern. All processing happens during build phase.

**Compliance**:

- Image processing integrated into build pipeline
- CSS purging and minification at build time
- Font subsetting at build time
- Critical CSS extraction at build time
- No runtime overhead introduced

### Principle III: Simplicity & Pragmatism ✅ PASS

**Evaluation**: Solution adds build tooling but no runtime complexity. Dependencies are minimal and purpose-specific.

**Compliance**:

- No frameworks or abstractions added
- Image optimization uses standalone Go tools or CLI utilities
- No ORM, state management, or complex patterns
- Each tool solves a specific, documented problem
- Remains maintainable by single developer

**New Dependencies (justified)**:

- Image optimization library (replaces manual optimization)
- CSS critical path extractor (measurable performance gain)
- All tools are build-time only, zero runtime impact

### Principle IV: Infrastructure as Code ✅ PASS

**Evaluation**: CloudFront and S3 configuration changes will be managed via Terraform.

**Compliance**:

- Cache header changes → `terraform/`
- CloudFront behaviors for compression → `terraform/`
- S3 bucket policies → `terraform/`
- Use `make tf-plan` before applying changes

### Principle V: Performance & Cost Efficiency ✅ PASS

**Evaluation**: This feature directly addresses constitutional performance requirements.

**Compliance**:

- Target: <2s load on 3G (constitutional requirement)
- Cost impact: Minimal or zero (reduces bandwidth via optimization)
- CloudFront caching reduces S3 requests (cost savings)
- No additional AWS services required

**Constitutional Alignment**: Feature strengthens constitutional compliance by achieving the mandated <2s load time on 3G.

### Summary

**Result**: ✅ ALL GATES PASSED

No constitutional violations. Feature is fully aligned with all core principles and actively strengthens constitutional compliance by achieving mandated performance targets.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Existing Project Structure (Go + Templ + Tailwind)
cmd/
├── main/          # Local dev server entry point
├── lambda/        # AWS Lambda entry point
└── build/         # Static site generator

internal/
├── handlers/      # HTTP request handlers
└── views/         # Templ templates (*.templ files)

static/
├── css/           # Generated Tailwind CSS
├── js/            # Client-side JavaScript
├── fonts/         # Custom fonts (to be optimized)
└── images/        # Source images (to be optimized)

dist/              # Generated static site output
└── static/        # Optimized static assets

src/
└── input.css      # Tailwind source CSS

terraform/         # Infrastructure as Code
└── *.tf           # CloudFront, S3, Lambda config

# New directories for this feature
build/
├── optimize-images.go    # Image optimization script
├── generate-responsive.go # Responsive image generator
└── extract-critical-css.go # Critical CSS extractor

.lighthouse/       # Lighthouse CI configuration
└── lighthouserc.json
```

**Structure Decision**: This is a web application using Go's static-first architecture. Performance optimization will be integrated into the existing build pipeline:

1. **Build scripts** in `build/` directory for image optimization and critical CSS extraction
2. **Static assets** remain in `static/` (source) and `dist/static/` (optimized)
3. **Templ templates** in `internal/views/` will be updated to reference optimized assets
4. **Terraform** configuration for CloudFront/S3 cache headers
5. **No new runtime code** - all optimizations happen at build time via `make` commands

## Complexity Tracking

**No violations** - all constitutional checks passed. No complexity justification required.

---

## Phase 1 Post-Design Constitution Re-Check

**Re-evaluation Date**: 2025-10-23
**Status**: ✅ ALL GATES STILL PASS

### Design Artifacts Review

**Created Documents**:

- `research.md` - Technology decisions and best practices
- `data-model.md` - Marked as N/A (no data entities)
- `contracts/README.md` - Marked as N/A (no API changes)
- `quickstart.md` - Implementation guide

**New Build Scripts Planned**:

- `cmd/optimize-images/main.go` - Go program for responsive image generation
- `cmd/generate-lqip/main.go` - Go program for LQIP placeholders
- Updated `Makefile` with `optimize-images` target

**CLI Tools (System-Level)**:

- `cwebp`, `avifenc` (image format conversion)
- `glyphhanger`, `pyftsubset` (font optimization)

### Constitutional Re-Evaluation

#### Principle I: Static-First Architecture ✅ PASS

**No changes** - Design confirms all optimization at build time, static HTML delivery maintained

#### Principle II: Build-Time Optimization ✅ PASS

**Confirmed** - All Go programs and CLI tools run during `make static-build`, zero runtime processing

#### Principle III: Simplicity & Pragmatism ✅ PASS

**Maintained Simplicity**:

- Rejected critical CSS extraction (too complex for benefit)
- Chose CLI tools over CGO/C library dependencies
- Simple Go programs using existing `imaging` library
- No new frameworks or abstractions

**Dependency Additions (Justified)**:

- System CLI tools (webp, libavif) - Industry standard, one-time install
- Font tools (glyphhanger, pyftsubset) - One-time setup for font processing
- All tools are build-time only, simple to understand and maintain

#### Principle IV: Infrastructure as Code ✅ PASS

**Terraform Updates Planned**:

- CloudFront cache behaviors for different asset types
- S3 bucket CORS configuration for fonts
- All changes will be codified in `terraform/cloudfront.tf`

#### Principle V: Performance & Cost Efficiency ✅ PASS

**Performance Alignment**:

- Designed to achieve <2s load on 3G (constitutional requirement)
- Expected 60-85% page weight reduction
- 70% font payload reduction

**Cost Impact**:

- Reduced CloudFront bandwidth (smaller files)
- Reduced S3 requests (better caching)
- No additional AWS services
- **Net effect**: Cost reduction

### Final Verdict: ✅ NO CONSTITUTIONAL VIOLATIONS

Design phase has **strengthened** constitutional alignment:

- Simplicity maintained by rejecting overly complex solutions
- Build-time optimization principle followed rigorously
- Static-first architecture preserved
- Cost efficiency improved through optimization

**Ready to proceed to Phase 2 (Tasks Generation)**
