# Navigation Components - Quick Start Guide

Get started with the navigation components in 5 minutes.

## Step 1: Add Dependencies to Your HTML

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My App</title>

    <!-- Tailwind CSS -->
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            darkMode: 'class',
        }
    </script>

    <!-- Alpine.js -->
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>

    <!-- HTMX (optional - only if using dynamic tabs) -->
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>

    <!-- Required styles -->
    <style>
        [x-cloak] { display: none !important; }
    </style>
</head>
<body>
    <!-- Your app content -->
</body>
</html>
```

## Step 2: Import the Navigation Package

```go
import "your-project/web/ui/components/navigation"
```

## Step 3: Use the Components

### Basic Navbar

```go
templ MyPage() {
    @navigation.NavbarComplete(navigation.NavbarCompleteProps{
        LogoText:    "My App",
        Subtitle:    "Welcome",
        ShowStatus:  true,
        StatusText:  "Live",
        StatusColor: "green",
        NavLinks: []navigation.NavLink{
            {Href: "/", Text: "Home", Label: "Go to home"},
            {Href: "/about", Text: "About", Label: "About page"},
        },
    })

    <main class="container mx-auto px-4 py-8">
        <h1>Welcome to My App</h1>
    </main>
}
```

### Add Breadcrumbs

```go
templ ProductPage() {
    @navigation.NavbarComplete(myNavbarProps)

    <main class="container mx-auto px-4 py-8">
        <!-- Breadcrumbs -->
        @navigation.Breadcrumb(navigation.BreadcrumbProps{
            Items: []navigation.BreadcrumbItem{
                {Text: "Products", Href: "/products"},
                {Text: "Electronics", Href: "/products/electronics"},
                {Text: "iPhone 15", Href: ""},
            },
        })

        <h1>iPhone 15</h1>
        <!-- Product content -->
    </main>
}
```

### Add Tabs

```go
templ ProductDetailPage() {
    @navigation.NavbarComplete(myNavbarProps)

    <main class="container mx-auto px-4 py-8">
        <h1>Product Details</h1>

        <!-- Simple Tabs -->
        @navigation.Tabs(navigation.TabsProps{
            DefaultTab: "description",
            Variant:    "underline",
            Size:       "md",
            UseHTMX:    false,
            Tabs: []navigation.TabItem{
                {
                    ID:      "description",
                    Label:   "Description",
                    Content: navigation.TabContent("Product description here..."),
                },
                {
                    ID:      "specs",
                    Label:   "Specifications",
                    Badge:   "New",
                    Content: navigation.TabContent("Technical specs here..."),
                },
                {
                    ID:      "reviews",
                    Label:   "Reviews",
                    Content: navigation.TabContent("Customer reviews here..."),
                },
            },
        })
    </main>
}
```

## Step 4: Test It

1. Run your Templ application
2. Navigate to your page
3. Test the theme toggle (should persist on reload)
4. Test mobile menu on small screens
5. Test tab switching
6. Test keyboard navigation (Tab, arrow keys, Escape)

## Common Patterns

### Responsive Breadcrumbs

```go
<!-- Desktop: full breadcrumb, Mobile: compact -->
<div class="hidden md:block">
    @navigation.Breadcrumb(props)
</div>
<div class="md:hidden">
    @navigation.BreadcrumbCompact(props)
</div>
```

### Tabs with HTMX

```go
@navigation.Tabs(navigation.TabsProps{
    DefaultTab:    "dashboard",
    Variant:       "underline",
    UseHTMX:       true,
    HTMXTarget:    "content-area",
    HTMXIndicator: "spinner",
    Tabs: []navigation.TabItem{
        {ID: "dashboard", Label: "Dashboard", Href: "/api/dashboard"},
        {ID: "reports", Label: "Reports", Href: "/api/reports"},
    },
})
```

### Pills Style Tabs

```go
@navigation.Tabs(navigation.TabsProps{
    DefaultTab: "all",
    Variant:    "pills", // Pills style
    Tabs: []navigation.TabItem{
        {ID: "all", Label: "All", Content: myContent},
        {ID: "active", Label: "Active", Badge: "5", Content: myContent},
    },
})
```

## Styling Tips

### Custom Accent Colors

Replace `blue` with your brand color in the component files:
- `text-blue-600` → `text-purple-600`
- `bg-blue-500` → `bg-purple-500`
- `ring-blue-500` → `ring-purple-500`

### Custom Logo

Replace the gradient logo:
```go
// In navbar_complete.templ, replace:
<div class="h-10 w-10 rounded-lg bg-gradient-to-br from-blue-600 to-blue-700...">

// With your custom classes:
<div class="h-10 w-10 rounded-lg bg-gradient-to-br from-purple-600 to-pink-600...">
```

## Troubleshooting

**Theme toggle not working?**
→ Check that Alpine.js loaded: `console.log(window.Alpine)`

**Mobile menu not sliding?**
→ Ensure `[x-cloak] { display: none !important; }` is in your CSS

**Tabs not switching?**
→ Verify tab IDs are unique and DefaultTab matches an existing ID

**Components look unstyled?**
→ Confirm Tailwind CSS is loaded and configured with `darkMode: 'class'`

## Next Steps

1. Read the full [README.md](./README.md) for detailed documentation
2. Check [examples.templ](./examples.templ) for more usage patterns
3. Customize colors and styles to match your brand
4. Add your own icons and badges
5. Test accessibility with keyboard and screen readers

## Need Help?

- Full documentation: [README.md](./README.md)
- Working examples: [examples.templ](./examples.templ)
- Component files are well-commented with usage instructions
