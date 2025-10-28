# Tasks: Wedding Party Page

**Input**: Design documents from `/specs/002-wedding-party-page/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: No test tasks included - feature spec specifies manual visual testing and Lighthouse CI only

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This project uses single web application structure at repository root:
- `cmd/` - Entry points (build, main, lambda)
- `internal/` - Application code (handlers, views, router)
- `static/` - Static assets (images, CSS)
- `dist/` - Generated static site output

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare directory structure and assets for wedding party feature

- [X] T001 Create directory for wedding party photos at static/images/wedding-party/
- [X] T002 [P] Create default avatar placeholder at static/images/default-avatar.svg
- [X] T003 [P] Verify existing Templ and Tailwind setup (no changes needed, validation only)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core changes to existing files that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Define WeddingPartyMember struct type in internal/views/wedding_party.templ
- [X] T005 Create helper functions getGroomsmen() and getBridesmaids() in internal/views/wedding_party.templ
- [X] T006 Create handler HandleWeddingPartyPage in internal/handlers/wedding_party.go
- [X] T007 Register /wedding-party route in internal/router.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - View Wedding Party Members (Priority: P1) 🎯 MVP

**Goal**: Display all wedding party members with photos, names, roles, and descriptions in responsive 2-column/single-column layout

**Independent Test**: Navigate to http://localhost:8080/wedding-party and verify:
- All wedding party members display with photos/names/roles/descriptions
- Desktop (>768px): 2-column grid layout
- Mobile (<768px): Single column layout
- Page loads without errors

### Implementation for User Story 1

- [X] T008 [US1] Convert WeddingPartySection() to full-page WeddingParty() component in internal/views/wedding_party.templ
- [X] T009 [P] [US1] Create renderMembers() helper template in internal/views/wedding_party.templ
- [X] T010 [P] [US1] Create renderMember(member) template for individual cards in internal/views/wedding_party.templ
- [X] T011 [US1] Implement 2-column responsive grid layout with Tailwind classes (grid md:grid-cols-2) in internal/views/wedding_party.templ
- [X] T012 [US1] Add conditional rendering for missing photos (default avatar fallback) in renderMember() template
- [X] T013 [US1] Populate getGroomsmen() with actual wedding party data (5 members: Ronnie Campbell as Best Man, Mike Alves, Dana Roy, Mike Silva, Pete Smith)
- [X] T014 [US1] Populate getBridesmaids() with actual wedding party data (4 members: Melissa Moylan and Ainsley Kelliher as Maids of Honor, Kasey Silva, Allison Chisholm)
- [X] T015 [US1] Run templ generate to compile templates to Go code
- [X] T016 [US1] Test local dev server with make server and verify page renders at /wedding-party

**Checkpoint**: At this point, User Story 1 should be fully functional - wedding party page displays with all members in responsive layout

---

## Phase 4: User Story 2 - Navigate Between Wedding Party Members (Priority: P2)

**Goal**: Ensure wedding party members are visually distinct with proper spacing and hierarchy for easy scanning

**Independent Test**: View http://localhost:8080/wedding-party and verify:
- Each member card has clear visual separation
- Photos and text align consistently
- Easy to distinguish one member from another
- Scrolling through list feels natural

### Implementation for User Story 2

- [X] T017 [US2] Add consistent spacing between member cards (mb-8 or similar) in renderMember() template
- [X] T018 [P] [US2] Add shadow and border styling to member cards for visual separation in internal/views/wedding_party.templ
- [X] T019 [P] [US2] Ensure consistent text hierarchy (role → name → description) with appropriate font sizes in renderMember() template
- [X] T020 [US2] Test visual scanning on various screen sizes (320px, 768px, 1024px, 1920px) using browser DevTools

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - page displays with good visual hierarchy

---

## Phase 5: User Story 3 - Responsive Photo Viewing (Priority: P3)

**Goal**: Optimize photo loading and display for all devices and connection speeds

**Independent Test**: Load http://localhost:8080/wedding-party and verify:
- Photos maintain aspect ratio on all screen sizes
- No distortion or stretching
- Images load efficiently (lazy loading)
- Page remains responsive during image loading

### Implementation for User Story 3

- [X] T021 [US3] Add wedding party photos to static/images/wedding-party/ (provided by couple: ronnie-campbell.jpg, mike-alves.jpg, etc.)
- [X] T022 [P] [US3] Add proper image attributes (width, height, loading) to renderMember() template
- [X] T023 [P] [US3] Set loading="lazy" for below-fold images and loading="eager" for above-fold in renderMember() template
- [X] T024 [US3] Apply rounded-full and object-cover classes for consistent photo display in internal/views/wedding_party.templ
- [X] T025 [US3] Test photo display on various devices and ensure no distortion

**Checkpoint**: All user stories should now be independently functional - photos load efficiently and display beautifully

---

## Phase 6: Navigation & Static Generation

**Purpose**: Integrate page into navigation and static build process

- [X] T026 Add "Wedding Party" link to navigation in internal/views/hero.templ
- [X] T027 Add wedding party page generation function to cmd/build/main.go (follow pattern from other pages)
- [X] T028 Call wedding party generation function in main build workflow in cmd/build/main.go
- [X] T029 Run make static-build to generate dist/wedding-party.html
- [X] T030 Verify generated HTML file exists and contains correct content at dist/wedding-party.html

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [X] T031 [P] Add text truncation handling for long descriptions (max 500 chars) in renderMember() template
- [X] T032 [P] Verify all edge cases: missing photos, odd number of members, long text, empty state
- [X] T033 Test responsive layout breakpoints (320px, 768px, 1024px, 1920px) using browser DevTools
- [X] T034 Run make server and manually test all acceptance scenarios from spec.md
- [X] T035 [P] Run npm run lighthouse for performance audit (verify >90 performance score)
- [X] T036 Run make static-build && make upload-static && make invalidate-cache to deploy

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Navigation & Static Generation (Phase 6)**: Depends on User Story 1 (P1) minimum for MVP
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Enhances US1 but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Enhances US1 but independently testable

### Within Each User Story

- Struct and helper functions before templates
- Templates before handler
- Handler before route registration
- Local testing before static build
- Static build before deployment

### Parallel Opportunities

- T002 and T003 in Phase 1 can run in parallel
- T009 and T010 in Phase 3 can run in parallel (different template components)
- T018 and T019 in Phase 4 can run in parallel (styling changes)
- T022 and T023 in Phase 5 can run in parallel (image attributes)
- T031 and T032 in Phase 7 can run in parallel
- T035 can run parallel with other polish tasks
- Once Foundational phase completes, User Stories 1, 2, and 3 could theoretically be worked on in parallel by different developers

---

## Parallel Example: User Story 1

```bash
# Launch template helper components together:
Task: "Create renderMembers() helper template in internal/views/wedding_party.templ"
Task: "Create renderMember(member) template for individual cards in internal/views/wedding_party.templ"
```

## Parallel Example: Phase 7 Polish

```bash
# Launch independent polish tasks together:
Task: "Add text truncation handling in renderMember() template"
Task: "Verify all edge cases"
Task: "Run npm run lighthouse for performance audit"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (create directories, placeholders) - **15 min**
2. Complete Phase 2: Foundational (struct, helpers, handler, router) - **45 min**
3. Complete Phase 3: User Story 1 (full page with responsive layout) - **2 hours**
4. Complete Phase 6: Navigation & Static Generation (integrate into site) - **30 min**
5. **STOP and VALIDATE**: Test User Story 1 independently
6. Deploy/demo if ready - **15 min**

