package response

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/javiorfo/go-microservice-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

type ErrorCode = string
type Message = string

// Error represents an error
type Error struct {
	HttpStatus int       `json:"-"`
	Code       ErrorCode `json:"code"`
	Message    Message   `json:"message"`
}

// Stringer interface implementation
func (e Error) String() string {
	return fmt.Sprintf("ERROR CODE: %s. MESSAGE: %s", e.Code, e.Message)
}

// ResponseError represents an array of errors
type ResponseError struct {
	Errors []Error `json:"errors"`
}

// Get the first error if exists
func (re ResponseError) Get() Error {
	if len(re.Errors) == 0 {
		return Error{}
	}
	return re.Errors[0]
}

// Method for Fiber reponse (implementa inteface backend.Error)
func (re *ResponseError) ToResponse(c *fiber.Ctx) error {
	status := re.Get().HttpStatus
	return c.Status(status).JSON(re)
}

func (re ResponseError) Error() string {
	if len(re.Errors) == 0 {
		return "unknown error"
	}
	return re.Errors[0].Message
}

// Add adds an error to ResponseError and logs it
func (rre *ResponseError) Add(span trace.Span, e Error) *ResponseError {
	log.Error(tracing.LogError(span, e.String()))

	rre.Errors = append(rre.Errors, e)
	return rre
}

// NewResponseError creates an error to ResponseError and logs it
func NewResponseError(span trace.Span, e Error) *ResponseError {
	log.Error(tracing.LogError(span, e.String()))

	return &ResponseError{
		Errors: []Error{e},
	}
}

// InternalServerError creates a generic internal server error response
func InternalServerError(span trace.Span, msg Message) *ResponseError {
	return NewResponseError(span, Error{fiber.StatusInternalServerError, "INTERNAL_ERROR", msg})
}
