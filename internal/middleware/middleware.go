package middleware

import (
	"eventapi/internal/session"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
)

func UseSession(c *fiber.Ctx) error {
	ss := session.NewSessionService()
	sId := c.Cookies("session")
	if sId != "" {
		sessionData, err := ss.GetSession(c, sId)

		if err != nil {
			if err == redis.Nil {
				ss.ClearSession(c, sId)
				return c.Redirect("/api/login", 303)
			} else {
				fmt.Println("Could not get session info")
				return err
			}

		}

		//not sure if this is going to work well, thinking i may just hold onto sessionId in context instead
		c.Locals("session", sessionData)
	}

	return c.Next()
}
