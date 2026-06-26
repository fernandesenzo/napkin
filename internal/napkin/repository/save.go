package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fernandesenzo/napkin/internal/napkin"
)

func (r *RedisRepository) Save(ctx context.Context, npk *napkin.Napkin, ttl time.Duration) error {
	if err := r.client.Set(ctx, napkinPrefix+npk.Code, npk.Text, ttl).Err(); err != nil {
		return fmt.Errorf("repository.Save: failed to set napkin value on redis: %w", err)
	}
	return nil
}
