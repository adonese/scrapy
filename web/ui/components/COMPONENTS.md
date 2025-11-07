# Templ Component Library

A comprehensive, reusable component library built with Templ, Tailwind CSS, and Alpine.js. **No Node.js dependencies required.**

## Component Organization

```
web/ui/components/
├── core/           # Core UI components (buttons, inputs, cards, etc.)
└── feedback/       # User feedback components (alerts, toasts, modals, etc.)
```

## Core Components

### Button (`core/button.templ`)

Flexible button component with multiple variants and sizes.

**Features:**
- Variants: `primary`, `secondary`, `danger`, `ghost`, `link`
- Sizes: `sm`, `md`, `lg`
- Full HTMX attribute support via `HxAttrs` map
- Alpine.js integration
- Accessible with ARIA labels
- Disabled state support

**Usage:**
```go
@components.ButtonText(components.ButtonProps{
    Variant: "primary",
    Size: "md",
    Type: "submit",
    HxAttrs: map[string]string{
        "hx-post": "/api/submit",
        "hx-target": "#result",
    },
}, "Submit")
```

---

### Input (`core/input.templ`)

Form input component with label and error message support.

**Features:**
- Types: `text`, `email`, `number`, `password`, `search`
- `InputGroup` component includes label, hint, and error message
- Validation attributes: `min`, `max`, `step`, `pattern`
- Full HTMX and Alpine.js support
- Accessible with proper label associations

**Usage:**
```go
@components.InputGroup(components.InputGroupProps{
    Label: "Email Address",
    LabelFor: "email",
    Error: errorMessage,
    Required: true,
    InputProps: components.InputProps{
        Type: "email",
        ID: "email",
        Name: "email",
        Placeholder: "you@example.com",
    },
})
```

---

### Select (`core/select.templ`)

Dropdown select component with option support.

**Features:**
- Dynamic option list with `SelectOption` struct
- Individual option disabled/selected states
- `SelectGroup` includes label, hint, and error message
- HTMX and Alpine.js support
- Accessible with proper ARIA attributes

**Usage:**
```go
@components.SelectGroup(components.SelectGroupProps{
    Label: "Country",
    LabelFor: "country",
    Required: true,
    SelectProps: components.SelectProps{
        ID: "country",
        Name: "country",
        Options: []components.SelectOption{
            {Value: "us", Label: "United States", Selected: true},
            {Value: "ca", Label: "Canada"},
            {Value: "mx", Label: "Mexico"},
        },
    },
})
```

---

### Card (`core/card.templ`)

Container component with header, body, and footer sections.

**Features:**
- Variants: `default`, `bordered`, `elevated`
- Separate `CardHeader`, `CardContent`, and `CardFooter` components
- Flexible content slots
- HTMX and Alpine.js support

**Usage:**
```go
@components.Card(components.CardProps{Variant: "elevated"}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    components.CardHeader(components.CardHeaderProps{
        Title: "Card Title",
        Description: "Card description text",
    }).Render(ctx, w)
    components.CardContent(components.CardContentProps{}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // Your content here
        return nil
    })).Render(ctx, w)
    return nil
}))
```

---

### Badge (`core/badge.templ`)

Small label component for status indicators.

**Features:**
- Variants: `default`, `success`, `warning`, `danger`, `info`
- Text or custom content support
- `BadgeDot` component for small status indicators
- HTMX and Alpine.js support

**Usage:**
```go
@components.Badge(components.BadgeProps{
    Variant: "success",
}, "Active")

@components.BadgeDot(components.BadgeDotProps{
    Variant: "warning",
})
```

---

## Feedback Components

### Toast (`feedback/toast.templ`)

Notification toast with auto-dismiss functionality.

**Features:**
- Types: `success`, `error`, `warning`, `info`
- Auto-dismiss with configurable duration (default: 5000ms)
- Manual dismiss option
- Animated enter/exit with Alpine.js
- `ToastContainer` for positioning (top-right, top-left, bottom-right, bottom-left, top-center, bottom-center)
- Accessible with ARIA live regions

**Usage:**
```go
@components.ToastContainer(components.ToastContainerProps{
    Position: "top-right",
}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    components.Toast(components.ToastProps{
        Type: "success",
        Title: "Success!",
        Message: "Your changes have been saved.",
        Duration: 5000,
        Dismissible: true,
    }).Render(ctx, w)
    return nil
}))
```

---

### Modal (`feedback/modal.templ`)

Dialog modal with backdrop and animations.

