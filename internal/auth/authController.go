package auth

import (
	"errors"
	"eventapi/internal/database"
	"eventapi/internal/session"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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
func Callback(oa oAuthService, u database.UserRepository, s session.SessionManager) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token, err := oa.exchange(c.Context(), c.FormValue("state"), c.FormValue("code"))
		if err != nil {
			return err
		}

		ui, err := getUserInfo(token.AccessToken)

		if err != nil {
			return fmt.Errorf("could not fetch user info: %w", err)
		}

		if _, e := u.Get(ui.Email); e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return c.SendStatus(403)
			}
			return fmt.Errorf("error retrieving user during callback: %w", e)
		}

		user, err := u.Update(ui.Email, &database.UpdateUser{
			ID:           ui.ID,
			FirstName:    ui.FirstName,
			LastName:     ui.LastName,
			Picture:      ui.Picture,
			Verified:     ui.Verified,
			RefreshToken: token.RefreshToken,
		})

		if err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		_, cErr := s.Create(c, token.Expiry, session.SessionData{
			Email:       user.Email,
			AccessToken: token.AccessToken,
		})

		if cErr != nil {
			//todo: add logger call
			println(cErr.Error())
			return c.SendStatus(500)
		}

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
