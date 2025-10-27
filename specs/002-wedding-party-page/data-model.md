# Data Model: Wedding Party Page

**Feature**: 002-wedding-party-page
**Phase**: 1 (Design & Contracts)
**Date**: 2025-10-27

## Overview

This document defines the data structures for the wedding party page feature. Since this is a static website with build-time generation, data is represented as Go structs compiled into the template at build time.

## Core Entities

### WeddingPartyMember

Represents an individual member of the wedding party.

**Attributes**:

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| Name | string | Yes | Full name of wedding party member | Non-empty string, max 100 characters |
| Role | string | Yes | Role in wedding (e.g., "Best Man", "Maid of Honor", "Bridesmaid", "Groomsman") | Non-empty string, max 50 characters |
| Photo | string | No | Relative path to photo (e.g., "/images/wedding-party/john-doe.jpg") | Valid path format, empty string if no photo |
| Description | string | Yes | Personal anecdote or relationship description | Non-empty string, max 500 characters |
| Side | string | Yes | Which side of wedding party ("groomsmen" or "bridesmaids") | Must be "groomsmen" or "bridesmaids" |

**Example**:
```go
WeddingPartyMember{
    Name:        "Ronnie Campbell",
    Role:        "Best Man",
    Photo:       "/images/wedding-party/ronnie-campbell.jpg",
    Description: "Andrew's longtime friend and college roommate. Known for his legendary barbecue skills and terrible golf swing.",
    Side:        "groomsmen",
}
```

**State Transitions**: N/A (static data, no lifecycle)

### WeddingPartyData

Container for all wedding party members, organized by side.

**Attributes**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Groomsmen | []WeddingPartyMember | Yes | List of groomsmen (including best man) |
| Bridesmaids | []WeddingPartyMember | Yes | List of bridesmaids (including maid(s) of honor) |

**Example**:
```go
WeddingPartyData{
    Groomsmen: []WeddingPartyMember{
        {Name: "Ronnie Campbell", Role: "Best Man", ...},
        {Name: "Mike Alves", Role: "Groomsman", ...},
        {Name: "Dana Roy", Role: "Groomsman", ...},
        {Name: "Mike Silva", Role: "Groomsman", ...},
        {Name: "Pete Smith", Role: "Groomsman", ...},
    },
    Bridesmaids: []WeddingPartyMember{
        {Name: "Melissa Moylan", Role: "Maid of Honor", ...},
        {Name: "Ainsley Kelliher", Role: "Maid of Honor", ...},
        {Name: "Kasey Silva", Role: "Bridesmaid", ...},
        {Name: "Allison Chisholm", Role: "Bridesmaid", ...},
    },
}
```

## Data Flow

### Build Time

```
1. Go source code (data-model.go or inline in template)
   └─> Compile-time: Go compiler validates struct fields
       └─> Templ template generation
           └─> Static HTML output (dist/wedding-party.html)
               └─> Upload to S3
                   └─> Served via CloudFront
```

### Runtime (User Request)

```
User requests /wedding-party.html
└─> CloudFront cache hit (or S3 fetch if miss)
    └─> Static HTML delivered to browser
        └─> Browser renders page
            └─> Images lazy loaded from CloudFront
```

**Note**: No database, no API calls, no server-side logic at request time.

## Data Storage Strategy

### Option 1: Inline in Template (Recommended)

Store data directly in `internal/views/wedding_party.templ`:

```go
package views

templ WeddingParty() {
    @App(weddingPartyContent())
}

templ weddingPartyContent() {
    {{
        groomsmen := []WeddingPartyMember{
            {Name: "Ronnie Campbell", Role: "Best Man", ...},
            // ... more members
        }

        bridesmaids := []WeddingPartyMember{
            {Name: "Melissa Moylan", Role: "Maid of Honor", ...},
            // ... more members
        }
    }}

    <section class="py-16 bg-white">
        <!-- Template rendering logic -->
    </section>
}
```

**Pros**:
- Simplest approach
- Type-safe at compile time
- No external files to manage
- Follows existing pattern in codebase

**Cons**:
- Content updates require template recompile
- Less separation of content and presentation

### Option 2: Separate Go File (Alternative)

Create `internal/views/wedding_party_data.go`:

```go
package views

type WeddingPartyMember struct {
    Name        string
    Role        string
    Photo       string
    Description string
    Side        string
}

func GetWeddingPartyData() (groomsmen, bridesmaids []WeddingPartyMember) {
    groomsmen = []WeddingPartyMember{
        {Name: "Ronnie Campbell", Role: "Best Man", ...},
        // ...
    }

    bridesmaids = []WeddingPartyMember{
        {Name: "Melissa Moylan", Role: "Maid of Honor", ...},
        // ...
    }

    return groomsmen, bridesmaids
}
```

