// services/cache.go
package services

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Cache struct {
	client *redis.Client
}

func NewCache(addr string) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Cache{client: rdb}
}

func (c *Cache) Set(key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}
