# Component Library Quick Start

## Installation

No installation needed! Just import and use.

## Import Components

```go
import "yourproject/web/ui/components"
```

## Quick Examples

### Basic Button
```go
@components.ButtonText(components.ButtonProps{
    Variant: "primary",
    Size: "md",
}, "Click Me")
```

### Form Input
```go
@components.InputGroup(components.InputGroupProps{
    Label: "Email",
    LabelFor: "email",
    InputProps: components.InputProps{
        Type: "email",
        ID: "email",
        Name: "email",
        Required: true,
    },
})
```

### Select Dropdown
```go
@components.SelectGroup(components.SelectGroupProps{
    Label: "Choose Option",
    SelectProps: components.SelectProps{
        Options: []components.SelectOption{
            {Value: "1", Label: "Option 1"},
            {Value: "2", Label: "Option 2"},
        },
    },
})
```

### Card
```go
@components.Card(components.CardProps{Variant: "elevated"}, templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    components.CardHeader(components.CardHeaderProps{
        Title: "Card Title",
    }).Render(ctx, w)
    return nil
}))
```

### Badge
```go
@components.Badge(components.BadgeProps{Variant: "success"}, "Active")
```

### Toast Notification
```go
@components.Toast(components.ToastProps{
    Type: "success",
    Title: "Success!",
    Message: "Operation completed.",
    Duration: 5000,
    Dismissible: true,
})
```

### Modal Dialog
```go
@components.Modal(components.ModalProps{
    Size: "md",
    CloseButton: true,
},
    components.ModalHeader(components.ModalHeaderProps{Title: "Confirm"}),
    components.ModalBody(components.ModalBodyProps{}, yourContent),
    components.ModalFooter(components.ModalFooterProps{}, yourButtons),
)
```

### Alert Message
```go
@components.Alert(components.AlertProps{
    Type: "warning",
    Title: "Warning",
    Message: "This is a warning message.",
    Icon: true,
    Dismissible: true,
})
```

### Loading Skeleton
```go
@components.SkeletonCard(components.SkeletonCardProps{
    ShowAvatar: true,
    Lines: 3,
})
```

## HTMX Integration

```go
@components.ButtonText(components.ButtonProps{
    HxAttrs: map[string]string{
        "hx-post": "/api/submit",
        "hx-target": "#result",
    },
}, "Submit")
```

## Alpine.js State

```go
@components.Modal(components.ModalProps{
    AlpineData: "{open: false}",
}, ...)

// Trigger: <button @click="open = true">Open Modal</button>
```

## Variants Reference

### Button Variants
- `primary` - Dark background
- `secondary` - Light background
- `danger` - Red background
- `ghost` - Transparent
- `link` - Underlined text

### Button Sizes
- `sm` - Small (h-8)
- `md` - Medium (h-10)
- `lg` - Large (h-12)

### Alert/Toast Types
- `success` - Green
- `error` - Red
- `warning` - Amber
- `info` - Blue

### Card Variants
- `default` - Border + shadow
- `bordered` - Border only
- `elevated` - Large shadow

## Need More?

See `/home/user/scrapy/web/ui/components/COMPONENTS.md` for complete documentation.
