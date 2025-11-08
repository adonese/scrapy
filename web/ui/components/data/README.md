# Data Display Components

HTMX-powered data display components for Templ + Go applications.

## Components

### 1. DataTable

A responsive data table with server-side sorting, filtering, and pagination using HTMX.

#### Features
- **Responsive Design**: Desktop shows full table, mobile shows stacked cards
- **Server-Side Sorting**: Click column headers to sort (HTMX request)
- **Server-Side Filtering**: Search box with debounced HTMX requests
- **Server-Side Pagination**: Page navigation via HTMX
- **Loading Indicators**: Built-in HTMX indicators

#### Usage

```go
import "github.com/adonese/cost-of-living/web/ui/components/data"

// In your handler
config := data.DataTableConfig{
    ID:                "users-table",
    Columns: []data.Column{
        {Key: "id", Label: "ID", Sortable: true, Width: "80px"},
        {Key: "name", Label: "Name", Sortable: true},
        {Key: "email", Label: "Email", Sortable: true},
    },
    Data:              pageData, // Current page data
    CurrentPage:       page,
    TotalPages:        totalPages,
    TotalRecords:      totalRecords,
    PageSize:          pageSize,
    SearchQuery:       search,
    SortColumn:        sortBy,
    SortDirection:     sortDir,
    SearchPlaceholder: "Search users...",
    EmptyMessage:      "No users found",
    BaseURL:           "/api/users",
}

return render.Component(c, http.StatusOK, data.DataTable(config))
```

#### HTMX Attributes Used
- `hx-get`: Fetch sorted/filtered/paginated data
- `hx-trigger`: Debounced search (500ms delay)
- `hx-target`: Update the table container
- `hx-swap`: Replace entire table (outerHTML)
- `hx-include`: Include search/sort parameters in requests
- `hx-vals`: Pass sort column/direction values

---

### 2. StatCard

Statistics display card with optional trend indicators and auto-refresh.

#### Features
- **Color Variants**: default, success, warning, danger, info
- **Trend Indicators**: Up/down arrows with percentage
- **Auto-Refresh**: Optional HTMX polling
- **Icon Support**: Pass SVG icons as props

#### Usage

```go
config := data.StatCardConfig{
    ID:          "total-users",
    Label:       "Total Users",
    Value:       "1,234",
    Icon:        data.IconUsers,
    Trend:       "up",
    TrendValue:  "+12.5%",
    Variant:     "success",
    Description: "from last month",
    URL:         "/api/stats/users", // Optional: for auto-refresh
    RefreshInterval: "30s",          // Optional: polling interval
}

return render.Component(c, http.StatusOK, data.StatCard(config))
```

#### Available Icons
- `data.IconUsers`
- `data.IconCurrency`
- `data.IconChart`
- `data.IconActivity`

#### HTMX Attributes Used
- `hx-get`: Auto-refresh endpoint
- `hx-trigger`: Polling interval (e.g., "every 30s")
- `hx-swap`: Replace card (outerHTML)

---

### 3. ChartWrapper

Container for Chart.js charts with HTMX refresh support.

#### Features
- **Chart.js Integration**: Automatic initialization
- **Auto-Refresh**: Optional HTMX polling for live data
- **Export Options**: PNG, SVG download and print
- **Responsive**: Maintains aspect ratio
- **Chart Types**: line, bar, pie, doughnut, radar

#### Setup

First, add Chart.js to your base layout:

```html
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
```

#### Usage

```go
config := data.ChartConfig{
    ID:          "sales-chart",
    Title:       "Monthly Sales",
    Description: "Sales over the last 6 months",
    Height:      "300px",
    Type:        "line",
    Data: map[string]interface{}{
        "labels": []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
        "datasets": []data.Dataset{
            {
                Label:           "Sales",
                Data:            []float64{12, 19, 15, 25, 22, 30},
                BorderColor:     "rgb(59, 130, 246)",
                BackgroundColor: "rgba(59, 130, 246, 0.1)",
                Fill:            true,
                Tension:         0.4,
            },
        },
    },
    Options:     data.getDefaultChartOptions("line"),
    ShowExport:  true,
    RefreshURL:  "/api/chart/sales",  // Optional
    RefreshInterval: "30s",           // Optional
}

return render.Component(c, http.StatusOK, data.ChartWrapper(config))
```

#### Simple Chart Helper

For quick charts:

```go
data.SimpleChart(
    "my-chart",
    "Sales Data",
    "bar",
    []string{"Jan", "Feb", "Mar"},
    []data.Dataset{{
        Label: "Sales",
        Data:  []float64{100, 200, 150},
    }},
    "300px",
)
```

#### HTMX Attributes Used
- `hx-get`: Refresh chart data
- `hx-trigger`: Polling interval
- `hx-swap`: Replace chart wrapper (outerHTML)

