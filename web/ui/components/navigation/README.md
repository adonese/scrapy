# Navigation Components

A comprehensive set of navigation components built with **Templ + Alpine.js + HTMX + Tailwind CSS**. All components are fully responsive, accessible, and support dark mode out of the box.

## Components Overview

### 1. Navbar (`navbar_complete.templ`)

A fully responsive navigation bar with theme toggle and mobile menu support.

**Features:**
- Responsive design (desktop full menu, mobile hamburger)
- Dark mode toggle with localStorage persistence
- Sticky header option
- Status badge display
- Accessible keyboard navigation
- ARIA labels for screen readers

**Usage:**
```go
@NavbarComplete(NavbarCompleteProps{
    LogoText:    "My Application",
    Subtitle:    "Built with Templ",
    ShowStatus:  true,
    StatusText:  "Live",
    StatusColor: "green", // green, blue, red, yellow
    NavLinks: []NavLink{
        {Href: "/", Text: "Home", Label: "Navigate to home page"},
        {Href: "/about", Text: "About", Label: "Navigate to about page"},
        {Href: "/products", Text: "Products", Label: "Navigate to products page"},
    },
})
```

**Dark Mode:**
The theme toggle uses Alpine.js to manage dark mode state with localStorage persistence and respects system preferences.

---

### 2. Mobile Menu (`mobile-menu-complete.templ`)

A slide-out mobile menu with smooth transitions and gestures.

**Features:**
- Slide-out drawer from left side
- Backdrop overlay with click-to-close
- Swipe gesture support (swipe left to close)
- Smooth CSS transitions
- Auto-close on navigation
- Keyboard accessible (ESC to close)

**Usage:**
```go
@MobileMenuComplete(
    []NavLink{
        {Href: "/", Text: "Home", Label: "Go to home"},
        {Href: "/products", Text: "Products", Label: "View products"},
    },
    "MyApp",        // Logo text
    "Navigation",   // Subtitle
)
```

**Variants:**
- `MobileMenuComplete` - Full-featured with branding header
- `MobileMenuSimple` - Minimal version without header

---

### 3. Breadcrumb (`breadcrumb.templ`)

Breadcrumb navigation with responsive design.

**Features:**
- Automatic home icon as first item
- Current page highlighting
- Responsive (compact mode for mobile)
- Support for icons/emojis
- Keyboard navigation
- Proper ARIA labels

**Usage:**
```go
@Breadcrumb(BreadcrumbProps{
    Items: []BreadcrumbItem{
        {Text: "Products", Href: "/products"},
        {Text: "Electronics", Href: "/products/electronics"},
        {Text: "Laptops", Href: "/products/electronics/laptops"},
        {Text: "MacBook Pro", Href: ""}, // Current page (no href)
    },
})
```

**Compact Variant:**
For mobile devices, use `BreadcrumbCompact` which shows only the last 2 items with an expandable dropdown for the rest:

```go
@BreadcrumbCompact(BreadcrumbProps{
    Items: []BreadcrumbItem{
        {Text: "Products", Href: "/products"},
        {Text: "Electronics", Href: "/products/electronics"},
        {Text: "Laptops", Href: "/products/electronics/laptops"},
        {Text: "MacBook Pro", Href: ""},
    },
})
```

---

### 4. Tabs (`tabs.templ`)

Flexible tab navigation with multiple variants and HTMX support.

**Features:**
- Three variants: `underline`, `pills`, `boxed`
- Three sizes: `sm`, `md`, `lg`
- Alpine.js state management
- HTMX support for dynamic content
- Badge support for notifications
- Icon support
- Keyboard navigation (arrow keys)
- Smooth transitions
- ARIA roles for accessibility

**Basic Usage (Inline Content):**
```go
@Tabs(TabsProps{
    DefaultTab: "overview",
    Variant:    "underline", // underline, pills, boxed
    Size:       "md",        // sm, md, lg
    UseHTMX:    false,
    Tabs: []TabItem{
        {
            ID:      "overview",
            Label:   "Overview",
            Icon:    "📋",
            Content: TabContent("This is the overview section."),
        },
        {
            ID:      "specs",
            Label:   "Specifications",
            Icon:    "⚙️",
            Badge:   "New",
            Content: TabContent("Technical specifications."),
        },
    },
})
```

**HTMX Usage (Dynamic Content):**
```go
@Tabs(TabsProps{
    DefaultTab:    "dashboard",
    Variant:       "underline",
    Size:          "md",
    UseHTMX:       true,
    HTMXTarget:    "tab-content",      // ID of target element
    HTMXIndicator: "loading-indicator", // Loading spinner selector
    Tabs: []TabItem{
        {
            ID:    "dashboard",
            Label: "Dashboard",
            Href:  "/api/tabs/dashboard",
        },
        {
            ID:    "analytics",
            Label: "Analytics",
            Href:  "/api/tabs/analytics",
        },
    },
})
```

**Helper Components:**
```go
// Simple tabs with defaults
@TabsSimple(tabs, "defaultTabID")

// HTMX tabs with defaults
@TabsHTMX(tabs, "defaultTabID", "targetElementID")
```

---

## Installation & Setup

### 1. Required Dependencies

Add these CDN links to your HTML `<head>`:

```html
<!-- Tailwind CSS -->
<script src="https://cdn.tailwindcss.com"></script>

<!-- Alpine.js -->
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>

<!-- HTMX (if using tabs with dynamic content) -->
<script src="https://unpkg.com/htmx.org@1.9.10"></script>

<!-- Alpine.js cloaking styles -->
<style>
    [x-cloak] { display: none !important; }
    .htmx-indicator { display: none; }
    .htmx-request .htmx-indicator { display: flex; }
</style>
```

