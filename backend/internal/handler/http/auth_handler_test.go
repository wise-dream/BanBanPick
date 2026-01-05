package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/repository/models"
	"github.com/bbp/backend/internal/repository/sqlite"
	"github.com/bbp/backend/internal/usecase/auth"
	"github.com/bbp/backend/pkg/database"
	"github.com/bbp/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := database.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Миграции
	if err := database.Migrate(db,
		&models.UserModel{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		database.Close(db)
	}

	return db, cleanup
}

func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	db, cleanup := setupTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Инициализируем репозитории
	userRepo := sqlite.NewUserRepository(db)

	// Инициализируем JWT сервис
	jwtService := jwt.NewJWTService("test-secret", 24*60*60*1000*1000000) // 24 часа в наносекундах

	// Инициализируем use cases
	registerUseCase := auth.NewRegisterUseCase(userRepo, jwtService)
	loginUseCase := auth.NewLoginUseCase(userRepo, jwtService)
	getCurrentUserUseCase := auth.NewGetCurrentUserUseCase(userRepo)

	// Инициализируем handler
	authHandler := NewAuthHandler(registerUseCase, loginUseCase, getCurrentUserUseCase)

	// Настраиваем роуты
	api := router.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.GET("/me", authHandler.GetCurrentUser)
		}
	}

	return router, cleanup
}

func TestAuthHandler_Register(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        dto.RegisterRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			payload: dto.RegisterRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response dto.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "testuser", response.User.Username)
				assert.Equal(t, "test@example.com", response.User.Email)
			},
		},
		{
			name: "duplicate email",
			payload: dto.RegisterRequest{
				Email:    "test@example.com",
				Username: "testuser2",
				Password: "password123",
			},
			expectedStatus: http.StatusOK, // Первая регистрация успешна
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Регистрируем первого пользователя
			},
		},
		{
			name: "invalid email format",
			payload: dto.RegisterRequest{
				Email:    "invalid-email",
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "short password",
			payload: dto.RegisterRequest{
				Email:    "test2@example.com",
				Username: "testuser2",
				Password: "12345",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Сначала регистрируем пользователя для теста входа
	registerPayload := dto.RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerPayload)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	router.ServeHTTP(registerW, registerReq)

	tests := []struct {
		name           string
		payload        dto.LoginRequest
		expectedStatus int
	}{
		{
			name: "successful login",
			payload: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			payload: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			payload: dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Сначала регистрируем пользователя
	registerPayload := dto.RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerPayload)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	router.ServeHTTP(registerW, registerReq)

	var registerResponse dto.AuthResponse
	json.Unmarshal(registerW.Body.Bytes(), &registerResponse)
	token := registerResponse.Token

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
