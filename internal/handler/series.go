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

type SeriesHandler struct {
	seriesSvc service.SeriesService
}

func NewSeriesHandler(seriesSvc service.SeriesService) *SeriesHandler {
	return &SeriesHandler{seriesSvc: seriesSvc}
}

// Create creates a new series
// @Summary Create a series
// @Tags Series
// @Accept json
// @Produce json
// @Param request body model.CreateSeriesRequest true "Series data"
// @Success 201 {object} response.SuccessResponse
// @Router /series [post]
func (h *SeriesHandler) Create(c *gin.Context) {
	var req model.CreateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	series, err := h.seriesSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrCategoryNotFound {
			response.Error(c, apierror.NotFound("Category not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create series"))
		return
	}

	response.Created(c, series)
}

// Get retrieves a series by ID
// @Summary Get a series
// @Tags Series
// @Produce json
// @Param id path string true "Series ID"
// @Success 200 {object} response.SuccessResponse
// @Router /series/{id} [get]
func (h *SeriesHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid series ID"))
		return
	}

	series, err := h.seriesSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Series not found"))
		return
	}

	response.Success(c, series)
}

// Update updates a series
// @Summary Update a series
// @Tags Series
// @Accept json
// @Produce json
// @Param id path string true "Series ID"
// @Param request body model.UpdateSeriesRequest true "Series data"
// @Success 200 {object} response.SuccessResponse
// @Router /series/{id} [put]
func (h *SeriesHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid series ID"))
		return
	}

	var req model.UpdateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	series, err := h.seriesSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrSeriesNotFound {
			response.Error(c, apierror.NotFound("Series not found"))
			return
		}
		if err == service.ErrCategoryNotFound {
			response.Error(c, apierror.NotFound("Category not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update series"))
		return
	}

	response.Success(c, series)
}

// Delete deletes a series
// @Summary Delete a series
// @Tags Series
// @Param id path string true "Series ID"
// @Success 204
// @Router /series/{id} [delete]
func (h *SeriesHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid series ID"))
		return
	}

	if err := h.seriesSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrSeriesNotFound {
			response.Error(c, apierror.NotFound("Series not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete series"))
		return
	}

	response.NoContent(c)
}

// List lists series with pagination
// @Summary List series
// @Tags Series
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Search term"
// @Param category_id query string false "Filter by category ID"
// @Param is_paid query bool false "Filter by paid status"
// @Success 200 {object} response.PaginatedResponse
// @Router /series [get]
func (h *SeriesHandler) List(c *gin.Context) {
	var req model.ListSeriesRequest
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

	series, total, err := h.seriesSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch series"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, series, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
