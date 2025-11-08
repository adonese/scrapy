# Alpine.js Quick Reference for Templ

## Quick Copy-Paste Patterns

### Dark Mode Toggle (Already Implemented)
```go
@components.ThemeToggleCompact()
```

### Form with Auto-Save
```html
<form
    x-data="{
        formData: JSON.parse(localStorage.getItem('myForm') || '{}'),
        save() { localStorage.setItem('myForm', JSON.stringify(this.formData)); }
    }"
>
    <input x-model="formData.name" @input.debounce.500ms="save()" />
</form>
```

### Required Field Validation
```html
<div
    x-data="{
        value: '',
        error: '',
        touched: false,
        validate() {
            this.touched = true;
            this.error = !this.value ? 'Required' : '';
            return !this.error;
        }
    }"
>
    <input
        x-model="value"
        @blur="validate()"
        @input="touched && validate()"
        :class="error ? 'border-red-500' : 'border-slate-300'"
    />
    <p x-show="error" x-cloak class="text-red-600 text-xs" x-text="error"></p>
</div>
```

### Email Validation
```javascript
validateEmail() {
    this.touched = true;
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    this.error = !regex.test(this.value) ? 'Invalid email' : '';
    return !this.error;
}
```

### Number Range Validation
```javascript
validateNumber() {
    this.touched = true;
    const num = parseFloat(this.value);
    if (isNaN(num) || num < this.min || num > this.max) {
        this.error = `Must be between ${this.min} and ${this.max}`;
        return false;
    }
    this.error = '';
    return true;
}
```

### Toast Notifications
```go
// In base layout (once):
@components.ToastStore()

// Anywhere in your app:
<button @click="window.toastStore.success('Success!')">Click</button>
<button @click="window.toastStore.error('Error!')">Click</button>
<button @click="window.toastStore.warning('Warning!')">Click</button>
<button @click="window.toastStore.info('Info!')">Click</button>
```

### Modal
```go
// In base layout (once):
@components.ModalStore()

// Anywhere in your app:
<button @click="window.modalStore.open('myModal', { content: '<h1>Hello</h1>' })">
    Open Modal
</button>
```

### Complete Form Pattern
```html
<form
    x-data="{
        data: JSON.parse(localStorage.getItem('form') || '{}'),
        errors: {},
        touched: {},

        validate(field) {
            this.touched[field] = true;
            if (!this.data[field]) {
                this.errors[field] = 'Required';
                return false;
            }
            delete this.errors[field];
            return true;
        },

        save() {
            localStorage.setItem('form', JSON.stringify(this.data));
        },

        submit() {
            if (this.validate('name') && this.validate('email')) {
                window.toastStore.success('Submitted!');
                return true;
            }
            window.toastStore.error('Fix errors');
            return false;
        }
    }"
    @submit.prevent="submit()"
>
    <input
        x-model="data.name"
        @blur="validate('name')"
        @input="touched.name && validate('name'); save()"
        :class="errors.name ? 'border-red-500' : 'border-slate-300'"
    />
    <p x-show="errors.name" x-text="errors.name"></p>

    <button type="submit">Submit</button>
</form>
```

### Dark Mode Classes Template
```html
class="
    bg-white dark:bg-slate-800
    text-slate-900 dark:text-slate-100
    border-slate-300 dark:border-slate-700
    hover:bg-slate-100 dark:hover:bg-slate-700
    focus:ring-slate-500 dark:focus:ring-slate-400
    transition-colors
"
```

## Common Modifiers

- `x-model.number` - Parse as number
- `x-model.debounce` - Debounce updates
- `@input.debounce.500ms` - Debounce 500ms
- `@click.prevent` - Prevent default
- `@submit.prevent` - Prevent form submission
- `:class` - Dynamic classes
- `x-show` - Toggle display
- `x-if` - Conditional render
- `x-cloak` - Hide until Alpine loads

## File Locations

All patterns available in:
- `/home/user/scrapy/web/ui/components/core/theme-toggle.templ`
- `/home/user/scrapy/web/ui/components/core/form-persist.templ`
- `/home/user/scrapy/web/ui/components/core/form-validation.templ`
- `/home/user/scrapy/web/ui/components/core/alpine-stores.templ`
- `/home/user/scrapy/web/ui/components/core/alpine-examples.templ`

## Full Documentation

See `/home/user/scrapy/web/ui/components/core/ALPINE_PATTERNS.md` for complete documentation.
