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

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      model.RegisterRequest  true  "Register request"
// @Success      200      {object}  response.SuccessResponse{data=model.LoginResponse}
// @Failure      400      {object}  response.ErrorResponse
// @Failure      409      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      model.LoginRequest  true  "Login request"
// @Success      200      {object}  response.SuccessResponse{data=model.LoginResponse}
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /auth/login [post]
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

// Me godoc
// @Summary      Get current user
// @Description  Get the authenticated user's profile
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.SuccessResponse{data=model.User}
// @Failure      401  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /auth/me [get]
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

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      model.RefreshTokenRequest  true  "Refresh token"
// @Success      200      {object}  response.SuccessResponse{data=model.LoginResponse}
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	resp, err := h.userSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidToken:
			response.Error(c, apierror.Unauthorized("Invalid or expired refresh token"))
		case service.ErrUserNotFound:
			response.Error(c, apierror.Unauthorized("User not found"))
		default:
			response.Error(c, apierror.Internal("Failed to refresh token"))
		}
		return
	}

	response.Success(c, resp)
}
