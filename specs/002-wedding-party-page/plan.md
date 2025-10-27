# Implementation Plan: Wedding Party Page

**Branch**: `002-wedding-party-page` | **Date**: 2025-10-27 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-wedding-party-page/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Create a dedicated wedding party page that displays wedding party members with photos, names, roles, and personal anecdotes. The page will be statically generated at build time and served from S3/CloudFront with responsive 2-column layout (desktop) and single-column layout (mobile).

## Technical Context

**Language/Version**: Go 1.23.3
**Primary Dependencies**: Templ v0.2.793 (templating), Tailwind CSS 3.4.14 (styling)
**Storage**: Static files (Go slice/array for wedding party data in template)
**Testing**: Manual visual testing, Lighthouse CI for performance
**Target Platform**: Static HTML served from S3/CloudFront, generated via `cmd/build/main.go`
**Project Type**: Web (static-first architecture, single codebase)
**Performance Goals**: Page load <3s on 3G, images display within 3s, responsive layout adapts automatically
**Constraints**: <2s load time, <$5/month operational cost, maintain mobile-first performance (320px-1920px)
**Scale/Scope**: Single page displaying 10-15 wedding party members, static HTML generation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Static-First Architecture ✓ PASS

**Requirement**: All user-facing content MUST be pre-generated as static HTML.

**Status**: PASS - Wedding party page will be generated as static HTML via `cmd/build/main.go` and served from S3/CloudFront. No Lambda involvement for page rendering.

### Principle II: Build-Time Optimization ✓ PASS

**Requirement**: All template compilation and asset optimization MUST happen at build time.

**Status**: PASS - Templ templates will be compiled to Go code at build time. Images will use existing optimization pipeline. Tailwind CSS purged and minified during build.

### Principle III: Simplicity & Pragmatism ✓ PASS

**Requirement**: Choose simplest solution. No unnecessary frameworks or abstractions.

**Status**: PASS - Reuses existing patterns (Templ components, Tailwind utility classes, Go handlers). Wedding party data stored as simple Go slice in template. No database, no ORM, no frontend framework needed.

### Principle IV: Infrastructure as Code ✓ PASS

**Requirement**: All AWS infrastructure managed via Terraform.

**Status**: PASS - No new infrastructure required. Page served from existing S3 bucket and CloudFront distribution.

### Principle V: Performance & Cost Efficiency ✓ PASS

**Requirement**: Load <2s on 3G, cost <$5/month.

**Status**: PASS - Static HTML eliminates Lambda cold starts. Images will use existing optimization pipeline (AVIF/WebP/JPEG with responsive sizes). No additional AWS costs beyond storage/bandwidth.

### Gate Evaluation: ✓ ALL CHECKS PASS

No constitution violations. Feature aligns with all five core principles. Proceed to Phase 0.

---

## Post-Phase 1 Constitution Re-evaluation

### Principle I: Static-First Architecture ✓ PASS (Confirmed)

**Design Decision**: Wedding party data stored as Go slice in template, compiled to static HTML at build time.

**Status**: PASS - No runtime data fetching, no API calls, no database queries. Pure static generation.

### Principle II: Build-Time Optimization ✓ PASS (Confirmed)

**Design Decision**: Templ templates compile to Go code, images use existing optimization pipeline.

**Status**: PASS - All optimization happens at build time. No runtime template parsing or image processing.

### Principle III: Simplicity & Pragmatism ✓ PASS (Confirmed)

**Design Decision**: Reuse existing patterns (handlers, router, Templ components), no new dependencies.

**Status**: PASS - Data model uses simple Go structs. No ORM, no external services, no complexity added.

### Principle IV: Infrastructure as Code ✓ PASS (Confirmed)

**Design Decision**: No infrastructure changes required.

**Status**: PASS - Uses existing S3 bucket, CloudFront distribution, build pipeline. Zero Terraform changes.

### Principle V: Performance & Cost Efficiency ✓ PASS (Confirmed)

**Design Decision**: Static HTML + optimized images + CloudFront CDN.

**Status**: PASS - Page load <2s guaranteed (static delivery). No additional AWS costs. Images optimized via existing pipeline.

### Final Gate Evaluation: ✓ ALL CHECKS PASS

Phase 1 design confirms no constitution violations. Implementation plan ready for Phase 2 (task generation).

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
cmd/
├── build/main.go          # Static site generator (update to include wedding party page)
├── main/main.go           # Local dev server (update router)
└── lambda/main.go         # AWS Lambda handler (no changes needed)

internal/
├── handlers/
│   ├── homepage.go        # Existing handler pattern
│   ├── venue.go           # Existing handler pattern
│   └── wedding_party.go   # NEW: Wedding party page handler
├── views/
│   ├── app.templ          # Layout wrapper (existing)
│   ├── wedding_party.templ # UPDATE: Convert section to full page
│   └── ...                # Other existing views
├── router.go              # UPDATE: Add /wedding-party route
└── assets/
    └── assets.go          # Static asset handling (existing)

