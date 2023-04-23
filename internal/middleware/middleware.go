package middleware

import (
	"eventapi/internal/session"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func UseSession(s session.SessionManager) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sId := c.Cookies("session")
		if sId != "" {
			sessionData, err := s.Get(c, sId)
			if err == redis.Nil {
				s.Clear(c, sId)
				return c.Redirect("/api/login", 303)
			}

			if err != nil {
				fmt.Println("Could not get session info")
				return err
			}

			//not sure if this is going to work well, thinking i may just hold onto sessionId in context instead
			c.Locals("session", sessionData)
		}

		return c.Next()
	}
}
