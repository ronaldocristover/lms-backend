package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ronaldocristover/lms-backend/pkg/apierror"
)

func setupResponseTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupResponseTest()

	Success(c, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"key":"value"`)
}

func TestCreated(t *testing.T) {
	c, w := setupResponseTest()

	Created(c, map[string]string{"id": "123"})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"id":"123"`)
}

func TestNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		NoContent(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestError(t *testing.T) {
	c, w := setupResponseTest()

	apiErr := apierror.BadRequest("test error")
	Error(c, apiErr)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"success":false`)
	assert.Contains(t, w.Body.String(), `"code":400`)
	assert.Contains(t, w.Body.String(), `"message":"test error"`)
}

func TestError_NotFound(t *testing.T) {
	c, w := setupResponseTest()

	apiErr := apierror.NotFound("not found")
	Error(c, apiErr)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestError_Internal(t *testing.T) {
	c, w := setupResponseTest()

	apiErr := apierror.Internal("internal error")
	Error(c, apiErr)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPaginated(t *testing.T) {
	c, w := setupResponseTest()

	items := []map[string]string{{"id": "1"}, {"id": "2"}}
	meta := &PaginationMeta{
		Page:       1,
		PageSize:   10,
		TotalItems: 2,
		TotalPages: 1,
	}

	Paginated(c, items, meta)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"page":1`)
	assert.Contains(t, w.Body.String(), `"page_size":10`)
	assert.Contains(t, w.Body.String(), `"total_items":2`)
}
