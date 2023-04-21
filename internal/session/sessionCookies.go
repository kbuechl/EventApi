package session

import (
	"eventapi/config"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateSessionCookie(c *fiber.Ctx, sId string, expires time.Time) {
	c.Cookie(buildCookie(config.New().Server.SessionCookieName,
		sId,
		expires,
	))
}

func ClearSessionCookie(c *fiber.Ctx) {
	c.Cookie(buildCookie(config.New().Server.SessionCookieName,
		"",
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	))
}

func buildCookie(name string, value string, expires time.Time) *fiber.Cookie {
	cookie := new(fiber.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.HTTPOnly = true
	cookie.Expires = expires

	return cookie
}
