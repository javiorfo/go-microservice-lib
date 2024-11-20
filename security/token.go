package security

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/javiorfo/go-microservice-lib/response"
	"github.com/javiorfo/go-microservice-lib/response/codes"
	"github.com/javiorfo/go-microservice-lib/tracing"
	"github.com/javiorfo/steams"
)

type TokenConfig struct {
	SecretKey []byte
	Issuer    string
	Enabled   bool
}

type TokenClaims struct {
	Username    string              `json:"username"`
	Issuer      string              `json:"iss"`
	Permissions map[string][]string `json:"permissions"`
	jwt.RegisteredClaims
}

func (t TokenConfig) Secure(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Infof("%s Path captured: %s", tracing.LogTraceAndSpan(c), c.Path())
		if !t.Enabled {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.Contains(authHeader, "Bearer") {
			authorizationHeaderError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: "Authorization header or Bearer missing",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(authorizationHeaderError)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
			return t.SecretKey, nil
		})

		if err != nil || !token.Valid {
			invalidTokenError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: "Invalid token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		claims, ok := token.Claims.(*TokenClaims)
		if !ok {
			invalidTokenError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: "Invalid or expired token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		if len(roles) > 0 {
			if ok := hasRole(claims.Permissions, roles); !ok {
				invalidTokenError := response.NewRestResponseError(c, response.ResponseError{
					Code:    codes.AUTH_ERROR,
					Message: "User does not have permission to access",
				})
				return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
			}
		}

		c.Locals("tokenUser", claims.Username)
		return c.Next()
	}
}

func hasRole(permissions map[string][]string, roles []string) bool {
	return steams.OfMap(permissions).ValuesToSteam().AnyMatch(func(param []string) bool {
		return steams.OfSlice(roles).AnyMatch(func(r string) bool {
			return slices.Contains(param, r)
		})
	})
}
