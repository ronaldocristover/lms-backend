package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/lms/pkg/apierror"
	"github.com/yourusername/lms/pkg/response"
)

type UploadHandler struct {
	dir      string
	maxSize  int64
	allowed  map[string]bool
}

func NewUploadHandler(dir string, maxSize int64) *UploadHandler {
	// Allow common file types
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".txt":  true,
		".csv":  true,
	}
	return &UploadHandler{dir: dir, maxSize: maxSize, allowed: allowed}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, apierror.BadRequest("No file provided"))
		return
	}

	if file.Size > h.maxSize {
		response.Error(c, apierror.BadRequest(fmt.Sprintf("File size exceeds limit of %d bytes", h.maxSize)))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !h.allowed[ext] {
		response.Error(c, apierror.BadRequest("File type not allowed"))
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(h.dir, filename)

	if err := os.MkdirAll(h.dir, 0755); err != nil {
		response.Error(c, apierror.Internal("Failed to create upload directory"))
		return
	}

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		response.Error(c, apierror.Internal("Failed to save file"))
		return
	}

	response.Success(c, gin.H{
		"filename": filename,
		"url":      fmt.Sprintf("/uploads/%s", filename),
	})
}

func (h *UploadHandler) Serve(c *gin.Context) {
	filename := c.Param("filename")
	filepath := filepath.Join(h.dir, filename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		response.Error(c, apierror.NotFound("File not found"))
		return
	}

	c.File(filepath)
}
