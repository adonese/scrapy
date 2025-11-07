# Usage Examples

Complete working examples for all data display components.

## Example 1: Basic Data Table

### Handler

```go
package handlers

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/adonese/cost-of-living/internal/ui/render"
    "github.com/adonese/cost-of-living/web/ui/components/data"
)

func (h *Handler) GetUsers(c echo.Context) error {
    // Parse parameters
    page := getIntParam(c, "page", 1)
    search := c.QueryParam("search")
    sortBy := c.QueryParam("sort_by")
    sortDir := c.QueryParam("sort_dir")

    // Fetch from database (example)
    users, total, err := h.userService.List(c.Request().Context(),
        page, 10, search, sortBy, sortDir)
    if err != nil {
        return err
    }

    // Convert to table data
    tableData := make([]map[string]interface{}, len(users))
    for i, user := range users {
        tableData[i] = map[string]interface{}{
            "id":       user.ID,
            "name":     user.Name,
            "email":    user.Email,
            "role":     user.Role,
            "status":   user.Status,
            "created":  user.CreatedAt.Format("2006-01-02"),
        }
    }

    // Configure table
    config := data.DataTableConfig{
        ID:      "users-table",
        Columns: []data.Column{
            {Key: "id", Label: "ID", Sortable: true, Width: "80px"},
            {Key: "name", Label: "Name", Sortable: true},
            {Key: "email", Label: "Email", Sortable: true},
            {Key: "role", Label: "Role", Sortable: true, Width: "120px"},
            {Key: "status", Label: "Status", Sortable: true, Width: "100px"},
            {Key: "created", Label: "Created", Sortable: true, Width: "120px"},
        },
        Data:              tableData,
        CurrentPage:       page,
        TotalPages:        (total + 9) / 10,
        TotalRecords:      total,
        PageSize:          10,
        SearchQuery:       search,
        SortColumn:        sortBy,
        SortDirection:     sortDir,
        SearchPlaceholder: "Search users by name, email, or role...",
        EmptyMessage:      "No users found",
        BaseURL:           "/api/users",
    }

    return render.Component(c, http.StatusOK, data.DataTable(config))
}
```

### Route

```go
e.GET("/api/users", handler.GetUsers)
```

### HTML Usage

```html
<!-- Initial page load -->
<div id="users-table-container">
    <!-- DataTable component renders here -->
</div>
```

---

## Example 2: Dashboard with Stat Cards

### Handler

```go
func (h *Handler) GetDashboardStats(c echo.Context) error {
    stats, err := h.analyticsService.GetStats(c.Request().Context())
    if err != nil {
        return err
    }

    cards := []templ.Component{
        data.StatCard(data.StatCardConfig{
            ID:          "total-users",
            Label:       "Total Users",
            Value:       fmt.Sprintf("%d", stats.TotalUsers),
            Icon:        data.IconUsers,
            Trend:       getTrend(stats.UserGrowth),
            TrendValue:  fmt.Sprintf("%+.1f%%", stats.UserGrowth),
            Variant:     "success",
            Description: "from last month",
            URL:         "/api/stats/users",
            RefreshInterval: "30s",
        }),
        data.StatCard(data.StatCardConfig{
            ID:          "revenue",
            Label:       "Monthly Revenue",
            Value:       fmt.Sprintf("$%s", formatCurrency(stats.Revenue)),
            Icon:        data.IconCurrency,
            Trend:       getTrend(stats.RevenueGrowth),
            TrendValue:  fmt.Sprintf("%+.1f%%", stats.RevenueGrowth),
            Variant:     "info",
            Description: "from last month",
        }),
        data.StatCard(data.StatCardConfig{
            ID:          "active-sessions",
            Label:       "Active Sessions",
            Value:       fmt.Sprintf("%d", stats.ActiveSessions),
            Icon:        data.IconActivity,
            Variant:     "default",
            Description: "right now",
            URL:         "/api/stats/sessions",
            RefreshInterval: "10s",
        }),
    }

    // Render all cards
    return render.Components(c, http.StatusOK, cards...)
}

func getTrend(value float64) string {
    if value > 0 {
        return "up"
    }
    if value < 0 {
        return "down"
    }
    return ""
}
```