**Total MVP Effort**: ~4 hours

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (**1 hour**)
2. Add User Story 1 → Test independently → Deploy/Demo (**3 hours** - MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (**1 hour**)
4. Add User Story 3 → Test independently → Deploy/Demo (**1.5 hours**)
5. Add Polish → Final validation → Deploy (**1 hour**)

**Total Full Implementation**: ~7.5 hours

### Single Developer Sequential Strategy

Recommended order for solo implementation:

1. **Day 1, Session 1** (4 hours): Setup + Foundational + User Story 1 + Navigation
   - Complete T001-T016 + T026-T030
   - Deploy MVP version with basic wedding party page
2. **Day 1, Session 2** (2 hours): User Story 2 + User Story 3
   - Complete T017-T025
   - Enhance visual design and photo handling
3. **Day 2, Session 1** (1 hour): Polish + Deploy
   - Complete T031-T036
   - Final testing and production deployment

---

## File Change Summary

### New Files (2)
- `internal/handlers/wedding_party.go` - Wedding party page handler
- `static/images/wedding-party/` - Directory with ~10-15 wedding party photos

### Modified Files (4)
- `internal/views/wedding_party.templ` - Convert section to full page with data
- `internal/router.go` - Add /wedding-party route
- `internal/views/app.templ` - Add navigation link
- `cmd/build/main.go` - Add static page generation

### Generated Files (2)
- `internal/views/wedding_party_templ.go` - Auto-generated from .templ file
- `dist/wedding-party.html` - Static HTML output

**Total files touched**: 8 files (2 new, 4 modified, 2 generated)

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label maps task to specific user story (US1, US2, US3) for traceability
- Each user story should be independently completable and testable
- No test tasks included - spec calls for manual testing only
- Commit after each phase or logical group of tasks
- Stop at any checkpoint to validate story independently
- Follow existing patterns from homepage.go and venue.go handlers
- Use existing Tailwind utilities, avoid custom CSS
- Wedding party data hardcoded in template (no database, no CMS)
- Images optimized using existing pipeline (make optimize-images if needed)
- Static-first architecture - no Lambda changes required
- Zero new dependencies, zero infrastructure changes

## Validation Checklist

Before marking feature complete, verify:

- [ ] All wedding party members display with correct information
- [ ] Responsive layout works (2-column desktop, 1-column mobile)
- [ ] Photos load without distortion
- [ ] Missing photo fallback works (default avatar)
- [ ] Page accessible from main navigation
- [ ] Static HTML generated in dist/wedding-party.html
- [ ] Lighthouse performance score >90
- [ ] No console errors in browser
- [ ] Works on mobile devices (test 320px width minimum)
- [ ] Deployed to production and accessible at thedrewzers.com/wedding-party.html
