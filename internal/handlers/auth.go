package handlers

import (
	"eventapi/internal/auth"
	"eventapi/internal/database"
	"eventapi/internal/session"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Login handler
func Auth(c *fiber.Ctx) error {
	path := auth.ConfigGoogle()
	url := path.AuthCodeURL("state1")
	return c.Redirect(url)

}

// Callback to receive google's response
func Callback(c *fiber.Ctx) error {
	us := database.NewUserService()
	ss := session.NewSessionService()

	fmt.Println("state from context", c.FormValue("state"))
	token, err := auth.ConfigGoogle().Exchange(c.Context(), c.FormValue("code"))
	if err != nil {
		return err
	}
	fmt.Println(token)

	u, err := auth.GetUserInfo(token.AccessToken)

	if err != nil {
		fmt.Println("error fetching user info", err.Error())
		return err
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
			return err
		}
	}

	ss.CreateSession(c, session.SessionData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})

	if err != nil {
		println("error with redis")
		panic(err)
	}

	if err != nil {
		fmt.Println("error converting token")
		return err
	}

	return c.Redirect("/")

}

func Logout(c *fiber.Ctx) error {
	ss := session.NewSessionService()
	if sId := c.Cookies("session"); sId != "" {
		ss.ClearSession(c, sId)
	}
	return c.Redirect("/", 303)
}