static/
├── images/
│   └── wedding-party/     # NEW: Wedding party member photos
├── css/
│   └── tailwind.css       # Existing (no changes needed)
└── ...

dist/
├── wedding-party.html     # NEW: Generated static page
├── images/
│   └── wedding-party/     # Optimized images copied here
└── ...
```

**Structure Decision**: Single web project following existing architecture. Reuses established patterns:
- Handler in `internal/handlers/wedding_party.go`
- Templ component in `internal/views/wedding_party.templ` (convert existing section to full page)
- Route registration in `internal/router.go`
- Static generation in `cmd/build/main.go`
- No new infrastructure or dependencies required

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**Status**: No complexity violations detected. Feature implementation follows simplest possible approach:
- Uses existing patterns and infrastructure
- No new dependencies
- No architectural changes
- Minimal code additions

This table is intentionally empty as no constitution violations require justification.

---

## Planning Summary

### Completed Phases

- ✅ **Phase 0: Outline & Research** - Analyzed existing codebase, researched implementation patterns, documented decisions in `research.md`
- ✅ **Phase 1: Design & Contracts** - Defined data model (`data-model.md`), documented lack of API contracts (`contracts/README.md`), created quickstart guide (`quickstart.md`)
- ✅ **Agent Context Update** - Updated CLAUDE.md with feature technologies

### Generated Artifacts

| Artifact | Status | Description |
|----------|--------|-------------|
| `spec.md` | ✅ Complete | Feature specification (from /speckit.specify) |
| `plan.md` | ✅ Complete | This implementation plan |
| `research.md` | ✅ Complete | Research findings and technology decisions |
| `data-model.md` | ✅ Complete | Data structures and storage strategy |
| `contracts/README.md` | ✅ Complete | API contracts (none required, explained why) |
| `quickstart.md` | ✅ Complete | Developer implementation guide |
| `tasks.md` | ⏳ Pending | Task breakdown (use /speckit.tasks command) |

### Next Steps

1. **Review planning artifacts** - Ensure all documents are accurate and complete
2. **Run `/speckit.tasks`** - Generate detailed implementation tasks with dependencies
3. **Begin implementation** - Follow quickstart.md for step-by-step guidance
4. **Test locally** - Use `make server` to verify functionality
5. **Deploy** - Use `make deploy` to push to production

### Key Implementation Points

1. **Convert existing component**: Transform `WeddingPartySection()` to full-page `WeddingParty()` component
2. **Create handler**: Add `internal/handlers/wedding_party.go` following existing pattern
3. **Update router**: Register `/wedding-party` route in `internal/router.go`
4. **Add navigation**: Include link in main navigation
5. **Update build script**: Add page generation to `cmd/build/main.go`
6. **Add photos**: Place wedding party photos in `static/images/wedding-party/`
7. **Test responsive layout**: Verify 2-column (desktop) and single-column (mobile) behavior
8. **Generate static site**: Run `make static-build` to create `dist/wedding-party.html`
9. **Deploy**: Run `make deploy` to push to S3/CloudFront

### Technical Approach Summary

- **Language**: Go 1.23.3
- **Templating**: Templ v0.2.793 (type-safe, compile-time HTML)
- **Styling**: Tailwind CSS 3.4.14 (utility-first, responsive)
- **Data Storage**: Go structs (inline in template or separate file)
- **Image Handling**: Existing optimization pipeline (AVIF/WebP/JPEG)
- **Static Generation**: `cmd/build/main.go` (follows existing pattern)
- **Deployment**: Static HTML to S3/CloudFront (no Lambda changes)

### Alignment with Constitution

This feature exemplifies all five constitution principles:

1. **Static-First**: Page pre-generated as HTML, no runtime rendering
2. **Build-Time Optimization**: Templates compiled, CSS purged, images optimized at build time
3. **Simplicity**: Reuses patterns, no new dependencies, minimal code
4. **Infrastructure as Code**: Zero infrastructure changes, existing Terraform unchanged
5. **Performance & Cost**: <2s load time, no additional AWS costs, optimized delivery

### Risk Assessment

**Low Risk**: Feature adds single page using proven patterns. No infrastructure changes, no new dependencies, no database, no API endpoints. Straightforward implementation with minimal surface area for issues.

### Estimated Effort

- **Implementation**: 2-4 hours (handler, template conversion, router update, build script)
- **Content**: 1-2 hours (adding photos, writing descriptions)
- **Testing**: 1 hour (local testing, responsive verification)
- **Deployment**: 15 minutes (build, upload, invalidate cache)
- **Total**: 4-7 hours

---

## Planning Complete

**Status**: ✅ Planning phases complete. Ready for task generation.

**Branch**: `002-wedding-party-page`
**Plan File**: `/Users/drewzer/Projects/thedrewzers/specs/002-wedding-party-page/plan.md`

**Next Command**: `/speckit.tasks` - Generate implementation tasks with dependency ordering
