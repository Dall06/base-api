package sorting

import "fmt"

// BuildOrder constructs a safe ORDER BY clause from sortBy and sortOrder.
// Only columns in the allowedColumns whitelist are accepted.
// Falls back to defaultOrder if sortBy is empty or not allowed.
func BuildOrder(sortBy, sortOrder, defaultOrder string, allowedColumns map[string]string) string {
	if sortBy == "" {
		return defaultOrder
	}

	column, ok := allowedColumns[sortBy]
	if !ok {
		return defaultOrder
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	return fmt.Sprintf("%s %s", column, sortOrder)
}
