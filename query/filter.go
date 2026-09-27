package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// Filter is one "field <operator> value" condition, as decoded from a
// request's `filters` query parameter (a JSON array of these).
type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// ParseFilters decodes a `filters` query string - a JSON array of Filter - or
// returns an empty, non-nil slice if raw is empty. Malformed JSON is a
// huma.Error400BadRequest.
func ParseFilters(raw string) ([]Filter, error) {
	filters := []Filter{}
	if raw == "" {
		return filters, nil
	}
	if err := json.Unmarshal([]byte(raw), &filters); err != nil {
		return nil, huma.Error400BadRequest(fmt.Sprintf("filters must be a JSON array of {field, operator, value} objects: %s", err.Error()))
	}
	return filters, nil
}

// OperatorFunc turns a filter's value into a SQL WHERE fragment (using `?`
// placeholders) and the bind arguments for those placeholders.
type OperatorFunc func(value string) (fragment string, args []any, err error)

// FieldSpec declares the operators valid for one field, keyed by operator
// name (e.g. "eq", "text-contains").
type FieldSpec struct {
	Operators map[string]OperatorFunc
}

// Translate validates each filter's Field and Operator against fields -
// owned by the caller, since only the caller knows which columns/operators
// its data supports - and returns the AND-joined WHERE fragments and their
// bind arguments, in filter order. An unknown field or operator is a
// huma.Error400BadRequest naming the offending field/operator.
func Translate(filters []Filter, fields map[string]FieldSpec) ([]string, []any, error) {
	var fragments []string
	var args []any
	for _, filter := range filters {
		spec, ok := fields[filter.Field]
		if !ok {
			return nil, nil, huma.Error400BadRequest(fmt.Sprintf("may not filter on field %q", filter.Field))
		}
		op, ok := spec.Operators[filter.Operator]
		if !ok {
			return nil, nil, huma.Error400BadRequest(fmt.Sprintf("may not filter on field %q with operator %q", filter.Field, filter.Operator))
		}
		fragment, values, err := op(filter.Value)
		if err != nil {
			return nil, nil, err
		}
		fragments = append(fragments, fragment)
		args = append(args, values...)
	}
	return fragments, args, nil
}

// Merge combines several single/partial operator maps (e.g. from EqualsOperator,
// InOperator, TextOperators, ...) into one FieldSpec.
func Merge(operatorMaps ...map[string]OperatorFunc) FieldSpec {
	spec := FieldSpec{Operators: map[string]OperatorFunc{}}
	for _, operators := range operatorMaps {
		for name, fn := range operators {
			spec.Operators[name] = fn
		}
	}
	return spec
}

// LikeEscaper escapes the LIKE wildcard characters '%' and '_' (and the
// escape character itself) so a user-supplied substring is matched
// literally rather than as a pattern. Combine with an ESCAPE '\\' clause.
var LikeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// NumericApproxEpsilon is the default tolerance used by NumericOperators'
// "numeric-aeq" (approximately-equal) comparison, since values are often
// floating point and prone to rounding error.
const NumericApproxEpsilon = 0.001

// EqualsOperator returns an "eq" OperatorFunc doing a plain equality
// comparison against column.
func EqualsOperator(column string) map[string]OperatorFunc {
	return map[string]OperatorFunc{
		"eq": func(value string) (string, []any, error) {
			return column + " = ?", []any{value}, nil
		},
	}
}

// InOperator returns an "in" OperatorFunc matching column against a
// comma-separated list of values, e.g. "1,2,3". Empty entries are rejected.
func InOperator(column string) map[string]OperatorFunc {
	return map[string]OperatorFunc{
		"in": func(value string) (string, []any, error) {
			rawValues := strings.Split(value, ",")
			args := make([]any, 0, len(rawValues))
			placeholders := make([]string, 0, len(rawValues))
			for _, rawValue := range rawValues {
				trimmed := strings.TrimSpace(rawValue)
				if trimmed == "" {
					return "", nil, huma.Error400BadRequest(fmt.Sprintf("%q is not a valid comma-separated list of values", value))
				}
				args = append(args, trimmed)
				placeholders = append(placeholders, "?")
			}
			return column + " IN (" + strings.Join(placeholders, ",") + ")", args, nil
		},
	}
}

// TextOperators returns "text-eq" (exact match) and "text-contains"
// (case-insensitive substring match) OperatorFuncs for column.
func TextOperators(column string) map[string]OperatorFunc {
	return map[string]OperatorFunc{
		"text-eq": func(value string) (string, []any, error) {
			return column + " = ?", []any{value}, nil
		},
		"text-contains": func(value string) (string, []any, error) {
			pattern := "%" + LikeEscaper.Replace(value) + "%"
			return "LOWER(" + column + ") LIKE LOWER(?) ESCAPE '\\\\'", []any{pattern}, nil
		},
	}
}

// NumericOperators returns "numeric-eq", "numeric-aeq" (approximately equal,
// within NumericApproxEpsilon), "numeric-lt", and "numeric-gt" OperatorFuncs
// for column. Each rejects a non-numeric value with a 400 error.
func NumericOperators(column string) map[string]OperatorFunc {
	parse := func(value string) (float64, error) {
		asFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, huma.Error400BadRequest(fmt.Sprintf("filter value %q must be numeric", value))
		}
		return asFloat, nil
	}
	return map[string]OperatorFunc{
		"numeric-eq": func(value string) (string, []any, error) {
			v, err := parse(value)
			if err != nil {
				return "", nil, err
			}
			return column + " = ?", []any{v}, nil
		},
		"numeric-aeq": func(value string) (string, []any, error) {
			v, err := parse(value)
			if err != nil {
				return "", nil, err
			}
			return "ABS(" + column + " - ?) < ?", []any{v, NumericApproxEpsilon}, nil
		},
		"numeric-lt": func(value string) (string, []any, error) {
			v, err := parse(value)
			if err != nil {
				return "", nil, err
			}
			return column + " < ?", []any{v}, nil
		},
		"numeric-gt": func(value string) (string, []any, error) {
			v, err := parse(value)
			if err != nil {
				return "", nil, err
			}
			return column + " > ?", []any{v}, nil
		},
	}
}

// BoolOperator returns a "bool-eq" OperatorFunc for column, accepting only
// the literal values "true"/"false".
func BoolOperator(column string) map[string]OperatorFunc {
	return map[string]OperatorFunc{
		"bool-eq": func(value string) (string, []any, error) {
			switch value {
			case "true":
				return column + " = ?", []any{true}, nil
			case "false":
				return column + " = ?", []any{false}, nil
			default:
				return "", nil, huma.Error400BadRequest(fmt.Sprintf("filter value %q must be \"true\" or \"false\"", value))
			}
		},
	}
}
