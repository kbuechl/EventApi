package session

import (
	"eventapi/internal/cache"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SessionService struct {
	cacheService *cache.CacheService
}

func createSessionKey(sId string) string {
	return fmt.Sprintf("session:%v", sId)
}

func NewSessionService() *SessionService {
	return &SessionService{
		cacheService: cache.NewCacheService(),
	}
}

func (s *SessionService) CreateSession(c *fiber.Ctx, sd SessionData) string {
	sId := uuid.New().String()
	key := createSessionKey(sId)
	s.cacheService.Set(c.Context(), key, sd, time.Until(sd.Expiry))
	fmt.Println("expiry", sd.Expiry)
	CreateSessionCookie(c, sId, sd.Expiry)
	return sId
}

func (s *SessionService) ClearSession(c *fiber.Ctx, sId string) {
	ClearSessionCookie(c)
	key := createSessionKey(sId)
	s.cacheService.Del(c.Context(), key)
}

func (s *SessionService) GetSession(c *fiber.Ctx, sId string) (*SessionData, error) {
	var sd SessionData
	data, err := s.cacheService.Get(c.Context(), createSessionKey(sId))
	if err != nil {
		return nil, err
	}

	sd.UnmarshalBinary([]byte(data))

	return &sd, nil
}
