package backend

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/javiorfo/go-microservice-lib/response"
	"go.opentelemetry.io/otel/trace"
)

// Interface for Fiber response
type Error interface {
	error
	ToResponse(c *fiber.Ctx) error
}

func InternalError(span trace.Span, err error) error {
	return InternalMsgError(span, err.Error())
}

func InternalMsgError(span trace.Span, msg string) error {
	return response.NewResponseError(span,
		response.Error{
			HttpStatus: http.StatusInternalServerError,
			Code:       response.ErrorCode("INTERNAL-ERROR"),
			Message:    response.Message(msg),
		},
	)
}

func ParseError(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(Error); ok {
		return fiberErr.ToResponse(c)
	}

	return c.Status(http.StatusInternalServerError).JSON(response.Error{
		HttpStatus: http.StatusInternalServerError,
		Code:       response.ErrorCode("INTERNAL-ERROR"),
		Message:    response.Message(err.Error()),
	})
}
