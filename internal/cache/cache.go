package cache

import (
	"context"
	"eventapi/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

func NewCacheService() *CacheService {
	config := config.New()

	return &CacheService{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Cache.Address,
			Password: "",
			DB:       0,
		}),
	}
}

func (s *CacheService) Get(c context.Context, key string) (string, error) {
	return s.client.Get(c, key).Result()
}

func (s *CacheService) Set(c context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(c, key, value, ttl).Err()
}

func (s *CacheService) Del(c context.Context, key string) {
	s.client.Del(c, key)
}
