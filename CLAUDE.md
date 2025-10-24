# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Wedding website built with Go, Templ templating, and Tailwind CSS. Deployable to AWS Lambda with static assets served from S3.

## Key Commands

### Development
- `make server` - Run local dev server with hot reload on port 8080
- `make all` - Build templates, compile CSS, and start server
- `make tpl` - Generate Go code from templ templates
- `make styles` - Watch and compile Tailwind CSS

### Static Site Generation
- `make static-build` - Generate static HTML files in ./dist/
- `make static-deploy` - Upload static site to S3 and invalidate CloudFront cache
- `make invalidate-cache` - Manually invalidate CloudFront cache

### Deployment
- `make deploy` - Build Lambda, generate static site, and deploy everything
- `make lambda-build` - Build Lambda binary (linux/amd64)
- `make upload-static` - Upload only static assets to S3
- `make tf-plan` - Preview Terraform changes
- `make tf-apply` - Apply Terraform changes

### Other Commands
- `make clean` - Remove build artifacts
- `npm run build` - Build production CSS
- `npm run watch` - Watch CSS changes

## Architecture

### Entry Points
- **Local Dev**: `cmd/main/main.go` - Runs on port 8080 with static file serving
- **Lambda**: `cmd/lambda/main.go` - AWS Lambda handler for API routes only (`/api/*`)
- **Static Build**: `cmd/build/main.go` - Generates static HTML files from templates

### Core Structure
- **Handlers**: `internal/handlers/` - HTTP request handlers (homepage.go, venue.go)
- **Templates**: `internal/views/*.templ` - Templ components compiled to Go
- **Static Assets**: `static/` - CSS, fonts, images served locally or from S3
- **Generated Site**: `dist/` - Static HTML output from build process

### Key Patterns
1. **Static-First Architecture**: 
   - HTML pages served directly from S3 via CloudFront
   - Lambda only handles API requests (`/api/*`)
   - No Lambda cold starts for page loads
2. **Templ Templates**: Type-safe HTML generation at build time
3. **Client-Side State**: First-time visitor overlay managed via localStorage
4. **API Configuration**: `static/js/config.js` defines API endpoint

### AWS Infrastructure (Terraform)
- **S3 Bucket**: Static website hosting for HTML/CSS/JS
- **CloudFront**: CDN serving static site, routes `/api/*` to Lambda
- **Lambda Function**: Handles API requests only (future RSVP functionality)
- **API Gateway**: HTTP API for Lambda integration

### Development Flow
1. `air` watches .go, .templ, .css files
2. Auto-runs `templ generate` and rebuilds on changes
3. Templates compile .templ → _templ.go files
4. Tailwind processes CSS via npm scripts

## Important Notes
- Lambda binary must be named `bootstrap` for custom runtime
- Static paths in Lambda redirect to S3 bucket
- Default S3 bucket: `thedrewzers-wedding-static`
- Default region: `us-east-1`
- CloudFront invalidation: Set `CLOUDFRONT_DISTRIBUTION_ID` env var or let it auto-detect from Terraform outputs

## Active Technologies
- Go 1.23.3, Node.js (Tailwind CSS 3.4.14) (001-performance-optimization)
- AWS S3 (static assets), CloudFront CDN (distribution), no database (001-performance-optimization)

## Recent Changes
- 001-performance-optimization: Added Go 1.23.3, Node.js (Tailwind CSS 3.4.14)
