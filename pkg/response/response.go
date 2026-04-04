package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/lms/pkg/apierror"
)

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   interface{} `json:"error"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Meta    *PaginationMeta `json:"meta"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, err *apierror.Error) {
	c.JSON(err.Code, ErrorResponse{
		Success: false,
		Error:   err,
	})
}

func Paginated(c *gin.Context, items interface{}, meta *PaginationMeta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    items,
		Meta:    meta,
	})
}
