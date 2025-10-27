# Quickstart Guide: Wedding Party Page

**Feature**: 002-wedding-party-page
**Branch**: `002-wedding-party-page`
**Date**: 2025-10-27

## Overview

This guide will help you implement, test, and deploy the wedding party page feature. Follow these steps in order for a smooth development experience.

## Prerequisites

- Go 1.23.3 installed
- Node.js and npm installed (for Tailwind CSS)
- Git repository cloned
- Existing wedding website codebase running locally
- Feature branch `002-wedding-party-page` checked out

## Quick Start (TL;DR)

```bash
# 1. Ensure you're on the feature branch
git checkout 002-wedding-party-page

# 2. Create wedding party handler
# (See Step-by-Step Implementation below)

# 3. Update router
# (See Step-by-Step Implementation below)

# 4. Update template
# (See Step-by-Step Implementation below)

# 5. Add navigation link
# (See Step-by-Step Implementation below)

# 6. Test locally
make server
open http://localhost:8080/wedding-party

# 7. Build static site
make static-build

# 8. Deploy
make deploy
```

## Step-by-Step Implementation

### Step 1: Create Wedding Party Handler

**File**: `internal/handlers/wedding_party.go`

```go
package handlers

import (
    "net/http"
    "github.com/apkiernan/thedrewzers/internal/views"
)

func HandleWeddingPartyPage(w http.ResponseWriter, r *http.Request) {
    views.App(views.WeddingParty()).Render(r.Context(), w)
}
```

### Step 2: Update Router

**File**: `internal/router.go`

```go
package router

import (
    "net/http"
    "github.com/apkiernan/thedrewzers/internal/handlers"
)

func Router() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handlers.HandleHomePage)
    mux.HandleFunc("/wedding-party", handlers.HandleWeddingPartyPage) // ADD THIS LINE
    return mux
}
```

### Step 3: Update Wedding Party Template

**File**: `internal/views/wedding_party.templ`

Convert the existing `WeddingPartySection()` component to a full-page `WeddingParty()` component.

**Before** (section component):
```go
templ WeddingPartySection() {
    <section id="wedding-party" class="py-16 bg-white">
        <!-- content -->
    </section>
}
```

**After** (full page component):
```go
package views

type WeddingPartyMember struct {
    Name        string
    Role        string
    Photo       string
    Description string
    Side        string
}

templ WeddingParty() {
    <div class="min-h-screen bg-white">
        <section id="wedding-party" class="py-16">
            <div class="max-w-4xl mx-auto px-6 mb-8">
                <h1 class="script text-5xl md:text-6xl text-blue-300 mb-12 font-extralight text-center">Our Wedding Party</h1>

                <div class="grid md:grid-cols-2 gap-16">
                    <!-- Groomsmen Column -->
                    <div>
                        <h2 class="text-lg uppercase tracking-wider text-gray-500 mb-4">Groomsmen</h2>
                        { renderMembers(getGroomsmen()) }
                    </div>

                    <!-- Bridesmaids Column -->
                    <div>
                        <h2 class="text-lg uppercase tracking-wider text-gray-500 mb-4">Bridesmaids</h2>
                        { renderMembers(getBridesmaids()) }
                    </div>
                </div>
            </div>
        </section>
    </div>
}

templ renderMembers(members []WeddingPartyMember) {
    for _, member := range members {
        @renderMember(member)
    }
}

templ renderMember(member WeddingPartyMember) {
    <div class="flex flex-col items-center text-center rounded-lg p-6 shadow-sm mb-8">
        if member.Photo != "" {
            <img
                src={ member.Photo }
                alt={ member.Name }
                class="rounded-full w-32 h-32 object-cover shadow-md"
                width="128"
                height="128"
                loading="lazy"
            />
        } else {
            <img
                src="/images/default-avatar.svg"
                alt="No photo available"
                class="rounded-full w-32 h-32 object-cover shadow-md"
                width="128"
                height="128"
            />
        }
        <div class="mt-4">
            if member.Role != "" {
                <p class="text-sm uppercase tracking-wider text-gray-500">{ member.Role }</p>
            }
            <p class="text-gray-700 text-lg font-medium mt-1">{ member.Name }</p>
            <p class="text-gray-500 text-sm mt-2 leading-relaxed">{ member.Description }</p>
        </div>
    </div>
}

func getGroomsmen() []WeddingPartyMember {
    return []WeddingPartyMember{
        {
            Name:        "Ronnie Campbell",
            Role:        "Best Man",
            Photo:       "/images/wedding-party/ronnie-campbell.jpg",
            Description: "Andrew's longtime friend and college roommate.",
            Side:        "groomsmen",
        },
        {
            Name:        "Mike Alves",
            Role:        "Groomsman",
            Photo:       "/images/wedding-party/mike-alves.jpg",
            Description: "Andrew's childhood friend from the neighborhood.",
            Side:        "groomsmen",
        },
        // Add more groomsmen...
    }
}

func getBridesmaids() []WeddingPartyMember {
    return []WeddingPartyMember{
        {
            Name:        "Melissa Moylan",
            Role:        "Maid of Honor",
            Photo:       "/images/wedding-party/melissa-moylan.jpg",
            Description: "Molly's closest confidant since college.",
            Side:        "bridesmaids",
        },
        {
            Name:        "Ainsley Kelliher",
            Role:        "Maid of Honor",
            Photo:       "/images/wedding-party/ainsley-kelliher.jpg",
            Description: "Molly's best friend since high school.",
            Side:        "bridesmaids",
        },
        // Add more bridesmaids...
    }
}
```

