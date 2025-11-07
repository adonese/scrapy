package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/adonese/cost-of-living/internal/ui/render"
	"github.com/adonese/cost-of-living/web/ui/components/data"
)

// DataTableExamplesHandler demonstrates HTMX-powered data table interactions
type DataTableExamplesHandler struct {
	// In a real app, you'd inject your data service here
}

// NewDataTableExamplesHandler creates a new handler instance
func NewDataTableExamplesHandler() *DataTableExamplesHandler {
	return &DataTableExamplesHandler{}
}

// GetDataTable handles GET requests for the data table with sorting, filtering, and pagination
// Route: GET /api/data-table
func (h *DataTableExamplesHandler) GetDataTable(c echo.Context) error {
	// Parse query parameters
	page := getIntParam(c, "page", 1)
	pageSize := getIntParam(c, "page_size", 10)
	search := c.QueryParam("search")
	sortBy := c.QueryParam("sort_by")
	sortDir := c.QueryParam("sort_dir")

	// Fetch data (in real app, this would come from a database)
	allData := h.fetchSampleData()

	// Apply search filter
	filteredData := h.filterData(allData, search)

	// Apply sorting
	sortedData := h.sortData(filteredData, sortBy, sortDir)

	// Calculate pagination
	totalRecords := len(sortedData)
	totalPages := (totalRecords + pageSize - 1) / pageSize
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	// Get page data
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalRecords {
		end = totalRecords
	}

	var pageData []map[string]interface{}
	if start < totalRecords {
		pageData = sortedData[start:end]
	} else {
		pageData = []map[string]interface{}{}
	}

	// Create table configuration
	config := data.DataTableConfig{
		ID:                "data-table-example",
		Columns:           h.getTableColumns(),
		Data:              pageData,
		CurrentPage:       page,
		TotalPages:        totalPages,
		TotalRecords:      totalRecords,
		PageSize:          pageSize,
		SearchQuery:       search,
		SortColumn:        sortBy,
		SortDirection:     sortDir,
		SearchPlaceholder: "Search data...",
		EmptyMessage:      "No data found",
		BaseURL:           "/api/data-table",
	}

	// Render the data table component
	return render.Component(c, http.StatusOK, data.DataTable(config))
}

// GetStatCards handles GET requests for stat cards (example with auto-refresh)
// Route: GET /api/stats
func (h *DataTableExamplesHandler) GetStatCards(c echo.Context) error {
	// Fetch stats (in real app, this would come from your analytics service)
	stats := []data.StatCardConfig{
		{
			ID:          "total-users",
			Label:       "Total Users",
			Value:       "1,234",
			Icon:        data.IconUsers,
			Trend:       "up",
			TrendValue:  "+12.5%",
			Variant:     "success",
			Description: "from last month",
		},
		{
			ID:          "revenue",
			Label:       "Revenue",
			Value:       "$45,678",
			Icon:        data.IconCurrency,
			Trend:       "up",
			TrendValue:  "+8.2%",
			Variant:     "info",
			Description: "from last month",
		},
		{
			ID:          "active-sessions",
			Label:       "Active Sessions",
			Value:       "892",
			Icon:        data.IconActivity,
			Trend:       "down",
			TrendValue:  "-3.1%",
			Variant:     "warning",
			Description: "from last hour",
		},
	}

	// For HTMX, you could return individual stat cards
	// This example shows how to return all cards
	// In practice, you might have separate endpoints for each stat
	return c.JSON(http.StatusOK, stats)
}

// GetChartData handles GET requests for chart data with HTMX refresh
// Route: GET /api/chart/:id
func (h *DataTableExamplesHandler) GetChartData(c echo.Context) error {
	chartID := c.Param("id")

	// Example chart configurations
	switch chartID {
	case "sales-chart":
		config := data.ChartConfig{
			ID:          "sales-chart",
			Title:       "Monthly Sales",
			Description: "Sales performance over the last 6 months",
			Height:      "300px",
			Type:        "line",
			Data: map[string]interface{}{
				"labels": []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
				"datasets": []data.Dataset{
					{
						Label:           "Sales",
						Data:            []float64{12000, 19000, 15000, 25000, 22000, 30000},
						BorderColor:     "rgb(59, 130, 246)",
						BackgroundColor: "rgba(59, 130, 246, 0.1)",
						BorderWidth:     2,
						Fill:            true,
						Tension:         0.4,
					},
				},
			},
			Options:     data.GetDefaultChartOptions("line"),
			ShowExport:  true,
			RefreshURL:  "/api/chart/sales-chart",
			RefreshInterval: "30s",
		}
		return render.Component(c, http.StatusOK, data.ChartWrapper(config))

	case "category-chart":
		config := data.ChartConfig{
			ID:          "category-chart",
			Title:       "Sales by Category",
			Description: "Distribution of sales across categories",
			Height:      "300px",
			Type:        "doughnut",
			Data: map[string]interface{}{
				"labels": []string{"Electronics", "Clothing", "Food", "Books", "Other"},
				"datasets": []data.Dataset{
					{
						Label: "Sales",
						Data:  []float64{35, 25, 20, 12, 8},
						BackgroundColor: []string{
							"rgb(59, 130, 246)",
							"rgb(16, 185, 129)",
							"rgb(245, 158, 11)",
							"rgb(239, 68, 68)",
							"rgb(156, 163, 175)",
						},
						BorderWidth: 0,
					},
				},
			},
			Options:    data.GetDefaultChartOptions("doughnut"),
			ShowExport: true,
		}
		return render.Component(c, http.StatusOK, data.ChartWrapper(config))

	default:
		return echo.NewHTTPError(http.StatusNotFound, "Chart not found")
	}
}

