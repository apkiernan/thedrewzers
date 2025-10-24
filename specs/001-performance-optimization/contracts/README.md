# API Contracts

**Feature**: Website Performance Optimization
**Branch**: 001-performance-optimization
**Status**: Not Applicable

---

## Overview

This performance optimization feature does not introduce new API endpoints or modify existing API contracts. All optimizations are focused on static asset delivery and browser rendering performance.

---

## Rationale

Performance optimization affects:

- **Static HTML/CSS/JS/Image delivery** (no API contracts)
- **CDN caching and compression** (infrastructure configuration)
- **Build-time asset processing** (CLI tools and build scripts)

The feature does not:

- Add new HTTP API endpoints
- Modify existing Lambda function contracts
- Change GraphQL schemas
- Introduce WebSocket connections
- Add new request/response formats

---

## Existing API Contracts (Unchanged)

The website currently has the following API structure:

### Lambda Function (API Routes Only)

**Base Path**: `/api/*`

**Current Status**: API routes are defined but not actively used (prepared for future RSVP functionality)

**Performance Optimization Impact**: None - Lambda handles dynamic API operations only, not static page delivery

---

## CloudFront Configuration (Infrastructure, Not API Contract)

While CloudFront behaviors are configured via Terraform, these are **infrastructure settings**, not API contracts:

- **Static Assets** (`/static/*`) → S3 bucket direct serving
- **HTML Pages** (`/*.html`, `/`) → S3 bucket direct serving
- **API Routes** (`/api/*`) → Lambda function integration

**Changes in This Feature**:

- Update cache headers for static assets
- Configure compression (gzip/brotli)
- Optimize TTL settings

These are **HTTP header configurations**, not API contract changes.

---

## Future Considerations

If performance monitoring or analytics features are added in the future (e.g., `/api/metrics`, `/api/performance-logs`), API contracts should be documented in this directory using OpenAPI/Swagger specification.

**Example structure for future API contracts**:

```
contracts/
├── openapi.yaml           # OpenAPI 3.0 specification
├── metrics-api.md         # Human-readable API documentation
└── performance-api.md     # Performance monitoring endpoints
```

---

## Conclusion

**No API contracts are defined or modified** by this performance optimization feature. This directory is included for template compliance but remains empty (except for this README) as no API changes are required.

---

**Document Version**: 1.0
**Last Updated**: 2025-10-23
**Status**: Not Applicable