### Step 4: Compile Templ Templates

```bash
# Generate Go code from .templ files
templ generate

# Or use make command (runs templ generate + builds)
make tpl
```

### Step 5: Add Navigation Link

**File**: `internal/views/app.templ`

Find the navigation section and add a link to the wedding party page:

```html
<!-- Existing nav -->
<nav>
    <a href="/">Home</a>
    <a href="/venue">Venue</a>
    <a href="/wedding-party">Wedding Party</a> <!-- ADD THIS -->
    <a href="/rsvp">RSVP</a>
    <!-- ... -->
</nav>
```

### Step 6: Add Wedding Party Photos

Create directory and add photos:

```bash
# Create directory
mkdir -p static/images/wedding-party

# Add photos (provided by couple)
# Format: firstname-lastname.jpg
# Example: ronnie-campbell.jpg
```

### Step 7: Update Build Script

**File**: `cmd/build/main.go`

Add wedding party page generation to the build process:

```go
// Existing imports...

func main() {
    // ... existing code ...

    // Generate wedding party page
    if err := generateWeddingPartyPage(); err != nil {
        log.Fatalf("Failed to generate wedding party page: %v", err)
    }

    // ... rest of build process ...
}

func generateWeddingPartyPage() error {
    file, err := os.Create("dist/wedding-party.html")
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer file.Close()

    component := views.App(views.WeddingParty())
    if err := component.Render(context.Background(), file); err != nil {
        return fmt.Errorf("render template: %w", err)
    }

    log.Println("✓ Generated wedding-party.html")
    return nil
}
```

## Testing

### Local Development Testing

```bash
# Start local dev server with hot reload
make server

# Open in browser
open http://localhost:8080/wedding-party

# Verify:
# - Page loads without errors
# - 2-column layout on desktop (>768px)
# - Single column layout on mobile (<768px)
# - All wedding party members display
# - Photos load correctly
# - Descriptions are readable
# - Navigation link works
```

### Static Build Testing

```bash
# Generate static site
make static-build

# Check output file exists
ls -lh dist/wedding-party.html

# Serve locally from dist/
cd dist
python3 -m http.server 8000

# Open in browser
open http://localhost:8000/wedding-party.html

# Verify same items as above
```

### Responsive Testing

Test the page at different breakpoints:

- **Mobile**: 320px, 375px, 414px (single column)
- **Tablet**: 768px, 1024px (2-column transition)
- **Desktop**: 1280px, 1920px (2-column)

Use browser DevTools device emulation or resize window.

### Performance Testing