**Pros**:
- Cleaner separation of data and presentation
- Easier to find and update content
- Could be moved to separate package if needed

**Cons**:
- Additional file to manage
- Still requires rebuild/redeploy for updates

**Decision**: Use Option 1 (inline) initially. If content updates become frequent or multiple editors need access, migrate to Option 2.

## Validation Rules

### Compile-Time Validation (Go)

- Non-nil slices (enforced by Go type system)
- Type safety for all fields

### Template-Time Validation (Templ)

- Check for empty Photo field, render default avatar
- Truncate Description if exceeds max length
- Validate Side field, default to "groomsmen" if invalid

### Example Validation Logic

```go
templ renderMember(member WeddingPartyMember) {
    <div class="member-card">
        if member.Photo != "" {
            <img src={ member.Photo } alt={ member.Name } />
        } else {
            <img src="/images/default-avatar.svg" alt="No photo available" />
        }
        <h3>{ member.Name }</h3>
        <p class="role">{ member.Role }</p>
        <p class="description">{ truncate(member.Description, 500) }</p>
    </div>
}
```

## Edge Cases

### Missing Photo
- **Scenario**: WeddingPartyMember with empty Photo field
- **Handling**: Display default avatar image (`/images/default-avatar.svg`)
- **Requirement**: FR-010

### Long Description
- **Scenario**: Description exceeds 500 characters
- **Handling**: Truncate with ellipsis, no scrolling
- **Requirement**: FR-007

### Odd Number of Members
- **Scenario**: One side has 5 members, other has 4
- **Handling**: CSS Grid auto-fills cells, empty space in column is acceptable
- **Requirement**: Edge case from spec

### Empty Wedding Party
- **Scenario**: No members defined (unlikely but possible during development)
- **Handling**: Display "Coming soon" message or hide section
- **Requirement**: Edge case from spec

## Relationships

No relationships between entities. Wedding party members are independent entries with no cross-references.

## Persistence

**Storage**: None (ephemeral at build time)
**Backup**: Version control (Git) serves as backup
**Updates**: Edit source code, rebuild, redeploy

## Performance Considerations

### Memory

- ~10-15 members × ~600 bytes each ≈ 6-9 KB total
- Negligible impact on build process
- Zero runtime memory (static HTML)

### Build Time

- Go compilation: <100ms for struct definitions
- Templ template generation: <500ms
- Static HTML generation: <1s total
- No database queries, no external API calls

## Security

### Injection Risks

- Templ provides automatic HTML escaping
- No user input (all data hardcoded at build time)
- No SQL injection risk (no database)

### Data Privacy

- All data public (served as static HTML)
- No PII beyond names and photos (voluntarily provided by wedding party)
- No GDPR concerns (public event information)

## Future Considerations

### Potential Enhancements (Out of Scope for This Feature)

1. **CMS Integration**: Move data to headless CMS for non-technical content updates
2. **Internationalization**: Add translation support for multilingual weddings
3. **Photo Gallery**: Link each member to dedicated photo gallery
4. **Social Links**: Add Instagram/Facebook profile links
5. **Filtering**: Allow guests to filter by side, role, or name

### Migration Path (if needed)

If content updates become frequent:
1. Extract data to separate package (`internal/data/wedding_party.go`)
2. Consider JSON/YAML for easier editing
3. Add validation tests for data integrity
4. Maintain backward compatibility with existing template

## Implementation Notes

### File Locations

- **Data structures**: `internal/views/wedding_party.templ` (inline) or `internal/views/wedding_party_data.go` (separate)
- **Template**: `internal/views/wedding_party.templ`
- **Handler**: `internal/handlers/wedding_party.go`
- **Generated HTML**: `dist/wedding-party.html`

### Dependencies

- `github.com/a-h/templ` (already in go.mod)
- No additional Go packages required

### Testing Strategy

- **Compile-time**: Go compiler catches type errors
- **Visual testing**: Manual review of generated HTML
- **Accessibility**: Screen reader testing (optional)
- **Performance**: Lighthouse CI (already configured)

## Summary

This data model follows the "simplest thing that works" principle from the project constitution. Wedding party data is stored as Go structs, compiled at build time, and rendered as static HTML. No database, no runtime overhead, no external dependencies. Easy to maintain, performant, and aligns with static-first architecture.
