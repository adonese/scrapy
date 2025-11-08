# Alpine.js Patterns for Templ Templates

This directory contains reusable Alpine.js patterns that can be used inline in Templ templates. All Alpine.js code is inline - no external JavaScript files required.

## Overview

All Alpine.js patterns in this project are:
- **Inline**: Written directly in Templ templates using x-data, x-init, x-on, etc.
- **Self-contained**: No external .js files needed
- **Dark mode compatible**: All components support Tailwind dark: classes
- **Accessible**: ARIA labels and semantic HTML included

## Available Patterns

### 1. Dark Mode Toggle (`theme-toggle.templ`)

**Usage:**
```go
import "github.com/adonese/cost-of-living/web/ui/components"

templ MyPage() {
    @components.ThemeToggle()         // Full toggle with background
    @components.ThemeToggleCompact()  // Compact version
}
```

**Features:**
- Persists theme preference to localStorage
- Detects system preference on first load
- Toggles 'dark' class on <html> element
- Icon changes based on current theme
- Smooth transitions

**How it works:**
```javascript
x-data="{
    darkMode: localStorage.getItem('theme') === 'dark' ||
              (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches),
    toggle() {
        this.darkMode = !this.darkMode;
        localStorage.setItem('theme', this.darkMode ? 'dark' : 'light');
        document.documentElement.classList.toggle('dark', this.darkMode);
    }
}"
```

### 2. Form Persistence (`form-persist.templ`)

**Pattern 1: Complete Form Persistence**
```html
<form
    x-data="{
        formData: JSON.parse(localStorage.getItem('myFormKey') || '{}'),
        saveForm() {
            localStorage.setItem('myFormKey', JSON.stringify(this.formData));
        }
    }"
>
    <input x-model="formData.name" @input.debounce.500ms="saveForm()" />
</form>
```

**Pattern 2: Inline Single Field**
```html
<input
    x-data="{ value: localStorage.getItem('myKey') || '' }"
    x-model="value"
    @input.debounce.500ms="localStorage.setItem('myKey', value)"
/>
```

**Pattern 3: Multi-Step Form**
See `form-persist.templ` for complete multi-step wizard with progress tracking.

**Key Features:**
- Auto-save with debouncing (500ms default)
- Last saved timestamp tracking
- Step persistence for wizards
- Clear/reset functionality

### 3. Form Validation (`form-validation.templ`)

**Pattern 1: Required Field**
```javascript
x-data="{
    value: '',
    error: '',
    touched: false,
    validate() {
        this.touched = true;
        if (!this.value) {
            this.error = 'This field is required';
            return false;
        }
        this.error = '';
        return true;
    }
}"
```

**Pattern 2: Email Validation**
```javascript
validateEmail() {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(this.email)) {
        this.error = 'Please enter a valid email';
        return false;
    }
    this.error = '';
    return true;
}
```

**Pattern 3: Number Range**
```javascript
validateNumber() {
    const num = parseInt(this.value);
    if (num < this.min || num > this.max) {
        this.error = `Value must be between ${this.min} and ${this.max}`;
        return false;
    }
    this.error = '';
    return true;
}
```

**Pattern 4: Custom Password Validation**
See `form-validation.templ` for advanced password strength validation.

**Key Features:**
- Touch tracking (only show errors after blur)
- Real-time validation on input
- Multiple validation rules per field
- Error message display with x-show

### 4. Alpine Stores (`alpine-stores.templ`)

#### Toast/Notification Queue

**Setup in base layout:**
```go
@components.ToastStore()
```

**Usage anywhere:**
```javascript
// Show success message
window.toastStore.success('Saved successfully!')

// Show error
window.toastStore.error('Failed to save')

// Show warning
window.toastStore.warning('Please be careful')

// Show info
window.toastStore.info('Here is some information')
```

**Features:**
- Auto-dismiss after duration (default 5s)
- Multiple toasts with queue
- Different types: success, error, warning, info
- Close button
- Smooth animations

#### Modal Manager

**Setup:**
```go
@components.ModalStore()
```

