package response

// PaginationResponse represents pagination details
type Pagination struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
}

// RestResponsePagination represents a paginated response
type PaginationResponse[T any] struct {
	Pagination Pagination `json:"pagination"`
	Elements   []T        `json:"elements"`
}

func NewPaginationResponse[T any](pagination Pagination, Elements []T) PaginationResponse[T] {
	return PaginationResponse[T]{pagination, Elements}
}
