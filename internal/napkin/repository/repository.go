package repository

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

const napkinPrefix = "napkin:napkin:"

var ErrNotFound = errors.New("repository: resource not found")

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}
