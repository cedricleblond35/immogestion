package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// =============================================================================
// STRATÉGIE 1: VARIABLES GLOBALES + FONCTIONS GLOBALES
// =============================================================================
//
// Description:
// - Utilise des variables globales pour stocker la configuration
// - Fonctions globales simples à utiliser
// - Pas besoin d'instancier de service
//
// Avantages:
// + Simplicité d'utilisation (pas d'instance à créer)
// + Compatible avec du code legacy
// + Accès rapide depuis n'importe où dans l'application
// + Moins de boilerplate code
//
// Inconvénients:
// - Variables globales mutables (risque de concurrence)
// - Difficile à tester (état global partagé)
// - Pas de configuration par instance
// - Risque de collision de configuration
// - Plus difficile à mocker dans les tests
//
// Usage:
//   Initialize("secret-key")
//   token, err := GenerateAccessToken(userID, email, ttl)
//   claims, err := ParseAndValidateAccessToken(token)
//
// =============================================================================

var jwtSecret []byte

// Initialize initialise le secret JWT global (à appeler au démarrage de l'application)
func Initialize(secret string) {
	jwtSecret = []byte(secret)
}

// ParseAndValidateAccessToken version globale compatible avec du code existant
func ParseAndValidateAccessToken(tok string) (*AccessClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("jwt secret not initialized - call Initialize() first")
	}

	// Parse et valider la signature
	token, err := jwt.ParseWithClaims(tok, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature pour éviter alg=none attacks
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Assertions et conversion de types
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims or token")
	}

	// Vérifier issuer (iss)
	if claims.Issuer != "immogestion-auth" {
		return nil, errors.New("invalid issuer")
	}

	// Vérifier audience (aud)
	requiredAud := "immogestion-gateway"
	audOK := false
	for _, a := range claims.Audience {
		if a == requiredAud {
			audOK = true
			break
		}
	}
	if !audOK {
		return nil, errors.New("invalid audience")
	}

	// Vérifier not before (nbf)
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		return nil, errors.New("token not yet valid (nbf)")
	}

	// Vérifier expiration (exp)
	if claims.ExpiresAt == nil {
		return nil, errors.New("token without expiration")
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// GenerateAccessToken version globale compatible avec du code existant
func GenerateAccessToken(userID uint, email string, ttl time.Duration) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("jwt secret not initialized - call Initialize() first")
	}

	now := time.Now()
	claims := AccessClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("%d_%d", userID, now.Unix()), // jti unique
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "immogestion-auth",
			Audience:  []string{"immogestion-gateway"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// =============================================================================
// STRATÉGIE 2: SERVICE ORIENTÉ OBJET (RECOMMANDÉ)
// =============================================================================
//
// Description:
// - Service structuré avec état encapsulé
// - Configuration par instance
// - Méthodes d'instance pour toutes les opérations
//
// Avantages:
// + Isolation de la configuration (plusieurs instances possibles)
// + Facilité de test (injection de dépendances)
// + État encapsulé et thread-safe
// + Configuration flexible par instance
// + Meilleure architecture (SOLID principles)
// + Facilité de mocking dans les tests
//
// Inconvénients:
// - Plus de boilerplate (création d'instances)
// - Légèrement plus complexe à utiliser
// - Besoin de passer l'instance partout
//
// Usage:
//   jwtService, err := NewJWTService(accessTTL, refreshTTL)
//   accessToken, refreshToken, tokenID, _, _, err := jwtService.GenerateTokenPair(userID, email, role)
//   claims, err := jwtService.ValidateAccessToken(token)
//
// =============================================================================

// AccessClaims représente les claims pour les access tokens
type AccessClaims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// RefreshClaims représente les claims pour les refresh tokens
type RefreshClaims struct {
	UserID    uint   `json:"uid"`
	TokenType string `json:"typ"` // "refresh"
	jwt.RegisteredClaims
}

// JWTService gère la création et validation des tokens JWT
type JWTService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTService crée une nouvelle instance de JWTService
func NewJWTService(accessTokenTTL, refreshTokenTTL time.Duration) (*JWTService, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET non configuré dans les variables d'environnement")
	}

	if len(secretKey) < 32 {
		return nil, errors.New("JWT_SECRET doit contenir au moins 32 caractères pour la sécurité")
	}

	return &JWTService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}, nil
}