---

### 4. Pagination

Standalone pagination controls with HTMX navigation.

#### Features
- **Responsive**: Shows page numbers on desktop, page counter on mobile
- **Smart Range**: Shows ellipsis for large page counts
- **Disabled States**: Prev/Next buttons disabled at boundaries
- **Configurable**: Customizable max buttons

#### Usage

```go
config := data.PaginationConfig{
    CurrentPage:  page,
    TotalPages:   totalPages,
    TotalRecords: totalRecords,
    PageSize:     pageSize,
    BaseURL:      "/api/data",
    TargetID:     "data-container",
    ShowInfo:     true,
    MaxButtons:   7,
}

return render.Component(c, http.StatusOK, data.Pagination(config))
```

#### HTMX Attributes Used
- `hx-get`: Navigate to page
- `hx-target`: Update target container
- `hx-swap`: Replace content (outerHTML)
- `hx-include`: Include search/sort parameters

---

## Handler Examples

See `/internal/ui/handlers/data_table_examples.go` for complete working examples of:

1. **GetDataTable**: Handle sorting, filtering, and pagination
2. **GetStatCards**: Return stat card data
3. **GetChartData**: Return chart configurations

### Example Route Setup

```go
// In your routes file
dataHandler := handlers.NewDataTableExamplesHandler()

e.GET("/api/data-table", dataHandler.GetDataTable)
e.GET("/api/stats", dataHandler.GetStatCards)
e.GET("/api/chart/:id", dataHandler.GetChartData)
```

---

## HTMX Patterns Summary

### Pattern 1: Server-Side Search with Debouncing

```html
<input
    type="text"
    name="search"
    hx-get="/api/data"
    hx-trigger="keyup changed delay:500ms"
    hx-target="#data-table"
    hx-swap="outerHTML"
    hx-include="[name='sort_by'],[name='sort_dir']"
/>
```

### Pattern 2: Server-Side Sorting

```html
<button
    hx-get="/api/data"
    hx-vals='{"sort_by": "name", "sort_dir": "asc"}'
    hx-target="#data-table"
    hx-swap="outerHTML"
    hx-include="[name='search']"
>
    Sort
</button>
```

### Pattern 3: Server-Side Pagination

```html
<button
    hx-get="/api/data?page=2"
    hx-target="#data-table"
    hx-swap="outerHTML"
    hx-include="[name='search'],[name='sort_by'],[name='sort_dir']"
>
    Next Page
</button>
```

### Pattern 4: Auto-Refresh (Polling)

```html
<div
    hx-get="/api/stats"
    hx-trigger="every 30s"
    hx-swap="outerHTML"
>
    <!-- Content refreshes every 30 seconds -->
</div>
```

### Pattern 5: Loading Indicators

```html
<!-- Add indicator to any HTMX element -->
<div hx-get="/api/data" hx-indicator="#spinner">
    <div id="spinner" class="htmx-indicator">
        Loading...
    </div>
</div>
```

---

## Best Practices

### 1. Always Return the Same Component Structure

When handling HTMX requests, always return the complete component with the same ID. This ensures proper replacement.

### 2. Include Context in Requests

Use `hx-include` to preserve search/sort state across pagination:

```html
hx-include="[name='search'],[name='sort_by'],[name='sort_dir']"
```

### 3. Handle Empty States

Always provide meaningful empty states:

```go
config := data.DataTableConfig{
    EmptyMessage: "No results found. Try adjusting your search.",
    // ...
}
```

### 4. Debounce Search Inputs

Add delay to search to avoid excessive requests:

```html
hx-trigger="keyup changed delay:500ms"
```

### 5. Validate Parameters Server-Side

Always validate page numbers, sort columns, etc.:

```go
if page < 1 {
    page = 1
}
if page > totalPages && totalPages > 0 {
    page = totalPages
}
```

---

## Browser Support

These components require:
- HTMX 1.9+
- Alpine.js 3.x (for dropdowns in ChartWrapper)
- Chart.js 4.x (for ChartWrapper only)
- Modern browser with ES6 support

---

## Testing

See `/internal/ui/handlers/data_table_examples_test.go` for test examples.

To test HTMX interactions:

```go
func TestGetDataTable(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/data-table?page=1&search=alice", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    handler := NewDataTableExamplesHandler()
    err := handler.GetDataTable(c)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), "alice")
}
```

---

## Troubleshooting

### Table not updating on search
- Check that `hx-target` points to correct element ID
- Verify `hx-include` includes all necessary form inputs
- Check browser console for HTMX errors

### Chart not rendering
- Ensure Chart.js is loaded before component
- Check console for Chart.js errors
- Verify data format matches Chart.js requirements

### Pagination not working
- Ensure page numbers are validated server-side
- Check that BaseURL matches your route
- Verify TargetID matches the container element

---

## License

MIT
