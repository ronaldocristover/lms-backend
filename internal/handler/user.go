package handler

import (
	"strconv"

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

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userSvc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch users"))
		return
	}

	meta := pagination.NewMeta(page, pageSize, total)
	response.Paginated(c, users, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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
		response.Error(c, apierror.Internal("Failed to update user"))
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid user ID"))
		return
	}

	if err := h.userSvc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, apierror.Internal("Failed to delete user"))
		return
	}

	response.Success(c, nil)
}
