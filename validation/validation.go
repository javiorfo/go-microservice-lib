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
	tag       Tag
	jsonField JsonField
	validate  validator.Func
	code      string
	message   string
}

type Tag = string
type JsonField = string

type FuncCode = func(response.ErrorCode) customValidator
type FuncCodeAndMessage = func(response.ErrorCode, response.Message) customValidator

func ValidateRequest[T any](c *fiber.Ctx, span trace.Span, code response.ErrorCode, customValidators ...customValidator) (*T, *response.ResponseError) {
	entity := new(T)

	if err := c.BodyParser(entity); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, response.NewResponseError(span, response.Error{
			HttpStatus: fiber.StatusBadRequest,
			Code:       code,
			Message:    response.Message("Invalid Request Body"),
		})
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

	if err := validate.Struct(entity); err != nil {
		span.SetStatus(codes.Error, err.Error())
		validationErrors := err.(validator.ValidationErrors)

		restResponseError := &response.ResponseError{
			Errors: make([]response.Error, 0),
		}

	errorsLoop:
		for _, e := range validationErrors {
			msg := e.Translate(translator)

			for _, cv := range customValidators {
				if strings.Contains(msg, cv.message) {
					restResponseError.Add(span, response.Error{
						HttpStatus: fiber.StatusBadRequest,
						Code:       response.ErrorCode(cv.code),
						Message:    response.Message(cv.message),
					})
					continue errorsLoop
				}
			}

			restResponseError.Add(span, response.Error{
				HttpStatus: fiber.StatusBadRequest,
				Code:       response.ErrorCode(code),
				Message:    response.Message(msg),
			})
		}
		return nil, restResponseError
	}
	return entity, nil
}

func NewCustomValidator(tag Tag, jsonField JsonField, validate validator.Func) FuncCodeAndMessage {
	return func(errorCode response.ErrorCode, message response.Message) customValidator {
		return customValidator{
			tag:       tag,
			jsonField: jsonField,
			validate:  validate,
			code:      errorCode,
			message:   message,
		}
	}
}

func NewEnumValidator(tag Tag, jsonField JsonField, enums ...string) FuncCode {
	return func(errorCode response.ErrorCode) customValidator {
		return NewCustomValidator(tag, jsonField, func(fl validator.FieldLevel) bool {
			return slices.Contains(enums, fl.Field().String())
		})(errorCode, fmt.Sprintf("Field %s must be one of %s", jsonField, strings.Join(enums, ", ")))
	}
}

func NewNotBlankValidator(jsonField JsonField) FuncCode {
	return func(errorCode response.ErrorCode) customValidator {
		return NewCustomValidator("notblank", jsonField, func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})(errorCode, fmt.Sprintf("Field %s must be not be empty", jsonField))
	}
}
