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

type MediaHandler struct {
	mediaSvc service.MediaService
}

func NewMediaHandler(mediaSvc service.MediaService) *MediaHandler {
	return &MediaHandler{mediaSvc: mediaSvc}
}

func (h *MediaHandler) Create(c *gin.Context) {
	var req model.CreateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	media, err := h.mediaSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if apiErr, ok := err.(*apierror.Error); ok && apiErr.Code == 400 {
			response.Error(c, apiErr)
			return
		}
		response.Error(c, apierror.Internal("Failed to create media"))
		return
	}

	response.Created(c, media)
}

func (h *MediaHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid media ID"))
		return
	}

	media, err := h.mediaSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Media not found"))
		return
	}

	response.Success(c, media)
}

func (h *MediaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid media ID"))
		return
	}

	var req model.UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	media, err := h.mediaSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrMediaNotFound {
			response.Error(c, apierror.NotFound("Media not found"))
			return
		}
		if apiErr, ok := err.(*apierror.Error); ok && apiErr.Code == 400 {
			response.Error(c, apiErr)
			return
		}
		response.Error(c, apierror.Internal("Failed to update media"))
		return
	}

	response.Success(c, media)
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid media ID"))
		return
	}

	if err := h.mediaSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrMediaNotFound {
			response.Error(c, apierror.NotFound("Media not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete media"))
		return
	}

	response.NoContent(c)
}

func (h *MediaHandler) List(c *gin.Context) {
	var req model.ListMediaRequest
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

	medias, total, err := h.mediaSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch media"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, medias, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
