package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenRepository définit l'interface pour les opérations sur les tokens
type TokenRepository interface {
	// StoreRefreshToken stocke un refresh token avec son ID et sa date d'expiration
	StoreRefreshToken(ctx context.Context, userID uint, tokenID string, token string, expiry int64) error

	// GetRefreshToken récupère un refresh token par userID et tokenID
	GetRefreshToken(ctx context.Context, userID uint, tokenID string) (string, error)

	// DeleteRefreshToken supprime un refresh token spécifique
	DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error

	// DeleteAllUserTokens supprime tous les refresh tokens d'un utilisateur
	// Utile pour "logout all devices"
	DeleteAllUserTokens(ctx context.Context, userID uint) error

	// IsTokenBlacklisted vérifie si un access token est sur la liste noire
	IsTokenBlacklisted(ctx context.Context, tokenID string) bool

	// BlacklistToken ajoute un access token à la liste noire avec expiration
	BlacklistToken(ctx context.Context, tokenID string, expiry int64) error
}

// RedisTokenRepository implémente l'interface TokenRepository avec Redis
type RedisTokenRepository struct {
	client *redis.Client
}

// NewRedisTokenRepository crée une nouvelle instance de RedisTokenRepository
// Cette fonction est celle que vous utilisez dans votre main.go
func NewRedisTokenRepository(client *redis.Client) TokenRepository {
	return &RedisTokenRepository{
		client: client,
	}
}

// StoreRefreshToken stocke un refresh token dans Redis
// Structure des clés Redis :
// - refresh_token:{userID}:{tokenID} -> le token lui-même
// - user_tokens:{userID} -> set de tous les tokenID de l'utilisateur
func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, userID uint, tokenID string, token string, expiry int64) error {
	// Clé pour stocker le refresh token
	key := r.getRefreshTokenKey(userID, tokenID)
	duration := time.Until(time.Unix(expiry, 0))

	// Stocker le token avec expiration automatique
	if err := r.client.Set(ctx, key, token, duration).Err(); err != nil {
		return fmt.Errorf("erreur lors du stockage du refresh token: %w", err)
	}

	// Ajouter le tokenID à la liste des tokens de l'utilisateur
	// Cela permet de retrouver tous les tokens d'un utilisateur pour les révoquer
	userTokensKey := r.getUserTokensKey(userID)
	if err := r.client.SAdd(ctx, userTokensKey, tokenID).Err(); err != nil {
		return fmt.Errorf("erreur lors de l'ajout à la liste des tokens: %w", err)
	}

	// Définir l'expiration pour la liste des tokens
	if err := r.client.Expire(ctx, userTokensKey, duration).Err(); err != nil {
		return fmt.Errorf("erreur lors de la définition de l'expiration: %w", err)
	}

	return nil
}

// GetRefreshToken récupère un refresh token depuis Redis
func (r *RedisTokenRepository) GetRefreshToken(ctx context.Context, userID uint, tokenID string) (string, error) {
	key := r.getRefreshTokenKey(userID, tokenID)

	token, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("refresh token non trouvé ou expiré")
		}
		return "", fmt.Errorf("erreur lors de la récupération du refresh token: %w", err)
	}

	return token, nil
}

// DeleteRefreshToken supprime un refresh token de Redis
func (r *RedisTokenRepository) DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error {
	key := r.getRefreshTokenKey(userID, tokenID)

	// Supprimer le token lui-même
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("erreur lors de la suppression du refresh token: %w", err)
	}

	// Supprimer de la liste des tokens de l'utilisateur
	userTokensKey := r.getUserTokensKey(userID)
	if err := r.client.SRem(ctx, userTokensKey, tokenID).Err(); err != nil {
		return fmt.Errorf("erreur lors de la suppression de la liste des tokens: %w", err)
	}

	return nil
}

