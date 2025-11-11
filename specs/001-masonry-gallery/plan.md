# Implementation Plan: Masonry Gallery Layout

**Branch**: `001-masonry-gallery` | **Date**: 2025-11-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-masonry-gallery/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a functional masonry layout for the gallery page that eliminates empty whitespace within image cards. The current implementation uses JavaScript-based absolute positioning with a column-balancing algorithm. This plan will analyze the existing implementation, identify gaps causing whitespace issues, and implement fixes to ensure zero visible whitespace within image cards while maintaining performance and responsiveness.

## Technical Context

**Language/Version**: Go 1.23.3 (backend), JavaScript ES6+ (frontend), Templ v0.2.793 (templating)
**Primary Dependencies**: Tailwind CSS 3.4.14, existing image optimization pipeline (AVIF/WebP/JPEG)
**Storage**: Static JSON metadata file (`static/gallery-metadata.json`) for image dimensions
**Testing**: Manual visual testing, Lighthouse CI performance audits
**Target Platform**: Modern browsers (Chrome 90+, Firefox 88+, Safari 14+, Edge 90+)
**Project Type**: Static web application with pre-generated HTML
**Performance Goals**: Zero Cumulative Layout Shift (CLS = 0), 60fps scrolling, <2s total render time
**Constraints**: No CSS Grid masonry (limited browser support), must work with existing static-first architecture, maintain existing image optimization integration
**Scale/Scope**: 21 gallery images (per metadata), responsive across mobile (320px+), tablet (768px+), desktop (1024px+)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Static-First Architecture ✅ PASS

**Assessment**: Masonry layout is purely client-side enhancement. HTML is pre-generated with all image metadata, JavaScript only repositions elements on the client. No server-side rendering at request time.

**Compliance**:
- Gallery HTML generated at build time by `cmd/build/main.go`
- Image metadata loaded from static JSON file
- JavaScript enhances static HTML (progressive enhancement)
- No Lambda involvement (gallery is static page)

### Principle II: Build-Time Optimization ✅ PASS

**Assessment**: All template compilation and metadata generation happens at build time.

**Compliance**:
- Templ templates compiled to Go (`gallery.templ` → `gallery_templ.go`)
- Image metadata pre-generated in `gallery-metadata.json`
- Tailwind CSS purged and minified at build time
- Static HTML output to `dist/gallery.html`

### Principle III: Simplicity & Pragmatism ✅ PASS

**Assessment**: Solution uses vanilla JavaScript, no frameworks. Leverages existing infrastructure.

**Compliance**:
- No new npm dependencies required
- Vanilla JavaScript ES6+ (no React, Vue, etc.)
- Reuses existing Tailwind classes
- Works with existing image optimization pipeline
- Single-purpose JavaScript class (~300 LOC)

**Justification**: JavaScript is necessary for true masonry layout as CSS Grid masonry has limited browser support. The implementation is minimal and focused.

### Principle IV: Infrastructure as Code ✅ PASS

**Assessment**: No infrastructure changes required. Gallery is static content served from S3.

**Compliance**:
- No Terraform changes needed
- No new AWS resources
- Deployment via existing `make static-deploy`

### Principle V: Performance & Cost Efficiency ✅ PASS

**Assessment**: Masonry layout enhances performance by eliminating layout shifts and optimizing image loading.

**Compliance**:
- Target CLS = 0 (zero layout shift)
- Images already optimized (AVIF/WebP/JPEG)
- JavaScript ~3-4KB minified
- No additional API calls or network requests
- Maintains <2s page load target

**Cost Impact**: Zero. No additional AWS resources, no Lambda invocations.

### Summary

**Status**: ✅ ALL GATES PASSED

No constitutional violations. Feature aligns with all core principles:
- Static-first: Gallery is pre-generated HTML
- Build-time: All optimization happens during build
- Simplicity: Vanilla JS, no frameworks
- Infrastructure: No changes required
- Performance: Improves CLS and visual experience

