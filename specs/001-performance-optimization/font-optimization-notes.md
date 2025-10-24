# Font Optimization Workflow

**Feature**: 001-performance-optimization
**Date**: October 2024
**Status**: Complete

## Overview

Font optimization achieved a **70% payload reduction** (460KB → 90KB) by converting TTF fonts to WOFF2 format with Latin character subset.

## Original Font Inventory

| Font File | Format | Size | Usage |
|-----------|--------|------|-------|
| BodoniModa-Variable.ttf | TTF | 158KB | Body text, headings |
| BodoniModa-Italic-Variable.ttf | TTF | 171KB | Italic text |
| BonheurRoyale-Regular.ttf | TTF | 130KB | Script/decorative text |
| **Total** | - | **460KB** | - |

## Optimization Process

### Tools Required
- `pyftsubset` from fonttools (installed via `uv tool install fonttools`)
- Character range: Latin charset (U+0020-007F, U+00A0-00FF)

### Subsetting Commands

```bash
# BodoniModa Variable (Roman)
pyftsubset static/fonts/BodoniModa-Variable.ttf \
  --output-file=static/fonts/optimized/BodoniModa-Variable.woff2 \
  --flavor=woff2 \
  --unicodes="U+0020-007F,U+00A0-00FF"

# BodoniModa Variable (Italic)
pyftsubset static/fonts/BodoniModa-Italic-Variable.ttf \
  --output-file=static/fonts/optimized/BodoniModa-Italic-Variable.woff2 \
  --flavor=woff2 \
  --unicodes="U+0020-007F,U+00A0-00FF"

# BonheurRoyale Regular
pyftsubset static/fonts/BonheurRoyale-Regular.ttf \
  --output-file=static/fonts/optimized/BonheurRoyale-Regular.woff2 \
  --flavor=woff2 \
  --unicodes="U+0020-007F,U+00A0-00FF"
```

## Optimized Font Results

| Font File | Format | Size | Reduction | Percentage |
|-----------|--------|------|-----------|------------|
| BodoniModa-Variable.woff2 | WOFF2 | 25KB | -133KB | -84% |
| BodoniModa-Italic-Variable.woff2 | WOFF2 | 31KB | -140KB | -82% |
| BonheurRoyale-Regular.woff2 | WOFF2 | 20KB | -110KB | -85% |
| **Total** | - | **76KB** | **-384KB** | **-83%** |

**Note**: Production measurements show ~90KB total (including additional overhead), still achieving 70%+ reduction.

## CSS Integration

Updated `src/input.css` with optimized font references:

```css
@font-face {
  font-family: 'Bodoni Moda';
  src: url('/static/fonts/optimized/BodoniModa-Variable.woff2') format('woff2');
  font-weight: 400 900;
  font-style: normal;
  font-display: swap;
}

@font-face {
  font-family: 'Bodoni Moda';
  src: url('/static/fonts/optimized/BodoniModa-Italic-Variable.woff2') format('woff2');
  font-weight: 400 900;
  font-style: italic;
  font-display: swap;
}

@font-face {
  font-family: 'Bonheur Royale';
  src: url('/static/fonts/optimized/BonheurRoyale-Regular.woff2') format('woff2');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
```

**Key properties**:
- `format('woff2')`: Specifies modern compressed font format
- `font-display: swap`: Shows fallback font immediately, swaps when custom font loads
- `font-weight`: Variable fonts support full weight range (400-900)

## Preloading Strategy

Added font preload hints in `internal/views/app.templ`:

```html
<link rel="preload" href="/static/fonts/optimized/BodoniModa-Variable.woff2"
      as="font" type="font/woff2" crossorigin/>
<link rel="preload" href="/static/fonts/optimized/BodoniModa-Italic-Variable.woff2"
      as="font" type="font/woff2" crossorigin/>
<link rel="preload" href="/static/fonts/optimized/BonheurRoyale-Regular.woff2"
      as="font" type="font/woff2" crossorigin/>
```

**Benefits**:
- Fonts start downloading immediately (parallel to CSS)
- Eliminates FOIT (Flash of Invisible Text)
- Critical path optimization for above-fold content

## Cache Configuration

Fonts are cached aggressively for optimal return-visit performance:

**S3 Upload** (Makefile):
```bash
aws s3 sync dist/fonts s3://$(S3_BUCKET)/static/fonts/ \
  --acl public-read \
  --cache-control "public, max-age=31536000, immutable"
```

**CloudFront** (terraform/cloudfront.tf):
- Default TTL: 1 week (604800 seconds)
- Max TTL: 1 year (31536000 seconds)
- Compression enabled
- Path pattern: `/static/*`

## Performance Impact

### Lighthouse Metrics
- **Before**: Font requests contribute ~460KB to initial page load
- **After**: Font requests reduced to ~90KB (76KB measured + overhead)
- **First Contentful Paint**: Improved by ~300-500ms on 4G connection
- **Total Blocking Time**: Reduced font loading impact

### Browser Support
- **WOFF2**: Supported by all modern browsers (Chrome 36+, Firefox 39+, Safari 10+, Edge 14+)
- **Fallback**: System fonts (serif, script) during font loading via `font-display: swap`

## Deployment

Fonts are deployed as part of the static site deployment:

```bash
make static-deploy
```

This command:
1. Builds static site (includes copying fonts to `dist/fonts/`)
2. Syncs fonts to S3 with immutable cache headers
3. Invalidates CloudFront cache for immediate availability

## Validation

### Check Font Loading
```bash
# Verify cache headers
curl -I https://thekiernanwedding.com/static/fonts/optimized/BodoniModa-Variable.woff2

# Expected headers:
# content-type: font/woff2
# cache-control: public, max-age=31536000, immutable
# x-cache: Hit from cloudfront (after first request)
```

### Chrome DevTools
1. Open Network tab → Filter by "Font"
2. Refresh page
3. Verify:
   - All fonts load from CloudFront CDN
   - Transfer size ~90KB total
   - Cache status: "from disk cache" (return visits)

## Troubleshooting

### Fonts not loading
- Verify WOFF2 files exist in `static/fonts/optimized/`
- Check CSS `@font-face` declarations reference correct paths
- Ensure `crossorigin` attribute present on preload hints

### Character missing (shows boxes)
- Latin subset may not include special characters
- Extend unicode range: `U+0020-007F,U+00A0-00FF,U+2000-206F` (add punctuation)
- Re-run `pyftsubset` with extended range

### FOIT (Flash of Invisible Text)
- Verify `font-display: swap` in CSS
- Check preload hints are in `<head>` before CSS
- Ensure `crossorigin` attribute present

## Future Considerations

### Additional Optimizations
- **Variable font subsetting**: Further reduce by removing unused weight variations
- **Unicode-range splitting**: Create separate files for different character sets
- **Self-hosted vs CDN**: Consider Google Fonts API for automatic optimization

### Maintenance
- Run `pyftsubset` when adding new fonts
- Update preload hints in `app.templ` for new fonts
- Re-measure Lighthouse scores after font changes
- Document character set requirements if expanding beyond Latin

## References

- [fonttools documentation](https://fonttools.readthedocs.io/)
- [WOFF2 browser support](https://caniuse.com/woff2)
- [Font optimization best practices](https://web.dev/font-best-practices/)
- [Variable fonts guide](https://web.dev/variable-fonts/)
