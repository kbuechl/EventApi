package session

import (
	"eventapi/internal/cache"
	"eventapi/internal/configuration"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SessionService struct {
	cacheService *cache.CacheService
	cfg          *configuration.Server
}

type SessionManager interface {
	Create(c *fiber.Ctx, expiry time.Time, sd SessionData) (string, error)
	Clear(c *fiber.Ctx, sId string)
	Get(c *fiber.Ctx, sId string) (*SessionData, error)
}

func createSessionKey(sId string) string {
	return fmt.Sprintf("session:%v", sId)
}

func NewSessionService(c *cache.CacheService, cfg *configuration.Server) (*SessionService, error) {
	if cfg.CookieSecret == "" {
		return nil, fmt.Errorf("failed to initialize session service: Cookie Secret not set")
	}

	return &SessionService{
		cfg:          cfg,
		cacheService: c,
	}, nil
}

func (s *SessionService) Create(c *fiber.Ctx, expiry time.Time, sd SessionData) (string, error) {
	sId := uuid.New().String()
	key := createSessionKey(sId)
	err := s.cacheService.Set(c.Context(), key, sd, time.Until(expiry))
	if err != nil {
		return "", fmt.Errorf("error setting user session in cache: %w", err)
	}
	createSessionCookie(c, s.cfg.SessionCookieName, sId, expiry)
	return sId, nil
}

func (s *SessionService) Clear(c *fiber.Ctx, sId string) {
	clearSessionCookie(c, s.cfg.SessionCookieName)
	key := createSessionKey(sId)
	s.cacheService.Del(c.Context(), key)
}

func (s *SessionService) Get(c *fiber.Ctx, sId string) (*SessionData, error) {
	var sd SessionData
	data, err := s.cacheService.Get(c.Context(), createSessionKey(sId))
	if err != nil {
		return nil, err
	}

	sd.UnmarshalBinary([]byte(data))

	return &sd, nil
}
