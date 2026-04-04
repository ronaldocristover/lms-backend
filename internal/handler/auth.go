package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/service"
	"github.com/yourusername/lms/pkg/apierror"
	"github.com/yourusername/lms/pkg/response"
)

type AuthHandler struct {
	userSvc service.UserService
}

func NewAuthHandler(userSvc service.UserService) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	resp, err := h.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			response.Error(c, apierror.Conflict("User with this email already exists"))
		default:
			response.Error(c, apierror.Internal("Failed to register user"))
		}
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	resp, err := h.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			response.Error(c, apierror.Unauthorized("Invalid email or password"))
		default:
			response.Error(c, apierror.Internal("Failed to login"))
		}
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Error(c, apierror.Unauthorized("User not authenticated"))
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Error(c, apierror.Unauthorized("Invalid user ID"))
		return
	}

	user, err := h.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, apierror.NotFound("User not found"))
		return
	}

	response.Success(c, user)
}
