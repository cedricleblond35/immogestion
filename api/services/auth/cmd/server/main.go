package main

import (
	"api/services/auth/internal/database"
	model "api/services/auth/internal/models"
	"api/services/auth/internal/repository"
	"api/services/auth/internal/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents the expected request structure for user registration
type RegisterRequest struct {
	Company   string `json:"company" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents the expected request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the request structure for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterResponse represents the response structure for registration endpoint
type RegisterResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TokenResponse represents the JWT token response
type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	User         interface{} `json:"user"`
}

func main() {
	// Logger avec Zap
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Info("Starting auth service...")

	ctx := context.Background()

	// Database
	db, err := database.GormOpen(ctx, false, sugar)
	if err != nil {
		sugar.Errorf("Failed to connect to database: %v", err)
		sugar.Fatal("Exiting due to DB error")
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			sugar.Errorf("Failed to get sql.DB for close: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			sugar.Errorf("Failed to close database connection: %v", err)
		} else {
			sugar.Info("Database connection pool closed successfully")
		}
	}()

	// Initialiser Redis pour les tokens
	redisURL := getEnv("REDIS_URL", "redis://:cleb76yebles77@redis:6379")
	redisClient := initRedis(redisURL, sugar)
	sugar.Infof("Connecting to Redis: %s", redisURL)



	// Tester la connexion Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		sugar.Errorf("Failed to connect to Redis: %v", err)
		sugar.Fatal("Exiting due to Redis error")
	}
	sugar.Info("Connected to Redis successfully")

	// Initialiser les services JWT
	accessTokenTTL := 15 * time.Minute
	refreshTokenTTL := 7 * 24 * time.Hour // 7 jours

	jwtService, err := services.NewJWTService(accessTokenTTL, refreshTokenTTL)
	if err != nil {
		sugar.Fatalf("Failed to initialize JWT service: %v", err)
	}
	sugar.Info("JWT service initialized successfully")

	// Créer les repositories avec adaptateur
	userRepo := repository.NewUserRepository(db, sugar)
	tokenRepo := database.NewRedisTokenRepository(redisClient)

	// Initialize Gin router
	sugar.Info("Setting up routes...")
	r := gin.Default()

	// Middleware CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
			"version": "1.0.0",
		})
	})

	// User registration endpoint avec JWT
	sugar.Info("Setting up /register endpoint...")
	r.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, RegisterResponse{
				Status:  "error",
				Message: "Invalid request: " + err.Error(),
			})
			return
		}

		// Hasher le mot de passe
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			sugar.Errorf("Failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, RegisterResponse{
				Status:  "error",
				Message: "Internal server error during password hashing",
			})
			return
		}

		// Créer l'utilisateur
		user := &model.User{
			Company:   req.Company,
			Lastname:  req.Lastname,
			Firstname: req.Firstname,
			Email:     req.Email,
			Password:  string(hashedPassword),
			Role:      "user", // Rôle par défaut
			IsActive:  true,
		}

		if err := userRepo.Create(ctx, user); err != nil {
			sugar.Errorf("Failed to create user in repository: %v", err)
			c.JSON(http.StatusInternalServerError, RegisterResponse{
				Status:  "error",
				Message: "Failed to create user: " + err.Error(),
			})
			return
		}

		// Générer les tokens JWT
		accessToken, refreshToken, tokenID, accessExp, refreshExp, err := jwtService.GenerateTokenPair(
			user.ID, user.Email, user.Role,
		)
		if err != nil {
			sugar.Errorf("Failed to generate tokens: %v", err)
			c.JSON(http.StatusInternalServerError, RegisterResponse{
				Status:  "error",
				Message: "Failed to generate authentication tokens",
			})
			return
		}

		// Stocker le refresh token dans Redis
		if err := tokenRepo.StoreRefreshToken(ctx, user.ID, tokenID, refreshToken, refreshExp); err != nil {
			sugar.Errorf("Failed to store refresh token: %v", err)
			// On continue quand même, l'utilisateur est créé
		}

		sugar.Infof("New user registration: ID=%d prenom=%s nom=%s mail=%s",
			user.ID, req.Firstname, req.Lastname, req.Email)

		// Préparer les données utilisateur pour la réponse
		userData := gin.H{
			"id":         user.ID,
			"company":    user.Company,
			"firstname":  user.Firstname,
			"lastname":   user.Lastname,
			"email":      user.Email,
			"role":       user.Role,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt.Format(time.RFC3339),
		}

		// Réponse avec tokens JWT
		tokenResponse := TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    accessExp,
			User:         userData,
		}

		c.Header("Location", fmt.Sprintf("/api/v1/auth/users/%d", user.ID))
		c.JSON(http.StatusCreated, RegisterResponse{
			Status:  "success",
			Message: "User registered successfully",
			Data:    tokenResponse,
		})
	})

	// // Login endpoint
	// sugar.Info("Setting up /login endpoint...")
	// r.POST("/login", func(c *gin.Context) {
	// 	var req LoginRequest
	// 	if err := c.ShouldBindJSON(&req); err != nil {
	// 		c.JSON(http.StatusBadRequest, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Invalid request: " + err.Error(),
	// 		})
	// 		return
	// 	}

	// 	// Récupérer l'utilisateur par email
	// 	user, err := userRepo.GetByEmail(ctx, req.Email)
	// 	if err != nil {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Email ou mot de passe incorrect",
	// 		})
	// 		return
	// 	}

	// 	// Vérifier si l'utilisateur est actif
	// 	if !user.IsActive {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Compte utilisateur désactivé",
	// 		})
	// 		return
	// 	}

	// 	// Vérifier le mot de passe
	// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Email ou mot de passe incorrect",
	// 		})
	// 		return
	// 	}

	// 	// Mettre à jour la dernière connexion
	// 	if err := userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
	// 		sugar.Errorf("Failed to update last login: %v", err)
	// 	}

	// 	// Générer les tokens JWT
	// 	accessToken, refreshToken, tokenID, accessExp, refreshExp, err := jwtService.GenerateTokenPair(
	// 		user.ID, user.Email, user.Role,
	// 	)
	// 	if err != nil {
	// 		sugar.Errorf("Failed to generate tokens: %v", err)
	// 		c.JSON(http.StatusInternalServerError, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Failed to generate authentication tokens",
	// 		})
	// 		return
	// 	}

	// 	// Stocker le refresh token dans Redis
	// 	if err := tokenRepo.StoreRefreshToken(ctx, user.ID, tokenID, refreshToken, refreshExp); err != nil {
	// 		sugar.Errorf("Failed to store refresh token: %v", err)
	// 	}

	// 	sugar.Infof("User login: ID=%d email=%s", user.ID, user.Email)

	// 	// Préparer les données utilisateur pour la réponse
	// 	userData := gin.H{
	// 		"id":         user.ID,
	// 		"company":    user.Company,
	// 		"firstname":  user.Firstname,
	// 		"lastname":   user.Lastname,
	// 		"email":      user.Email,
	// 		"role":       user.Role,
	// 		"is_active":  user.IsActive,
	// 		"last_login": user.LastLogin,
	// 	}

	// 	// Réponse avec tokens JWT
	// 	tokenResponse := TokenResponse{
	// 		AccessToken:  accessToken,
	// 		RefreshToken: refreshToken,
	// 		TokenType:    "Bearer",
	// 		ExpiresIn:    accessExp,
	// 		User:         userData,
	// 	}

	// 	c.JSON(http.StatusOK, RegisterResponse{
	// 		Status:  "success",
	// 		Message: "Login successful",
	// 		Data:    tokenResponse,
	// 	})
	// })

	// // Refresh token endpoint
	// sugar.Info("Setting up /refresh endpoint...")
	// r.POST("/refresh", func(c *gin.Context) {
	// 	var req RefreshTokenRequest
	// 	if err := c.ShouldBindJSON(&req); err != nil {
	// 		c.JSON(http.StatusBadRequest, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Refresh token required",
	// 		})
	// 		return
	// 	}

	// 	// Valider le refresh token
	// 	claims, err := jwtService.ValidateRefreshToken(req.RefreshToken)
	// 	if err != nil {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Invalid refresh token: " + err.Error(),
	// 		})
	// 		return
	// 	}

	// 	// Vérifier si le token existe dans Redis
	// 	storedToken, err := tokenRepo.GetRefreshToken(ctx, claims.UserID, claims.TokenID)
	// 	if err != nil || storedToken != req.RefreshToken {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Refresh token not found or invalid",
	// 		})
	// 		return
	// 	}

	// 	// Récupérer l'utilisateur pour vérifier qu'il est toujours actif
	// 	user, err := userRepo.GetByID(ctx, claims.UserID)
	// 	if err != nil {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "User not found",
	// 		})
	// 		return
	// 	}

	// 	if !user.IsActive {
	// 		c.JSON(http.StatusUnauthorized, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "User account is disabled",
	// 		})
	// 		return
	// 	}

	// 	// Générer de nouveaux tokens
	// 	accessToken, newRefreshToken, tokenID, accessExp, refreshExp, err := jwtService.GenerateTokenPair(
	// 		user.ID, user.Email, user.Role,
	// 	)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, RegisterResponse{
	// 			Status:  "error",
	// 			Message: "Failed to generate new tokens",
	// 		})
	// 		return
	// 	}

	// 	// Supprimer l'ancien refresh token et stocker le nouveau
	// 	tokenRepo.DeleteRefreshToken(ctx, claims.UserID, claims.TokenID)
	// 	tokenRepo.StoreRefreshToken(ctx, user.ID, tokenID, newRefreshToken, refreshExp)

	// 	sugar.Infof("Token refreshed for user: ID=%d email=%s", user.ID, user.Email)

	// 	// Préparer les données utilisateur pour la réponse
	// 	userData := gin.H{
	// 		"id":         user.ID,
	// 		"company":    user.Company,
	// 		"firstname":  user.Firstname,
	// 		"lastname":   user.Lastname,
	// 		"email":      user.Email,
	// 		"role":       user.Role,
	// 		"is_active":  user.IsActive,
	// 	}

	// 	tokenResponse := TokenResponse{
	// 		AccessToken:  accessToken,
	// 		RefreshToken: newRefreshToken,
	// 		TokenType:    "Bearer",
	// 		ExpiresIn:    accessExp,
	// 		User:         userData,
	// 	}

	// 	c.JSON(http.StatusOK, RegisterResponse{
	// 		Status:  "success",
	// 		Message: "Tokens refreshed successfully",
	// 		Data:    tokenResponse,
	// 	})
	// })

	// // Endpoint de validation de token (pour la gateway)
	// sugar.Info("Setting up /validate endpoint...")
	// r.POST("/validate", func(c *gin.Context) {
	// 	// Extraire le token de l'en-tête Authorization
	// 	authHeader := c.GetHeader("Authorization")
	// 	if authHeader == "" {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Authorization header required",
	// 		})
	// 		return
	// 	}

	// 	// Vérifier le format "Bearer <token>"
	// 	parts := strings.SplitN(authHeader, " ", 2)
	// 	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Invalid authorization header format",
	// 		})
	// 		return
	// 	}

	// 	token := parts[1]

	// 	// Valider le token
	// 	claims, err := jwtService.ValidateAccessToken(token)
	// 	if err != nil {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Invalid token: " + err.Error(),
	// 		})
	// 		return
	// 	}

	// 	// Vérifier si le token n'est pas sur la liste noire
	// 	if tokenRepo.IsTokenBlacklisted(ctx, claims.ID) {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Token has been revoked",
	// 		})
	// 		return
	// 	}

	// 	// Token valide
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status": "success",
	// 		"data": gin.H{
	// 			"user_id":  claims.UserID,
	// 			"email":    claims.Email,
	// 			"role":     claims.Role,
	// 			"token_id": claims.TokenID,
	// 		},
	// 	})
	// })

	// // Logout endpoint
	// sugar.Info("Setting up /logout endpoint...")
	// r.POST("/logout", func(c *gin.Context) {
	// 	// Extraire le token de l'en-tête Authorization
	// 	authHeader := c.GetHeader("Authorization")
	// 	if authHeader == "" {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Authorization header required",
	// 		})
	// 		return
	// 	}

	// 	// Vérifier le format "Bearer <token>"
	// 	parts := strings.SplitN(authHeader, " ", 2)
	// 	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Invalid authorization header format",
	// 		})
	// 		return
	// 	}

	// 	token := parts[1]

	// 	// Valider le token pour extraire les informations
	// 	claims, err := jwtService.ValidateAccessToken(token)
	// 	if err != nil {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Invalid token: " + err.Error(),
	// 		})
	// 		return
	// 	}

	// 	// Supprimer le refresh token associé
	// 	if err := tokenRepo.DeleteRefreshToken(ctx, claims.UserID, claims.TokenID); err != nil {
	// 		sugar.Errorf("Failed to delete refresh token: %v", err)
	// 	}

	// 	// Ajouter l'access token à la liste noire
	// 	expiry, err := jwtService.GetTokenExpiry(token)
	// 	if err != nil {
	// 		sugar.Errorf("Failed to get token expiry: %v", err)
	// 	} else {
	// 		if err := tokenRepo.BlacklistToken(ctx, claims.ID, expiry); err != nil {
	// 			sugar.Errorf("Failed to blacklist token: %v", err)
	// 		}
	// 	}

	// 	sugar.Infof("User logout: ID=%d email=%s", claims.UserID, claims.Email)

	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status":  "success",
	// 		"message": "Logged out successfully",
	// 	})
	// })

	// // Profile endpoint
	// sugar.Info("Setting up /profile endpoint...")
	// r.GET("/profile", func(c *gin.Context) {
	// 	// Extraire le token de l'en-tête Authorization
	// 	authHeader := c.GetHeader("Authorization")
	// 	if authHeader == "" {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Authorization header required",
	// 		})
	// 		return
	// 	}

	// 	// Vérifier le format "Bearer <token>"
	// 	parts := strings.SplitN(authHeader, " ", 2)
	// 	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Invalid authorization header format",
	// 		})
	// 		return
	// 	}

	// 	token := parts[1]

	// 	// Valider le token
	// 	claims, err := jwtService.ValidateAccessToken(token)
	// 	if err != nil {
	// 		c.JSON(http.StatusUnauthorized, gin.H{
	// 			"status":  "error",
	// 			"message": "Invalid token: " + err.Error(),
	// 		})
	// 		return
	// 	}

	// 	// Récupérer l'utilisateur
	// 	user, err := userRepo.GetByID(ctx, claims.UserID)
	// 	if err != nil {
	// 		c.JSON(http.StatusNotFound, gin.H{
	// 			"status":  "error",
	// 			"message": "User not found",
	// 		})
	// 		return
	// 	}

	// 	// Préparer les données utilisateur pour la réponse
	// 	userData := gin.H{
	// 		"id":         user.ID,
	// 		"company":    user.Company,
	// 		"firstname":  user.Firstname,
	// 		"lastname":   user.Lastname,
	// 		"email":      user.Email,
	// 		"role":       user.Role,
	// 		"is_active":  user.IsActive,
	// 		"last_login": user.LastLogin,
	// 		"created_at": user.CreatedAt.Format(time.RFC3339),
	// 		"updated_at": user.UpdatedAt.Format(time.RFC3339),
	// 	}

	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status": "success",
	// 		"data":   userData,
	// 	})
	// })

	port := getEnv("AUTH_SERVICE_PORT", "8081")
	sugar.Infof("Auth-service started on port %s", port)

	if err := r.Run(":" + port); err != nil {
		sugar.Fatalf("Server startup error: %v", err)
	}
}

// getEnv récupère une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Remplacez cette partie dans votre main.go
func initRedis(redisURL string, sugar *zap.SugaredLogger) *redis.Client {
	// Si vous avez une URL Redis complète avec mot de passe
	if strings.HasPrefix(redisURL, "redis://") {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			sugar.Fatalf("Failed to parse Redis URL: %v", err)
		}
		return redis.NewClient(opt)
	}
	sugar.Info("Using manual Redis configuration from environment variables")
	sugar.Infof("Redis Addr: %s", getEnv("REDIS_ADDR", "localhost:6379"))
	sugar.Infof("Redis DB: %d", getIntEnv("REDIS_DB", 0))
	sugar.Infof("Redis Pool Size: %d", getIntEnv("REDIS_POOL_SIZE", 10))

	// Configuration manuelle avec mot de passe
	client := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""), // Mot de passe
		DB:       getIntEnv("REDIS_DB", 0),     // Base de données
		PoolSize: getIntEnv("REDIS_POOL_SIZE", 10),
	})

	return client
}

// Fonction helper pour les entiers
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
