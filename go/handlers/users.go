package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-keycloak-app/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, keycloak_id, email, first_name, last_name, created_at FROM app_users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var user models.User
			err := rows.Scan(&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			users = append(users, user)
		}

		c.JSON(http.StatusOK, users)
	}
}

func GetUserByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var user models.User
		err = db.QueryRow(
			"SELECT id, keycloak_id, email, first_name, last_name, created_at FROM app_users WHERE id = $1",
			id,
		).Scan(&user.ID, &user.KeycloakID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user struct {
			Email     string `json:"email" binding:"required"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id int
		err := db.QueryRow(
			"INSERT INTO app_users (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING id",
			user.Email, user.FirstName, user.LastName,
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "User created successfully"})
	}
}
