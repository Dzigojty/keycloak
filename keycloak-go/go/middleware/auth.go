package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go-keycloak-app/go/handlers"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(kc *handlers.KeycloakClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Валидация токена через Keycloak
		userInfo, err := validateToken(kc, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Сохраняем информацию о пользователе в контексте
		c.Set("user_id", userInfo["sub"])
		c.Set("user_email", userInfo["email"])
		c.Set("user_roles", userInfo["realm_access"])

		c.Next()
	}
}

func validateToken(kc *handlers.KeycloakClient, tokenString string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", kc.BaseURL, kc.Realm)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed")
	}

	var userInfo map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &userInfo)

	return userInfo, nil
}

func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("user_email")

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"user_email": userEmail,
		"message":    "This is protected data",
	})
}