**Usage:**
```javascript
// Open modal with content
window.modalStore.open('myModal', {
    content: '<h1>Hello</h1>'
})

// Close specific modal
window.modalStore.close('myModal')

// Close all modals
window.modalStore.closeAll()

// Check if modal is open
window.modalStore.isOpen('myModal')
```

**Features:**
- Multiple modal support with stacking
- Backdrop with click-outside to close
- Body scroll lock
- Smooth animations
- Close button

#### Form Store Pattern

See `alpine-stores.templ` for a complete form state management pattern with:
- Field value tracking
- Error tracking
- Touch tracking
- Validation
- Async submission

### 5. Complete Examples (`alpine-examples.templ`)

#### Example 1: Form with Validation + Persistence + Toast
A complete contact form demonstrating:
- Real-time validation
- Auto-save to localStorage
- Toast notifications on submit/error
- Error display inline
- Clear/reset functionality

#### Example 2: Estimator Form Pattern
A pattern specifically for the UAE Cost estimator:
- Persona data structure
- Number validation
- Select dropdowns
- HTMX compatibility
- Auto-save

## Common Patterns & Best Practices

### Debouncing Input
```javascript
@input.debounce.500ms="saveForm()"  // Wait 500ms after typing stops
```

### Touch Tracking
```javascript
// Only validate and show errors after field is touched
@blur="validate()"
@input="touched && validate()"
```

### Conditional Classes
```javascript
:class="error ? 'border-red-500' : 'border-slate-300'"
```

### Conditional Rendering
```javascript
x-show="error"        // Show/hide with display
x-if="error"          // Remove from DOM
x-cloak               // Hide until Alpine loads
```

### Dark Mode Classes
Always include both light and dark variants:
```html
class="bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
```

### Model Modifiers
```javascript
x-model.number="value"    // Parse as number
x-model.debounce="value"  // Debounce updates
```

## Integration with HTMX

Alpine.js works seamlessly with HTMX:

```html
<form
    x-data="{ formData: {...} }"
    hx-post="/api/estimate"
    hx-target="#result"
    @htmx:before-request="saveForm()"
    @htmx:after-request="clearForm()"
>
```

## File Structure

```
web/ui/components/core/
├── theme-toggle.templ       # Dark mode toggle component
├── form-persist.templ       # Form persistence patterns
├── form-validation.templ    # Validation patterns
├── alpine-stores.templ      # Global stores (toast, modal, form)
├── alpine-examples.templ    # Complete working examples
└── ALPINE_PATTERNS.md       # This file
```

## Quick Start

1. **Add dark mode support to your base layout:**
   The base layout already includes dark mode classes and x-cloak styles.

2. **Add theme toggle to navbar:**
   Already integrated in `navbar.templ`

3. **Initialize global stores (optional):**
   ```go
   // In your base layout or app root
   @components.ToastStore()
   @components.ModalStore()
   ```

4. **Use patterns in your forms:**
   Copy patterns from the example files directly into your Templ templates.

5. **Test the examples:**
   - Navigate to the components in your UI to see them in action
   - Inspect localStorage to see persistence
   - Toggle dark mode to see theme support

## Browser Support

- All modern browsers (Chrome, Firefox, Safari, Edge)
- Alpine.js 3.x loaded from CDN
- LocalStorage API required for persistence
- CSS Grid and Flexbox for layouts

## Tips

1. **Start simple**: Begin with single-field patterns, then combine them
2. **Copy patterns**: Use the examples as templates, modify as needed
3. **Test persistence**: Clear localStorage during development to test defaults
4. **Dark mode first**: Always add dark: variants when styling
5. **Validate early**: Add validation to any user input
6. **Debounce saves**: Use 500ms debounce for auto-save to avoid excessive writes

## Examples in Action

See these files for working implementations:
- `/home/user/scrapy/web/ui/components/core/form-persist.templ`
- `/home/user/scrapy/web/ui/components/core/form-validation.templ`
- `/home/user/scrapy/web/ui/components/core/alpine-stores.templ`
- `/home/user/scrapy/web/ui/components/core/alpine-examples.templ`

Each file includes multiple examples with different patterns and use cases.
