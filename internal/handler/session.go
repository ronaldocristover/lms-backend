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

type SessionHandler struct {
	sessionSvc service.SessionService
}

func NewSessionHandler(sessionSvc service.SessionService) *SessionHandler {
	return &SessionHandler{sessionSvc: sessionSvc}
}

// Create creates a new session
// @Summary Create a session
// @Tags Sessions
// @Accept json
// @Produce json
// @Param request body model.CreateSessionRequest true "Session data"
// @Success 201 {object} response.SuccessResponse
// @Router /sessions [post]
func (h *SessionHandler) Create(c *gin.Context) {
	var req model.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	session, err := h.sessionSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrSeriesNotFound {
			response.Error(c, apierror.NotFound("Series not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create session"))
		return
	}

	response.Created(c, session)
}

// Get retrieves a session by ID
// @Summary Get a session
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} response.SuccessResponse
// @Router /sessions/{id} [get]
func (h *SessionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid session ID"))
		return
	}

	session, err := h.sessionSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Session not found"))
		return
	}

	response.Success(c, session)
}

// Update updates a session
// @Summary Update a session
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body model.UpdateSessionRequest true "Session data"
// @Success 200 {object} response.SuccessResponse
// @Router /sessions/{id} [put]
func (h *SessionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid session ID"))
		return
	}

	var req model.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	session, err := h.sessionSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrSessionNotFound {
			response.Error(c, apierror.NotFound("Session not found"))
			return
		}
		if err == service.ErrSeriesNotFound {
			response.Error(c, apierror.NotFound("Series not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update session"))
		return
	}

	response.Success(c, session)
}

// Delete deletes a session
// @Summary Delete a session
// @Tags Sessions
// @Param id path string true "Session ID"
// @Success 204
// @Router /sessions/{id} [delete]
func (h *SessionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid session ID"))
		return
	}

	if err := h.sessionSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrSessionNotFound {
			response.Error(c, apierror.NotFound("Session not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete session"))
		return
	}

	response.NoContent(c)
}

// List lists sessions with pagination
// @Summary List sessions
// @Tags Sessions
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Search term"
// @Param series_id query string false "Filter by series ID"
// @Success 200 {object} response.PaginatedResponse
// @Router /sessions [get]
func (h *SessionHandler) List(c *gin.Context) {
	var req model.ListSessionsRequest
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

	sessions, total, err := h.sessionSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch sessions"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, sessions, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
