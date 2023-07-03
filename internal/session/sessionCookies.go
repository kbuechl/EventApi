package session

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func createSessionCookie(c *fiber.Ctx, name string, sId string, expires time.Time) {
	c.Cookie(buildCookie(name,
		sId,
		expires,
	))
}

func clearSessionCookie(c *fiber.Ctx, name string) {
	c.Cookie(buildCookie(name,
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
