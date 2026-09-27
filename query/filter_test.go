package query

import "testing"

func TestParseFilters(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		filters, err := ParseFilters("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filters) != 0 {
			t.Fatalf("expected no filters, got %v", filters)
		}
	})

	t.Run("valid", func(t *testing.T) {
		filters, err := ParseFilters(`[{"field":"name","operator":"text-contains","value":"kitchen"}]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filters) != 1 || filters[0] != (Filter{Field: "name", Operator: "text-contains", Value: "kitchen"}) {
			t.Fatalf("unexpected filters: %+v", filters)
		}
	})

	t.Run("malformed", func(t *testing.T) {
		if _, err := ParseFilters("not json"); err == nil {
			t.Fatal("expected an error for malformed JSON")
		}
	})
}

func TestTranslate(t *testing.T) {
	fields := map[string]FieldSpec{
		"id":   Merge(EqualsOperator("id"), InOperator("id")),
		"name": Merge(TextOperators("name")),
	}

	t.Run("valid", func(t *testing.T) {
		fragments, args, err := Translate([]Filter{
			{Field: "id", Operator: "eq", Value: "5"},
			{Field: "name", Operator: "text-contains", Value: "kit"},
		}, fields)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(fragments) != 2 || len(args) != 2 {
			t.Fatalf("expected 2 fragments/args, got fragments=%v args=%v", fragments, args)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		if _, _, err := Translate([]Filter{{Field: "nope", Operator: "eq", Value: "x"}}, fields); err == nil {
			t.Fatal("expected an error for an unknown field")
		}
	})

	t.Run("unknown operator", func(t *testing.T) {
		if _, _, err := Translate([]Filter{{Field: "id", Operator: "contains", Value: "x"}}, fields); err == nil {
			t.Fatal("expected an error for an unknown operator")
		}
	})
}

func TestInOperator(t *testing.T) {
	ops := InOperator("id")
	t.Run("valid", func(t *testing.T) {
		fragment, args, err := ops["in"]("1, 2,3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fragment != "id IN (?,?,?)" || len(args) != 3 {
			t.Fatalf("unexpected result: fragment=%q args=%v", fragment, args)
		}
	})
	t.Run("empty entry", func(t *testing.T) {
		if _, _, err := ops["in"]("1,,3"); err == nil {
			t.Fatal("expected an error for an empty entry")
		}
	})
}

func TestNumericOperators(t *testing.T) {
	ops := NumericOperators("value")
	t.Run("valid", func(t *testing.T) {
		fragment, args, err := ops["numeric-lt"]("3.5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fragment != "value < ?" || len(args) != 1 {
			t.Fatalf("unexpected result: fragment=%q args=%v", fragment, args)
		}
	})
	t.Run("non-numeric", func(t *testing.T) {
		if _, _, err := ops["numeric-eq"]("nope"); err == nil {
			t.Fatal("expected an error for a non-numeric value")
		}
	})
}

func TestDateOperators(t *testing.T) {
	ops := DateOperators("created")
	t.Run("valid", func(t *testing.T) {
		fragment, args, err := ops["date-gt"]("2026-09-27T00:00:00Z")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fragment != "created > ?" || len(args) != 1 {
			t.Fatalf("unexpected result: fragment=%q args=%v", fragment, args)
		}
	})
	t.Run("malformed", func(t *testing.T) {
		if _, _, err := ops["date-eq"]("not a date"); err == nil {
			t.Fatal("expected an error for a non-RFC3339 value")
		}
	})
}

func TestBoolOperator(t *testing.T) {
	ops := BoolOperator("flag")
	if _, _, err := ops["bool-eq"]("true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := ops["bool-eq"]("maybe"); err == nil {
		t.Fatal("expected an error for a non-bool value")
	}
}
