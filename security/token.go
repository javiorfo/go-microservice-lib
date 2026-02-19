package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/javiorfo/go-microservice-lib/response"
	"github.com/javiorfo/go-microservice-lib/tracing"
	"github.com/javiorfo/steams/v2"
	"go.opentelemetry.io/otel"
)

var tokenTracer = otel.Tracer("TokenSecurity")

type TokenSecurity struct {
	Enabled bool
}

func NewTokenSecurity() TokenSecurity {
	var isEnabled = true
	securityEnabled := os.Getenv("SECURITY_ENABLED")

	if securityEnabled != "" {
		isEnabled = strings.ToLower(securityEnabled) == "true"
	}
	return TokenSecurity{isEnabled}
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
func (t TokenSecurity) Secure(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, span := tokenTracer.Start(c.UserContext(), "JWT Security")
		defer span.End()

		if !t.Enabled {
			log.Warn(tracing.LogInfo(span, "security disabled!"))
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.Contains(authHeader, "Bearer") {
			authorizationHeaderError := response.NewResponseError(span, response.Error{
				Code:    "AUTH_ERROR",
				Message: "Authorization header or Bearer missing",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(authorizationHeaderError)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil || !token.Valid {
			invalidTokenError := response.NewResponseError(span, response.Error{
				Code:    "AUTH_ERROR",
				Message: "Invalid or expired token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		claims, ok := token.Claims.(*TokenClaims)
		if !ok {
			invalidTokenError := response.NewResponseError(span, response.Error{
				Code:    "AUTH_ERROR",
				Message: "Invalid token",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(invalidTokenError)
		}

		if len(roles) > 0 {
			if ok := hasRole(claims.Permission, roles); !ok {
				invalidTokenError := response.NewResponseError(span, response.Error{
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
	return steams.FromSlice(roles).Any(func(r string) bool {
		return slices.Contains(permission.Roles, r)
	})
}

func CreateToken(permission TokenPermission, username string) (string, error) {
	duration, err := strconv.Atoi(os.Getenv("JWT_DURATION"))
	if err != nil {
		return "", err
	}
	return CreateTokenWithDuration(permission, username, time.Duration(duration*int(time.Second)))
}

func CreateTokenWithDuration(permission TokenPermission, username string, duration time.Duration) (string, error) {
	claims := TokenClaims{
		Permission: permission,
		Audience:   os.Getenv("JWT_AUDIENCE"),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    os.Getenv("JWT_ISSUER"),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Subject:   username,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

func RefreshToken(oldToken string) (string, error) {
	token, _ := jwt.ParseWithClaims(oldToken, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return "", errors.New("Invalid token")
	}
	return CreateToken(claims.Permission, claims.Subject)
}

const (
	passwordLength = 8
	charset        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
)

func GenerateRandomPassword() (string, error) {
	password := make([]byte, passwordLength)

	_, err := rand.Read(password)
	if err != nil {
		return "", err
	}

	for i := range passwordLength {
		password[i] = charset[int(password[i])%len(charset)]
	}

	return string(password), nil
}

func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func Hash(password, salt string) string {
	saltedPassword := password + salt
	hash := sha256.Sum256([]byte(saltedPassword))
	return base64.StdEncoding.EncodeToString(hash[:])
}
