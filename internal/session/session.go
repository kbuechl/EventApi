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
	Create(c *fiber.Ctx, sd SessionData) string
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

func (s *SessionService) Create(c *fiber.Ctx, sd SessionData) string {
	sId := uuid.New().String()
	key := createSessionKey(sId)
	s.cacheService.Set(c.Context(), key, sd, time.Until(sd.Expiry))
	fmt.Println("expiry", sd.Expiry)
	createSessionCookie(c, s.cfg.SessionCookieName, sId, sd.Expiry)
	return sId
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
