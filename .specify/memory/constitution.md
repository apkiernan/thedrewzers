<!--
SYNC IMPACT REPORT
==================
Version Change: none → 1.0.0
Change Type: INITIAL - First constitution ratification

Modified Principles: N/A (initial creation)
Added Sections:
  - Static-First Architecture
  - Build-Time Optimization
  - Simplicity & Pragmatism
  - Infrastructure as Code
  - Performance & Cost Efficiency
  - Development Workflow
  - Deployment Strategy
  - Governance

Templates Requiring Updates:
  ✅ plan-template.md - Validated (Constitution Check section compatible)
  ✅ spec-template.md - Validated (Requirements structure compatible)
  ✅ tasks-template.md - Validated (Task organization compatible)
  ✅ checklist-template.md - Not reviewed (not critical for constitution)
  ✅ agent-file-template.md - Not reviewed (not critical for constitution)

Follow-up TODOs: None
-->

# The Drewzers Wedding Website Constitution

## Core Principles

### I. Static-First Architecture

**Rule**: All user-facing content MUST be pre-generated as static HTML and served from S3/CloudFront. Lambda functions SHALL ONLY handle dynamic API operations (e.g., RSVP submissions, contact forms).

**Rationale**: Static-first architecture eliminates Lambda cold starts for page loads, reduces cost, improves performance, and simplifies deployment. Users get instant page loads; dynamic operations are isolated to `/api/*` routes.

**Application**:
- Homepage, venue info, schedule, registry pages → static HTML in `dist/`
- RSVP submission, admin actions → Lambda via API Gateway
- Never render HTML server-side at request time unless absolutely necessary

### II. Build-Time Optimization

**Rule**: All template compilation and asset optimization MUST happen at build time, never at runtime. This includes:
- Templ templates compiled to Go code (`.templ` → `_templ.go`)
- Tailwind CSS purged and minified
- Images optimized and resized
- Static HTML generated via `cmd/build/main.go`

**Rationale**: Build-time optimization ensures zero runtime overhead, predictable performance, and faster deployments. Development remains fast with hot-reload (`air`); production is optimized.

**Application**:
- Use `make static-build` to generate production artifacts
- Never compile templates or process CSS at request time
- Verify build artifacts before deployment (`dist/` directory)

### III. Simplicity & Pragmatism

**Rule**: Choose the simplest solution that works. Avoid frameworks, abstractions, and dependencies unless they solve a clear, documented problem. The project MUST remain maintainable by a single developer.

**Rationale**: This is a wedding website, not a SaaS platform. Over-engineering increases maintenance burden, deployment complexity, and cognitive load. Go's standard library + Templ + Tailwind is sufficient.

**Application**:
- No ORMs (direct SQL if database needed)
- No frontend frameworks (vanilla JS + Tailwind)
- No complex state management (localStorage for simple client state)
- Question every new dependency

### IV. Infrastructure as Code

**Rule**: All AWS infrastructure MUST be managed via Terraform. Manual console changes are prohibited except for emergency debugging (and must be reverted to Terraform afterward).

**Rationale**: Terraform ensures reproducibility, version control, and prevents configuration drift. Disasters are recoverable; changes are auditable.

**Application**:
- S3 bucket configuration → `terraform/`
- CloudFront distribution → `terraform/`
- Lambda function settings → `terraform/`
- Use `make tf-plan` before `make tf-apply`
- Document manual changes in PR and replicate in Terraform

### V. Performance & Cost Efficiency

**Rule**: The website MUST load in under 2 seconds on 3G connections and cost less than $5/month to operate.

**Rationale**: Wedding guests may have limited bandwidth. Static hosting on S3 + CloudFront is nearly free; Lambda invocations only occur for API calls. Poor performance reflects badly on the hosts.

**Application**:
- Images optimized and served via CloudFront CDN
- CSS minified and purged of unused classes
- No large JavaScript bundles (current JS is <5KB)
- Monitor CloudWatch costs monthly

## Development Workflow

### Local Development

- `make server` runs local dev server with hot reload
- `air` watches `.go`, `.templ`, `.css` files
- Changes rebuild automatically
- Test locally before deploying

### Code Quality

- Go code formatted with `gofmt`
- Templates validated with `templ generate`
- CSS built with `npm run build`
- Git commits follow conventional format (e.g., `feat:`, `fix:`, `docs:`)

### Testing Philosophy

- Static HTML generation is tested by building (`make static-build`)
- Visual testing is manual (screenshots in PRs for UI changes)
- API endpoints (when added) require integration tests
- No unit tests for template rendering (Templ is type-safe)

## Deployment Strategy

### Deployment Process

1. **Build**: `make static-build` generates `dist/` artifacts
2. **Upload**: `make upload-static` pushes to S3
3. **Invalidate**: `make invalidate-cache` clears CloudFront
4. **Lambda** (if changed): `make lambda-build` + Terraform deploy

Shortcut: `make deploy` runs all steps.

### Deployment Gates

- Static build completes without errors
- All pages render correctly in `dist/`
- Local testing passed (`make server`)
- Git branch is clean (no uncommitted changes)

### Rollback Strategy

- S3 versioning enabled for static assets
- CloudFront invalidation reverses by re-uploading previous version
- Lambda versions managed via Terraform state
- Keep previous Terraform plan for rollback reference

## Governance

### Amendment Process

1. Propose change via GitHub issue or PR
2. Document rationale and impact on templates
3. Update constitution version per semantic versioning:
   - **MAJOR**: Breaking principle removal/redefinition
   - **MINOR**: New principle or section added
   - **PATCH**: Clarification, wording, typo fixes
4. Update dependent templates (plan/spec/tasks) if needed
5. Merge after approval

### Constitution Supersedes

This constitution supersedes all other practices, documentation, and preferences. If a conflict arises between this document and external guidance:
1. Consult this constitution first
2. Justify any deviation in writing
3. Update constitution if deviation becomes permanent

### Compliance Review

- Review constitution compliance before merging features
- Flag complexity violations (Principle III) in PR comments
- Update constitution when patterns evolve (avoid drift)

### Living Document

This constitution is a living document. Update it when:
- New AWS services are adopted (add to Principle IV)
- Performance budgets change (update Principle V)
- Project complexity grows (revisit Principle III limits)

**Version**: 1.0.0 | **Ratified**: 2025-10-23 | **Last Amended**: 2025-10-23
