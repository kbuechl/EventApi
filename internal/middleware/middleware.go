package middleware

import (
	"eventapi/internal/session"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// UseSession loads session data into fiber Locals if a valid session cookie is present.
// Expired/missing sessions are cleared but do NOT redirect — use RequireAuth for that.
func UseSession(s session.SessionManager) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sId := c.Cookies("session")
		if sId == "" {
			return c.Next()
		}

		sessionData, err := s.Get(c, sId)
		if err == redis.Nil {
			// Session expired — clear the stale cookie and continue unauthenticated
			s.Clear(c, sId)
			return c.Next()
		}
		if err != nil {
			return fmt.Errorf("error reading session: %w", err)
		}

		c.Locals("session", sessionData)
		return c.Next()
	}
}

// RequireAuth returns 401 JSON if no session is loaded. Apply after UseSession.
func RequireAuth() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if GetSession(c) == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "authentication required",
			})
		}
		return c.Next()
	}
}

// GetSession returns the session data from fiber Locals, or nil if not authenticated.
func GetSession(c *fiber.Ctx) *session.SessionData {
	sd, _ := c.Locals("session").(*session.SessionData)
	return sd
}
