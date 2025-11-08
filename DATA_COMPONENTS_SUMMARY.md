# Data Display Components - Implementation Summary

HTMX-powered data display components for Templ + Go, created for the Scrapy Cost of Living project.

## Files Created

### Components (Templ)

1. **`/web/ui/components/data/data-table.templ`**
   - Responsive data table with desktop (full table) and mobile (card) views
   - Server-side sorting via HTMX
   - Server-side filtering with debounced search (500ms)
   - Server-side pagination
   - Loading indicators
   - Empty states

2. **`/web/ui/components/data/stat-card.templ`**
   - Statistics display cards
   - 5 color variants (default, success, warning, danger, info)
   - Trend indicators (up/down arrows with percentages)
   - Optional auto-refresh via HTMX polling
   - Predefined SVG icons (Users, Currency, Chart, Activity)

3. **`/web/ui/components/data/chart-wrapper.templ`**
   - Container for Chart.js charts
   - Inline script for Chart.js initialization
   - Support for all Chart.js chart types (line, bar, pie, doughnut, radar)
   - Optional auto-refresh via HTMX
   - Export options (PNG, SVG, Print)
   - Responsive canvas element
   - SimpleChart helper for quick chart creation

4. **`/web/ui/components/data/pagination.templ`**
   - HTMX-powered pagination controls
   - Responsive: page numbers on desktop, page counter on mobile
   - Smart page range with ellipsis for large datasets
   - Disabled states for boundaries
   - Results info display
   - Preserves search/sort state via hx-include

### Handlers (Go)

5. **`/internal/ui/handlers/data_table_examples.go`**
   - Complete working examples for all components
   - `GetDataTable()`: Handles sorting, filtering, and pagination
   - `GetStatCards()`: Returns stat card configurations
   - `GetChartData()`: Returns chart configurations with data
   - Helper functions for filtering, sorting, and parameter parsing

### Documentation

6. **`/web/ui/components/data/README.md`**
   - Comprehensive component documentation
   - HTMX patterns explained
   - Best practices
   - Troubleshooting guide
   - Browser requirements

7. **`/web/ui/components/data/USAGE_EXAMPLES.md`**
   - 5 complete working examples
   - Testing examples
   - Performance tips
   - Common pitfalls

8. **`/DATA_COMPONENTS_SUMMARY.md`** (this file)
   - Implementation overview
   - Quick reference

## Component Features

### DataTable

**Props:**
- `ID` - Unique identifier for HTMX targeting
- `Columns` - Array of column definitions
- `Data` - Current page data
- `CurrentPage`, `TotalPages`, `TotalRecords`, `PageSize` - Pagination state
- `SearchQuery`, `SortColumn`, `SortDirection` - Filter/sort state
- `BaseURL` - Endpoint for HTMX requests
- `SearchPlaceholder`, `EmptyMessage` - UI text

**HTMX Attributes:**
- Search input: `hx-get`, `hx-trigger="keyup changed delay:500ms"`, `hx-target`, `hx-swap`, `hx-include`
- Sort buttons: `hx-get`, `hx-vals`, `hx-target`, `hx-swap`, `hx-include`
- Pagination: Embedded Pagination component

**Responsive Behavior:**
- Desktop (sm+): Full table with sortable columns
- Mobile (<sm): Stacked cards with label-value pairs

### StatCard

**Props:**
- `ID`, `Label`, `Value` - Basic info
- `Icon` - SVG icon string
- `Trend`, `TrendValue` - Trend indicator (up/down with percentage)
- `Variant` - Color scheme (default, success, warning, danger, info)
- `Description` - Additional context
- `URL`, `RefreshInterval` - Optional auto-refresh

**HTMX Attributes:**
- `hx-get` - Refresh endpoint
- `hx-trigger` - Polling interval (e.g., "every 30s")
- `hx-swap` - Replace strategy

**Predefined Icons:**
- `data.IconUsers`
- `data.IconCurrency`
- `data.IconChart`
- `data.IconActivity`

### ChartWrapper

**Props:**
- `ID`, `Title`, `Description` - Chart metadata
- `Height` - Canvas height (e.g., "300px")
- `Type` - Chart.js type (line, bar, pie, doughnut, radar)
- `Data` - Chart.js data object
- `Options` - Chart.js options object
- `ShowExport` - Enable export menu
- `RefreshURL`, `RefreshInterval` - Optional auto-refresh

