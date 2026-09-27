package query

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

// TestPaginationEmbedding is a smoke test proving huma promotes tags from
// Pagination/TotalCount when they're embedded (anonymously) into a service's
// own input/output structs, rather than requiring every service to retype
// the offset/limit/X-Total-Count fields by hand.
func TestPaginationEmbedding(t *testing.T) {
	_, api := humatest.New(t)

	type input struct {
		Pagination
	}
	type output struct {
		TotalCount
		Body []string
	}

	var gotOffset, gotLimit int
	huma.Register(api, huma.Operation{
		OperationID: "list-things",
		Method:      "GET",
		Path:        "/things",
	}, func(ctx context.Context, in *input) (*output, error) {
		gotOffset, gotLimit = in.Offset, in.Limit
		return &output{TotalCount: TotalCount{TotalCount: 123}, Body: []string{"a"}}, nil
	})

	resp := api.Get("/things")
	if resp.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if gotOffset != 0 || gotLimit != 50 {
		t.Errorf("expected defaults Offset=0 Limit=50 promoted from embedded Pagination, got Offset=%d Limit=%d", gotOffset, gotLimit)
	}
	if got := resp.Header().Get("X-Total-Count"); got != "123" {
		t.Fatalf("expected X-Total-Count promoted from embedded TotalCount to be \"123\", got %q", got)
	}

	resp = api.Get("/things?offset=10&limit=20")
	if resp.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if gotOffset != 10 || gotLimit != 20 {
		t.Errorf("expected Offset=10 Limit=20 promoted from embedded Pagination, got Offset=%d Limit=%d", gotOffset, gotLimit)
	}
}
