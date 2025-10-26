package response

import "strconv"

// PaginationResponse represents pagination details
type Pagination struct {
	PageNumber int   `json:"pageNumber"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
}

// RestResponsePagination represents a paginated response
type PaginationResponse[T any] struct {
	Pagination Pagination `json:"pagination"`
	Elements   []T        `json:"elements"`
}

func NewPaginationResponse[T any](pageNumber, pageSize string, total int64, Elements []T) PaginationResponse[T] {
	pn, _ := strconv.Atoi(pageNumber)
	ps, _ := strconv.Atoi(pageSize)
	p := Pagination{
		PageNumber: pn,
		PageSize:   ps,
		Total:      total,
	}
	return PaginationResponse[T]{p, Elements}
}