### Templ Page

```templ
templ DashboardPage(stats StatsData) {
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold mb-8">Dashboard</h1>

        <!-- Stats Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            @data.StatCard(data.StatCardConfig{
                ID:          "total-users",
                Label:       "Total Users",
                Value:       fmt.Sprintf("%d", stats.TotalUsers),
                Icon:        data.IconUsers,
                Trend:       "up",
                TrendValue:  "+12.5%",
                Variant:     "success",
                Description: "from last month",
            })

            @data.StatCard(data.StatCardConfig{
                ID:          "revenue",
                Label:       "Revenue",
                Value:       "$45,678",
                Icon:        data.IconCurrency,
                Trend:       "up",
                TrendValue:  "+8.2%",
                Variant:     "info",
                Description: "from last month",
            })

            @data.StatCard(data.StatCardConfig{
                ID:          "active-sessions",
                Label:       "Active Sessions",
                Value:       "892",
                Icon:        data.IconActivity,
                Variant:     "warning",
                Description: "right now",
            })
        </div>
    </div>
}
```

---

## Example 3: Charts with Real-Time Updates

### Handler

```go
func (h *Handler) GetSalesChart(c echo.Context) error {
    // Fetch last 6 months of sales data
    salesData, err := h.salesService.GetMonthlySales(c.Request().Context(), 6)
    if err != nil {
        return err
    }

    labels := make([]string, len(salesData))
    values := make([]float64, len(salesData))
    for i, data := range salesData {
        labels[i] = data.Month
        values[i] = data.Amount
    }

    config := data.ChartConfig{
        ID:          "sales-chart",
        Title:       "Monthly Sales",
        Description: "Sales performance over the last 6 months",
        Height:      "400px",
        Type:        "line",
        Data: map[string]interface{}{
            "labels": labels,
            "datasets": []data.Dataset{
                {
                    Label:           "Sales ($)",
                    Data:            values,
                    BorderColor:     "rgb(59, 130, 246)",
                    BackgroundColor: "rgba(59, 130, 246, 0.1)",
                    BorderWidth:     2,
                    Fill:            true,
                    Tension:         0.4,
                },
            },
        },
        Options: map[string]interface{}{
            "responsive":          true,
            "maintainAspectRatio": false,
            "plugins": map[string]interface{}{
                "legend": map[string]interface{}{
                    "position": "top",
                },
                "tooltip": map[string]interface{}{
                    "mode":      "index",
                    "intersect": false,
                    "callbacks": map[string]string{
                        "label": "function(ctx) { return '$' + ctx.parsed.y.toLocaleString(); }",
                    },
                },
            },
            "scales": map[string]interface{}{
                "y": map[string]interface{}{
                    "beginAtZero": true,
                    "ticks": map[string]interface{}{
                        "callback": "function(value) { return '$' + value.toLocaleString(); }",
                    },
                },
            },
        },
        ShowExport:      true,
        RefreshURL:      "/api/charts/sales",
        RefreshInterval: "60s", // Refresh every minute
    }

    return render.Component(c, http.StatusOK, data.ChartWrapper(config))
}
```

### Templ Page

```templ
templ AnalyticsPage() {
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold mb-8">Analytics</h1>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- Line Chart -->
            <div id="sales-chart-container">
                @data.ChartWrapper(data.ChartConfig{
                    ID:     "sales-chart",
                    Title:  "Monthly Sales",
                    Height: "300px",
                    Type:   "line",
                    // ... config
                })
            </div>

            <!-- Pie Chart -->
            <div id="category-chart-container">
                @data.ChartWrapper(data.ChartConfig{
                    ID:     "category-chart",
                    Title:  "Sales by Category",
                    Height: "300px",
                    Type:   "doughnut",
                    // ... config
                })
            </div>
        </div>
    </div>
}
```

