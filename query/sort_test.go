package query

import "testing"

func TestParseSort(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		sorts, err := ParseSort("")
		if err != nil || sorts != nil {
			t.Fatalf("expected nil, nil, got %v, %v", sorts, err)
		}
	})

	t.Run("defaults to asc", func(t *testing.T) {
		sorts, err := ParseSort(`[{"field":"name"}]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sorts) != 1 || sorts[0].Direction != "asc" {
			t.Fatalf("unexpected result: %+v", sorts)
		}
	})

	t.Run("valid desc", func(t *testing.T) {
		sorts, err := ParseSort(`[{"field":"name","direction":"desc"}]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sorts) != 1 || sorts[0].Direction != "desc" {
			t.Fatalf("unexpected result: %+v", sorts)
		}
	})

	t.Run("invalid direction", func(t *testing.T) {
		if _, err := ParseSort(`[{"field":"name","direction":"sideways"}]`); err == nil {
			t.Fatal("expected an error for an invalid direction")
		}
	})

	t.Run("malformed", func(t *testing.T) {
		if _, err := ParseSort("not json"); err == nil {
			t.Fatal("expected an error for malformed JSON")
		}
	})
}

func TestBuildOrderBy(t *testing.T) {
	allowed := map[string]string{"name": "name", "id": "id"}

	t.Run("empty falls back", func(t *testing.T) {
		clause, err := BuildOrderBy(nil, allowed, "id ASC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if clause != "id ASC" {
			t.Fatalf("expected fallback, got %q", clause)
		}
	})

	t.Run("valid", func(t *testing.T) {
		clause, err := BuildOrderBy([]Sort{{Field: "name", Direction: "asc"}, {Field: "id", Direction: "desc"}}, allowed, "id ASC")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if clause != "name ASC, id DESC" {
			t.Fatalf("unexpected clause: %q", clause)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		if _, err := BuildOrderBy([]Sort{{Field: "nope", Direction: "asc"}}, allowed, "id ASC"); err == nil {
			t.Fatal("expected an error for an unknown field")
		}
	})
}
