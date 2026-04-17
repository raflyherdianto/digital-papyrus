package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/pkg/response"
)

// UploadHandler handles file uploads.
type UploadHandler struct{}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadImage handles POST /api/v1/upload
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Parse max 2MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20) // 2 MB
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		response.BadRequest(c, "File too large. Max size is 2MB", nil)
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "Image file is required", nil)
		return
	}
	defer file.Close()

	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		response.BadRequest(c, "Only .png, .jpg, .jpeg files are allowed", nil)
		return
	}

	// Read first 512 bytes to sniff content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		response.InternalError(c, "Failed to read file")
		return
	}

	contentType := http.DetectContentType(buffer)
	if contentType != "image/png" && contentType != "image/jpeg" {
		response.BadRequest(c, "File is not a valid image", nil)
		return
	}

	// Reset file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.InternalError(c, "Failed to process file")
		return
	}

	uploadDir := filepath.Join("frontend", "public", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.InternalError(c, "Failed to create upload directory: " + err.Error())
		return
	}

	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(uploadDir, newFilename)

	out, err := os.Create(dst)
	if err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}

	// Return the relative URL starting from /uploads/... (since it's in public/uploads for Vite/Astro)
	imageURL := fmt.Sprintf("/uploads/%s", newFilename)

	response.OK(c, "Image uploaded successfully", map[string]string{
		"url": imageURL,
	})
}
