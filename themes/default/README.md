# OpenHost Default Theme

A modern, responsive theme for OpenHost with clean design and great user experience.

## Features

- 🎨 **Modern Design** - Clean, professional look with gradient accents
- 📱 **Fully Responsive** - Works perfectly on desktop, tablet, and mobile
- 🌓 **Dark Mode Ready** - Automatic dark mode support via CSS variables
- ⚡ **Performance Optimized** - Minimal CSS/JS for fast loading
- ♿ **Accessible** - WCAG compliant with proper semantic HTML
- 🎯 **Developer Friendly** - Well-organized, easy to customize

## Structure

```
themes/default/
├── layouts/
│   └── base.html          # Base layout with navigation and footer
├── pages/
│   ├── home.html          # Homepage with hero and features
│   ├── products.html      # Product catalog with filtering
│   └── dashboard.html     # Customer dashboard
├── partials/
│   └── (future components)
└── assets/
    ├── css/
    │   └── main.css       # Complete stylesheet
    ├── js/
    │   └── main.js        # Interactive features
    └── images/
        └── (theme images)
```

## Customization

### Colors

Edit CSS variables in `assets/css/main.css`:

```css
:root {
    --primary: #667eea;      /* Primary brand color */
    --secondary: #764ba2;    /* Secondary brand color */
    --accent: #f093fb;       /* Accent color */
    /* ... more colors */
}
```

### Typography

Change fonts by updating the Google Fonts link in `layouts/base.html` and the CSS variable:

```css
:root {
    --font-sans: 'Inter', sans-serif;
}
```

### Logo

Replace the SVG logo in `layouts/base.html` or add your own logo image to `assets/images/`.

## Pages

### Home Page (`pages/home.html`)

Features:
- Hero section with CTA buttons
- Feature cards grid
- Stats section
- Call-to-action section

### Products Page (`pages/products.html`)

Features:
- Product grid with cards
- Category filtering
- Comparison table
- FAQ section with accordion

### Dashboard Page (`pages/dashboard.html`)

Features:
- Statistics cards
- Active services list
- Recent invoices
- Support tickets table

## Components

### Cards

```html
<div class="card">
    <div class="card-header">
        <h3 class="card-title">Title</h3>
    </div>
    <div class="card-body">
        Content
    </div>
    <div class="card-footer">
        Footer
    </div>
</div>
```

### Buttons

```html
<button class="btn btn-primary">Primary</button>
<button class="btn btn-outline">Outline</button>
<button class="btn btn-secondary">Secondary</button>
<button class="btn btn-lg">Large</button>
<button class="btn btn-sm">Small</button>
```

### Grid

```html
<div class="grid grid-3">
    <div>Column 1</div>
    <div>Column 2</div>
    <div>Column 3</div>
</div>
```

### Forms

```html
<div class="form-group">
    <label class="form-label">Label</label>
    <input type="text" class="form-control" placeholder="Placeholder">
</div>
```

### Alerts

```html
<div class="alert alert-success">Success message</div>
<div class="alert alert-error">Error message</div>
<div class="alert alert-warning">Warning message</div>
<div class="alert alert-info">Info message</div>
```

## JavaScript Features

- Mobile navigation toggle
- Auto-dismissing alerts
- Form validation helper
- Price calculator
- Copy to clipboard
- Toast notifications
- Smooth scrolling

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers (iOS Safari, Chrome Mobile)

## Accessibility

- Semantic HTML5 elements
- ARIA labels where needed
- Keyboard navigation support
- High contrast mode support
- Screen reader friendly

## Performance

- Minimal CSS (~14KB)
- Optimized JavaScript (~6KB)
- No jQuery dependency
- Lazy loading ready
- Critical CSS inline option

## Creating a Custom Theme

1. Copy the `default` theme folder:
```bash
cp -r themes/default themes/mytheme
```

2. Update colors, fonts, and styles in `assets/css/main.css`

3. Modify layouts and pages as needed

4. Set your theme in the application config:
```go
web.SetRenderer(web.NewRenderer("mytheme"))
```

## Tips

- Use CSS variables for easy theming
- Follow the existing component structure
- Test on mobile devices
- Check accessibility with screen readers
- Optimize images before adding

## License

This theme is part of OpenHost and follows the same MIT license.
