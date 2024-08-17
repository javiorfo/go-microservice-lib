package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/javiorfo/go-microservice-lib/tracing"
	"github.com/javiorfo/go-microservice-lib/response/codes"
)

// ResponseError represents an error
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

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

// restResponseError represents an array of errors
type restResponseError struct {
	Errors []ResponseError `json:"errors"`
}

// AddError adds an error to restResponseError and logs it
func (rre *restResponseError) AddError(c *fiber.Ctx, re ResponseError) *restResponseError {
	log.Errorf("%s Code: %s Message: %s", tracing.LogTraceAndSpan(c), re.Code, re.Message)
	rre.Errors = append(rre.Errors, re)
	return rre
}

// AddErrorWithCodeAndMsg adds an error to restResponseError
func (rre *restResponseError) AddErrorWithCodeAndMsg(c *fiber.Ctx, code, msg string) *restResponseError {
	return rre.AddError(c, ResponseError{code, msg})
}

// NewRestResponseError creates an error to restResponseError and logs it
func NewRestResponseError(c *fiber.Ctx, re ResponseError) *restResponseError {
	log.Errorf("%s Code: %s Message: %s", tracing.LogTraceAndSpan(c), re.Code, re.Message)
	return &restResponseError{
		Errors: []ResponseError{re},
	}
}

// NewRestResponseErrorWithCodeAndMsg creates a new restResponseError
func NewRestResponseErrorWithCodeAndMsg(c *fiber.Ctx, code, msg string) *restResponseError {
	return NewRestResponseError(c, ResponseError{code, msg})
}

// InternalServerError creates a generic internal server error response
func InternalServerError(c *fiber.Ctx, msg string) *restResponseError {
	return NewRestResponseError(c, ResponseError{codes.INTERNAL_ERROR, msg})
}
