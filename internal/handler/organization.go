package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/service"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
	"github.com/ronaldocristover/lms-backend/pkg/pagination"
	"github.com/ronaldocristover/lms-backend/pkg/response"
)

type OrganizationHandler struct {
	orgSvc service.OrganizationService
}

func NewOrganizationHandler(orgSvc service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgSvc: orgSvc}
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	var req model.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	org, err := h.orgSvc.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, org)
}

func (h *OrganizationHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	org, err := h.orgSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, org)
}

func (h *OrganizationHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	var req model.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	org, err := h.orgSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, org)
}

func (h *OrganizationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	if err := h.orgSvc.Delete(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *OrganizationHandler) List(c *gin.Context) {
	var req model.ListOrganizationsRequest
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

	orgs, total, err := h.orgSvc.List(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, orgs, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}

func (h *OrganizationHandler) AddUser(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	var req model.AddOrgUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	orgUser, err := h.orgSvc.AddUser(c.Request.Context(), orgID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, orgUser)
}

func (h *OrganizationHandler) UpdateUserRole(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	orgUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid user ID"))
		return
	}

	var req model.UpdateOrgUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	orgUser, err := h.orgSvc.UpdateUserRole(c.Request.Context(), orgID, orgUserID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, orgUser)
}

func (h *OrganizationHandler) RemoveUser(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	orgUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid user ID"))
		return
	}

	if err := h.orgSvc.RemoveUser(c.Request.Context(), orgID, orgUserID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *OrganizationHandler) ListUsers(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid organization ID"))
		return
	}

	var req model.ListOrgUsersRequest
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

	users, total, err := h.orgSvc.ListUsers(c.Request.Context(), orgID, &req)
	if err != nil {
		h.handleError(c, err)
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

func (h *OrganizationHandler) handleError(c *gin.Context, err error) {
	switch err {
	case service.ErrOrganizationNotFound:
		response.Error(c, apierror.NotFound("Organization not found"))
	case service.ErrOrganizationExists:
		response.Error(c, apierror.Conflict("Organization with this name already exists"))
	case service.ErrUserAlreadyInOrg:
		response.Error(c, apierror.Conflict("User already in organization"))
	case service.ErrUserNotInOrg:
		response.Error(c, apierror.NotFound("User not in organization"))
	case service.ErrCannotRemoveOwner:
		response.Error(c, apierror.BadRequest("Cannot remove organization owner"))
	case service.ErrInvalidUserID, service.ErrInvalidOrganizationID:
		response.Error(c, apierror.BadRequest(err.Error()))
	case service.ErrOwnerNotFound:
		response.Error(c, apierror.NotFound("Owner not found"))
	case service.ErrUserNotFound:
		response.Error(c, apierror.NotFound("User not found"))
	default:
		response.Error(c, apierror.Internal("An unexpected error occurred"))
	}
}
