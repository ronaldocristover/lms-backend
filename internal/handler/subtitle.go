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

type SubtitleHandler struct {
	subtitleSvc service.SubtitleService
}

func NewSubtitleHandler(subtitleSvc service.SubtitleService) *SubtitleHandler {
	return &SubtitleHandler{subtitleSvc: subtitleSvc}
}

func (h *SubtitleHandler) Create(c *gin.Context) {
	var req model.CreateSubtitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	subtitle, err := h.subtitleSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrSubtitleExists {
			response.Error(c, apierror.Conflict("Subtitle already exists for this media and language"))
			return
		}
		if apiErr, ok := err.(*apierror.Error); ok && apiErr.Code == 400 {
			response.Error(c, apiErr)
			return
		}
		response.Error(c, apierror.Internal("Failed to create subtitle"))
		return
	}

	response.Created(c, subtitle)
}

func (h *SubtitleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid subtitle ID"))
		return
	}

	subtitle, err := h.subtitleSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Subtitle not found"))
		return
	}

	response.Success(c, subtitle)
}

func (h *SubtitleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid subtitle ID"))
		return
	}

	var req model.UpdateSubtitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	subtitle, err := h.subtitleSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrSubtitleNotFound {
			response.Error(c, apierror.NotFound("Subtitle not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update subtitle"))
		return
	}

	response.Success(c, subtitle)
}

func (h *SubtitleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid subtitle ID"))
		return
	}

	if err := h.subtitleSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrSubtitleNotFound {
			response.Error(c, apierror.NotFound("Subtitle not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete subtitle"))
		return
	}

	response.NoContent(c)
}

func (h *SubtitleHandler) List(c *gin.Context) {
	var req model.ListSubtitlesRequest
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

	subtitles, total, err := h.subtitleSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch subtitles"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, subtitles, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
