package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-keycloak-app/go/models"

	"github.com/gin-gonic/gin"
)

// UpdateUser обновляет данные пользователя
func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var updateData struct {
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}

		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем существование пользователя
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM app_users WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Обновляем данные
		result, err := db.Exec(
			"UPDATE app_users SET email = $1, first_name = $2, last_name = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4",
			updateData.Email, updateData.FirstName, updateData.LastName, id,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

// DeleteUser удаляет пользователя
func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Проверяем существование пользователя
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM app_users WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Удаляем пользователя
		result, err := db.Exec("DELETE FROM app_users WHERE id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

// GetProfile получает профиль пользователя (с данными из user_profiles)
func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// В реальном приложении здесь будет извлечение ID пользователя из JWT токена или сессии
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		id, err := strconv.Atoi(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var profile struct {
			User    models.User `json:"user"`
			Profile struct {
				Bio       string         `json:"bio"`
				AvatarURL sql.NullString `json:"avatar_url"`
				Settings  sql.NullString `json:"settings"`
			} `json:"profile"`
		}

		// Получаем данные пользователя
		err = db.QueryRow(
			"SELECT id, keycloak_id, email, first_name, last_name, created_at FROM app_users WHERE id = $1",
			id,
		).Scan(&profile.User.ID, &profile.User.KeycloakID, &profile.User.Email, &profile.User.FirstName, &profile.User.LastName, &profile.User.CreatedAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Получаем данные профиля (если есть)
		err = db.QueryRow(
			"SELECT bio, avatar_url, settings FROM user_profiles WHERE user_id = $1",
			id,
		).Scan(&profile.Profile.Bio, &profile.Profile.AvatarURL, &profile.Profile.Settings)

		// Если профиль не найден - это не ошибка, просто вернем пустой профиль
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

// CreateOrUpdateProfile создает или обновляет профиль пользователя
func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var profileData struct {
			UserID    int    `json:"user_id" binding:"required"`
			Bio       string `json:"bio"`
			AvatarURL string `json:"avatar_url"`
			Settings  string `json:"settings"`
		}

		if err := c.ShouldBindJSON(&profileData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем существование пользователя
		var userExists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM app_users WHERE id = $1)", profileData.UserID).Scan(&userExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !userExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Проверяем существование профиля
		var profileExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_profiles WHERE user_id = $1)", profileData.UserID).Scan(&profileExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if profileExists {
			// Обновляем существующий профиль
			_, err = db.Exec(
				"UPDATE user_profiles SET bio = $1, avatar_url = $2, settings = $3 WHERE user_id = $4",
				profileData.Bio, profileData.AvatarURL, profileData.Settings, profileData.UserID,
			)
		} else {
			// Создаем новый профиль
			_, err = db.Exec(
				"INSERT INTO user_profiles (user_id, bio, avatar_url, settings) VALUES ($1, $2, $3, $4)",
				profileData.UserID, profileData.Bio, profileData.AvatarURL, profileData.Settings,
			)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		action := "updated"
		if !profileExists {
			action = "created"
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile " + action + " successfully"})
	}
}