```bash
# Run Lighthouse CI
npm run lighthouse

# Check metrics:
# - Performance score > 90
# - LCP < 2.5s
# - CLS < 0.1
# - No accessibility errors
```

## Deployment

### Full Deployment

```bash
# Build static site, upload to S3, invalidate CloudFront
make deploy

# Verify deployment
curl -I https://thedrewzers.com/wedding-party.html
# Expected: Status 200, Content-Type: text/html
```

### Incremental Deployment

```bash
# If only static files changed (no Lambda updates)
make static-build
make upload-static
make invalidate-cache
```

### Rollback

If issues occur after deployment:

```bash
# 1. Revert to previous commit
git revert HEAD

# 2. Rebuild and redeploy
make deploy

# 3. Or: manually restore previous version from S3 versioning
aws s3api list-object-versions --bucket thedrewzers-wedding-static --prefix wedding-party.html
aws s3api copy-object --copy-source ... --bucket ...
```

## Common Issues & Solutions

### Issue: Page returns 404

**Cause**: Route not registered or static file not generated

**Solution**:
```bash
# Check router.go has /wedding-party route
grep "wedding-party" internal/router.go

# Check build generated the file
ls dist/wedding-party.html

# Rebuild if missing
make static-build
```

### Issue: Photos not loading

**Cause**: Photos not in correct directory or incorrect paths

**Solution**:
```bash
# Check photo paths
ls static/images/wedding-party/

# Verify paths in template match actual filenames
# Example: "ronnie-campbell.jpg" not "Ronnie Campbell.jpg"

# Ensure photos are copied to dist/ during build
ls dist/images/wedding-party/
```

### Issue: Layout broken on mobile

**Cause**: Tailwind CSS not compiled or breakpoints incorrect

**Solution**:
```bash
# Rebuild Tailwind CSS
npm run build

# Check responsive classes in template
# Should use: grid md:grid-cols-2 (not just grid-cols-2)
```

### Issue: Templ compilation errors

**Cause**: Syntax errors in .templ file

**Solution**:
```bash
# Run templ generate to see errors
templ generate

# Common issues:
# - Missing closing braces
# - Invalid Go syntax in template functions
# - Type mismatches in struct fields
```

## Next Steps

After implementation and testing:

1. **Review**: Check code against constitution principles
2. **Commit**: `git add . && git commit -m "feat: add wedding party page"`
3. **Push**: `git push origin 002-wedding-party-page`
4. **PR**: Create pull request for review
5. **Deploy**: Merge and deploy to production

## Development Tips

### Hot Reload

The `air` tool (used by `make server`) watches for changes to:
- `*.go` files
- `*.templ` files
- `*.css` files

Save any file to trigger automatic rebuild and browser refresh.

### Templ Best Practices

- Keep components small and focused
- Use helper functions for repeated logic
- Extract data to separate functions (getGroomsmen, getBridesmaids)
- Leverage Go's type safety for compile-time checks

### Tailwind Best Practices

- Use utility classes (don't write custom CSS)
- Mobile-first responsive design (base → md: → lg:)
- Consistent spacing scale (p-4, p-6, p-8, etc.)
- Reuse existing color palette (text-blue-300, text-gray-500)

### Performance Optimization

- Use `loading="lazy"` for below-fold images
- Use `loading="eager"` for above-fold images
- Consider `fetchpriority="high"` for critical images
- Optimize images before adding to static/images/ (or use make optimize-images)

## Resources

- **Templ Docs**: https://templ.guide/
- **Tailwind CSS Docs**: https://tailwindcss.com/docs
- **Go Templates**: https://pkg.go.dev/html/template
- **Project README**: /README.md
- **CLAUDE.md**: Project-specific instructions for Claude Code

## Support

For questions or issues:
1. Review this quickstart guide
2. Check CLAUDE.md for project conventions
3. Review existing handlers (homepage.go, venue.go) for patterns
4. Consult feature spec (spec.md) for requirements
5. Ask team for help if stuck

---

**Happy coding!** This feature is a great example of the project's static-first architecture and simplicity principles in action.
