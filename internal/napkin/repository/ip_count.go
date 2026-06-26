package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (r *RedisRepository) GetIPCounter(ctx context.Context, ip string) (int, error) {
	count, err := r.client.Get(ctx, ipPrefix+ip).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, fmt.Errorf("repository.GetIPCounter: failed to get ip counter from redis: %w", err)
	}
	intCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, fmt.Errorf("repository.GetIPCounter: got an invalid value for the ip counter: %w", err)
	}
	return intCount, nil
}

func (r *RedisRepository) IncrementIPCounter(ctx context.Context, ip string) error {
	if _, err := r.client.Incr(ctx, ipPrefix+ip).Result(); err != nil {
		return fmt.Errorf("repository.IncrementIPCounter: failed to increment IP counter :%w", err)
	}
	if err := r.client.Expire(ctx, ipPrefix+ip, time.Minute).Err(); err != nil {
		slog.WarnContext(ctx, "repository.IncrementIPCounter: failed to set ip key to expire", "error", err, "ip", ip)
	}
	return nil
}
