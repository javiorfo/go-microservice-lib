package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/javiorfo/go-microservice-lib/response"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func ValidateRequest[T any](c *fiber.Ctx, span trace.Span, code string) (*T, *response.RestResponseError) {
	entity := new(T)

	if err := c.BodyParser(entity); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, response.NewRestResponseErrorWithCodeAndMsg(span, code, "Invalid Request Body")
	}

	validate := validator.New()
	_ = validate.RegisterValidation("notblank", notBlank)

	if err := validate.Struct(entity); err != nil {
		span.SetStatus(codes.Error, err.Error())
		validationErrors := err.(validator.ValidationErrors)

		var translatedErrors []string
		for _, e := range validationErrors {
			if e.Tag() == "notblank" {
				translatedErrors = append(translatedErrors, e.Field()+" should not be empty")
			} else {
				translatedErrors = append(translatedErrors, e.Error())
			}
		}
		return nil, response.NewRestResponseErrorWithCodeAndMsg(span, code, strings.Join(translatedErrors, ", "))
	}
	return entity, nil
}

func notBlank(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	return strings.TrimSpace(val) != ""
}
