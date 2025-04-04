package response

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/javiorfo/go-microservice-lib/response/codes"
	"github.com/javiorfo/go-microservice-lib/tracing"
	"go.opentelemetry.io/otel/trace"
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
func (rre *restResponseError) AddError(span trace.Span, re ResponseError) *restResponseError {
	log.Errorf("%s Code: %s Message: %s", tracing.LogTraceAndSpan(span), re.Code, re.Message)
	rre.Errors = append(rre.Errors, re)
	return rre
}

// AddErrorWithCodeAndMsg adds an error to restResponseError
func (rre *restResponseError) AddErrorWithCodeAndMsg(span trace.Span, code, msg string) *restResponseError {
	return rre.AddError(span, ResponseError{code, msg})
}

// NewRestResponseError creates an error to restResponseError and logs it
func NewRestResponseError(span trace.Span, re ResponseError) *restResponseError {
	log.Errorf("%s Code: %s Message: %s", tracing.LogTraceAndSpan(span), re.Code, re.Message)
	return &restResponseError{
		Errors: []ResponseError{re},
	}
}

// NewRestResponseErrorWithCodeAndMsg creates a new restResponseError
func NewRestResponseErrorWithCodeAndMsg(span trace.Span, code, msg string) *restResponseError {
	return NewRestResponseError(span, ResponseError{code, msg})
}

// InternalServerError creates a generic internal server error response
func InternalServerError(span trace.Span, msg string) *restResponseError {
	return NewRestResponseError(span, ResponseError{codes.INTERNAL_ERROR, msg})
}
