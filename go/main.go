package main

import (
	"log"
	"os"

	"go-keycloak-app/database"
	"go-keycloak-app/handlers"
	"go-keycloak-app/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Инициализация базы данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Инициализация Keycloak клиента
	authClient, err := handlers.NewKeycloakClient()
	if err != nil {
		log.Fatal("Failed to initialize Keycloak client:", err)
	}

	// Настройка роутера
	r := gin.Default()

	// Публичные маршруты
	r.POST("/auth/login", authClient.Login)
	r.POST("/auth/register", authClient.Register)
	r.POST("/auth/refresh", authClient.RefreshToken)

	// Защищенные маршруты
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authClient))
	{
		api.GET("/users", handlers.GetUsers(db))
		api.GET("/users/:id", handlers.GetUserByID(db))
		api.POST("/users", handlers.CreateUser(db))
		api.PUT("/users/:id", handlers.UpdateUser(db))
		api.DELETE("/users/:id", handlers.DeleteUser(db))
		api.GET("/profile", handlers.GetProfile)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "OK",
			"database": "Connected",
			"keycloak": "Available",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
