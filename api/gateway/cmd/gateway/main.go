package main

/**
API Gateway for Immogestion
- Handles CORS
- Proxies requests to auth-service for user registration
- Provides health check endpoint
- Implements rate limiting

Environment Variables:
- PORT: Port for the gateway (default: 8080)
- CORS_ORIGINS: Comma-separated list of allowed CORS origins (default: http://localhost:4200,http://localhost:4201)
- RATE_LIMIT_REQUESTS: Number of requests allowed (default: 100)
- RATE_LIMIT_DURATION: Duration for rate limiting (default: 1m)

Production Deployment:
- Use Docker and Kubernetes for deployment
- Ensure auth-service is reachable at http://auth-service:8081

- Monitor logs and performance
- Scale services as needed

PROD EXAMPLE:
export CORS_ORIGINS="https://monfrontend.com,https://admin.monfrontend.com"
go run gateway.go


*/

import (
	"io"
	"log"
	"net/http"
	"os"

	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	//"github.com/spf13/viper"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin" // Updated import for v3 middleware
	"github.com/ulule/limiter/v3/drivers/store/memory"        // v3 memory store
)

func main() {

	// Initialize Gin router with default middleware (logging, recovery)
	r := gin.Default()

	// Récupération des origines CORS depuis ENV
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		// Valeur par défaut si non défini
		corsOrigins = "http://localhost:4200,http://localhost:4201"
	}
	origins := strings.Split(corsOrigins, ",")

	log.Printf("🌐 CORS Origins autorisés: %v", origins)

	// Middleware CORS avec liste blanche
	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-Id", "Access-Control-Allow-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiting middleware
	// Configure rate limiting middleware using v3 API
	rate, err := limiter.NewRateFromFormatted(os.Getenv("RATE_LIMIT_REQUESTS") + "-" + os.Getenv("RATE_LIMIT_DURATION"))
	if err != nil {
		log.Fatalf("Failed to configure rate limiting: %v", err)
	}
	store := memory.NewStore() // v3 memory store
	limiter := limiter.New(store, rate)
	limiterMiddleware := mgin.NewMiddleware(limiter) // v3 middleware constructor
	r.Use(limiterMiddleware)

	// Health check endpoint for API Gateway
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Endpoint to handle user registration requests from Angular
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		// Create a new HTTP request to forward to auth-service
		req, err := http.NewRequest(http.MethodPost, "http://auth-service:"+os.Getenv("AUTH_SERVICE_PORT")+"/register", c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
			return
		}

		// Copy relevant headers from the incoming request
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))

		// Send the request to auth-service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy to auth-service: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		// Read the response body from auth-service
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response: " + err.Error()})
			return
		}

		// Forward the response back to the client (Angular)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	})

	// Endpoint to handle user login requests from Angular
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		// Create a new HTTP request to forward to auth-service
		req, err := http.NewRequest(http.MethodPost, "http://auth-service:"+os.Getenv("AUTH_SERVICE_PORT")+"/login", c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
			return
		}

		// Copy relevant headers from the incoming request
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))

		// Send the request to auth-service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy to auth-service: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		// Read the response body from auth-service
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response: " + err.Error()})
			return
		}

		// Forward the response back to the client (Angular)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	})

	// Endpoint to handle token refresh requests from Angular
	r.POST("/api/v1/auth/refresh", func(c *gin.Context) {
		// Create a new HTTP request to forward to auth-service
		req, err := http.NewRequest(http.MethodPost, "http://auth-service:"+os.Getenv("AUTH_SERVICE_PORT")+"/refresh", c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
			return
		}

		// Copy relevant headers from the incoming request
		req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
		req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))

		// Send the request to auth-service
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy to auth-service: " + err.Error()})
			return
		}
		defer resp.Body.Close()
		// Read the response body from auth-service

	})

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	err = r.Run(":" + port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
