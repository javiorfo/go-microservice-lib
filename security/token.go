package security

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/javiorfo/go-microservice-lib/response"
	"github.com/javiorfo/steams"
	"go.opentelemetry.io/otel"
)

var jwtTracer = otel.Tracer("JWT")

type TokenConfig struct {
	SecretKey []byte
	Issuer    string
	Enabled   bool
}

type TokenClaims struct {
	Permission TokenPermission `json:"permission"`
	Audience   string          `json:"aud"`
	jwt.RegisteredClaims
}

type TokenPermission struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// Secure method with role validation. If no role is specified
// no role validation is executed
func (t TokenConfig) Secure(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := keycloakTracer.Start(c.UserContext(), "secure")
		defer span.End()

		if !t.Enabled {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.Contains(authHeader, "Bearer") {
			authorizationHeaderError := response.NewRestResponseError(span, response.ResponseError{
				Code:    "AUTH_ERROR",
				Message: "Authorization header or Bearer missing",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(authorizationHeaderError)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
			return t.SecretKey, nil
		})

		if err != nil || !token.Valid {
			invalidTokenError := response.NewRestResponseError(span, response.ResponseError{
				Code:    "AUTH_ERROR",
				Message: "Invalid or expired token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		claims, ok := token.Claims.(*TokenClaims)
		if !ok {
			invalidTokenError := response.NewRestResponseError(span, response.ResponseError{
				Code:    "AUTH_ERROR",
				Message: "Invalid token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		if len(roles) > 0 {
			if ok := hasRole(claims.Permission, roles); !ok {
				invalidTokenError := response.NewRestResponseError(span, response.ResponseError{
					Code:    "AUTH_ERROR",
					Message: "User does not have permission to access",
				})
				return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
			}
		}

		c.Locals("tokenUser", claims.Subject)
		return c.Next()
	}
}

func hasRole(permission TokenPermission, roles []string) bool {
	return steams.OfSlice(roles).AnyMatch(func(r string) bool {
		return slices.Contains(permission.Roles, r)
	})
}
