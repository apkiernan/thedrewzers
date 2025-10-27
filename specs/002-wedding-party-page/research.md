# Research Document: Wedding Party Page

**Feature**: 002-wedding-party-page
**Phase**: 0 (Outline & Research)
**Date**: 2025-10-27

## Overview

This document consolidates research findings for implementing a dedicated wedding party page. The feature converts an existing section component into a standalone page with responsive layout and static generation.

## Research Areas

### 1. Existing Wedding Party Implementation

**Current State**:
- Exists as `WeddingPartySection()` component in `internal/views/wedding_party.templ`
- Uses CSS animations for slideshow effect (cycle-25, cycle-20 classes)
- 2-column grid layout (md:grid-cols-2)
- Hardcoded wedding party data in template
- Placeholder images (picsum.photos)

**Decision**: Convert section to full-page component, preserve layout structure

**Rationale**:
- Existing layout already implements 2-column/single-column responsiveness
- Proven patterns for wedding party display
- Minimal refactoring needed

**Alternatives Considered**:
- Complete rewrite: Rejected - unnecessarily complex, existing layout works well
- Keep as section: Rejected - spec requires dedicated page

### 2. Responsive Layout Approach

**Decision**: Use Tailwind's responsive grid utilities (grid md:grid-cols-2)

**Rationale**:
- Already implemented in existing component
- Mobile-first approach (single column default, 2-column at md breakpoint)
- No additional CSS or JavaScript needed
- Proven performance in existing codebase

**Alternatives Considered**:
- CSS Grid with custom media queries: Rejected - Tailwind utilities simpler
- Flexbox: Rejected - Grid better for equal-height columns
- CSS columns property: Rejected - less control over item distribution

### 3. Wedding Party Data Management

**Decision**: Store wedding party members as Go struct slice in template file

**Rationale**:
- Aligns with Constitution Principle III (Simplicity & Pragmatism)
- No database needed for static content
- Type-safe at compile time via Templ
- Easy content updates without deployment complexity

**Implementation Pattern**:
```go
type WeddingPartyMember struct {
    Name        string
    Role        string
    Photo       string
    Description string
    Side        string // "groomsmen" or "bridesmaids"
}
```

**Alternatives Considered**:
- JSON/YAML data files: Rejected - adds parsing overhead, less type-safe
- Database: Rejected - violates simplicity principle, unnecessary for ~10-15 entries
- CMS integration: Rejected - over-engineering for wedding website

### 4. Image Handling Strategy

**Decision**: Use existing image optimization pipeline

**Rationale**:
- Project already has `cmd/optimize-images/main.go` for AVIF/WebP/JPEG generation
- Responsive image sizes already supported
- `<picture>` elements with `srcset` for optimal loading
- Follows established patterns from hero/gallery images

**Implementation**:
- Store original photos in `static/images/wedding-party/`
- Run `make optimize-images` during build
- Output to `dist/images/wedding-party/`
- Use same lazy loading patterns as existing images

**Alternatives Considered**:
- Direct S3 upload without optimization: Rejected - poor performance
- Client-side image resizing: Rejected - increases bundle size, slower initial load
- Third-party CDN (Cloudinary, Imgix): Rejected - adds cost and external dependency

### 5. Navigation Integration

**Decision**: Add link to existing navigation in `internal/views/app.templ`

**Rationale**:
- Navigation system already exists
- Other pages (venue, gallery, etc.) follow same pattern
- Accessibility requirement (FR-009) satisfied with standard link

**Implementation**:
- Add "Wedding Party" link to nav menu
- Update active state styling for current page
- Mobile menu already responsive

**Alternatives Considered**:
- Dropdown submenu: Rejected - unnecessary complexity for single page
- Footer-only link: Rejected - fails accessibility requirement (>2 clicks)

### 6. Static Generation Pattern

**Decision**: Follow existing pattern in `cmd/build/main.go`

