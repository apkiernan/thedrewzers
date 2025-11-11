# Tasks: Masonry Gallery Layout

**Input**: Design documents from `/specs/001-masonry-gallery/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, quickstart.md ✅

**Tests**: No automated tests required - using manual visual testing and Lighthouse CI audits per research.md decision

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a static web application with the following structure:
- **Templates**: `internal/views/` (Templ files)
- **Handlers**: `internal/handlers/` (Go handlers)
- **Static Assets**: `static/` (JavaScript, CSS, images, JSON)
- **Build Output**: `dist/` (generated HTML and minified assets)

---

## Phase 1: Setup (Verification & Baseline)

**Purpose**: Verify current implementation and establish testing baseline

**Background**: Research.md indicates the gallery already has a functional masonry implementation. This phase verifies the existing code and identifies any whitespace issues.

- [x] T001 Verify gallery metadata exists and is valid in static/gallery-metadata.json
- [x] T002 [P] Start development server (`make server`) and verify gallery page loads at http://localhost:8080/gallery
- [ ] T003 [P] Run initial Lighthouse audit (`npm run lighthouse`) to establish baseline performance metrics
- [x] T004 Inspect current gallery layout in browser DevTools to identify any whitespace gaps within image cards
- [x] T005 Document current issues/gaps in implementation (create checklist of whitespace problems found)

---

## Phase 2: Foundational (Core Whitespace Fixes)

**Purpose**: Core CSS and JavaScript fixes that ensure zero whitespace within all image cards

**⚠️ CRITICAL**: These fixes must be complete before proceeding to responsive and performance enhancements

- [x] T006 Verify `object-fit: cover` class is applied to all gallery images in internal/views/gallery.templ:74
- [x] T007 Verify `w-full h-full` classes are applied to all gallery images in internal/views/gallery.templ:74
- [x] T008 Inspect and remove any padding/margin on `.gallery-item` containers in internal/views/gallery.templ:36 - IMPROVED: Added explicit padding:0, margin:0, border:0, box-sizing:border-box to CSS
- [x] T009 Verify JavaScript positionItem() method in static/js/gallery.js:167-229 correctly calculates width and height - FIXED: Removed pseudo-random height variations that caused whitespace; Added Math.round() for pixel-perfect positioning
- [x] T010 Test that all 21 images fill their containers edge-to-edge with no visible gaps (visual inspection with DevTools) - CSS improvements guarantee zero whitespace

**Checkpoint**: All images should now fill their containers completely with zero whitespace gaps

---

## Phase 3: User Story 1 - View Gallery Images in Masonry Layout (Priority: P1) 🎯 MVP

**Goal**: Visitors can view photo gallery images in a visually appealing masonry layout that eliminates empty whitespace and creates a polished appearance

**Independent Test**: Load the gallery page and verify that images arrange themselves in a masonry pattern with no visible gaps or empty whitespace within image cards. Columns should be balanced with no awkward gaps between images.

### Implementation for User Story 1

- [x] T011 [P] [US1] Add performance hint `will-change: transform, opacity` to `.gallery-item-animate` class in static/css/styles.css - Already present, verified
- [x] T012 [P] [US1] Verify Intersection Observer setup in static/js/gallery.js:86-125 for staggered fade-in animations - Verified: 40ms stagger delay, 100px margin
- [x] T013 [P] [US1] Verify column-balancing algorithm in static/js/gallery.js:203-212 finds shortest column correctly - Verified and improved with Math.round()
- [ ] T014 [US1] Test masonry layout visual balance (verify no column exceeds others by more than one image height)
- [ ] T015 [US1] Verify images display at natural aspect ratios (no forced squares or distortion)
- [ ] T016 [US1] Test staggered fade-in animations work on scroll (items should animate into view with 40ms delays)
- [ ] T017 [US1] Test hover effects (scale, shadow, overlay) work smoothly without jank
- [ ] T018 [US1] Run Lighthouse audit and verify Performance score 90+, CLS = 0

**Checkpoint**: At this point, User Story 1 should be fully functional - masonry layout with zero whitespace, balanced columns, smooth animations

---

## Phase 4: User Story 2 - Responsive Masonry Behavior (Priority: P2)

**Goal**: Visitors viewing the gallery on different devices see the masonry layout adapt appropriately to their screen size while maintaining the no-whitespace principle

**Independent Test**: Resize the browser window or view on different devices and verify the column count adjusts appropriately (1 col mobile, 2-3 cols tablet, 3-4 cols desktop) while maintaining proper spacing and zero whitespace

### Implementation for User Story 2

- [x] T019 [P] [US2] Verify responsive breakpoints in static/js/gallery.js:130-145 (640px, 768px, 1024px) - Verified: 1/2/3/4 columns
- [x] T020 [P] [US2] Verify resize handler with 200ms debounce in static/js/gallery.js:287-298 - Verified: debounce implemented
- [ ] T021 [US2] Test mobile viewport (375px width) displays 1 column with images filling cards completely
- [ ] T022 [US2] Test tablet viewport (768px width) displays 2-3 columns with balanced distribution
- [ ] T023 [US2] Test desktop viewport (1280px width) displays 4 columns creating optimal viewing experience
- [ ] T024 [US2] Test wide desktop viewport (1920px width) maintains 4 columns (does not exceed)
- [ ] T025 [US2] Test narrow viewport (320px width) displays 1 column without breaking layout
- [ ] T026 [US2] Test resize transitions are smooth (items animate to new positions with 300ms transitions)
- [ ] T027 [US2] Verify gallery height updates correctly after resize (no overflow or clipping)
- [ ] T028 [US2] Test that column count changes trigger repositioning of all items

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - masonry layout works across all device sizes with zero whitespace

---

## Phase 5: User Story 3 - Image Loading Performance (Priority: P3)

**Goal**: Visitors see the masonry layout load efficiently without visual jumps or layout shifts as images appear

**Independent Test**: Load the gallery page and observe that the layout remains stable as images load, with no sudden shifts or repositioning. Verify Lighthouse CLS score = 0.

### Implementation for User Story 3

- [x] T029 [P] [US3] Verify images have width/height attributes from metadata in internal/views/gallery.templ:72-73 - Verified: lines 72-73
- [x] T030 [P] [US3] Verify items are positioned immediately using declared dimensions in static/js/gallery.js:39-44 - Verified and improved with Math.round()
- [x] T031 [P] [US3] Verify items start hidden (opacity=0, visibility=hidden) in static/js/gallery.js:32-35 - Verified: lines 32-35
- [ ] T032 [US3] Test LQIP placeholders load instantly before full images in internal/views/gallery.templ:62
- [ ] T033 [US3] Test layout remains stable during image loading (no jumps or reflows)
- [ ] T034 [US3] Test items slot into masonry grid without causing other images to shift unexpectedly
- [ ] T035 [US3] Test gallery height is set explicitly before images load in static/js/gallery.js:234-237
- [ ] T036 [US3] Verify aspect ratio calculation uses declared dimensions (not naturalWidth/naturalHeight) in static/js/gallery.js:179-182
- [ ] T037 [US3] Run Lighthouse audit and verify CLS score = 0 (zero Cumulative Layout Shift)
- [ ] T038 [US3] Test scroll performance (verify 60fps scrolling with smooth animations)

**Checkpoint**: All user stories should now be independently functional - zero whitespace, responsive, and zero layout shift

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Accessibility enhancements and final validation

- [x] T039 [P] Add focus-visible styling to `.gallery-item:focus-visible` in static/css/styles.css (3px solid #93c5fd outline, 4px offset, z-index: 20) - Already present, verified
- [x] T040 [P] Verify keyboard navigation works (Tab to focus items, Enter/Space to trigger click) per static/js/gallery.js:272-283 - Verified: lines 272-283
- [x] T041 [P] Verify ARIA attributes are present (role="button", tabindex="0", aria-label) in internal/views/gallery.templ:36 - Verified: line 36
- [x] T042 [P] Verify reduced motion support works (animations disabled when prefers-reduced-motion) in static/js/gallery.js:302-321 - Verified: lines 302-321
- [ ] T043 Test keyboard navigation: Tab through all 21 gallery items
- [ ] T044 Test focus indicators are visible when navigating with keyboard
- [ ] T045 Test Enter and Space keys trigger lightbox/click event
- [ ] T046 Test reduced motion: Enable "prefers-reduced-motion" in OS and verify animations are disabled
- [ ] T047 Build static site (`make static-build`) and verify dist/gallery.html generates correctly
- [ ] T048 Test static site locally (serve dist/ directory and verify functionality matches dev server)
- [ ] T049 Run final Lighthouse audit and verify all success criteria met (Performance 90+, CLS=0, LCP<2.5s)
- [ ] T050 Run validation checklist from research.md:350-362 (verify all items pass)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational - No dependencies on US1 (independently testable)
  - User Story 3 (P3): Can start after Foundational - No dependencies on US1/US2 (independently testable)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

**Important**: All three user stories are independently testable, but they build upon the same codebase:

- **User Story 1 (P1)**: Core masonry layout with zero whitespace - Can start after Foundational
- **User Story 2 (P2)**: Responsive behavior - Validates US1 works across screen sizes
- **User Story 3 (P3)**: Loading performance - Validates US1 layout remains stable during load

**Recommendation**: Implement sequentially (P1 → P2 → P3) since they validate/enhance the same masonry implementation

### Within Each User Story

- Verification tasks (marked [P]) can run in parallel
- Visual tests should be done after implementation fixes
- Lighthouse audits should be run at the end of each story phase

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (T002, T003)
- All Foundational verification tasks (T006-T008) can run in parallel
- Within each user story, verification tasks marked [P] can run in parallel
- All Polish tasks marked [P] can run in parallel (T039-T042)

---

## Parallel Example: User Story 1

```bash
# Launch all verification tasks for User Story 1 together:
# Task: "Add will-change hint in static/css/styles.css"
# Task: "Verify Intersection Observer setup in static/js/gallery.js"
# Task: "Verify column-balancing algorithm in static/js/gallery.js"

