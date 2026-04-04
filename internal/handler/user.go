package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/service"
	"github.com/yourusername/lms/pkg/apierror"
	"github.com/yourusername/lms/pkg/pagination"
	"github.com/yourusername/lms/pkg/response"
)

type UserHandler struct {
	userSvc service.UserService
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// CreateUser godoc
// @Summary      Create user
// @Description  Create a new user with role and optional organization
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.CreateUserRequest  true  "User creation payload"
// @Success      201  {object}  response.SuccessResponse{data=model.User}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      409  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	user, err := h.userSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrUserExists {
			response.Error(c, apierror.Conflict("User already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create user"))
		return
	}

	response.Created(c, user)
}

// ListUsers godoc
// @Summary      List users
// @Description  Get paginated list of users with optional filters
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page           query  int     false  "Page number"       minimum(1)  default(1)
// @Param        page_size      query  int     false  "Items per page"    minimum(1)  maximum(100)  default(20)
// @Param        role_id        query  string  false  "Filter by role ID (UUID)"
// @Param        organization_id query string false  "Filter by organization ID (UUID)"
// @Param        search         query  string  false  "Search name/email"
// @Success      200  {object}  response.PaginatedResponse{data=[]model.User}
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req model.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	users, total, err := h.userSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch users"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, users, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Get a single user by their UUID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  response.SuccessResponse{data=model.User}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid user ID"))
		return
	}

	user, err := h.userSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("User not found"))
		return
	}

	response.Success(c, user)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update user profile (name, role, avatar, status)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                  true  "User ID (UUID)"
// @Param        request  body      model.UpdateUserRequest  true  "Update fields"
// @Success      200  {object}  response.SuccessResponse{data=model.User}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid user ID"))
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	user, err := h.userSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.Error(c, apierror.NotFound("User not found"))
			return
		}
		if err == service.ErrUserExists {
			response.Error(c, apierror.Conflict("User already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update user"))
		return
	}

	response.Success(c, user)
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Delete a user by ID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid user ID"))
		return
	}

	if err := h.userSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrUserNotFound {
			response.Error(c, apierror.NotFound("User not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete user"))
		return
	}

	response.NoContent(c)
}