---

## Example 4: Combined Dashboard

Full dashboard with stats, charts, and data table.

### Templ Page

```templ
package ui

import (
    "github.com/adonese/cost-of-living/web/ui/components/data"
    "fmt"
)

templ FullDashboard(dashData DashboardData) {
    @BaseLayout("Dashboard") {
        <div class="container mx-auto px-4 py-8">
            <!-- Page Header -->
            <div class="mb-8">
                <h1 class="text-3xl font-bold text-slate-900">Dashboard</h1>
                <p class="text-slate-600 mt-2">Real-time overview of your application</p>
            </div>

            <!-- Stats Cards -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                @data.StatCard(data.StatCardConfig{
                    ID:          "stat-users",
                    Label:       "Total Users",
                    Value:       fmt.Sprintf("%d", dashData.Stats.TotalUsers),
                    Icon:        data.IconUsers,
                    Trend:       "up",
                    TrendValue:  "+12.5%",
                    Variant:     "success",
                    Description: "from last month",
                    URL:         "/api/stats/users",
                    RefreshInterval: "30s",
                })

                @data.StatCard(data.StatCardConfig{
                    ID:          "stat-revenue",
                    Label:       "Revenue",
                    Value:       fmt.Sprintf("$%.2f", dashData.Stats.Revenue),
                    Icon:        data.IconCurrency,
                    Trend:       "up",
                    TrendValue:  "+8.2%",
                    Variant:     "info",
                    Description: "from last month",
                })

                @data.StatCard(data.StatCardConfig{
                    ID:          "stat-sessions",
                    Label:       "Active Sessions",
                    Value:       fmt.Sprintf("%d", dashData.Stats.ActiveSessions),
                    Icon:        data.IconActivity,
                    Variant:     "default",
                    Description: "right now",
                    URL:         "/api/stats/sessions",
                    RefreshInterval: "10s",
                })
            </div>

            <!-- Charts -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                @data.SimpleChart(
                    "sales-trend",
                    "Sales Trend",
                    "line",
                    []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
                    []data.Dataset{
                        {
                            Label:           "Sales",
                            Data:            []float64{12000, 19000, 15000, 25000, 22000, 30000},
                            BorderColor:     "rgb(59, 130, 246)",
                            BackgroundColor: "rgba(59, 130, 246, 0.1)",
                            Fill:            true,
                            Tension:         0.4,
                        },
                    },
                    "300px",
                )

                @data.SimpleChart(
                    "category-distribution",
                    "Sales by Category",
                    "doughnut",
                    []string{"Electronics", "Clothing", "Food", "Books"},
                    []data.Dataset{
                        {
                            Label: "Sales",
                            Data:  []float64{35, 25, 20, 20},
                            BackgroundColor: []string{
                                "rgb(59, 130, 246)",
                                "rgb(16, 185, 129)",
                                "rgb(245, 158, 11)",
                                "rgb(239, 68, 68)",
                            },
                        },
                    },
                    "300px",
                )
            </div>

            <!-- Recent Transactions Table -->
            <div class="bg-white rounded-lg border border-slate-200 p-6">
                <h2 class="text-xl font-semibold mb-4">Recent Transactions</h2>
                @data.DataTable(data.DataTableConfig{
                    ID:      "transactions-table",
                    Columns: []data.Column{
                        {Key: "id", Label: "ID", Sortable: true, Width: "80px"},
                        {Key: "customer", Label: "Customer", Sortable: true},
                        {Key: "amount", Label: "Amount", Sortable: true, Width: "120px"},
                        {Key: "status", Label: "Status", Sortable: true, Width: "100px"},
                        {Key: "date", Label: "Date", Sortable: true, Width: "120px"},
                    },
                    Data:              dashData.Transactions,
                    CurrentPage:       1,
                    TotalPages:        10,
                    TotalRecords:      100,
                    PageSize:          10,
                    SearchPlaceholder: "Search transactions...",
                    EmptyMessage:      "No transactions found",
                    BaseURL:           "/api/transactions",
                })
            </div>
        </div>
    }
}
```

