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

type RoleHandler struct {
	roleSvc service.RoleService
}

func NewRoleHandler(roleSvc service.RoleService) *RoleHandler {
	return &RoleHandler{roleSvc: roleSvc}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	role, err := h.roleSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrRoleExists {
			response.Error(c, apierror.Conflict("Role already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create role"))
		return
	}

	response.Created(c, role)
}

func (h *RoleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid role ID"))
		return
	}

	role, err := h.roleSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Role not found"))
		return
	}

	response.Success(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid role ID"))
		return
	}

	var req model.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	role, err := h.roleSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrRoleNotFound {
			response.Error(c, apierror.NotFound("Role not found"))
			return
		}
		if err == service.ErrRoleExists {
			response.Error(c, apierror.Conflict("Role already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update role"))
		return
	}

	response.Success(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid role ID"))
		return
	}

	if err := h.roleSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrRoleNotFound {
			response.Error(c, apierror.NotFound("Role not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete role"))
		return
	}

	response.NoContent(c)
}

func (h *RoleHandler) List(c *gin.Context) {
	var req model.ListRolesRequest
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

	roles, total, err := h.roleSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch roles"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, roles, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