**Rationale**:
- Build command already generates static HTML for other pages
- Proven pattern: render component, write to `dist/wedding-party.html`
- No Lambda cold starts, aligns with static-first architecture

**Implementation Pattern** (from existing pages):
```go
// In cmd/build/main.go
weddingPartyHTML, err := renderToString(views.App(views.WeddingParty()))
if err != nil {
    return err
}
writeFile("dist/wedding-party.html", weddingPartyHTML)
```

**Alternatives Considered**:
- Server-side rendering at request time: Rejected - violates static-first principle
- Client-side rendering: Rejected - poor SEO, slower initial load

### 7. Placeholder Image Handling

**Decision**: Use CSS background or default avatar image

**Rationale**:
- Edge case: wedding party member without photo (spec requirement FR-010)
- Maintain layout consistency
- Graceful degradation

**Implementation**:
- Default avatar SVG in `static/images/`
- Conditional rendering in Templ template
- Alt text indicates missing photo

**Alternatives Considered**:
- Hide member entry: Rejected - loses information about person
- Text-only card: Rejected - inconsistent visual hierarchy

### 8. Content Ordering Strategy

**Decision**: Manual ordering via slice index in Go code

**Rationale**:
- Couple controls display order (typically by importance/closeness)
- Spec assumption: "predefined order"
- Simple implementation

**Implementation**:
- Define slice in desired order
- Optional: Add `Order int` field for future flexibility

**Alternatives Considered**:
- Alphabetical sorting: Rejected - may not reflect importance
- Role-based sorting: Rejected - assumes hierarchy within roles
- Configurable order field: Accepted as future enhancement if needed

## Technology Stack Summary

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Templating | Templ v0.2.793 | Type-safe, compile-time HTML generation |
| Styling | Tailwind CSS 3.4.14 | Utility-first, mobile-responsive, existing patterns |
| Layout | CSS Grid (via Tailwind) | 2-column/single-column responsive behavior |
| Images | AVIF/WebP/JPEG + picture element | Optimal performance across browsers |
| Data Storage | Go struct slice | Simplest solution for static content |
| Routing | Go http.ServeMux | Standard library, no external dependencies |
| Static Generation | cmd/build/main.go | Existing pattern, proven approach |

## Performance Considerations

### Image Optimization
- Use existing pipeline: AVIF (best compression), WebP (good fallback), JPEG (universal)
- Responsive sizes: 640w, 768w, 1024w, 1280w, 1920w, 2560w
- Lazy loading for below-fold images
- Eager loading for above-fold (first 2-4 members)

### Layout Performance
- CSS Grid with Tailwind utilities (no runtime JS)
- Mobile-first approach (single column default)
- Predictable breakpoint at 768px (md:)

### Build-Time Optimization
- Templ templates compiled to Go (zero template parsing overhead)
- Tailwind CSS purged of unused classes
- Static HTML eliminates server-side rendering cost

## Open Questions Resolved

1. **Q**: Should wedding party page be server-rendered or static?
   **A**: Static HTML generated at build time (aligns with constitution)

2. **Q**: How to handle content updates?
   **A**: Update Go slice in template, rebuild, redeploy (standard process)

3. **Q**: Do we need animation for wedding party members?
   **A**: Existing implementation has slideshow animation; preserve for visual interest but ensure it doesn't impact accessibility

4. **Q**: What breakpoint for 2-column layout?
   **A**: 768px (Tailwind md: breakpoint) - standard tablet width, already used in existing component

5. **Q**: Should bridesmaids and groomsmen be separate columns or mixed?
   **A**: Separate columns (existing pattern, clearer organization)

## Dependencies

- No new dependencies required
- Reuses existing tools and libraries
- Existing image optimization pipeline
- Existing Tailwind CSS setup
- Existing static generation workflow

## Next Steps (Phase 1)

1. Generate data-model.md defining WeddingPartyMember struct
2. Define any API contracts (N/A for this feature - purely static)
3. Generate quickstart.md for local development
4. Update agent context with any new patterns (minimal - reuses existing)
