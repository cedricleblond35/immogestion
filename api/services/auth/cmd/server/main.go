package main

import (
	"api/services/auth/internal/database"
	model "api/services/auth/internal/models"
	"api/services/auth/internal/repository"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents the expected request structure for user registration
type RegisterRequest struct {
	Company   string `json:"company" binding:"required"`
	Lastname  string `json:"lastname" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
}

// RegisterResponse represents the response structure for registration endpoint
type RegisterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func main() {
	// Logger avec Zap
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Info("")

	ctx := context.Background()

	// Database
	db, err := database.GormOpen(ctx, false, sugar)
	if err != nil {
		sugar.Errorf("Failed to connect to database: %v", err) // Utilisez sugar pour logger l'erreur !
		sugar.Fatal("Exiting due to DB error")                 // Fatal log + exit
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

	// Initialize Gin router
	sugar.Info("Setting up routes...")
	r := gin.Default()

	// Health check endpoint
	sugar.Info("Setting up /register endpoint...")
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// User registration endpoint
	sugar.Info("Setting up /register endpoint...")
	r.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid request: " + err.Error(),
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14) // Cost 14 = bon équilibre sécurité/vitesse
		if err != nil {
			sugar.Errorf("Failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Internal server error during password hashing",
			})
			return
		}

		user := &model.User{
			Company:   req.Company,
			Lastname:  req.Lastname,
			Firstname: req.Firstname,
			Email:     req.Email,
			Password:  string(hashedPassword),
			IsActive:  true,
		}
		userRepo := repository.NewUserRepository(db, sugar)
		if err := userRepo.Create(ctx, user); err != nil {
			sugar.Errorf("Failed to create user in repository: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		sugar.Infof("New user registration: ID=%s prenom=%s nom=%s mail=%s", user.ID, req.Firstname, req.Lastname, req.Email)

		// Here you could insert the user into database, hash the password, etc.
		// For testing purposes, we generate a mock user ID

		// Response to gateway
		c.Header("Location", fmt.Sprintf("/api/v1/auth/users/%d", user.ID))
		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "User registered successfully",
			"data": gin.H{
				"id":         user.ID,
				"company":    user.Company,
				"firstname":  user.Firstname,
				"lastname":   user.Lastname,
				"email":      user.Email,
				"role":       user.Role,
				"is_active":  user.IsActive,
				"created_at": user.CreatedAt.Format(time.RFC3339),
			},
		})
	})

	log.Println("Auth-service started on port 8081")
	if err := r.Run(":" + os.Getenv("AUTH_SERVICE_PORT") + ""); err != nil {
		log.Fatalf("Server startup error: %v", err)
	}
}