**Features:**
- Sizes: `sm`, `md`, `lg`, `xl`, `full`
- Backdrop click to close
- Escape key to close
- Animated entrance/exit with Alpine.js
- Separate `ModalHeader`, `ModalBody`, and `ModalFooter` components
- Close button option
- Accessible with proper ARIA attributes and focus management

**Usage:**
```go
@components.Modal(components.ModalProps{
    ID: "example-modal",
    Size: "md",
    CloseButton: true,
},
    components.ModalHeader(components.ModalHeaderProps{
        Title: "Modal Title",
        Subtitle: "Optional subtitle",
    }),
    templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        components.ModalBody(components.ModalBodyProps{}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
            // Your modal content
            return nil
        })).Render(ctx, w)
        return nil
    }),
    templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        components.ModalFooter(components.ModalFooterProps{}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
            // Footer buttons
            return nil
        })).Render(ctx, w)
        return nil
    }),
)
```

---

### Alert (`feedback/alert.templ`)

Alert boxes for important messages.

**Features:**
- Types: `success`, `error`, `warning`, `info`
- Optional icon display with built-in SVG icons
- Dismissible with Alpine.js
- Title and message support
- Custom content via `AlertWithContent`
- Animated entrance/exit
- Accessible with proper ARIA roles

**Usage:**
```go
// Simple alert
@components.Alert(components.AlertProps{
    Type: "warning",
    Title: "Warning",
    Message: "This action cannot be undone.",
    Icon: true,
    Dismissible: true,
})

// Alert with custom content
@components.AlertWithContent(components.AlertProps{
    Type: "info",
    Icon: true,
}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    // Custom content
    return nil
}))
```

---

### Skeleton (`feedback/skeleton.templ`)

Loading skeleton components with pulse animation.

**Features:**
- Basic `Skeleton` with configurable width, height, rounded corners
- `SkeletonText` for multi-line text skeletons
- `SkeletonCircle` for avatar placeholders
- `SkeletonCard` for complete card loading states
- `SkeletonTable` for table loading states
- `SkeletonList` for list loading states
- Pulse animation (configurable)
- Fully customizable with Tailwind classes

**Usage:**
```go
// Basic skeleton
@components.Skeleton(components.SkeletonProps{
    Width: "w-full",
    Height: "h-4",
    Rounded: "rounded",
    Animate: true,
})

// Text skeleton (3 lines)
@components.SkeletonText(components.SkeletonTextProps{
    Lines: 3,
    LastLine: "w-3/4",
})

// Circle avatar skeleton
@components.SkeletonCircle(components.SkeletonCircleProps{
    Size: "h-12 w-12",
    Animate: true,
})

// Card skeleton
@components.SkeletonCard(components.SkeletonCardProps{
    ShowAvatar: true,
    ShowImage: true,
    Lines: 4,
})

// Table skeleton
@components.SkeletonTable(components.SkeletonTableProps{
    Rows: 5,
    Columns: 4,
})

// List skeleton
@components.SkeletonList(components.SkeletonListProps{
    Items: 5,
    ShowAvatar: true,
})
```

---

## Design Principles

### 1. **No Node.js Dependencies**
All components use only:
- **Templ** for templating
- **Tailwind CSS** for styling (utility classes only)
- **Alpine.js** for client-side interactivity (inline, no build step)

### 2. **Component Props Pattern**
Every component accepts a props struct for configuration:
```go
type ButtonProps struct {
    Variant    string
    Size       string
    Disabled   bool
    Class      string
    HxAttrs    map[string]string
    AlpineData string
    AriaLabel  string
}
```

### 3. **HTMX Integration**
All components support HTMX attributes via the `HxAttrs` map:
```go
HxAttrs: map[string]string{
    "hx-get": "/api/data",
    "hx-trigger": "click",
    "hx-target": "#result",
}
```

### 4. **Alpine.js for Interactivity**
Interactive components use Alpine.js inline directives:
- `x-data` for component state
- `x-show` for conditional rendering
- `x-transition` for animations
- `@click` for event handling
- `@keydown.escape` for keyboard interactions

### 5. **Accessibility First**
All components include:
- Proper ARIA labels and roles
- Keyboard navigation support
- Screen reader announcements
- Focus management
- Semantic HTML

### 6. **Flexible Styling**
Every component accepts a `Class` prop for additional Tailwind classes:
```go
Class: "mt-4 shadow-lg" // Add custom spacing, shadows, etc.
```

