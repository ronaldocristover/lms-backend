package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupUploadTest() (*UploadHandler, *gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	dir, _ := os.MkdirTemp("", "uploads")
	handler := NewUploadHandler(dir, 10*1024*1024)
	r := gin.New()
	r.POST("/upload", handler.Upload)
	r.GET("/uploads/:filename", handler.Serve)
	w := httptest.NewRecorder()
	return handler, r, w
}

func createMultipartRequest(fieldName string, fileContent []byte, filename string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(fieldName, filename)
	part.Write(fileContent)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadHandler_Upload_Success(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	content := []byte("test file content")
	req := createMultipartRequest("file", content, "test.txt")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"filename"`)
	assert.Contains(t, w.Body.String(), `txt`)
}

func TestUploadHandler_Upload_NoFile(t *testing.T) {
	_, r, w := setupUploadTest()

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"No file provided"`)
}

func TestUploadHandler_Upload_FileTooLarge(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	handler.maxSize = 100
	content := make([]byte, 200)
	for i := range content {
		content[i] = 'a'
	}
	req := createMultipartRequest("file", content, "test.txt")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "File size exceeds limit")
}

func TestUploadHandler_Upload_InvalidFileType(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	content := []byte("test content")
	req := createMultipartRequest("file", content, "test.exe")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"File type not allowed"`)
}

func TestUploadHandler_Upload_AllowedFileTypes(t *testing.T) {
	handler, r, _ := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".txt", ".csv"}

	for _, ext := range allowedTypes {
		w := httptest.NewRecorder()
		content := []byte("test content")
		req := createMultipartRequest("file", content, "test"+ext)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Failed for extension: %s", ext)
	}
}

func TestUploadHandler_Serve_Success(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	testFile := filepath.Join(handler.dir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	req := httptest.NewRequest(http.MethodGet, "/uploads/test.txt", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUploadHandler_Serve_NotFound(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	req := httptest.NewRequest(http.MethodGet, "/uploads/nonexistent.txt", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"File not found"`)
}

func TestUploadHandler_NewUploadHandler(t *testing.T) {
	handler := NewUploadHandler("/tmp/uploads", 5*1024*1024)

	assert.Equal(t, "/tmp/uploads", handler.dir)
	assert.Equal(t, int64(5*1024*1024), handler.maxSize)
	assert.True(t, handler.allowed[".jpg"])
	assert.True(t, handler.allowed[".pdf"])
	assert.True(t, handler.allowed[".txt"])
	assert.False(t, handler.allowed[".exe"])
}

func TestUploadHandler_Upload_CaseInsensitiveExtension(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	content := []byte("test content")
	req := createMultipartRequest("file", content, "test.TXT")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUploadHandler_Upload_ActuallySavesFile(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	content := []byte("test file content for saving")
	req := createMultipartRequest("file", content, "test.txt")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	files, _ := os.ReadDir(handler.dir)
	assert.Equal(t, 1, len(files))
	assert.True(t, files[0].Name() != "")

	savedContent, _ := os.ReadFile(filepath.Join(handler.dir, files[0].Name()))
	assert.Equal(t, content, savedContent)
}

func TestUploadHandler_Serve_DirectoryTraversal(t *testing.T) {
	handler, r, w := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	req := httptest.NewRequest(http.MethodGet, "/uploads/../../../etc/passwd", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUploadHandler_Upload_MultipleFiles(t *testing.T) {
	handler, r, _ := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		content := []byte("test content")
		req := createMultipartRequest("file", content, "test.txt")

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}

	files, _ := os.ReadDir(handler.dir)
	assert.Equal(t, 3, len(files))
}

func TestUploadHandler_Upload_Serve_Download(t *testing.T) {
	handler, r, _ := setupUploadTest()
	defer os.RemoveAll(handler.dir)

	content := []byte("original content")
	uploadW := httptest.NewRecorder()
	req := createMultipartRequest("file", content, "test.txt")
	r.ServeHTTP(uploadW, req)
	assert.Equal(t, http.StatusOK, uploadW.Code)

	files, _ := os.ReadDir(handler.dir)
	filename := files[0].Name()

	serveW := httptest.NewRecorder()
	serveReq := httptest.NewRequest(http.MethodGet, "/uploads/"+filename, nil)
	r.ServeHTTP(serveW, serveReq)

	assert.Equal(t, http.StatusOK, serveW.Code)
	body, _ := io.ReadAll(serveW.Body)
	assert.Equal(t, content, body)
}
