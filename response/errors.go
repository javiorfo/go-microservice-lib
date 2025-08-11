package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/javiorfo/go-microservice-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

type ErrorCode string
type Message string

// Error represents an error
type Error struct {
	HttpStatus int       `json:"-"`
	Code       ErrorCode `json:"code"`
	Message    Message   `json:"message"`
}

// ResponseError represents an array of errors
type ResponseError struct {
	Errors []Error `json:"errors"`
}

// Método para responder el error con Fiber (implementa inteface backend.Error)
func (re *ResponseError) ToResponse(c *fiber.Ctx) error {
	status := re.Errors[0].HttpStatus
	return c.Status(status).JSON(re)
}

// Add adds an error to ResponseError and logs it
func (rre *ResponseError) Add(span trace.Span, e Error) *ResponseError {
	log.Errorf("%s Code: %s Message: %s", tracing.LogTraceAndSpan(span), e.Code, e.Message)
	rre.Errors = append(rre.Errors, e)
	return rre
}

// NewResponseError creates an error to ResponseError and logs it
func NewResponseError(span trace.Span, e Error) *ResponseError {
	log.Errorf("%s Code: %s Message: %s", tracing.LogTraceAndSpan(span), e.Code, e.Message)
	return &ResponseError{
		Errors: []Error{e},
	}
}

// InternalServerError creates a generic internal server error response
func InternalServerError(span trace.Span, msg Message) *ResponseError {
	return NewResponseError(span, Error{fiber.StatusInternalServerError, "INTERNAL_ERROR", msg})
}
