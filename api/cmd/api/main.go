package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Configuration selon l'environnement
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes de base
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "immobilier-api",
			"version": "1.0.0",
		})
	})

	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "API Go fonctionne!",
			"framework": "Gin",
			"version":   "1.10.0",
		})
	})

	// Routes API basiques
	api := r.Group("/api")
	{
		api.GET("/properties", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"properties": []map[string]interface{}{
					{"id": 1, "title": "Appartement Centre", "price": 1200},
					{"id": 2, "title": "Maison Banlieue", "price": 1800},
				},
				"count": 2,
			})
		})

		api.GET("/tenants", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"tenants": []map[string]interface{}{
					{"id": 1, "name": "Jean Dupont", "email": "jean@example.com"},
					{"id": 2, "name": "Marie Martin", "email": "marie@example.com"},
				},
				"count": 2,
			})
		})

		api.GET("/contracts", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"contracts": []map[string]interface{}{
					{"id": 1, "property_id": 1, "tenant_id": 1, "status": "active"},
					{"id": 2, "property_id": 2, "tenant_id": 2, "status": "active"},
				},
				"count": 2,
			})
		})


	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Serveur API Immobilier démarré sur le port %s", port)
	log.Printf("🌍 Environment: %s", os.Getenv("ENV"))

	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Erreur lors du démarrage du serveur:", err)
	}
}