**Proceed to Phase 0 Research**: APPROVED

## Project Structure

### Documentation (this feature)

```text
specs/001-masonry-gallery/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (in progress)
├── research.md          # Phase 0 output (pending)
├── data-model.md        # Phase 1 output (pending)
├── quickstart.md        # Phase 1 output (pending)
├── contracts/           # Phase 1 output (N/A for this feature)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── handlers/
│   └── gallery.go              # Gallery page handler (existing, minimal changes)
└── views/
    ├── gallery.templ           # Gallery template (modifications needed)
    └── gallery_templ.go        # Generated Go code from templ

static/
├── gallery-metadata.json       # Image dimensions/aspect ratios (existing)
├── css/
│   └── styles.css              # Tailwind styles (minimal additions)
└── js/
    ├── gallery.js              # Masonry layout logic (existing, needs fixes)
    └── lightbox.js             # Image lightbox (existing)

dist/
├── gallery.html                # Pre-generated static HTML (build output)
├── gallery-metadata.json       # Copied metadata
└── js/
    └── gallery.*.min.js        # Minified JavaScript (build output)

cmd/
└── build/
    └── main.go                 # Static site generator (existing)
```

**Structure Decision**: This is a static web application using Go + Templ for build-time HTML generation. The masonry layout is implemented entirely client-side with vanilla JavaScript. No backend API changes required - all modifications are to the gallery template and JavaScript positioning logic.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitutional violations detected. This section is not applicable.

---

## Post-Phase 1 Constitution Re-Evaluation

**Date**: 2025-11-10
**Status**: ✅ ALL PRINCIPLES STILL COMPLIANT

### Re-Assessment After Design Phase

**Phase 1 Outputs**:
- `research.md`: Confirmed JavaScript absolute positioning approach
- `data-model.md`: Documented existing ImageMetadata structure (no changes)
- `quickstart.md`: Developer guide for implementation

**Constitution Compliance Review**:

1. **Principle I (Static-First)**: ✅ PASS
   - Design maintains static HTML generation
   - No server-side rendering introduced
   - Client-side JavaScript remains optional enhancement

2. **Principle II (Build-Time Optimization)**: ✅ PASS
   - No new build-time processes required
   - Reuses existing image optimization and metadata generation
   - Tailwind compilation unchanged

3. **Principle III (Simplicity)**: ✅ PASS
   - Solution confirmed to use vanilla JavaScript (no new dependencies)
   - Research validated existing approach is industry-standard
   - Total JS payload remains <5KB minified

4. **Principle IV (Infrastructure as Code)**: ✅ PASS
   - No infrastructure changes in design
   - Deployment remains via existing Terraform/S3/CloudFront
   - No new AWS resources

5. **Principle V (Performance & Cost)**: ✅ PASS
   - Research confirms CLS = 0 target achievable
   - Performance optimizations identified (will-change hints)
   - Zero cost increase (no new API calls or resources)

**New Technical Decisions from Phase 1**:
- Continue with JavaScript absolute positioning (not CSS Grid masonry)
- Use `object-fit: cover` for zero-whitespace guarantee
- Add `will-change` hints for 60fps animations
- Maintain existing responsive breakpoints (1/2/3/4 columns)

**Impact on Constitution**: None. All decisions align with existing principles.

**Approval for Implementation**: ✅ APPROVED

Proceed to `/speckit.tasks` to generate implementation task list.

---

## Summary

**Planning Phase**: COMPLETE

**Artifacts Generated**:
- ✅ `plan.md` - Implementation plan (this file)
- ✅ `research.md` - Technical research and decisions
- ✅ `data-model.md` - Data structures and flow
- ✅ `quickstart.md` - Developer quickstart guide
- ✅ Agent context updated (CLAUDE.md)

**Constitution Compliance**: ✅ ALL GATES PASSED (pre and post Phase 1)

**Next Command**: `/speckit.tasks`

**Ready for Implementation**: YES
