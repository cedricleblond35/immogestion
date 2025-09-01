package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// User registration endpoint
	r.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid request: " + err.Error(),
			})
			return
		}

		log.Printf("New user registration: %s %s (%s)", req.Firstname, req.Lastname, req.Email)

		// Here you could insert the user into database, hash the password, etc.
		// For testing purposes, we generate a mock user ID
		userID := "user_" + req.Lastname

		// Response to gateway
		c.JSON(http.StatusCreated, RegisterResponse{
			Status:  "success",
			Message: "User registered successfully",
			UserID:  userID,
		})
	})

	log.Println("Auth-service started on port 8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Server startup error: %v", err)
	}
}
