package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// RedisChannelValidationEvents is the pub/sub channel for real-time event streaming.
const RedisChannelValidationEvents = "validation_events"

// NewRedisClient creates a configured Redis client.
// addr can be a full URL (redis://...) or a plain host:port.
func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
	var opts *redis.Options
	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		var err error
		opts, err = redis.ParseURL(addr)
		if err != nil {
			return nil, fmt.Errorf("parsing redis URL: %w", err)
		}
	} else {
		opts = &redis.Options{Addr: addr}
	}
	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return client, nil
}
