package auth

import (
	"eventapi/internal/database"
	"eventapi/internal/session"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

var us *database.UserService
var ss *session.SessionService

func init() {
	us = database.NewUserService()
	ss = session.NewSessionService()
}

// Login handler
func Login(c *fiber.Ctx) error {
	url := getEntryPoint()
	return c.Redirect(url)
}

// oauth callback
func Callback(c *fiber.Ctx) error {
	token, err := exchange(c.Context(), c.FormValue("state"), c.FormValue("code"))
	if err != nil {
		return err
	}

	u, err := getUserInfo(token.AccessToken)

	if err != nil {
		return fmt.Errorf("could not fetch user info: %w", err)
	}

	if exists := us.Exists(u.ID); !exists {
		println("User does not exist, creating")

		_, err := us.Create(database.NewUser{
			ID:           u.ID,
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			Email:        u.Email,
			Picture:      u.Picture,
			Verified:     u.Verified,
			RefreshToken: token.RefreshToken,
		})

		if err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}
	}

	ss.CreateSession(c, session.SessionData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})

	return c.Redirect("/")

}

func Logout(c *fiber.Ctx) error {
	ss := session.NewSessionService()
	if sId := c.Cookies("session"); sId != "" {
		ss.ClearSession(c, sId)
	}
	return c.Redirect("/", 303)
}
