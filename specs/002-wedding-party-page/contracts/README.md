# API Contracts: Wedding Party Page

**Feature**: 002-wedding-party-page
**Date**: 2025-10-27

## No API Contracts Required

This feature has **no API contracts** because it is a purely static page with no dynamic API endpoints.

### Why No APIs?

1. **Static-First Architecture** (Constitution Principle I):
   - Page is pre-generated as static HTML at build time
   - Served directly from S3/CloudFront
   - No server-side rendering at request time

2. **No Dynamic Operations**:
   - Wedding party data is hardcoded in Go templates
   - No CRUD operations (Create, Read, Update, Delete)
   - No user interactions requiring API calls
   - No form submissions

3. **Build-Time Generation**:
   - All data compiled into static HTML via `cmd/build/main.go`
   - Content updates require rebuild and redeploy (not API calls)

## HTTP Routes

While there are no API endpoints, the feature does add one **static route**:

### GET /wedding-party.html

**Type**: Static HTML file
**Description**: Wedding party page displaying all wedding party members
**Response**: Static HTML document
**Cache**: CloudFront CDN (1-day TTL for HTML, 1-year for assets)

**Local Development Route**:
- URL: `http://localhost:8080/wedding-party`
- Handler: `internal/handlers/wedding_party.go::HandleWeddingPartyPage()`
- Renders: `views.App(views.WeddingParty())`

**Production Route**:
- URL: `https://thedrewzers.com/wedding-party.html`
- Source: `dist/wedding-party.html` (pre-generated)
- Served by: S3 bucket + CloudFront

## Static Assets

### Image URLs

Wedding party member photos are served as static assets:

- **Pattern**: `/images/wedding-party/{member-name}.{ext}`
- **Example**: `/images/wedding-party/ronnie-campbell.jpg`
- **Formats**: AVIF, WebP, JPEG (via picture element)
- **Sizes**: 640w, 768w, 1024w, 1280w (responsive srcset)

### Default Avatar

- **URL**: `/images/default-avatar.svg`
- **Usage**: Fallback when WeddingPartyMember.Photo is empty
- **Type**: SVG (scalable, small file size)

## Future API Considerations

If dynamic functionality is added in the future, potential APIs might include:

### Not in Scope (Current Feature)
- `/api/wedding-party` - GET list of members (not needed, data is static)
- `/api/wedding-party/:id` - GET individual member (not needed, no detail pages)
- `/api/wedding-party` - POST/PUT/DELETE (not needed, admin updates via rebuild)

### Possible Future Enhancements
- CMS integration for content updates
- RSVP integration showing guest's friends in wedding party
- Social media links for each member

## Testing

### Local Testing
```bash
# Start dev server
make server

# Visit page
open http://localhost:8080/wedding-party
```

### Static Build Testing
```bash
# Generate static site
make static-build

# Check output
ls -lh dist/wedding-party.html

# Serve locally
cd dist && python3 -m http.server 8000
open http://localhost:8000/wedding-party.html
```

### Production Testing
```bash
# Deploy
make deploy

# Test live
curl -I https://thedrewzers.com/wedding-party.html
# Verify: Status 200, Content-Type: text/html
```

## Summary

No API contracts are defined for this feature because it follows the project's static-first architecture. All content is pre-generated at build time and served as static files. This approach:

- Eliminates Lambda cold starts
- Reduces operational costs
- Improves performance
- Simplifies deployment

API contracts would only be needed if future requirements introduce dynamic operations (e.g., user-submitted content, real-time updates, or administrative interfaces).