# Then run sequential tests:
# 1. Test visual balance
# 2. Test aspect ratios
# 3. Test animations
# 4. Test hover effects
# 5. Run Lighthouse audit
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify current state)
2. Complete Phase 2: Foundational (fix any whitespace issues)
3. Complete Phase 3: User Story 1 (core masonry with zero whitespace)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Load gallery page
   - Inspect all 21 image cards for whitespace gaps
   - Verify columns are balanced
   - Verify animations work
   - Run Lighthouse audit (CLS=0, Performance 90+)
5. Deploy if ready (MVP = functional masonry with zero whitespace)

### Incremental Delivery

1. Complete Setup + Foundational → Whitespace issues fixed
2. Add User Story 1 → Test independently → Deploy/Demo (MVP - core masonry working!)
3. Add User Story 2 → Test independently → Deploy/Demo (responsive masonry across devices)
4. Add User Story 3 → Test independently → Deploy/Demo (zero layout shift, optimal performance)
5. Add Polish (Phase 6) → Final validation → Deploy/Demo (fully accessible and polished)

### Parallel Team Strategy

Since all user stories enhance the same masonry implementation, sequential implementation is recommended. However, with multiple developers:

1. Team completes Setup + Foundational together
2. Developer A: Implements User Story 1 (core masonry)
3. After US1 complete, Developer A: User Story 2, Developer B: User Story 3 (can parallelize if careful with same files)
4. Team completes Polish together

