package tokenstore

import (
	"context"
	"fmt"
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
func (s *TokenStore) SaveRefreshToken(ctx context.Context, jti string, token string, ttl time.Duration) error {
	key := "refresh:" + jti
	return s.rdb.Set(ctx, key, token, ttl).Err()
}

func (s *TokenStore) ValidateRefreshToken(ctx context.Context, jti, token string) error {
	key := "refresh:" + jti
	res, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if res != token {
		return fmt.Errorf("refresh token mismatch")
	}
	return nil
}

// Store un refresh token
func (s *TokenStore) StoreRefreshToken(ctx context.Context, userID uint, jti string, token string, ttl time.Duration) error {
    key := fmt.Sprintf("refresh_token:%d:%s", userID, jti)
    return s.rdb.Set(ctx, key, token, ttl).Err()
}

// Récupère un refresh token
func (s *TokenStore) GetRefreshToken(ctx context.Context, userID uint, jti string) (string, error) {
    key := fmt.Sprintf("refresh_token:%d:%s", userID, jti)
    return s.rdb.Get(ctx, key).Result()
}

// Supprime un refresh token
func (s *TokenStore) DeleteRefreshToken(ctx context.Context, userID uint, jti string) error {
    key := fmt.Sprintf("refresh_token:%d:%s", userID, jti)
    return s.rdb.Del(ctx, key).Err()
}