**Features:**
- Automatic Chart.js initialization via inline script
- Destroys existing chart instance before re-rendering
- Default options helper: `getDefaultChartOptions(chartType)`
- SimpleChart helper for quick charts
- Dataset struct for type-safe dataset configuration

**HTMX Attributes:**
- `hx-get` - Refresh endpoint
- `hx-trigger` - Polling interval
- `hx-swap` - Replace wrapper

### Pagination

**Props:**
- `CurrentPage`, `TotalPages`, `TotalRecords`, `PageSize` - Pagination state
- `BaseURL` - Endpoint for page requests
- `TargetID` - Container to update
- `ShowInfo` - Display results info
- `MaxButtons` - Max page buttons (default: 7)

**HTMX Attributes:**
- Previous/Next buttons: `hx-get`, `hx-target`, `hx-swap`, `hx-include`
- Page buttons: `hx-get`, `hx-target`, `hx-swap`, `hx-include`

**Smart Features:**
- Ellipsis for large page counts
- First/last page always visible (when >MaxButtons)
- Disabled states at boundaries
- Mobile: Shows "X / Y" counter
- Desktop: Shows page numbers

## HTMX Patterns Used

### 1. Server-Side Search with Debouncing
```html
<input
    hx-get="/api/data"
    hx-trigger="keyup changed delay:500ms"
    hx-target="#data-table"
    hx-swap="outerHTML"
    hx-include="[name='sort_by'],[name='sort_dir']"
/>
```

### 2. Server-Side Sorting
```html
<button
    hx-get="/api/data"
    hx-vals='{"sort_by": "name", "sort_dir": "asc"}'
    hx-target="#data-table"
    hx-swap="outerHTML"
    hx-include="[name='search']"
>
```

### 3. Server-Side Pagination
```html
<button
    hx-get="/api/data?page=2"
    hx-target="#data-table"
    hx-swap="outerHTML"
    hx-include="[name='search'],[name='sort_by'],[name='sort_dir']"
>
```

### 4. Auto-Refresh (Polling)
```html
<div
    hx-get="/api/stats"
    hx-trigger="every 30s"
    hx-swap="outerHTML"
>
```

### 5. Loading Indicators
```html
<div class="htmx-indicator">
    <!-- Spinner SVG -->
</div>
```

## Handler Pattern

All HTMX handlers follow this pattern:

```go
func (h *Handler) GetData(c echo.Context) error {
    // 1. Parse query parameters
    page := getIntParam(c, "page", 1)
    search := c.QueryParam("search")
    sortBy := c.QueryParam("sort_by")
    sortDir := c.QueryParam("sort_dir")

    // 2. Fetch data (from database/service)
    data, total, err := h.service.List(ctx, page, pageSize, search, sortBy, sortDir)
    if err != nil {
        return err
    }

    // 3. Transform to component format
    config := data.DataTableConfig{
        ID:            "my-table",
        Data:          data,
        CurrentPage:   page,
        TotalRecords:  total,
        // ... more config
    }

    // 4. Render component
    return render.Component(c, http.StatusOK, data.DataTable(config))
}
```

## Route Setup

```go
// Create handler
dataHandler := handlers.NewDataTableExamplesHandler()

// Register routes
e.GET("/api/data-table", dataHandler.GetDataTable)
e.GET("/api/stats", dataHandler.GetStatCards)
e.GET("/api/chart/:id", dataHandler.GetChartData)
```

## Quick Start

### 1. Add Chart.js to Base Layout (Already Done)

```html
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>
```

### 2. Use DataTable

```go
import "github.com/adonese/cost-of-living/web/ui/components/data"

config := data.DataTableConfig{
    ID:      "users-table",
    Columns: []data.Column{
        {Key: "id", Label: "ID", Sortable: true},
        {Key: "name", Label: "Name", Sortable: true},
    },
    Data:     myData,
    BaseURL:  "/api/users",
    // ... other config
}

return render.Component(c, http.StatusOK, data.DataTable(config))
```

### 3. Use StatCard

```go
config := data.StatCardConfig{
    Label:       "Total Users",
    Value:       "1,234",
    Icon:        data.IconUsers,
    Trend:       "up",
    TrendValue:  "+12.5%",
    Variant:     "success",
}

return render.Component(c, http.StatusOK, data.StatCard(config))
```

### 4. Use ChartWrapper

```go
config := data.ChartConfig{
    ID:     "my-chart",
    Title:  "Sales",
    Height: "300px",
    Type:   "line",
    Data: map[string]interface{}{
        "labels": []string{"Jan", "Feb", "Mar"},
        "datasets": []data.Dataset{
            {
                Label: "Sales",
                Data:  []float64{100, 200, 150},
            },
        },
    },
    Options: data.getDefaultChartOptions("line"),
}

return render.Component(c, http.StatusOK, data.ChartWrapper(config))
```

