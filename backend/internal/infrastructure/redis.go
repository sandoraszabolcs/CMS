package infrastructure

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisChannelValidationEvents is the pub/sub channel for real-time event streaming.
const RedisChannelValidationEvents = "validation_events"

// NewRedisClient creates a configured Redis client.
func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return client, nil
}
