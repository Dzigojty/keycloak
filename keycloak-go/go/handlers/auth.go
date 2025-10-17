package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go-keycloak-app/go/database"
	"go-keycloak-app/go/models"

	"github.com/gin-gonic/gin"
)

type KeycloakClient struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
}

func NewKeycloakClient() *KeycloakClient {
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
	}
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
	var refreshReq struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := kc.refreshToken(refreshReq.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (kc *KeycloakClient) getToken(username, password string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.BaseURL, kc.Realm)

	var data *strings.Reader
	if kc.ClientSecret != "" {
		data = strings.NewReader(fmt.Sprintf(
			"client_id=%s&client_secret=%s&username=%s&password=%s&grant_type=password",
			kc.ClientID, kc.ClientSecret, username, password,
		))
	} else {
		data = strings.NewReader(fmt.Sprintf(
			"client_id=%s&username=%s&password=%s&grant_type=password",
			kc.ClientID, username, password,
		))
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("keycloak error: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token models.TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (kc *KeycloakClient) createKeycloakUser(req models.CreateUserRequest) (string, error) {
	// Получаем admin token
	adminToken, err := kc.getAdminToken()
	if err != nil {
		return "", fmt.Errorf("failed to get admin token: %v", err)
	}

	// Создаем пользователя в Keycloak
	userURL := fmt.Sprintf("%s/admin/realms/%s/users", kc.BaseURL, kc.Realm)
	userData := map[string]interface{}{
		"username":  req.Email,
		"email":     req.Email,
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"enabled":   true,
	}

	// Добавляем пароль только если он предоставлен
	if req.Password != "" {
		userData["credentials"] = []map[string]interface{}{
			{
				"type":      "password",
				"value":     req.Password,
				"temporary": false,
			},
		}
	}

	jsonData, err := json.Marshal(userData)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", userURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+adminToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create user in Keycloak: %s", string(body))
	}

	// Извлекаем ID пользователя из Location header
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("failed to get user location from response")
	}

	parts := strings.Split(location, "/")
	userID := parts[len(parts)-1]

	return userID, nil
}

func (kc *KeycloakClient) getAdminToken() (string, error) {
	// В реальном приложении используйте сервисный аккаунт или клиент credentials grant
	adminUsername := os.Getenv("KEYCLOAK_ADMIN_USERNAME")
	adminPassword := os.Getenv("KEYCLOAK_ADMIN_PASSWORD")

	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin"
	}

	token, err := kc.getToken(adminUsername, adminPassword)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (kc *KeycloakClient) refreshToken(refreshToken string) (*models.TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.BaseURL, kc.Realm)

	var data string
	if kc.ClientSecret != "" {
		data = fmt.Sprintf(
			"client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s",
			kc.ClientID, kc.ClientSecret, refreshToken,
		)
	} else {
		data = fmt.Sprintf(
			"client_id=%s&grant_type=refresh_token&refresh_token=%s",
			kc.ClientID, refreshToken,
		)
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("keycloak error: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token models.TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// GetUserInfo получает информацию о пользователе из Keycloak
func (kc *KeycloakClient) GetUserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
		return
	}

	userInfo, err := kc.getUserInfo(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

func (kc *KeycloakClient) getUserInfo(accessToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", kc.BaseURL, kc.Realm)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
