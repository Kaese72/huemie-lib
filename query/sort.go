package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// Sort is one "order by field, direction" entry, as decoded from a request's
// `sort` query parameter (a JSON array of these).
type Sort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" (default, if empty) or "desc"
}

// ParseSort decodes a `sort` query string - a JSON array of Sort - or
// returns nil, nil if raw is empty. Malformed JSON, or a Direction other
// than "", "asc", or "desc", is a huma.Error400BadRequest.
func ParseSort(raw string) ([]Sort, error) {
	if raw == "" {
		return nil, nil
	}
	var sorts []Sort
	if err := json.Unmarshal([]byte(raw), &sorts); err != nil {
		return nil, huma.Error400BadRequest(fmt.Sprintf("sort must be a JSON array of {field, direction} objects: %s", err.Error()))
	}
	for i, sort := range sorts {
		switch sort.Direction {
		case "":
			sorts[i].Direction = "asc"
		case "asc", "desc":
			// already valid
		default:
			return nil, huma.Error400BadRequest(fmt.Sprintf("sort direction %q for field %q must be \"asc\" or \"desc\"", sort.Direction, sort.Field))
		}
	}
	return sorts, nil
}

// BuildOrderBy validates each Sort's Field against allowed - a service-owned
// map from the wire field name to the column/expression to order by - and
// returns a "column ASC|DESC[, ...]" clause (without the "ORDER BY" prefix).
// If sorts is empty, fallback is returned unchanged, so callers can preserve
// their existing default ordering when no `sort` param is given. An unknown
// field is a huma.Error400BadRequest.
func BuildOrderBy(sorts []Sort, allowed map[string]string, fallback string) (string, error) {
	if len(sorts) == 0 {
		return fallback, nil
	}
	clauses := make([]string, 0, len(sorts))
	for _, sort := range sorts {
		column, ok := allowed[sort.Field]
		if !ok {
			return "", huma.Error400BadRequest(fmt.Sprintf("may not sort on field %q", sort.Field))
		}
		direction := "ASC"
		if sort.Direction == "desc" {
			direction = "DESC"
		}
		clauses = append(clauses, column+" "+direction)
	}
	return strings.Join(clauses, ", "), nil
}