### 7. **Composable Architecture**
Components are designed to be nested and composed:
```go
@components.Card(cardProps, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    components.CardHeader(headerProps).Render(ctx, w)
    components.CardContent(contentProps, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        components.Alert(alertProps).Render(ctx, w)
        return nil
    })).Render(ctx, w)
    return nil
}))
```

---

## Common Patterns

### Error Handling
```go
@components.InputGroup(components.InputGroupProps{
    Label: "Email",
    Error: if err != nil { err.Error() } else { "" },
    InputProps: components.InputProps{
        Type: "email",
        Name: "email",
        Value: formData.Email,
    },
})
```

### HTMX Form Submission
```go
@components.ButtonText(components.ButtonProps{
    Variant: "primary",
    Type: "submit",
    HxAttrs: map[string]string{
        "hx-post": "/api/submit",
        "hx-target": "#result",
        "hx-swap": "outerHTML",
        "hx-indicator": "#spinner",
    },
}, "Submit")
```

### Modal Toggle
```go
// Open modal button
<button @click="$refs.myModal.open = true">Open Modal</button>

// Modal with ref
@components.Modal(components.ModalProps{
    AlpineData: "{open: false}",
}, ...)
```

### Toast Notification
```go
// In your handler, add toast to response
@components.Toast(components.ToastProps{
    Type: "success",
    Title: "Saved!",
    Message: "Your changes have been saved.",
    Duration: 3000,
    Dismissible: true,
})
```

### Loading States
```go
<div x-data="{loading: true}" x-init="setTimeout(() => loading = false, 2000)">
    <div x-show="loading">
        @components.SkeletonCard(components.SkeletonCardProps{
            ShowAvatar: true,
            Lines: 3,
        })
    </div>
    <div x-show="!loading" x-cloak>
        // Actual content
    </div>
</div>
```

---

## Setup Requirements

### 1. Tailwind CSS
Ensure your `tailwind.config.js` includes the components directory:
```js
module.exports = {
  content: [
    "./web/ui/**/*.templ",
    "./web/ui/**/*.go",
  ],
  // ... rest of config
}
```

### 2. Alpine.js
Include Alpine.js in your base template:
```html
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
```

### 3. HTMX (Optional)
If using HTMX features:
```html
<script src="https://unpkg.com/htmx.org@1.9.x"></script>
```

### 4. Alpine.js x-cloak
Add to your CSS to prevent flash of unstyled content:
```css
[x-cloak] { display: none !important; }
```

---

## File Structure

```
web/ui/components/
├── core/
│   ├── button.templ          # Button component
│   ├── input.templ           # Input with label/error
│   ├── select.templ          # Select dropdown
│   ├── card.templ            # Card container
│   └── badge.templ           # Badge labels
├── feedback/
│   ├── toast.templ           # Toast notifications
│   ├── modal.templ           # Modal dialogs
│   ├── alert.templ           # Alert messages
│   └── skeleton.templ        # Loading skeletons
└── COMPONENTS.md             # This documentation
```

---

## Component Status

| Component | Status | Features |
|-----------|--------|----------|
| Button | ✅ Complete | Variants, sizes, HTMX, Alpine.js |
| Input | ✅ Complete | Types, validation, label, error |
| Select | ✅ Complete | Options, label, error |
| Card | ✅ Complete | Variants, header/body/footer |
| Badge | ✅ Complete | Variants, dot indicator |
| Toast | ✅ Complete | Auto-dismiss, animations |
| Modal | ✅ Complete | Sizes, backdrop, ESC key |
| Alert | ✅ Complete | Icons, dismissible |
| Skeleton | ✅ Complete | Multiple variants |

---

## Best Practices

1. **Always provide ARIA labels** for better accessibility
2. **Use semantic HTML** - buttons for actions, links for navigation
3. **Keep Alpine.js state minimal** - use server-side rendering when possible
4. **Leverage HTMX** for dynamic updates without full page reloads
5. **Use skeleton loaders** for better perceived performance
6. **Provide error feedback** with clear, actionable messages
7. **Test keyboard navigation** in all interactive components
8. **Use appropriate toast positions** to avoid covering important UI
9. **Keep modals focused** - one clear action per modal
10. **Validate on both client and server** for security

---

## Contributing

When adding new components:
1. Follow the existing props pattern
2. Include HTMX and Alpine.js support
3. Add proper ARIA attributes
4. Use only Tailwind utility classes
5. Document usage in this file
6. Test accessibility with screen readers
7. Ensure responsive design

---

## License

This component library is part of the Scrapy project.
