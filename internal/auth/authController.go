package auth

import (
	"eventapi/internal/database"
	"eventapi/internal/session"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// Login handler
func Login(a *AuthService) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		url := a.getEntryPoint()
		return c.Redirect(url)
	}

}

// oauth callback
func Callback(oa oAuthService, u database.UserRepository, s *session.SessionService) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token, err := oa.exchange(c.Context(), c.FormValue("state"), c.FormValue("code"))
		if err != nil {
			return err
		}

		ui, err := getUserInfo(token.AccessToken)

		if err != nil {
			return fmt.Errorf("could not fetch user info: %w", err)
		}

		if exists := u.Exists(ui.ID); !exists {
			println("User does not exist, creating")

			_, err := u.Create(database.NewUser{
				ID:           ui.ID,
				FirstName:    ui.FirstName,
				LastName:     ui.LastName,
				Email:        ui.Email,
				Picture:      ui.Picture,
				Verified:     ui.Verified,
				RefreshToken: token.RefreshToken,
			})

			if err != nil {
				return fmt.Errorf("error creating user: %w", err)
			}
		}

		s.Create(c, session.SessionData{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       token.Expiry,
		})

		return c.Redirect("/")
	}

}

func Logout(s *session.SessionService) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if sId := c.Cookies("session"); sId != "" {
			s.Clear(c, sId)
		}
		return c.Redirect("/", 303)
	}
}
