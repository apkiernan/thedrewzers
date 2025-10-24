# Performance Optimization Setup Instructions

**Status**: Partially Complete - Manual Installation Required

## Required Tool Installation

### Image Optimization Tools

You need to install the following CLI tools for image optimization:

```bash
# Install AVIF encoder (MISSING - REQUIRED)
brew install libavif

# Verify installation
cwebp --version   # ✅ Already installed
avifenc --version # ⚠️  Needs installation
cjpeg --version   # ✅ Already installed (mozjpeg)
```

### Font Optimization Tools

```bash
# Install glyphhanger globally via npm
npm install -g glyphhanger

# Install Python fonttools for font subsetting (using uv - much faster than pip)
uv pip install fonttools brotli

# Or install as a standalone tool
# uv tool install fonttools

# Verify installation
glyphhanger --version
pyftsubset --help
```

### Performance Monitoring

```bash
# Install Lighthouse CI
npm install -g @lhci/cli

# Verify installation
lhci --version
```

## Verification Commands

After installing all tools, run:

```bash
# Verify all tools are available
which cwebp avifenc cjpeg glyphhanger lhci
pyftsubset --help

# All commands should return paths or help text
```

## Next Steps

Once all tools are installed:
1. Mark T001-T003 as complete in `tasks.md`
2. Continue with directory structure creation (T004)
3. Run final verification (T005)

---

**Current Status**:
- ✅ cwebp installed
- ✅ cjpeg (mozjpeg) installed
- ⚠️  avifenc needs installation
- ⚠️  glyphhanger needs installation
- ⚠️  fonttools needs installation
- ⚠️  Lighthouse CI needs installation
