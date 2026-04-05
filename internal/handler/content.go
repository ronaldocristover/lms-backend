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

type ContentHandler struct {
	contentSvc service.ContentService
}

func NewContentHandler(contentSvc service.ContentService) *ContentHandler {
	return &ContentHandler{contentSvc: contentSvc}
}

// Create creates new content
// @Summary Create content
// @Tags Contents
// @Accept json
// @Produce json
// @Param request body model.CreateContentRequest true "Content data"
// @Success 201 {object} response.SuccessResponse
// @Router /contents [post]
func (h *ContentHandler) Create(c *gin.Context) {
	var req model.CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	content, err := h.contentSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrSessionNotFound {
			response.Error(c, apierror.NotFound("Session not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create content"))
		return
	}

	response.Created(c, content)
}

// Get retrieves content by ID
// @Summary Get content
// @Tags Contents
// @Produce json
// @Param id path string true "Content ID"
// @Success 200 {object} response.SuccessResponse
// @Router /contents/{id} [get]
func (h *ContentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid content ID"))
		return
	}

	content, err := h.contentSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Content not found"))
		return
	}

	response.Success(c, content)
}

// Update updates content
// @Summary Update content
// @Tags Contents
// @Accept json
// @Produce json
// @Param id path string true "Content ID"
// @Param request body model.UpdateContentRequest true "Content data"
// @Success 200 {object} response.SuccessResponse
// @Router /contents/{id} [put]
func (h *ContentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid content ID"))
		return
	}

	var req model.UpdateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	content, err := h.contentSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrContentNotFound {
			response.Error(c, apierror.NotFound("Content not found"))
			return
		}
		if err == service.ErrSessionNotFound {
			response.Error(c, apierror.NotFound("Session not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update content"))
		return
	}

	response.Success(c, content)
}

// Delete deletes content
// @Summary Delete content
// @Tags Contents
// @Param id path string true "Content ID"
// @Success 204
// @Router /contents/{id} [delete]
func (h *ContentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid content ID"))
		return
	}

	if err := h.contentSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrContentNotFound {
			response.Error(c, apierror.NotFound("Content not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete content"))
		return
	}

	response.NoContent(c)
}

// List lists contents with pagination
// @Summary List contents
// @Tags Contents
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param session_id query string false "Filter by session ID"
// @Param type query string false "Filter by content type"
// @Success 200 {object} response.PaginatedResponse
// @Router /contents [get]
func (h *ContentHandler) List(c *gin.Context) {
	var req model.ListContentsRequest
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

	contents, total, err := h.contentSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch contents"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, contents, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