// ExportDataTableCSV exports the data table as CSV
// Route: GET /api/data-table/export
func (h *DataTableExamplesHandler) ExportDataTableCSV(c echo.Context) error {
	// Parse query parameters for filtering/sorting
	search := c.QueryParam("search")
	sortBy := c.QueryParam("sort_by")
	sortDir := c.QueryParam("sort_dir")

	// Fetch all data
	allData := h.fetchSampleData()

	// Apply filters and sorting
	filteredData := h.filterData(allData, search)
	sortedData := h.sortData(filteredData, sortBy, sortDir)

	// Get columns
	columns := h.getTableColumns()

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header row
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Label
	}
	if err := writer.Write(headers); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to write CSV header")
	}

	// Write data rows
	for _, row := range sortedData {
		record := make([]string, len(columns))
		for i, col := range columns {
			record[i] = fmt.Sprintf("%v", row[col.Key])
		}
		if err := writer.Write(record); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to write CSV row")
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to flush CSV writer")
	}

	// Set CSV headers
	filename := fmt.Sprintf("data-export-%s.csv", time.Now().Format("2006-01-02-15-04-05"))
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Blob(http.StatusOK, "text/csv", buf.Bytes())
}

// Helper functions

func (h *DataTableExamplesHandler) fetchSampleData() []map[string]interface{} {
	// Sample data - in a real app, this would come from a database
	return []map[string]interface{}{
		{"id": 1, "name": "Alice Johnson", "email": "alice@example.com", "role": "Admin", "status": "Active", "joined": "2023-01-15"},
		{"id": 2, "name": "Bob Smith", "email": "bob@example.com", "role": "User", "status": "Active", "joined": "2023-02-20"},
		{"id": 3, "name": "Charlie Brown", "email": "charlie@example.com", "role": "User", "status": "Inactive", "joined": "2023-03-10"},
		{"id": 4, "name": "Diana Prince", "email": "diana@example.com", "role": "Manager", "status": "Active", "joined": "2023-01-05"},
		{"id": 5, "name": "Eve Davis", "email": "eve@example.com", "role": "User", "status": "Active", "joined": "2023-04-12"},
		{"id": 6, "name": "Frank Miller", "email": "frank@example.com", "role": "Admin", "status": "Active", "joined": "2023-02-28"},
		{"id": 7, "name": "Grace Lee", "email": "grace@example.com", "role": "User", "status": "Inactive", "joined": "2023-05-18"},
		{"id": 8, "name": "Henry Wilson", "email": "henry@example.com", "role": "Manager", "status": "Active", "joined": "2023-03-22"},
		{"id": 9, "name": "Ivy Chen", "email": "ivy@example.com", "role": "User", "status": "Active", "joined": "2023-06-01"},
		{"id": 10, "name": "Jack Ryan", "email": "jack@example.com", "role": "User", "status": "Active", "joined": "2023-01-30"},
		{"id": 11, "name": "Kate Morgan", "email": "kate@example.com", "role": "Admin", "status": "Active", "joined": "2023-02-14"},
		{"id": 12, "name": "Leo Parker", "email": "leo@example.com", "role": "User", "status": "Inactive", "joined": "2023-04-25"},
		{"id": 13, "name": "Mia Santos", "email": "mia@example.com", "role": "Manager", "status": "Active", "joined": "2023-03-08"},
		{"id": 14, "name": "Noah Kim", "email": "noah@example.com", "role": "User", "status": "Active", "joined": "2023-05-03"},
		{"id": 15, "name": "Olivia Green", "email": "olivia@example.com", "role": "User", "status": "Active", "joined": "2023-01-20"},
	}
}

func (h *DataTableExamplesHandler) getTableColumns() []data.Column {
	return []data.Column{
		{Key: "id", Label: "ID", Sortable: true, Width: "80px"},
		{Key: "name", Label: "Name", Sortable: true, Width: ""},
		{Key: "email", Label: "Email", Sortable: true, Width: ""},
		{Key: "role", Label: "Role", Sortable: true, Width: "120px"},
		{Key: "status", Label: "Status", Sortable: true, Width: "100px"},
		{Key: "joined", Label: "Joined", Sortable: true, Width: "120px"},
	}
}

func (h *DataTableExamplesHandler) filterData(data []map[string]interface{}, search string) []map[string]interface{} {
	if search == "" {
		return data
	}

	search = strings.ToLower(strings.TrimSpace(search))
	var filtered []map[string]interface{}

	for _, row := range data {
		// Search across all fields
		for _, value := range row {
			if strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), search) {
				filtered = append(filtered, row)
				break
			}
		}
	}

	return filtered
}

func (h *DataTableExamplesHandler) sortData(data []map[string]interface{}, sortBy, sortDir string) []map[string]interface{} {
	if sortBy == "" {
		return data
	}

	// Create a copy to avoid modifying the original
	sorted := make([]map[string]interface{}, len(data))
	copy(sorted, data)

	sort.Slice(sorted, func(i, j int) bool {
		a := sorted[i][sortBy]
		b := sorted[j][sortBy]

		// Handle nil values
		if a == nil {
			return sortDir != "desc"
		}
		if b == nil {
			return sortDir == "desc"
		}

		// Compare based on type
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)

		// Try to parse as numbers
		aNum, aErr := strconv.ParseFloat(aStr, 64)
		bNum, bErr := strconv.ParseFloat(bStr, 64)

		var result bool
		if aErr == nil && bErr == nil {
			result = aNum < bNum
		} else {
			result = aStr < bStr
		}

		if sortDir == "desc" {
			return !result
		}
		return result
	})

	return sorted
}

// Helper to get integer query parameter with default value
func getIntParam(c echo.Context, key string, defaultValue int) int {
	val := c.QueryParam(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}
