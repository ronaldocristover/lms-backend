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

type CategoryHandler struct {
	categorySvc service.CategoryService
}

func NewCategoryHandler(categorySvc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categorySvc: categorySvc}
}

// Create creates a new category
// @Summary Create a category
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body model.CreateCategoryRequest true "Category data"
// @Success 201 {object} response.SuccessResponse
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	category, err := h.categorySvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrCategoryExists {
			response.Error(c, apierror.Conflict("Category already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create category"))
		return
	}

	response.Created(c, category)
}

// Get retrieves a category by ID
// @Summary Get a category
// @Tags Categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} response.SuccessResponse
// @Router /categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid category ID"))
		return
	}

	category, err := h.categorySvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Category not found"))
		return
	}

	response.Success(c, category)
}

// Update updates a category
// @Summary Update a category
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body model.UpdateCategoryRequest true "Category data"
// @Success 200 {object} response.SuccessResponse
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid category ID"))
		return
	}

	var req model.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	category, err := h.categorySvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrCategoryNotFound {
			response.Error(c, apierror.NotFound("Category not found"))
			return
		}
		if err == service.ErrCategoryExists {
			response.Error(c, apierror.Conflict("Category already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update category"))
		return
	}

	response.Success(c, category)
}

// Delete deletes a category
// @Summary Delete a category
// @Tags Categories
// @Param id path string true "Category ID"
// @Success 204
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid category ID"))
		return
	}

	if err := h.categorySvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrCategoryNotFound {
			response.Error(c, apierror.NotFound("Category not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete category"))
		return
	}

	response.NoContent(c)
}

// List lists categories with pagination
// @Summary List categories
// @Tags Categories
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Param search query string false "Search term"
// @Success 200 {object} response.PaginatedResponse
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	var req model.ListCategoriesRequest
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

	categories, total, err := h.categorySvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch categories"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, categories, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