// DeleteAllUserTokens supprime tous les refresh tokens d'un utilisateur
// Utile pour déconnecter un utilisateur de tous ses appareils
func (r *RedisTokenRepository) DeleteAllUserTokens(ctx context.Context, userID uint) error {
	userTokensKey := r.getUserTokensKey(userID)

	// Récupérer tous les tokenID de l'utilisateur
	tokenIDs, err := r.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération des tokens: %w", err)
	}

	// Supprimer chaque token individuellement
	for _, tokenID := range tokenIDs {
		key := r.getRefreshTokenKey(userID, tokenID)
		if err := r.client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("erreur lors de la suppression du token %s: %w", tokenID, err)
		}
	}

	// Supprimer la liste des tokens
	if err := r.client.Del(ctx, userTokensKey).Err(); err != nil {
		return fmt.Errorf("erreur lors de la suppression de la liste des tokens: %w", err)
	}

	return nil
}

// IsTokenBlacklisted vérifie si un access token est sur la liste noire
// Les tokens sont ajoutés à la blacklist lors du logout
func (r *RedisTokenRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) bool {
	key := r.getBlacklistKey(tokenID)

	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		// En cas d'erreur, considérer comme non blacklisté pour éviter de bloquer les utilisateurs
		return false
	}

	return result > 0
}

// BlacklistToken ajoute un access token à la liste noire
// Utilisé lors du logout pour invalider le token immédiatement
func (r *RedisTokenRepository) BlacklistToken(ctx context.Context, tokenID string, expiry int64) error {
	key := r.getBlacklistKey(tokenID)
	duration := time.Until(time.Unix(expiry, 0))

	// Stocker une valeur simple pour marquer le token comme blacklisté
	if err := r.client.Set(ctx, key, "blacklisted", duration).Err(); err != nil {
		return fmt.Errorf("erreur lors de l'ajout du token à la liste noire: %w", err)
	}

	return nil
}

// Méthodes utilitaires pour générer les clés Redis

// getRefreshTokenKey génère la clé Redis pour un refresh token
// Format: "refresh_token:{userID}:{tokenID}"
func (r *RedisTokenRepository) getRefreshTokenKey(userID uint, tokenID string) string {
	return fmt.Sprintf("refresh_token:%d:%s", userID, tokenID)
}

// getUserTokensKey génère la clé Redis pour la liste des tokens d'un utilisateur
// Format: "user_tokens:{userID}"
func (r *RedisTokenRepository) getUserTokensKey(userID uint) string {
	return fmt.Sprintf("user_tokens:%d", userID)
}

// getBlacklistKey génère la clé Redis pour un token blacklisté
// Format: "blacklist:{tokenID}"
func (r *RedisTokenRepository) getBlacklistKey(tokenID string) string {
	return fmt.Sprintf("blacklist:%s", tokenID)
}

// Méthodes utilitaires supplémentaires

// GetUserTokenCount retourne le nombre de tokens actifs pour un utilisateur
func (r *RedisTokenRepository) GetUserTokenCount(ctx context.Context, userID uint) (int64, error) {
	userTokensKey := r.getUserTokensKey(userID)
	return r.client.SCard(ctx, userTokensKey).Result()
}

// GetAllUserTokens retourne tous les tokenID actifs pour un utilisateur
func (r *RedisTokenRepository) GetAllUserTokens(ctx context.Context, userID uint) ([]string, error) {
	userTokensKey := r.getUserTokensKey(userID)
	return r.client.SMembers(ctx, userTokensKey).Result()
}

// CleanupExpiredTokens nettoie les tokens expirés (à appeler périodiquement)
// Note: Redis fait cela automatiquement, mais cette méthode peut être utile pour le monitoring
func (r *RedisTokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	// Cette méthode pourrait parcourir les clés et supprimer celles qui sont expirées
	// Mais Redis le fait automatiquement, donc cette méthode est optionnelle
	return nil
}

// GetTokenInfo retourne des informations sur un token (pour debug/monitoring)
func (r *RedisTokenRepository) GetTokenInfo(ctx context.Context, userID uint, tokenID string) (map[string]interface{}, error) {
	key := r.getRefreshTokenKey(userID, tokenID)

	// Vérifier si le token existe
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"exists": exists > 0,
		"key":    key,
	}

	if exists > 0 {
		// Obtenir le TTL
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		info["ttl_seconds"] = ttl.Seconds()
		info["expires_at"] = time.Now().Add(ttl).Unix()
	}

	return info, nil
}