### 2. Dark Mode Configuration

Ensure your Tailwind config supports dark mode:

```javascript
// tailwind.config.js
module.exports = {
    darkMode: 'class', // Enable class-based dark mode
    // ... rest of config
}
```

### 3. Import Components

```go
import "your-project/web/ui/components/navigation"
```

---

## Component Variants

### Navbar Variants
- **NavbarComplete** - Full-featured navbar with all options
- **Navbar** - Original simpler version (in navbar.templ)

### Mobile Menu Variants
- **MobileMenuComplete** - Full drawer with header and footer
- **MobileMenuSimple** - Minimal drawer without branding

### Breadcrumb Variants
- **Breadcrumb** - Full breadcrumb trail
- **BreadcrumbCompact** - Mobile-optimized with ellipsis

### Tabs Variants
- **Tabs** - Configurable component with all options
- **TabsSimple** - Quick setup with defaults
- **TabsHTMX** - Pre-configured for HTMX usage

---

## Accessibility Features

All components include:
- ✅ Proper ARIA labels and roles
- ✅ Keyboard navigation support
- ✅ Focus management
- ✅ Screen reader announcements
- ✅ Semantic HTML elements
- ✅ Color contrast compliance

### Keyboard Shortcuts

**Navbar:**
- `Tab` - Navigate through links
- `Enter/Space` - Activate links
- `Esc` - Close mobile menu

**Tabs:**
- `→` - Next tab
- `←` - Previous tab
- `Tab` - Move to tab content

**Breadcrumb:**
- `Tab` - Navigate through links
- `Enter/Space` - Follow link

---

## Dark Mode

Dark mode is fully supported across all components using Tailwind's dark mode classes.

**Theme Toggle Behavior:**
1. Checks localStorage for saved preference
2. Falls back to system preference
3. Persists selection in localStorage
4. Applies `dark` class to `<html>` element

**Manual Dark Mode Control:**
```javascript
// Enable dark mode
document.documentElement.classList.add('dark');
localStorage.setItem('theme', 'dark');

// Disable dark mode
document.documentElement.classList.remove('dark');
localStorage.setItem('theme', 'light');
```

---

## Examples

See `examples.templ` for complete working examples of:
- ✅ Full page layout with all components
- ✅ Different tab variants (underline, pills, boxed)
- ✅ HTMX integration examples
- ✅ Responsive breadcrumb usage
- ✅ Mobile menu configurations

Run the example page to see all components in action:
```go
@ExampleFullPageLayout()
```

---

## File Structure

```
navigation/
├── navbar_complete.templ           # Complete navbar with theme toggle
├── navbar.templ                    # Original simpler navbar
├── mobile-menu-complete.templ      # Full mobile menu implementation
├── mobile-menu.templ               # Menu items component
├── mobile-nav.templ                # Original mobile navigation
├── breadcrumb.templ                # Breadcrumb navigation
├── tabs.templ                      # Tab navigation component
├── examples.templ                  # Usage examples
└── README.md                       # This file
```

---

## Customization

### Colors
All components use Tailwind's color system. To customize:

1. **Navbar status colors:** Pass `StatusColor` prop (green, blue, red, yellow)
2. **Accent colors:** Modify `text-blue-600`, `bg-blue-500` classes
3. **Dark mode colors:** Adjust `dark:` prefixed classes

### Sizes
Tabs support three sizes via the `Size` prop:
- `sm` - Compact tabs
- `md` - Default medium size
- `lg` - Large tabs

### Transitions
All transitions use Tailwind's transition utilities and can be customized via classes.

---

## Browser Support

- ✅ Chrome/Edge (latest)
- ✅ Firefox (latest)
- ✅ Safari (latest)
- ✅ Mobile browsers (iOS Safari, Chrome Mobile)

**Requirements:**
- Modern browser with ES6+ support
- JavaScript enabled
- CSS Grid and Flexbox support

---

## Best Practices

1. **Always provide ARIA labels** for accessibility
2. **Use semantic HTML** elements (nav, header, main)
3. **Test keyboard navigation** on all components
4. **Provide alt text** for icons and images
5. **Test with screen readers** (NVDA, JAWS, VoiceOver)
6. **Ensure color contrast** meets WCAG AA standards
7. **Test responsive breakpoints** on real devices

---

## Troubleshooting

### Dark mode not working
- Ensure `darkMode: 'class'` is set in Tailwind config
- Check that Alpine.js is loaded
- Verify `x-data` directives are present

### Mobile menu not closing
- Check Alpine.js is loaded before page render
- Verify `@click.away` directive is present
- Ensure `[x-cloak]` styles are defined

### Tabs not switching
- Verify Alpine.js is loaded
- Check that tab IDs are unique
- Ensure `DefaultTab` matches an existing tab ID

### HTMX not loading content
- Confirm HTMX is loaded
- Check that endpoints return valid HTML
- Verify `HTMXTarget` ID exists in the DOM

---

## Contributing

When adding new navigation components:
1. Follow existing naming conventions
2. Include ARIA labels and roles
3. Support dark mode
4. Add responsive breakpoints
5. Include usage examples
6. Update this README

---

## License

These components are part of the Scrapy project and follow the project's license terms.

---

## Credits

Built with:
- [Templ](https://templ.guide/) - Go templating
- [Alpine.js](https://alpinejs.dev/) - Lightweight JavaScript framework
- [HTMX](https://htmx.org/) - High power tools for HTML
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