## Key Design Decisions

1. **Server-Side Everything**: All sorting, filtering, pagination happens server-side for scalability
2. **HTMX for Interactivity**: No JavaScript needed for data operations
3. **Component IDs**: Each component needs unique ID for HTMX targeting
4. **State Preservation**: Use `hx-include` to preserve search/sort state across pagination
5. **Debouncing**: 500ms delay on search to reduce server load
6. **Responsive**: Mobile-first design with card fallback
7. **Loading States**: Built-in indicators for better UX
8. **Empty States**: Always provide meaningful empty messages
9. **Type Safety**: Go structs for all configurations
10. **Reusability**: All components accept config structs

## Browser Requirements

- HTMX 1.9+ (included in base layout)
- Alpine.js 3.x (for dropdown menus in export)
- Chart.js 4.x (for ChartWrapper only)
- Modern browser with ES6 support

## Performance Considerations

1. **Pagination**: Always paginate large datasets (default: 10 items/page)
2. **Debouncing**: Search inputs debounced by 500ms
3. **Polling**: Use reasonable intervals (30s+ for most stats)
4. **Caching**: Cache chart data when possible
5. **Database**: Use proper indexes on sortable columns

## Testing

See `/internal/ui/handlers/data_table_examples.go` for implementation patterns.

Test HTMX handlers like regular Echo handlers:

```go
func TestHandler(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/data?page=1", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    handler := NewHandler()
    err := handler.GetData(c)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## Next Steps

1. Generate templ files: `templ generate web/ui/components/data/*.templ`
2. Add routes to your Echo router
3. Create handlers for your specific use cases
4. Test HTMX interactions in browser
5. Add CSS animations for smooth transitions (optional)

## Common Issues & Solutions

### Issue: Table doesn't update on search
**Solution**: Check `hx-target` matches table ID, verify `hx-include` includes all inputs

### Issue: Pagination loses search/sort
**Solution**: Add `hx-include="[name='search'],[name='sort_by'],[name='sort_dir']"` to pagination buttons

### Issue: Chart not rendering
**Solution**: Ensure Chart.js loaded before component, check browser console for errors

### Issue: HTMX request not firing
**Solution**: Check browser network tab, verify route exists, check for JavaScript errors

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Browser (Client)                    │
├─────────────────────────────────────────────────────────┤
│  User Action (search, sort, paginate)                   │
│         ↓                                               │
│  HTMX sends HTTP request with parameters                │
│         ↓                                               │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                   Server (Go + Echo)                    │
├─────────────────────────────────────────────────────────┤
│  Handler receives request                               │
│         ↓                                               │
│  Parse parameters (page, search, sort)                  │
│         ↓                                               │
│  Query database with filters                            │
│         ↓                                               │
│  Build component config                                 │
│         ↓                                               │
│  Render Templ component                                 │
│         ↓                                               │
│  Return HTML fragment                                   │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                     Browser (Client)                    │
├─────────────────────────────────────────────────────────┤
│  HTMX receives HTML                                     │
│         ↓                                               │
│  Swaps content (outerHTML)                              │
│         ↓                                               │
│  Chart.js initializes (if chart component)              │
│         ↓                                               │
│  User sees updated content                              │
└─────────────────────────────────────────────────────────┘
```

## File Locations

```
scrapy/
├── web/ui/components/data/
│   ├── data-table.templ          # Responsive data table
│   ├── stat-card.templ           # Statistics card
│   ├── chart-wrapper.templ       # Chart.js wrapper
│   ├── pagination.templ          # Pagination controls
│   ├── README.md                 # Component docs
│   └── USAGE_EXAMPLES.md         # Usage examples
├── internal/ui/handlers/
│   └── data_table_examples.go   # Handler examples
└── DATA_COMPONENTS_SUMMARY.md   # This file
```

## Credits

Built with:
- [HTMX](https://htmx.org/) - High power tools for HTML
- [Templ](https://templ.guide/) - Type-safe Go templating
- [Chart.js](https://www.chartjs.org/) - Simple yet flexible JavaScript charting
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [Echo](https://echo.labstack.com/) - High performance Go web framework

---

**Created**: 2025-11-07
**Project**: Scrapy - UAE Cost of Living Tracker
**Stack**: Go + Echo + Templ + HTMX + Alpine.js + Chart.js