// GenerateTokenPair génère une paire de tokens (access + refresh) avec un TokenID commun
func (j *JWTService) GenerateTokenPair(userID uint, email, role string) (accessToken, refreshToken, tokenID string, accessExp, refreshExp int64, err error) {
	// Générer un ID unique pour cette session de token
	tokenID, err = j.generateTokenID()
	if err != nil {
		return "", "", "", 0, 0, fmt.Errorf("erreur lors de la génération de l'ID de token: %w", err)
	}

	now := time.Now()
	accessExp = now.Add(j.accessTokenTTL).Unix()
	refreshExp = now.Add(j.refreshTokenTTL).Unix()

	// Générer l'access token
	accessToken, err = j.generateAccessToken(userID, email, role, tokenID, now, accessExp)
	if err != nil {
		return "", "", "", 0, 0, fmt.Errorf("erreur lors de la génération de l'access token: %w", err)
	}

	// Générer le refresh token
	refreshToken, err = j.generateRefreshToken(userID, email, role, tokenID, now, refreshExp)
	if err != nil {
		return "", "", "", 0, 0, fmt.Errorf("erreur lors de la génération du refresh token: %w", err)
	}

	return accessToken, refreshToken, tokenID, accessExp, refreshExp, nil
}

// ValidateAccessToken parse et valide un access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, j.keyFunc)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims or token")
	}

	// Validations supplémentaires
	if err := j.validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// GetTokenExpiry extrait la date d'expiration du token sans validation
func (j *JWTService) GetTokenExpiry(tokenString string) (int64, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &AccessClaims{})
	if err != nil {
		return 0, fmt.Errorf("erreur lors de l'extraction de l'expiration du token: %w", err)
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && claims.ExpiresAt != nil {
		return claims.ExpiresAt.Time.Unix(), nil
	}

	return 0, fmt.Errorf("impossible d'extraire l'expiration du token")
}

// IsTokenExpired vérifie si un token est expiré sans validation complète
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return true
	}

	if claims.ExpiresAt == nil {
		return true
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}

// ExtractClaims extrait les claims d'un token sans validation complète
func (j *JWTService) ExtractClaims(tokenString string) (*AccessClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &AccessClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if claims, ok := token.Claims.(*AccessClaims); ok {
		return claims, nil
	}

	return nil, errors.New("unable to extract claims")
}

// GetAccessTokenTTL retourne la durée de vie de l'access token
func (j *JWTService) GetAccessTokenTTL() time.Duration {
	return j.accessTokenTTL
}

// GetRefreshTokenTTL retourne la durée de vie du refresh token
func (j *JWTService) GetRefreshTokenTTL() time.Duration {
	return j.refreshTokenTTL
}

// =============================================================================
// STRATÉGIE 3: APPROCHE HYBRIDE/FACTORY
// =============================================================================
//
// Description:
// - Combine les avantages des deux approches précédentes
// - Utilise le service structuré mais avec des méthodes de création simplifiées
// - Factory pattern pour créer des instances pré-configurées
//
// Avantages:
// + Flexibilité d'usage (global ou instance)
// + Configuration par défaut simple
// + Possibilité de personnalisation avancée
// + Transition facile entre les approches
//
// Inconvénients:
// - Plus de complexité dans le code
// - Peut créer de la confusion sur quelle méthode utiliser
// - Duplicitation de logique potentielle
//
// Usage:
//   // Approche simple
//   token, err := CreateAccessToken(ctx, userID, email, role)
//
//   // Approche avec configuration
//   service := NewJWTServiceWithSecret("custom-secret", ttl1, ttl2)
//   token, err := service.GenerateAccessToken(userID, email, role)
//
// =============================================================================

// CreateAccessToken crée un token JWT avec configuration par défaut
func CreateAccessToken(ctx context.Context, userID uint, email, role string) (string, error) {
	// Utiliser des valeurs par défaut
	service, err := NewJWTService(15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return "", err
	}

	return service.GenerateAccessToken(userID, email, role)
}

// NewJWTServiceWithSecret crée un service JWT avec un secret personnalisé
func NewJWTServiceWithSecret(secret string, accessTokenTTL, refreshTokenTTL time.Duration) (*JWTService, error) {
	if len(secret) < 32 {
		return nil, errors.New("secret must be at least 32 characters long")
	}

	return &JWTService{
		secretKey:       []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}, nil
}

