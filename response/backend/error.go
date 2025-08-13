package backend

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/javiorfo/go-microservice-lib/response"
	"go.opentelemetry.io/otel/trace"
)

// Interface para responder con Fiber
type Error interface {
	ToResponse(c *fiber.Ctx) error
}

func InternalError(span trace.Span, err error) Error {
	return InternalMsgError(span, err.Error())
}

func InternalMsgError(span trace.Span, msg string) Error {
	return response.NewResponseError(span,
		response.Error{
			HttpStatus: http.StatusInternalServerError,
			Code:       response.ErrorCode("INTERNAL-ERROR"),
			Message:    response.Message(msg),
		},
	)
}
