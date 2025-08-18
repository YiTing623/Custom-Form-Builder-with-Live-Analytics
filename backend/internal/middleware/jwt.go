package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthOptional(secret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return c.Next()
		}
		claims, err := parseJWT(token, secret)
		if err == nil && claims != nil && claims.Subject != "" {
			c.Locals("userId", claims.Subject)
		}
		return c.Next()
	}
}

func AuthRequired(secret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}
		claims, err := parseJWT(token, secret)
		if err != nil || claims.Subject == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		c.Locals("userId", claims.Subject)
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	h := c.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if v := c.Cookies("token"); v != "" {
		return v
	}
	return ""
}

func parseJWT(token string, secret []byte) (*jwt.RegisteredClaims, error) {
	tok, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(*jwt.RegisteredClaims); ok && tok.Valid {
		return claims, nil
	}
	return nil, fiber.ErrUnauthorized
}
