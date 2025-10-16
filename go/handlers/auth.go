package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go-keycloak-app/database"
	"go-keycloak-app/models"

	"github.com/gin-gonic/gin"
)

type KeycloakClient struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
}

func NewKeycloakClient() (*KeycloakClient, error) {
	baseURL := os.Getenv("KEYCLOAK_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	realm := os.Getenv("KEYCLOAK_REALM")
	if realm == "" {
		realm = "master"
	}

	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	if clientID == "" {
		clientID = "go-app"
	}

	return &KeycloakClient{
		BaseURL:      baseURL,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	}, nil
}

func (kc *KeycloakClient) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := kc.getToken(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (kc *KeycloakClient) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создание пользователя в Keycloak
	userID, err := kc.createKeycloakUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сохранение пользователя в PostgreSQL
	db := database.GetDB()
	_, err = db.Exec(
		"INSERT INTO app_users (keycloak_id, email, first_name, last_name) VALUES ($1, $2, $3, $4)",
		userID, req.Email, req.FirstName, req.LastName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user in database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user_id": userID})
}

func (kc *KeycloakClient) RefreshToken(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	token, err := kc.refreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (kc *KeycloakClient) getToken(username, password string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.BaseURL, kc.Realm)

	data := strings.NewReader(fmt.Sprintf(
		"client_id=%s&username=%s&password=%s&grant_type=password",
		kc.ClientID, username, password,
	))

	if kc.ClientSecret != "" {
		data = strings.NewReader(fmt.Sprintf(
			"client_id=%s&client_secret=%s&username=%s&password=%s&grant_type=password",
			kc.ClientID, kc.ClientSecret, username, password,
		))
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var token models.TokenResponse
	json.Unmarshal(body, &token)

	return &token, nil
}

func (kc *KeycloakClient) createKeycloakUser(req models.CreateUserRequest) (string, error) {
	// Получаем admin token
	adminToken, err := kc.getAdminToken()
	if err != nil {
		return "", err
	}

	// Создаем пользователя в Keycloak
	userURL := fmt.Sprintf("%s/admin/realms/%s/users", kc.BaseURL, kc.Realm)
	userData := map[string]interface{}{
		"username":  req.Email,
		"email":     req.Email,
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"enabled":   true,
		"credentials": []map[string]string{
			{
				"type":  "password",
				"value": req.Password,
			},
		},
	}

	jsonData, _ := json.Marshal(userData)
	req, err := http.NewRequest("POST", userURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Извлекаем ID пользователя из Location header
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("failed to create user")
	}

	parts := strings.Split(location, "/")
	return parts[len(parts)-1], nil
}

func (kc *KeycloakClient) getAdminToken() (string, error) {
	// Упрощенная реализация - в продакшене используйте сервисный аккаунт
	token, err := kc.getToken("admin", "admin")
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (kc *KeycloakClient) refreshToken(refreshToken string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.BaseURL, kc.Realm)

	data := fmt.Sprintf(
		"client_id=%s&grant_type=refresh_token&refresh_token=%s",
		kc.ClientID, refreshToken,
	)

	if kc.ClientSecret != "" {
		data = fmt.Sprintf(
			"client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s",
			kc.ClientID, kc.ClientSecret, refreshToken,
		)
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var token models.TokenResponse
	json.Unmarshal(body, &token)

	return &token, nil
}
