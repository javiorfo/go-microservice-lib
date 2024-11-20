package security

import (
	"github.com/gofiber/fiber/v2"
)

type Securizer interface {
	Secure(roles ...string) fiber.Handler
}

func GetTokenUsername(c *fiber.Ctx) string {
	if tokenUser := c.Locals("tokenUser"); tokenUser != nil {
		return tokenUser.(string)
	}
	return "unknown"
}
