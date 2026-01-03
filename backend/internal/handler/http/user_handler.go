package http

import (
	"net/http"

	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/middleware"
	"github.com/bbp/backend/internal/usecase/user"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	getProfileUseCase    *user.GetProfileUseCase
	updateProfileUseCase *user.UpdateProfileUseCase
	getSessionsUseCase   *user.GetSessionsUseCase
	getRoomsUseCase      *user.GetRoomsUseCase
}

func NewUserHandler(
	getProfileUseCase *user.GetProfileUseCase,
	updateProfileUseCase *user.UpdateProfileUseCase,
	getSessionsUseCase *user.GetSessionsUseCase,
	getRoomsUseCase *user.GetRoomsUseCase,
) *UserHandler {
	return &UserHandler{
		getProfileUseCase:    getProfileUseCase,
		updateProfileUseCase: updateProfileUseCase,
		getSessionsUseCase:   getSessionsUseCase,
		getRoomsUseCase:      getRoomsUseCase,
	}
}

// GetProfile обрабатывает GET /api/users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userCtx, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.getProfileUseCase.Execute(userCtx.ID)
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
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

// UpdateProfile обрабатывает PUT /api/users/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userCtx, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.updateProfileUseCase.Execute(user.UpdateProfileInput{
		UserID:   userCtx.ID,
		Email:    req.Email,
		Username: req.Username,
	})
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case user.ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case user.ErrUsernameAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
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

// GetSessions обрабатывает GET /api/users/sessions
func (h *UserHandler) GetSessions(c *gin.Context) {
	userCtx, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.getSessionsUseCase.Execute(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.UserSessionsResponse{
		Sessions: dto.ToVetoSessionResponseList(result.Sessions),
	})
}

// GetRooms обрабатывает GET /api/users/rooms
func (h *UserHandler) GetRooms(c *gin.Context) {
	userCtx, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.getRoomsUseCase.Execute(userCtx.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.UserRoomsResponse{
		Rooms: dto.ToRoomResponseList(result.Rooms),
	})
}