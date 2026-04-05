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

type LanguageHandler struct {
	languageSvc service.LanguageService
}

func NewLanguageHandler(languageSvc service.LanguageService) *LanguageHandler {
	return &LanguageHandler{languageSvc: languageSvc}
}

func (h *LanguageHandler) Create(c *gin.Context) {
	var req model.CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	language, err := h.languageSvc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrLanguageExists {
			response.Error(c, apierror.Conflict("Language already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create language"))
		return
	}

	response.Created(c, language)
}

func (h *LanguageHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid language ID"))
		return
	}

	language, err := h.languageSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Language not found"))
		return
	}

	response.Success(c, language)
}

func (h *LanguageHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid language ID"))
		return
	}

	var req model.UpdateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	language, err := h.languageSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrLanguageNotFound {
			response.Error(c, apierror.NotFound("Language not found"))
			return
		}
		if err == service.ErrLanguageExists {
			response.Error(c, apierror.Conflict("Language already exists"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update language"))
		return
	}

	response.Success(c, language)
}

func (h *LanguageHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid language ID"))
		return
	}

	if err := h.languageSvc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrLanguageNotFound {
			response.Error(c, apierror.NotFound("Language not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete language"))
		return
	}

	response.NoContent(c)
}

func (h *LanguageHandler) List(c *gin.Context) {
	var req model.ListLanguagesRequest
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

	languages, total, err := h.languageSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch languages"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, languages, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
