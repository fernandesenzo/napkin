package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/fernandesenzo/napkin/internal/napkin"
	"github.com/redis/go-redis/v9"
)

func (r *RedisRepository) Get(ctx context.Context, code string) (*napkin.Napkin, error) {
	content, err := r.client.Get(ctx, napkinPrefix+code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository.Get: failed to get napkin from redis client: %w", err)
	}
	return &napkin.Napkin{Code: code, Text: content}, nil
}
