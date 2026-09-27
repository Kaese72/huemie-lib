// Package query provides reusable filter, sort, and pagination building
// blocks for huma-based REST list endpoints. It owns the wire format and the
// parsing/validation mechanics; each service still owns which fields/columns
// are valid for its own data and how a validated filter/sort turns into a
// query against its own storage.
package query

// Pagination is an offset/limit query parameter block. Embed it in a huma
// input struct to add "offset"/"limit" query params without retyping the
// tags at every endpoint:
//
//	type input struct {
//		query.Pagination
//		Filters string `query:"filters"`
//	}
type Pagination struct {
	Offset int `query:"offset" default:"0" minimum:"0" doc:"number of matching results to skip"`
	Limit  int `query:"limit" default:"50" minimum:"1" maximum:"200" doc:"maximum number of results to return"`
}

// TotalCount is a response header block reporting how many rows matched a
// filtered query before pagination was applied. Embed it in a huma output
// struct alongside Pagination on the request side:
//
//	type output struct {
//		query.TotalCount
//		Body []Thing
//	}
type TotalCount struct {
	TotalCount int `header:"X-Total-Count" doc:"total number of matching results, ignoring pagination"`
}
