package cache

import (
	"context"
	"eventapi/internal/configuration"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type config struct {
	Address  string
	Password string
}

type CacheService struct {
	client *redis.Client
	config *config
}

func NewCacheService() *CacheService {
	c := configure()

	return &CacheService{
		config: c,
		client: redis.NewClient(&redis.Options{
			Addr:     c.Address,
			Password: "",
			DB:       0,
		}),
	}
}

func (s *CacheService) Get(c context.Context, key string) (string, error) {
	return s.client.Get(c, key).Result()
}

func (s *CacheService) Set(c context.Context, key string, value any, ttl time.Duration) error {
	return s.client.Set(c, key, value, ttl).Err()
}

func (s *CacheService) Del(c context.Context, key string) {
	s.client.Del(c, key)
}

func configure() *config {
	return &config{
		Address:  fmt.Sprintf("%v:6379", configuration.GetEnv("REDIS_HOST", "redis")),
		Password: configuration.GetEnv("REDIS_PASSWORD", ""),
	}
}
