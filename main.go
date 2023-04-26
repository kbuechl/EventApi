package main

import (
	"errors"
	"eventapi/internal/auth"
	"eventapi/internal/cache"
	"eventapi/internal/configuration"
	"eventapi/internal/database"
	"eventapi/internal/middleware"
	"eventapi/internal/session"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func main() {
	config := configuration.New()

	db, err := database.New(&config.DB)
	if err != nil {
		log.Fatal(err)
	}

	authService, err := auth.NewAuthService(&config.Oauth)
	if err != nil {
		log.Fatal(err)
	}

	cacheService := cache.NewCacheService(&config.Cache)

	userRepo := database.NewUserRepo(db)
	if err != nil {
		log.Fatal(err)
	}

	sessionService, err := session.NewSessionService(cacheService, &config.Server)
	if err != nil {
		log.Fatal(err)
	}

	//todo: recover from panics

	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := err.Error()

			//todo: unwrap error futher?

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

	//todo: find a way to move this to middleware / session service
	ag.Use(encryptcookie.New(encryptcookie.Config{
		Key:    config.Server.CookieSecret,
		Except: []string{"csrf_1"},
	}))

	ag.Use(middleware.UseSession(sessionService))

	// todo: add csrf header
	// https://docs.gofiber.io/api/middleware/csrf

	ag.Get("/callback", auth.Callback(authService, userRepo, sessionService))
	ag.Get("/login", auth.Login(authService))
	ag.Get("/logout", auth.Logout(sessionService))
	// ag.Get("/albums", handlers.GetAlbums)
	// ag.Post("/albums", handlers.CreateAlbum)

	app.Listen(":3000")
}
