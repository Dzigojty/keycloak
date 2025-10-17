package main

import (
	"go-keycloak-app/go/database"
	"go-keycloak-app/go/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализация базы данных
	if _, err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	keycloakClient := handlers.NewKeycloakClient()

	router := gin.Default()

	// Keycloak routes
	router.POST("/auth/login", keycloakClient.Login)
	router.POST("/auth/register", keycloakClient.Register)
	router.POST("/auth/refresh", keycloakClient.RefreshToken)
	router.GET("/auth/userinfo", keycloakClient.GetUserInfo)

	// User management routes
	api := router.Group("/api")
	{
		api.GET("/users", handlers.GetUsers(database.GetDB()))
		api.POST("/users", handlers.CreateUser(database.GetDB()))
		api.PUT("/users/:id", handlers.UpdateUser(database.GetDB()))
		api.DELETE("/users/:id", handlers.DeleteUser(database.GetDB()))
	}

	log.Println("Server starting on :8080")
	router.Run(":8080")
}
