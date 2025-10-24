# Tasks: Website Performance Optimization

**Feature Branch**: `001-performance-optimization`
**Input**: Design documents from `/specs/001-performance-optimization/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Performance optimization is inherently testable through automated Lighthouse audits, so explicit test tasks are included.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions

**Project Structure** (Go + Templ + Tailwind static site):

- Build scripts: `cmd/optimize-images/`, `cmd/generate-lqip/`
- Templates: `internal/views/*.templ`
- Static assets: `static/fonts/`, `static/css/`, `static/js/`, `static/images/`
- Generated output: `dist/`
- Infrastructure: `terraform/*.tf`
- Build automation: `Makefile`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Install tools and initialize performance optimization infrastructure

- [ ] T001 Install image optimization CLI tools (cwebp, avifenc, mozjpeg) via brew/apt
- [ ] T002 [P] Install font optimization tools (glyphhanger via npm, fonttools/brotli via pip)
- [ ] T003 [P] Install Lighthouse CI for performance monitoring (npm install -g @lhci/cli)
- [ ] T004 [P] Create directory structure: cmd/optimize-images/, cmd/generate-lqip/, .lighthouse/
- [ ] T005 Verify all CLI tools installed correctly (cwebp --version, avifenc --version, glyphhanger --version, pyftsubset --help)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Font optimization and baseline performance infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: Font optimization must complete before image optimization begins, as it provides immediate Lighthouse improvements and is a prerequisite for accurate performance baselining

### Font Optimization (Foundational)

- [ ] T006 Subset BodoniModa-Variable.ttf to WOFF2 using pyftsubset with Latin charset (U+0020-007F,U+00A0-00FF) → static/fonts/BodoniModa-Variable.woff2
- [ ] T007 [P] Subset BodoniModa-Italic-Variable.ttf to WOFF2 using pyftsubset with Latin charset → static/fonts/BodoniModa-Italic-Variable.woff2
- [ ] T008 [P] Subset BonheurRoyale-Regular.ttf to WOFF2 using pyftsubset with Latin charset → static/fonts/BonheurRoyale-Regular.woff2
- [ ] T009 Update @font-face declarations in src/input.css to reference WOFF2 files and add font-display: swap
- [ ] T010 Add font preloading for BodoniModa-Variable.woff2 in internal/views/app.templ with crossorigin attribute
- [ ] T011 Rebuild CSS and verify WOFF2 fonts load correctly (npm run build && make server)
- [ ] T012 Deploy font optimizations (make static-build && make upload-static && make invalidate-cache)

### Performance Baseline

- [ ] T013 Configure Lighthouse CI in .lighthouse/lighthouserc.json with performance budgets (90+ mobile, 95+ desktop, LCP ≤2.5s, CLS ≤0.1)
- [ ] T014 Run initial Lighthouse audit to establish baseline scores (npm run lighthouse or lhci autorun)
- [ ] T015 Document baseline metrics in specs/001-performance-optimization/baseline-metrics.md

**Checkpoint**: Fonts optimized (70% payload reduction achieved), baseline metrics established - user story implementation can now begin

---

## Phase 3: User Story 1 - Fast Initial Page Load (Priority: P1) 🎯 MVP

**Goal**: Wedding guests visiting for the first time see page content and images load instantly without waiting. Page interactive within 2s on 4G, 1s on cable.

**Independent Test**: Measure load time from fresh browser (no cache) on throttled 4G/3G/cable networks. Verify page interactive within 2s (4G) and images display within 3s. Use Lighthouse Performance score 90+ and LCP ≤2.5s as success criteria.

### US1: Image Optimization - Responsive Sizes & Formats

- [ ] T016 [P] [US1] Create cmd/optimize-images/main.go to generate responsive image sizes (640w, 768w, 1024w, 1280w, 1920w, 2560w) using imaging.Resize with Lanczos filter
- [ ] T017 [P] [US1] Create cmd/generate-lqip/main.go to generate 20px blurred JPEG placeholders (2-5KB each) using imaging.Blur
- [ ] T018 [US1] Add optimize-images target to Makefile that runs Go programs then converts JPG → WebP (cwebp -q 85) and JPG → AVIF (avifenc --cq-level 18)
- [ ] T019 [US1] Update static-build target in Makefile to depend on optimize-images
- [ ] T020 [US1] Test image optimization pipeline (make optimize-images) and verify output files in dist/images/ (640w.jpg/webp/avif, 1280w._, 1920w._, lqip.jpg)

### US1: Template Updates for Responsive Images

- [ ] T021 [US1] Update internal/views/hero.templ to use <picture> element with AVIF/WebP/JPEG sources, srcset, sizes, and LQIP placeholder
- [ ] T022 [P] [US1] Update internal/views/gallery.templ to use <picture> elements for all gallery images with lazy loading and LQIP
- [ ] T023 [P] [US1] Update internal/views/wedding_details.templ for venue photos with responsive images
- [ ] T024 [US1] Add CSS for LQIP blur effect (filter: blur(20px); transition: filter 0.3s) in src/input.css
- [ ] T025 [US1] Add JavaScript for LQIP swap (onload: remove blur, swap src) in static/js/image-loader.js

### US1: Build & Validation

- [ ] T026 [US1] Run full build with image optimization (make static-build) and verify dist/ contains optimized images
- [ ] T027 [US1] Test locally (make server) on throttled 4G connection and verify images load progressively with blur → sharp transition
- [ ] T028 [US1] Run Lighthouse audit for homepage and verify Performance score 85+ (mobile), LCP ≤2.5s, CLS ≤0.1
- [ ] T029 [US1] Deploy optimized images (make upload-static && make invalidate-cache)
- [ ] T030 [US1] Run Lighthouse audit on production URL and verify 2s load time on 4G, 1s on cable

**Checkpoint**: User Story 1 complete - initial page load fast, images optimized, 60-85% page weight reduction achieved

---

## Phase 4: User Story 3 - Smooth Image Display (Priority: P1)

**Goal**: Images load progressively without layout shifts, blurry flashes, or jarring visual changes. CLS score ≤0.1.

**Independent Test**: Observe image loading on throttled connections. Measure CLS using Lighthouse. Verify no layout shifts, LQIP → full image transition is smooth, and lazy loading works for below-fold images.

### US3: Prevent Layout Shifts

- [ ] T031 [US3] Add explicit width/height attributes to all <img> tags in internal/views/hero.templ based on actual image dimensions
- [ ] T032 [P] [US3] Add explicit width/height attributes to gallery images in internal/views/gallery.templ
- [ ] T033 [P] [US3] Add explicit width/height attributes to venue images in internal/views/wedding_details.templ
- [ ] T034 [US3] Update CSS in src/input.css to add aspect-ratio property for images to prevent reflow

### US3: Lazy Loading Implementation

- [ ] T035 [US3] Add loading="lazy" attribute to all below-fold images in internal/views/gallery.templ
- [ ] T036 [US3] Verify hero images (above-fold) do NOT have loading="lazy" in internal/views/hero.templ
- [ ] T037 [US3] Test lazy loading behavior by scrolling gallery page and verifying images load just-in-time

### US3: LQIP Refinement

- [ ] T038 [US3] Verify LQIP quality settings produce 2-5KB placeholders without visible pixelation in cmd/generate-lqip/main.go
- [ ] T039 [US3] Ensure LQIP aspect ratio matches full image to prevent layout shift during swap
- [ ] T040 [US3] Test LQIP → full image transition on 3G connection and verify smooth blur-to-sharp effect

### US3: Validation

- [ ] T041 [US3] Run Lighthouse audit and verify CLS ≤0.1 (target: 0.05 or better)
- [ ] T042 [US3] Visual test on multiple devices (iPhone, Android, desktop) to verify no layout shifts or jarring transitions
- [ ] T043 [US3] Deploy smooth image loading improvements (make deploy)

**Checkpoint**: User Story 3 complete - smooth image loading, zero layout shifts, excellent visual experience

---

## Phase 5: User Story 4 - Optimal Mobile Performance (Priority: P1)

**Goal**: Mobile devices experience fast, smooth performance without excessive data usage. Total homepage <1MB, appropriately sized images served via srcset, 60fps interactions.

**Independent Test**: Test on real mobile devices or emulators with throttled 4G. Measure data transfer (Network tab), verify correct image sizes loaded via srcset, test 60fps scrolling performance.

### US4: Mobile Image Optimization

- [ ] T044 [US4] Verify srcset in templates serves 640w images to mobile devices (test via Chrome DevTools mobile emulation)
- [ ] T045 [US4] Verify sizes attribute in <picture> elements correctly calculates viewport-based image selection
- [ ] T046 [US4] Test on real iPhone/Android device and confirm 640w-1280w images load (not 1920w-2560w)
- [ ] T047 [US4] Measure total page weight on mobile (4G throttled) and verify <1MB for initial homepage load

### US4: Mobile JavaScript Performance

- [ ] T048 [US4] Audit JavaScript execution time using Chrome DevTools Performance tab on mobile emulation
- [ ] T049 [US4] Defer non-critical JavaScript in internal/views/app.templ (slideshow.js, flip-cards.js with defer attribute)
- [ ] T050 [US4] Verify hero-lightbox.js uses type="module" for better performance
- [ ] T051 [US4] Test scroll performance (60fps) using Chrome DevTools rendering panel with Paint Flashing enabled

### US4: Mobile-Specific Optimizations

- [ ] T052 [US4] Add viewport meta tag optimization in internal/views/app.templ (width=device-width, initial-scale=1)
- [ ] T053 [US4] Test touch interactions (tapping, scrolling) on real mobile device and verify responsive feel with no lag
- [ ] T054 [US4] Run Lighthouse mobile audit and verify Performance 90+, TTI <3s, TBT <200ms

**Checkpoint**: User Story 4 complete - mobile performance excellent, <1MB page weight, 60fps interactions

---

## Phase 6: User Story 2 - Instant Navigation for Return Visitors (Priority: P2)

**Goal**: Return visitors experience near-instant page loads (<0.5s) using cached resources. Navigation between pages feels instant (<0.3s).

**Independent Test**: Visit site, close tab (keep browser open), revisit site and measure load time. Verify cached resources used (Network tab shows "from disk cache"). Navigate between pages and verify instant transitions.

### US2: CloudFront Caching Configuration

- [ ] T055 [US2] Update terraform/cloudfront.tf to add cache behavior for /static/\* with 1-year TTL and compression enabled
- [ ] T056 [US2] Add cache behavior for /\*.html with 1-day default TTL and revalidation in terraform/cloudfront.tf
- [ ] T057 [US2] Add cache behavior for fonts (/static/fonts/\*.woff2) with immutable cache-control in terraform/cloudfront.tf
- [ ] T058 [US2] Run terraform plan (make tf-plan) and review caching changes
- [ ] T059 [US2] Apply Terraform changes (make tf-apply) to update CloudFront distribution

### US2: S3 Cache Headers

- [ ] T060 [US2] Update Makefile upload-static target to set Cache-Control: public, max-age=31536000, immutable for _.avif, _.webp, \*.woff2
- [ ] T061 [US2] Set Cache-Control: public, max-age=86400 for \*.html files in upload-static
- [ ] T062 [US2] Set correct Content-Type headers for AVIF (image/avif), WebP (image/webp), WOFF2 (font/woff2) during S3 upload
- [ ] T063 [US2] Deploy with new cache headers (make upload-static)

### US2: Cache Validation

- [ ] T064 [US2] Test cache headers using curl -I https://yoursite.com/static/fonts/BodoniModa-Variable.woff2 and verify Cache-Control present
- [ ] T065 [US2] Visit site, clear tab, revisit and verify cached resources via Network tab (from disk cache)
- [ ] T066 [US2] Measure return visit load time and verify <0.5s
- [ ] T067 [US2] Navigate between pages (homepage → venue → gallery) and verify instant transitions (<0.3s)
- [ ] T068 [US2] Test cache revalidation by waiting 24h and verifying only modified resources re-download

**Checkpoint**: User Story 2 complete - return visits instant, aggressive caching working, 95%+ cache hit rate

---

## Phase 7: User Story 5 - Efficient Resource Loading (Priority: P2)

**Goal**: Only required resources loaded per page. Critical resources load first, non-critical deferred. Modern formats (WebP/AVIF) served to supporting browsers.

**Independent Test**: Analyze Network tab for unused resources. Use Coverage tool in Chrome DevTools to detect unused CSS/JS. Verify modern browsers receive AVIF, fallback browsers receive JPEG.

### US5: CSS Optimization

- [ ] T069 [US5] Verify Tailwind purging is active by checking tailwind.config.js content paths include ./internal/views/\*.templ
- [ ] T070 [US5] Run npm run build and verify output CSS file size is minimal (~18KB minified)
- [ ] T071 [US5] Add preload hint for critical CSS in internal/views/app.templ (<link rel="preload" href="/static/css/tailwind.css" as="style">)
- [ ] T072 [US5] Use Chrome DevTools Coverage tool to verify no unused CSS (target <10% unused)

### US5: JavaScript Optimization

- [ ] T073 [US5] Audit JavaScript files (slideshow.js, flip-cards.js, hero-lightbox.js, gallery.js) for unused code
- [ ] T074 [US5] Ensure all JavaScript uses defer or type="module" in internal/views/app.templ and internal/views/app_static.templ
- [ ] T075 [US5] Verify no JavaScript is blocking initial render (Lighthouse audit shows no render-blocking scripts)

### US5: Resource Hints & Preloading

- [ ] T076 [US5] Add dns-prefetch for external resources (if any) in internal/views/app.templ
- [ ] T077 [US5] Add preconnect for critical external resources (if any) in internal/views/app.templ
- [ ] T078 [US5] Verify hero image is preloaded in internal/views/app.templ (<link rel="preload" as="image">)

### US5: Format Negotiation Validation

- [ ] T079 [US5] Test in Chrome/Edge and verify AVIF images load via Network tab
- [ ] T080 [US5] Test in Safari and verify WebP images load (or AVIF in Safari 16.1+)
- [ ] T081 [US5] Test in older browser (e.g., IE11 emulation) and verify JPEG fallback loads
- [ ] T082 [US5] Run Lighthouse audit and verify "Serve images in next-gen formats" passes

**Checkpoint**: User Story 5 complete - efficient resource loading, no waste, modern formats served correctly

---

## Phase 8: Performance Monitoring & Polish

**Purpose**: Establish ongoing performance monitoring and final optimizations

### Performance Monitoring Setup

- [ ] T083 [P] Add npm script "lighthouse": "lhci autorun" to package.json
- [ ] T084 [P] Document how to run Lighthouse audits in CLAUDE.md
- [ ] T085 Create performance budget documentation in specs/001-performance-optimization/performance-budget.md (90+ mobile, 95+ desktop, <1MB homepage, etc.)

### Final Validation

- [ ] T086 Run comprehensive Lighthouse audit on all pages (homepage, venue, gallery, travel, wedding party, FAQ)
- [ ] T087 Verify all pages achieve Performance 90+ (mobile), 95+ (desktop)
- [ ] T088 Verify Core Web Vitals: LCP ≤2.5s, FID ≤100ms, CLS ≤0.1, TTI <3s on all pages
- [ ] T089 Test on real devices (iPhone, Android, desktop) across different network conditions (4G, 3G, cable)
- [ ] T090 Verify total homepage weight <1MB on mobile, font payload <150KB, image payload optimized

### Documentation & Cleanup

- [ ] T091 [P] Update CLAUDE.md with new make commands (make optimize-images) and performance optimization notes
- [ ] T092 [P] Document font optimization workflow in specs/001-performance-optimization/font-optimization-notes.md
- [ ] T093 [P] Document image optimization workflow in specs/001-performance-optimization/image-optimization-notes.md
- [ ] T094 Create baseline vs. final metrics comparison table in specs/001-performance-optimization/results.md
- [ ] T095 Run quickstart.md validation steps to ensure all phases work correctly

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - install tools immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
  - Font optimization completes first
  - Performance baseline established
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - **US1 (Fast Initial Load) - P1**: Must complete first (foundational for other stories)
  - **US3 (Smooth Images) - P1**: Depends on US1 image optimization
  - **US4 (Mobile Performance) - P1**: Depends on US1 and US3
  - **US2 (Return Visitors) - P2**: Can run parallel to US4, or after US1/US3/US4
  - **US5 (Efficient Loading) - P2**: Can run parallel to US2, or after all P1 stories
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

```
Setup (Phase 1)
    ↓
Foundational (Phase 2: Fonts + Baseline)
    ↓
    ├→ US1 (Fast Initial Load) - P1 [MUST COMPLETE FIRST]
    │       ↓
    │       ├→ US3 (Smooth Images) - P1 [Depends on US1]
    │       │       ↓
    │       │       └→ US4 (Mobile Performance) - P1 [Depends on US1 + US3]
    │       │
    │       └→ US2 (Return Visitors) - P2 [Can run parallel to US3/US4 or after]
    │               ↓
    │               └→ US5 (Efficient Loading) - P2 [Can run parallel to US2 or after US4]
    ↓
Polish (Phase 8)
```

### Parallel Opportunities

**Within Phase 1 (Setup)**:

- T002, T003, T004 can all run in parallel after T001

**Within Phase 2 (Foundational)**:

- T007, T008 can run in parallel with T006 (different font files)

**Within User Story 1**:

- T016, T017 can run in parallel (different Go programs)
- T022, T023 can run in parallel with T021 (different template files)

**Within User Story 3**:

- T032, T033 can run in parallel with T031 (different template files)

**Within User Story 4**:

- T044, T045, T046, T047 can be tested in parallel (different test scenarios)

**Cross-Story Parallel** (with multiple developers):

- After US1 completes, US2 and US3 can start in parallel
- US5 tasks (T069-T082) can run in parallel with US2 tasks (T055-T068) if different developers

---

## Parallel Example: User Story 1 (Fast Initial Page Load)

```bash
# Phase 1: Create build scripts in parallel
Task T016: "Create cmd/optimize-images/main.go to generate responsive image sizes"
Task T017: "Create cmd/generate-lqip/main.go to generate LQIP placeholders"

# Phase 2: Update templates in parallel
Task T021: "Update internal/views/hero.templ with <picture> elements"
Task T022: "Update internal/views/gallery.templ with <picture> elements"
Task T023: "Update internal/views/wedding_details.templ with <picture> elements"
```

---

## Implementation Strategy

### MVP First (Font + Image Optimization)

**Recommended approach for maximum impact with minimal scope:**

1. ✅ **Complete Phase 1**: Setup (install tools)
2. ✅ **Complete Phase 2**: Foundational (fonts + baseline) - **70% font reduction**
3. ✅ **Complete Phase 3**: User Story 1 (image optimization) - **60-85% page weight reduction**
4. 🎯 **STOP and VALIDATE**: Run Lighthouse, measure load times, verify 90+ scores
5. ✅ **Complete Phase 4**: User Story 3 (smooth images) - **CLS ≤0.1**
6. 🎯 **STOP and VALIDATE**: Test on real devices, verify smooth loading
7. **Deploy MVP**: At this point, you have fast initial loads + smooth images = huge UX improvement

**Expected Lighthouse Improvement at MVP Stage**:

- Performance (Mobile): 65 → 90+ (+25 points)
- Performance (Desktop): 80 → 95+ (+15 points)
- LCP: 4.5s → 2.0s
- CLS: 0.3 → 0.05
- Page Weight: 7.5MB → 1MB (87% reduction)

### Incremental Delivery (Recommended Full Implementation)

1. **Phase 1 + 2**: Setup + Foundational → Fonts optimized (70% reduction)
2. **Phase 3 (US1)**: Fast initial load → Test independently → Deploy/Demo (MVP Stage 1)
3. **Phase 4 (US3)**: Smooth images → Test independently → Deploy/Demo (MVP Stage 2)
4. **Phase 5 (US4)**: Mobile performance → Test independently → Deploy/Demo
5. **Phase 6 (US2)**: Caching for return visitors → Test independently → Deploy/Demo
6. **Phase 7 (US5)**: Efficient resource loading → Test independently → Deploy/Demo
7. **Phase 8**: Polish + monitoring → Final validation → Production ready

Each phase adds measurable value without breaking previous optimizations.

### Parallel Team Strategy

**If you have 2-3 developers available:**

**Week 1**: Team completes Setup + Foundational together

- Developer A: Font optimization (T006-T012)
- Developer B: Tool installation and verification (T001-T005)
- Developer C: Lighthouse CI setup (T013-T015)

**Week 2**: Parallel user story work begins

- Developer A: US1 - Image optimization (T016-T030)
- Developer B: US2 - CloudFront caching (T055-T068) [starts after US1 builds pipeline]
- Developer C: US5 - Resource loading efficiency (T069-T082)

**Week 3**: Complete remaining P1 stories

- Developer A: US3 - Smooth images (T031-T043)
- Developer B: US4 - Mobile performance (T044-T054)

**Week 4**: Final validation and polish

- All developers: Phase 8 (T083-T095)

---

## Testing Checklist

### Automated Testing (Lighthouse CI)

- [ ] Lighthouse Performance score 90+ (mobile), 95+ (desktop)
- [ ] Lighthouse Best Practices score 100
- [ ] Lighthouse Accessibility score 100
- [ ] Lighthouse SEO score 100
- [ ] Core Web Vitals in "Good" range (LCP ≤2.5s, FID ≤100ms, CLS ≤0.1)
- [ ] Time to Interactive (TTI) <3s on 4G mobile

### Manual Testing (Real Devices)

- [ ] Test on iPhone (Safari) - verify WebP/AVIF support, smooth scrolling
- [ ] Test on Android (Chrome) - verify AVIF loading, 60fps interactions
- [ ] Test on desktop (Chrome, Firefox, Safari) - verify optimal performance
- [ ] Test on throttled 4G connection - verify 2s initial load
- [ ] Test on throttled 3G connection - verify critical content <3s
- [ ] Test on cable connection - verify <1s initial load

### Visual Regression Testing

- [ ] Verify no layout shifts during image loading (CLS check)
- [ ] Verify LQIP → full image transition is smooth (no flash)
- [ ] Verify lazy loading works correctly (below-fold images load on scroll)
- [ ] Verify fonts load with no invisible text (FOIT)
- [ ] Verify responsive images display correctly across breakpoints

### Cache Testing

- [ ] First visit: Measure load time, verify all resources downloaded
- [ ] Return visit (same session): Verify cached resources used, <0.5s load
- [ ] Return visit (after 24h): Verify cache revalidation, only changed files re-downloaded
- [ ] Navigate between pages: Verify instant transitions (<0.3s)

---

## Success Metrics

### Lighthouse Scores (Target)

| Metric                | Before | Target | Expected After |
| --------------------- | ------ | ------ | -------------- |
| Performance (Mobile)  | 65     | 90+    | 92-95          |
| Performance (Desktop) | 80     | 95+    | 97-99          |
| Best Practices        | 95     | 100    | 100            |
| Accessibility         | 100    | 100    | 100            |
| SEO                   | 100    | 100    | 100            |

### Core Web Vitals (Target)

| Metric                         | Before | Target | Expected After |
| ------------------------------ | ------ | ------ | -------------- |
| LCP (Largest Contentful Paint) | 4.5s   | ≤2.5s  | 1.8-2.2s       |
| FID (First Input Delay)        | 80ms   | ≤100ms | 50-80ms        |
| CLS (Cumulative Layout Shift)  | 0.3    | ≤0.1   | 0.03-0.05      |
| TTI (Time to Interactive)      | 5.2s   | <3s    | 2.5-2.8s       |

### Page Weight (Target)

| Asset Type      | Before      | Target | Expected After        |
| --------------- | ----------- | ------ | --------------------- |
| Total Homepage  | 7.5MB       | <1MB   | 800KB-950KB           |
| Fonts           | 460KB (TTF) | <150KB | 90-140KB (WOFF2)      |
| Images (Mobile) | 6MB         | <500KB | 350-450KB (AVIF/WebP) |
| CSS             | 29KB        | <20KB  | 18-19KB (minified)    |
| JavaScript      | 15KB        | <15KB  | 12-14KB (deferred)    |

---

## Notes

- **[P] tasks** = different files, no dependencies, can run in parallel
- **[Story] label** maps task to specific user story for traceability (US1, US2, US3, US4, US5)
- **Performance is testable** via Lighthouse CI - each user story has clear success metrics
- **Independent testing** means each user story can be validated without others being complete
- **Build-time optimization** means no runtime complexity or Lambda changes
- **Constitutional alignment** maintained throughout - simplicity, static-first, build-time only
- **Commit after each task** or logical group for easy rollback
- **Stop at checkpoints** to validate user story independently before proceeding
- **Tools are system-level** (brew install) not Go dependencies to maintain simplicity

---

## Task Summary

**Total Tasks**: 95 tasks

- **Phase 1 (Setup)**: 5 tasks
- **Phase 2 (Foundational)**: 10 tasks (fonts + baseline)
- **Phase 3 (US1 - Fast Initial Load)**: 15 tasks (image optimization foundation)
- **Phase 4 (US3 - Smooth Images)**: 13 tasks (layout shift prevention)
- **Phase 5 (US4 - Mobile Performance)**: 11 tasks (mobile optimization)
- **Phase 6 (US2 - Return Visitors)**: 14 tasks (caching optimization)
- **Phase 7 (US5 - Efficient Loading)**: 14 tasks (resource efficiency)
- **Phase 8 (Polish)**: 13 tasks (monitoring + documentation)

**Parallelizable Tasks**: 23 tasks marked [P]
**User Story Tasks**: 67 tasks (70% of total) directly mapped to user stories
**Estimated Timeline**: 3-5 weeks (solo), 2-3 weeks (team of 2-3)

**MVP Scope** (Recommended for fastest value):

- Phase 1 + 2 + 3 + 4 = 43 tasks
- Delivers: Font optimization + Fast initial loads + Smooth images
- Expected result: 90+ mobile score, 95+ desktop score, 60-85% page weight reduction