---

## Example 5: Custom Chart with Advanced Options

```go
func (h *Handler) GetAdvancedChart(c echo.Context) error {
    config := data.ChartConfig{
        ID:     "advanced-chart",
        Title:  "Multi-Dataset Comparison",
        Height: "400px",
        Type:   "bar",
        Data: map[string]interface{}{
            "labels": []string{"Q1", "Q2", "Q3", "Q4"},
            "datasets": []data.Dataset{
                {
                    Label:           "2023",
                    Data:            []float64{45, 52, 48, 60},
                    BackgroundColor: "rgba(59, 130, 246, 0.8)",
                    BorderColor:     "rgb(59, 130, 246)",
                    BorderWidth:     1,
                },
                {
                    Label:           "2024",
                    Data:            []float64{50, 58, 55, 70},
                    BackgroundColor: "rgba(16, 185, 129, 0.8)",
                    BorderColor:     "rgb(16, 185, 129)",
                    BorderWidth:     1,
                },
            },
        },
        Options: map[string]interface{}{
            "responsive":          true,
            "maintainAspectRatio": false,
            "interaction": map[string]interface{}{
                "mode":      "index",
                "intersect": false,
            },
            "plugins": map[string]interface{}{
                "legend": map[string]interface{}{
                    "position": "top",
                },
                "title": map[string]interface{}{
                    "display": true,
                    "text":    "Quarterly Performance Comparison",
                },
            },
            "scales": map[string]interface{}{
                "y": map[string]interface{}{
                    "beginAtZero": true,
                    "title": map[string]interface{}{
                        "display": true,
                        "text":    "Revenue ($K)",
                    },
                },
                "x": map[string]interface{}{
                    "title": map[string]interface{}{
                        "display": true,
                        "text":    "Quarter",
                    },
                },
            },
        },
        ShowExport: true,
    }

    return render.Component(c, http.StatusOK, data.ChartWrapper(config))
}
```

---

## Testing Examples

### Testing Data Table Handler

```go
func TestGetDataTable(t *testing.T) {
    e := echo.New()

    tests := []struct {
        name       string
        query      string
        wantStatus int
        contains   string
    }{
        {
            name:       "basic table",
            query:      "?page=1",
            wantStatus: http.StatusOK,
            contains:   "data-table-wrapper",
        },
        {
            name:       "with search",
            query:      "?page=1&search=alice",
            wantStatus: http.StatusOK,
            contains:   "alice",
        },
        {
            name:       "with sorting",
            query:      "?page=1&sort_by=name&sort_dir=asc",
            wantStatus: http.StatusOK,
            contains:   "data-table-wrapper",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/api/data"+tt.query, nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)

            handler := NewDataTableExamplesHandler()
            err := handler.GetDataTable(c)

            assert.NoError(t, err)
            assert.Equal(t, tt.wantStatus, rec.Code)
            assert.Contains(t, rec.Body.String(), tt.contains)
        })
    }
}
```

---

## Performance Tips

1. **Pagination**: Always paginate large datasets
2. **Debouncing**: Use 300-500ms delay for search inputs
3. **Caching**: Cache frequently accessed chart data
4. **Lazy Loading**: Load charts only when visible
5. **Connection Pooling**: Reuse database connections

---

## Common Pitfalls

1. **Forgetting hx-include**: Sort/search state gets lost during pagination
2. **Wrong target ID**: Table doesn't update
3. **Missing validation**: Invalid page numbers cause errors
4. **No empty states**: Poor UX when no data
5. **Over-polling**: Too frequent refreshes waste resources
