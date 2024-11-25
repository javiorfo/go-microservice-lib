package security

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/javiorfo/go-microservice-lib/response"
	"github.com/javiorfo/go-microservice-lib/response/codes"
	"github.com/javiorfo/steams"
)

type TokenConfig struct {
	SecretKey []byte
	Issuer    string
	Enabled   bool
}

type TokenClaims struct {
	Permissions map[string][]string `json:"permissions"`
	Audience    string              `json:"aud"`
	jwt.RegisteredClaims
}

// Secure method with role validation. If no role is specified
// no role validation is executed
func (t TokenConfig) Secure(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
				Message: "Invalid or expired token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		claims, ok := token.Claims.(*TokenClaims)
		if !ok { 
			invalidTokenError := response.NewRestResponseError(c, response.ResponseError{
				Code:    codes.AUTH_ERROR,
				Message: "Invalid token",
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

		c.Locals("tokenUser", claims.Subject)
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
