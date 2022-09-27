package dbrepo

import (
	"context"
	"time"

	"github.com/go-redis/cache/v9"
)

func (m *redisRepo) Set(key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.RedisCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
	})
	return err
}

func (m *redisRepo) Get(key string, dest interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.RedisCache.Get(ctx, key, dest)
	return err
}

func (m *redisRepo) Exists(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.RedisCache.Exists(ctx, key)
}
