package http

import (
	"net/http"

	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/middleware"
	"github.com/bbp/backend/internal/usecase/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	registerUseCase      *auth.RegisterUseCase
	loginUseCase         *auth.LoginUseCase
	getCurrentUserUseCase *auth.GetCurrentUserUseCase
}

func NewAuthHandler(
	registerUseCase *auth.RegisterUseCase,
	loginUseCase *auth.LoginUseCase,
	getCurrentUserUseCase *auth.GetCurrentUserUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUseCase:       registerUseCase,
		loginUseCase:          loginUseCase,
		getCurrentUserUseCase: getCurrentUserUseCase,
	}
}

// Register обрабатывает POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.registerUseCase.Execute(auth.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		switch err {
		case auth.ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case auth.ErrUsernameAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: result.Token,
		User:  dto.ToUserResponse(result.User),
	})
}

// Login обрабатывает POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.loginUseCase.Execute(auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: result.Token,
		User: dto.UserResponse{
			ID:        result.User.ID,
			Email:     result.User.Email,
			Username:  result.User.Username,
			CreatedAt: result.User.CreatedAt,
		},
	})
}

// GetCurrentUser обрабатывает GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.getCurrentUserUseCase.Execute(user.ID)
	if err != nil {
		switch err {
		case auth.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:        result.User.ID,
		Email:     result.User.Email,
		Username:  result.User.Username,
		CreatedAt: result.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}