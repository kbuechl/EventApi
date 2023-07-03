package auth

import (
	"errors"
	"eventapi/internal/database"
	"eventapi/internal/session"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// Login handler
func Login(a *AuthService) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		nonce, err := randString(16)

		if err != nil {
			return fmt.Errorf("error generating nonce value: %w", err)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "nonce",
			Value:    nonce,
			HTTPOnly: true,
			Secure:   c.Context().IsTLS(),
			MaxAge:   int(time.Hour.Seconds()),
		})

		url := a.getEntryPoint(nonce)
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

		rawIdToken, ok := token.Extra("id_token").(string)

		if !ok {
			return errors.New("error getting id_token from request")
		}

		idToken, err := oa.verify(c.Context(), rawIdToken)

		nonce := c.Cookies("nonce")

		log.Debug().Msgf("nonce: %v", idToken.Nonce)
		if nonce == "" || nonce != idToken.Nonce {
			log.Warn().Msg("nonce value did not match supplied")
			return c.SendStatus(http.StatusForbidden)
		}
		if err != nil {
			return c.SendStatus(http.StatusForbidden)
		}

		ui, err := oa.getUserInfo(c.Context(), token)

		if err != nil {
			return fmt.Errorf("could not get user info: %w", err)
		}

		if _, e := u.Get(ui.Email); e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return c.SendStatus(http.StatusForbidden)
			}
			return fmt.Errorf("error retrieving user during callback: %w", e)
		}

		user, err := u.Update(ui.Email, ui, token.RefreshToken)

		if err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		_, err = s.Create(c, token.Expiry, session.SessionData{
			Email:       user.Email,
			AccessToken: token.AccessToken,
		})

		if err != nil {
			log.Error().Err(err).Msg("error creating user session")
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
