# Node-Free Frontend Setup

This project uses a **Node.js-free** frontend stack with:
- **Tailwind CSS Standalone CLI** (single binary, no npm)
- **Alpine.js via CDN** (v3.14.3)
- **HTMX via CDN** (v2.0.4)
- **Chart.js via CDN** (v4.4.7)

## Quick Start

### 1. Install Tailwind CSS Standalone CLI

The easiest way to install is using the Makefile:

```bash
make install-tailwind
```

Or manually download the binary:

**Linux x64:**
```bash
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
chmod +x tailwindcss-linux-x64
sudo mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss
```

**Linux ARM64:**
```bash
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64
chmod +x tailwindcss-linux-arm64
sudo mv tailwindcss-linux-arm64 /usr/local/bin/tailwindcss
```

**macOS (Apple Silicon):**
```bash
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64
sudo mv tailwindcss-macos-arm64 /usr/local/bin/tailwindcss
```

**macOS (Intel):**
```bash
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64
chmod +x tailwindcss-macos-x64
sudo mv tailwindcss-macos-x64 /usr/local/bin/tailwindcss
```

**Windows:**
Download from: https://github.com/tailwindlabs/tailwindcss/releases/latest

### 2. Build CSS

**One-time build:**
```bash
make css-build
# or
./scripts/build-css.sh
```

**Watch mode (for development):**
```bash
make css-watch
# or
./scripts/build-css.sh --watch
```

## Development Workflow

### Option 1: Using Make (Recommended)

Run in two separate terminals:

**Terminal 1 - CSS Watch:**
```bash
make css-watch
```

**Terminal 2 - Go App:**
```bash
make run
```

### Option 2: Manual Commands

**Terminal 1:**
```bash
tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch
```

**Terminal 2:**
```bash
go run cmd/api/main.go
```

## Project Structure

```
web/
├── ui/
│   ├── base.templ              # Base layout with CDN scripts
│   ├── home.templ
│   ├── estimator.templ
│   └── components/             # Reusable components
│       ├── core/
│       ├── data/
│       ├── feedback/
│       └── navigation/
└── static/
    └── css/
        ├── input.css           # Source CSS with Tailwind directives
        └── output.css          # Generated CSS (gitignored)
```

## Configuration Files

### tailwind.config.js
- Uses CommonJS (`module.exports`) - works with standalone CLI
- NO Node.js plugins (removed @tailwindcss/forms and @tailwindcss/typography)
- Scans `.templ` files and generated `*_templ.go` files

### web/static/css/input.css
- Contains Tailwind directives: `@tailwind base`, `@tailwind components`, `@tailwind utilities`
- Includes custom CSS classes (buttons, stats, panels, etc.)

## CDN Libraries

All JavaScript libraries are loaded via CDN (no build step required):

| Library | Version | CDN |
|---------|---------|-----|
| HTMX | 2.0.4 | unpkg.com |
| Alpine.js | 3.14.3 | cdn.jsdelivr.net |
| Chart.js | 4.4.7 | cdn.jsdelivr.net |

## Makefile Commands

```bash
make help              # Show all available commands
make install-tailwind  # Install Tailwind CLI
make css-build         # Build CSS once (minified)
make css-watch         # Watch and rebuild CSS on changes
make css               # Alias for css-build
make dev               # Show dev setup instructions
```

## Benefits of This Approach

1. ✅ **No Node.js required** - Just Go + single binary
2. ✅ **Fast builds** - Standalone CLI is written in Rust
3. ✅ **Simple CI/CD** - One binary to install
4. ✅ **No package.json** - No dependency hell
5. ✅ **CDN for JS** - Always up-to-date libraries
6. ✅ **Templ-native** - Works seamlessly with Go templates

## Troubleshooting

### Tailwind CLI not found
```bash
which tailwindcss
# If not found, run: make install-tailwind
```

### Output CSS not updating
```bash
# Check if watch mode is running
ps aux | grep tailwindcss

# Restart watch mode
make css-watch
```

### Styles not applying
1. Ensure `output.css` is generated
2. Check browser console for 404 errors
3. Verify base.templ links to `/static/css/output.css`
4. Clear browser cache

## Production Build

For production, build minified CSS once:

```bash
make css-build
```

Then deploy:
- Go binary
- `web/static/css/output.css`
- Other static assets

The CDN scripts will be loaded by the browser at runtime.

## Why No Node.js?

This project deliberately avoids Node.js to:
- Keep the tech stack simple (Go + Templ)
- Reduce build complexity
- Avoid npm security issues
- Make deployment easier
- Leverage CDN for JS libraries

The Tailwind standalone CLI provides all CSS processing capabilities without Node.js.
