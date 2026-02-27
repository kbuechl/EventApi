package main

import (
	"errors"
	"eventapi/internal/albums"
	"eventapi/internal/auth"
	"eventapi/internal/cache"
	"eventapi/internal/configuration"
	"eventapi/internal/contacts"
	"eventapi/internal/database"
	"eventapi/internal/events"
	"eventapi/internal/middleware"
	"eventapi/internal/session"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/rs/zerolog/log"
)

func main() {
	config := configuration.New()

	db, err := database.New(&config.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	authService, err := auth.NewAuthService(&config.Oauth)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize auth service")
	}

	cacheService := cache.NewCacheService(&config.Cache)

	userRepo := database.NewUserRepo(db)
	eventRepo := database.NewEventRepo(db)
	contactRepo := database.NewContactRepo(db)

	sessionService, err := session.NewSessionService(cacheService, &config.Server)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize session service")
	}

	if m := os.Getenv("MIGRATE_ENABLED"); m == "true" {
		if err := userRepo.Migrate(); err != nil {
			log.Fatal().Err(err).Msg("user migration failed")
		}
		if err := eventRepo.Migrate(); err != nil {
			log.Fatal().Err(err).Msg("event migration failed")
		}
		if err := contactRepo.Migrate(); err != nil {
			log.Fatal().Err(err).Msg("contact migration failed")
		}
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			log.Error().Err(err).Msg("unhandled exception")
			return ctx.Status(code).JSON(fiber.Map{"message": err.Error()})
		},
	})

	// Serve built frontend
	app.Static("/", "./dist")

	ag := app.Group("api")
	ag.Use(compress.New())
	ag.Use(encryptcookie.New(encryptcookie.Config{
		Key:    config.Server.CookieSecret,
		Except: []string{"csrf_1"},
	}))
	ag.Use(middleware.UseSession(sessionService))

	// ── Public routes ──────────────────────────────────────────────────────────
	ag.Get("/login", auth.Login(authService))
	ag.Get("/callback", auth.Callback(authService, userRepo, sessionService))
	ag.Get("/logout", auth.Logout(sessionService))

	// ── Authenticated routes ───────────────────────────────────────────────────
	protected := ag.Group("", middleware.RequireAuth())

	protected.Get("/me", auth.Me(userRepo))

	eventCtrl := events.NewController(eventRepo, userRepo)
	protected.Get("/events", eventCtrl.List)
	protected.Post("/events", eventCtrl.Create)
	protected.Get("/events/:id", eventCtrl.Get)
	protected.Put("/events/:id", eventCtrl.Update)
	protected.Delete("/events/:id", eventCtrl.Delete)
	protected.Post("/events/:id/invite", eventCtrl.Invite)
	protected.Put("/events/:id/rsvp", eventCtrl.RSVP)

	albumCtrl := albums.NewController(eventRepo)
	protected.Post("/events/:id/photos", albumCtrl.UploadPhotos)
	protected.Delete("/events/:id/photos/:photoId", albumCtrl.DeletePhoto)

	contactCtrl := contacts.NewController(contactRepo, userRepo)
	protected.Get("/contacts", contactCtrl.List)
	protected.Post("/contacts", contactCtrl.Add)
	protected.Delete("/contacts/:email", contactCtrl.Remove)

	if err := app.Listen(":3000"); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}
