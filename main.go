package main

import (
	"errors"
	"eventapi/config"
	"eventapi/internal/handlers"
	"eventapi/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func main() {
	//todo: recover from panics
	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := err.Error()

			var e *fiber.Error

			if errors.As(err, &e) {
				code = e.Code
			}

			return ctx.Status(code).JSON(ErrorMessage{Message: message})
		},
	})

	ag := app.Group("api")

	app.Static("/", "./dist")

	ag.Use(compress.New())

	ag.Use(encryptcookie.New(encryptcookie.Config{
		Key:    config.New().Server.CookieSecret,
		Except: []string{"csrf_1"},
	}))

	ag.Use(middleware.UseSession)

	// todo: add csrf header
	// https://docs.gofiber.io/api/middleware/csrf

	ag.Get("/", handlers.HelloWorld)
	ag.Get("/test", handlers.AnotherFunction)
	ag.Get("/callback", handlers.Callback)
	ag.Get("/login", handlers.Auth)
	ag.Get("/logout", handlers.Logout)
	ag.Get("/albums", handlers.GetAlbums)
	ag.Post("/albums", handlers.CreateAlbum)

	//todo: can we use context in handlers to pass them down to database calls

	app.Listen(":3000")
}
