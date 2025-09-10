package tokenstore

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	rdb *redis.Client
}

func NewTokenStore(addr, password string) *TokenStore {
	return &TokenStore{
		rdb: redis.NewClient(&redis.Options{
			Addr: addr, Password: password, DB: 0,
		}),
	}
}

// Save a refresh token jti mapped to userID with TTL
func (s *TokenStore) SaveRefreshToken(ctx context.Context, jti string, userID uint, ttl time.Duration) error {
	key := "refresh:" + jti
	return s.rdb.Set(ctx, key, userID, ttl).Err()
}

func (s *TokenStore) ValidateRefreshToken(ctx context.Context, jti string) (uint, error) {
	key := "refresh:" + jti
	res, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// parse userID (stored as integer)
	id64, err := strconv.ParseUint(res, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}

func (s *TokenStore) DeleteRefreshToken(ctx context.Context, jti string) error {
	key := "refresh:" + jti
	return s.rdb.Del(ctx, key).Err()
}
