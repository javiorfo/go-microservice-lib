package response

// PaginationResponse represents pagination details
type PaginationResponse struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
}

// RestResponsePagination represents a paginated response
type RestResponsePagination[T any] struct {
	Pagination PaginationResponse `json:"pagination"`
	Elements   []T                `json:"elements"`
}
