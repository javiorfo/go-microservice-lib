package validation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gofiber/fiber/v2"
	"github.com/javiorfo/go-microservice-lib/response"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type customValidator struct {
	tag       string
	jsonField string
	validate  validator.Func
	code      string
	message   string
}

type FuncCode = func(string) customValidator

type FuncCodeAndMsg = func(string, string) customValidator

func ValidateRequest[T any](c *fiber.Ctx, span trace.Span, code string, customValidators ...customValidator) (*T, *response.RestResponseError) {
	entity := new(T)

	if err := c.BodyParser(entity); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, response.NewRestResponseErrorWithCodeAndMsg(span, code, "Invalid Request Body")
	}

	validate := validator.New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	translator, _ := uni.GetTranslator("en")

	en_translations.RegisterDefaultTranslations(validate, translator)

	for _, cv := range customValidators {
		validate.RegisterValidation(cv.tag, cv.validate)
		validate.RegisterTranslation(cv.tag, translator, func(ut ut.Translator) error {
			return ut.Add(cv.tag, "{0} "+cv.message, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(cv.tag, cv.jsonField)
			return t
		})
	}

	_ = validate.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return strings.TrimSpace(val) != ""
	})

	if err := validate.Struct(entity); err != nil {
		span.SetStatus(codes.Error, err.Error())
		validationErrors := err.(validator.ValidationErrors)

		restResponseError := &response.RestResponseError{
			Errors: make([]response.ResponseError, 0),
		}

	errorsLoop:
		for _, e := range validationErrors {
			msg := e.Translate(translator)

			for _, cv := range customValidators {
				if strings.Contains(msg, cv.message) {
					restResponseError.AddError(span, response.ResponseError{Code: cv.code, Message: cv.message})
					continue errorsLoop
				}
			}

			if e.Tag() == "notblank" {
				msg = e.Field() + " should not be empty"
			}
			restResponseError.AddError(span, response.ResponseError{Code: code, Message: msg})
		}
		return nil, restResponseError
	}
	return entity, nil
}

func NewCustomValidator(tag, jsonField string, validate validator.Func) FuncCodeAndMsg {
	return func(code, message string) customValidator {
		return customValidator{
			tag:       tag,
			jsonField: jsonField,
			validate:  validate,
			code:      code,
			message:   message,
		}
	}
}

func NewEnumValidator(tag, jsonField string, enums ...string) FuncCode {
	return func(code string) customValidator {
		return NewCustomValidator(tag, jsonField, func(fl validator.FieldLevel) bool {
			return slices.Contains(enums, fl.Field().String())
		})(code, fmt.Sprintf("Field %s must be one of %s", jsonField, strings.Join(enums, ", ")))
	}
}
