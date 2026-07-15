package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type ICache interface {
	Save(ctx context.Context, key string, data any, duration time.Duration) error
	Get(ctx context.Context, key string, data any) error
}

type cache struct {
	client *redis.Client
}

var ErrNotFound = errors.New("redis: not found")

func New(ctx context.Context, address string) (ICache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})

	ping := client.Ping(ctx)

	log.Println("client.Ping():", ping)

	pong, err := ping.Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed: %v", err)
	}

	log.Println("client.Ping():", pong)

	return &cache{
		client: client,
	}, nil
}

func (c *cache) Save(ctx context.Context, key string, data any, duration time.Duration) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	err = c.client.Set(ctx, key, dataBytes, duration).Err()
	if err != nil {
		return fmt.Errorf("client.Set: %w", err)
	}

	return nil
}

func (c *cache) Get(ctx context.Context, key string, data any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return fmt.Errorf("client.Get: %w", err)
	}

	err = json.Unmarshal([]byte(val), data)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	return nil
}