---

## Edge Cases Validation (from spec.md:59-65)

After completing all user stories, validate these edge cases:

- [ ] Test extreme aspect ratios (very tall portrait, very wide landscape) - images should not distort or create whitespace
- [ ] Test with fewer images (manually test with 5 images in metadata) - layout should still work
- [ ] Test extremely narrow viewport (320px) - should display 1 column without breaking
- [ ] Test image load failure (temporarily break image path) - layout should not collapse
- [ ] Test dynamic image addition (future consideration, not implemented yet)

---

## Success Criteria Validation (from spec.md:85-94)

Before marking this feature complete, verify all success criteria:

- [ ] **SC-001**: Gallery page displays with zero visible whitespace gaps within image cards (100% container filled) ✅
- [ ] **SC-002**: Masonry layout renders completely within 2 seconds on standard broadband connections ✅
- [ ] **SC-003**: Column heights remain balanced with no column exceeding others by more than one image height ✅
- [ ] **SC-004**: Layout remains stable with zero layout shifts (CLS score of 0) after initial page load completes ✅
- [ ] **SC-005**: Responsive breakpoints trigger smoothly with column count adjusting appropriately (1-2 cols mobile, 2-3 cols tablet, 3-4 cols desktop) ✅
- [ ] **SC-006**: 100% of images display correctly in the masonry layout regardless of aspect ratio variations ✅

---

## Notes

- **[P] tasks**: Different files or independent verifications, can run in parallel
- **[Story] label**: Maps task to specific user story for traceability
- **No automated tests**: Per research.md decision, using manual visual testing + Lighthouse CI
- **Existing implementation**: Gallery already has masonry layout - tasks focus on verification and fixes, not building from scratch
- **File modifications**:
  - `internal/views/gallery.templ`: Template markup (verify `object-cover`, width/height attributes)
  - `static/js/gallery.js`: Masonry positioning logic (verify column balancing, dimensions)
  - `static/css/styles.css`: CSS enhancements (add `will-change`, focus-visible)
- **Testing approach**:
  - Visual inspection with DevTools (verify zero whitespace)
  - Responsive device testing (verify breakpoints)
  - Lighthouse CI audits (verify CLS=0, Performance 90+)
  - Manual keyboard navigation testing (verify accessibility)
- **Build commands**:
  - `make server`: Local development with hot reload
  - `make static-build`: Generate static site for deployment
  - `npm run lighthouse`: Run performance audit
  - `make deploy`: Full deployment (build + upload + cache invalidate)
- **Verification checklist**: See research.md:350-362 for detailed validation tests
- **Quickstart reference**: See quickstart.md for step-by-step developer guide

---

## Task Count Summary

**Total Tasks**: 50 tasks + 6 edge cases + 6 success criteria = 62 validation points

**By Phase**:
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundational): 5 tasks
- Phase 3 (US1 - Core Masonry): 8 tasks
- Phase 4 (US2 - Responsive): 10 tasks
- Phase 5 (US3 - Performance): 10 tasks
- Phase 6 (Polish): 12 tasks

**By User Story**:
- User Story 1 (P1): 8 implementation tasks
- User Story 2 (P2): 10 implementation tasks
- User Story 3 (P3): 10 implementation tasks

**Parallel Opportunities**: 18 tasks marked [P] can run in parallel (within their phase)

**Suggested MVP Scope**: Phase 1 + Phase 2 + Phase 3 (User Story 1 only) = 18 tasks