// SetSecret définit le secret JWT pour une instance existante
func (j *JWTService) SetSecret(secret string) error {
	if len(secret) < 32 {
		return errors.New("secret must be at least 32 characters long")
	}
	j.secretKey = []byte(secret)
	return nil
}

// GenerateAccessToken génère un access token simple (version hybride)
func (j *JWTService) GenerateAccessToken(userID uint, email, role string) (string, error) {
	if len(j.secretKey) == 0 {
		return "", errors.New("jwt secret not configured")
	}

	now := time.Now()
	claims := &AccessClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("%d_%d", userID, now.Unix()), // jti unique
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "immogestion-auth",
			Audience:  []string{"immogestion-gateway"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// =============================================================================
// MÉTHODES PRIVÉES PARTAGÉES (utilisées par toutes les stratégies)
// =============================================================================

// generateAccessToken génère un access token
func (j *JWTService) generateAccessToken(userID uint, email, role, tokenID string, issuedAt time.Time, expiresAt int64) (string, error) {
	claims := AccessClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID, // JTI (JWT ID)
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expiresAt, 0)),
			NotBefore: jwt.NewNumericDate(issuedAt),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "immogestion-auth",
			Audience:  []string{"immogestion-gateway"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// generateRefreshToken génère un refresh token
func (j *JWTService) generateRefreshToken(userID uint, email, role, tokenID string, issuedAt time.Time, expiresAt int64) (string, error) {
	claims := RefreshClaims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID, // <— plus de suffixe
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expiresAt, 0)),
			NotBefore: jwt.NewNumericDate(issuedAt),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "immogestion-auth",
			Audience:  []string{"immogestion-gateway"},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// generateTokenID génère un ID unique sécurisé pour le token
func (j *JWTService) generateTokenID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("erreur lors de la génération de l'ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// keyFunc retourne la clé de signature pour la validation JWT
func (j *JWTService) keyFunc(token *jwt.Token) (interface{}, error) {
	// Vérifier la méthode de signature pour éviter les attaques alg=none
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
	}

	if len(j.secretKey) == 0 {
		return nil, errors.New("clé secrète JWT non configurée")
	}

	return j.secretKey, nil
}

// validateClaims valide les claims standards
func (j *JWTService) validateClaims(claims *AccessClaims) error {
	now := time.Now()

	// Vérifier l'issuer
	if claims.Issuer != "immogestion-auth" {
		return fmt.Errorf("invalid issuer: expected 'immogestion-auth', got '%s'", claims.Issuer)
	}

	// Vérifier l'audience
	requiredAud := "immogestion-gateway"
	audOK := false
	for _, a := range claims.Audience {
		if a == requiredAud {
			audOK = true
			break
		}
	}
	if !audOK {
		return errors.New("invalid audience")
	}

	// Vérifier not before (nbf)
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
		return errors.New("token not yet valid (nbf)")
	}

	// Vérifier expiration (exp)
	if claims.ExpiresAt == nil {
		return errors.New("token without expiration")
	}
	if claims.ExpiresAt.Time.Before(now) {
		return errors.New("token expired")
	}

	// Vérifier issued at (iat)
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(now.Add(5*time.Minute)) {
		return errors.New("token issued in the future")
	}

	return nil
}

// validateRefreshTokenClaims valide les claims d'un refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, j.keyFunc)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du parsing du refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("claims invalides ou refresh token invalide")
	}

	// validations de base + typ=refresh
	if err := j.validateRefreshClaims(claims); err != nil {
		return nil, err
	}
	return claims, nil
}
func (j *JWTService) validateRefreshClaims(claims *RefreshClaims) error {
	now := time.Now()

	if claims.Issuer != "immogestion-auth" {
		return fmt.Errorf("invalid issuer: expected 'immogestion-auth', got '%s'", claims.Issuer)
	}

	requiredAud := "immogestion-gateway"
	audOK := false
	for _, a := range claims.Audience {
		if a == requiredAud {
			audOK = true
			break
		}
	}
	if !audOK {
		return errors.New("invalid audience")
	}

	if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
		return errors.New("refresh token not yet valid (nbf)")
	}

	if claims.ExpiresAt == nil {
		return errors.New("refresh token without expiration")
	}
	if claims.ExpiresAt.Time.Before(now) {
		return errors.New("refresh token expired")
	}

	if claims.TokenType != "refresh" {
		return errors.New("invalid refresh token format") // typ manquant ou incorrect
	}

	return nil
}
