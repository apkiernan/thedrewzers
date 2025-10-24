# Data Model

**Feature**: Website Performance Optimization
**Branch**: 001-performance-optimization
**Status**: Not Applicable

---

## Overview

This performance optimization feature does not introduce new data entities or modify existing data models. All optimizations occur at the asset delivery and rendering layer without requiring data persistence or entity relationships.

---

## Rationale

Performance optimization focuses on:
- **Build-time asset processing** (image optimization, font subsetting)
- **HTTP delivery optimization** (caching headers, compression)
- **Browser rendering optimization** (critical CSS, preloading)

None of these require:
- Database models or schemas
- Entity relationships
- Data validation rules
- State transitions
- Persistence logic

---

## Asset Metadata (Informational Only)

While not true "data entities," the following metadata structures are generated during the build process:

### Image Metadata (Generated in gallery-metadata.json)

**Purpose**: Track responsive image variants and dimensions for template rendering

**Structure** (example):
```json
{
  "filename": "photo-1",
  "width": 4000,
  "height": 6000,
  "aspectRatio": 0.667,
  "gridRowSpan": 12,
  "srcset": {
    "jpeg": "photo-1-640w.jpg 640w, photo-1-1280w.jpg 1280w, ...",
    "webp": "photo-1-640w.webp 640w, photo-1-1280w.webp 1280w, ...",
    "avif": "photo-1-640w.avif 640w, photo-1-1280w.avif 1280w, ..."
  },
  "lqip": "photo-1-lqip.jpg"
}
```

**Lifecycle**:
- Generated during `make optimize-images`
- Consumed by Templ templates during `make static-build`
- Not persisted beyond build artifacts

**Not a Data Model Because**:
- Ephemeral (regenerated on every build)
- No persistence layer or database
- No CRUD operations
- Build-time only metadata

---

## Cache Headers (Configuration, Not Data)

**CloudFront/S3 Cache Configuration**:
- Cache-Control headers for static assets
- ETag generation for cache validation
- TTL settings for different asset types

**Managed via**: Terraform (Infrastructure as Code), not data models

---

## Conclusion

**No data model is required** for this feature. All changes are infrastructure, build pipeline, and asset delivery optimizations that operate independently of application data.

If future performance features require data tracking (e.g., analytics, performance metrics storage), this document should be updated with appropriate entity definitions.

---

**Document Version**: 1.0
**Last Updated**: 2025-10-23
**Status**: Not Applicable
