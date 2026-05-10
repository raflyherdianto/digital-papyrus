package handler

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepteams/webp"
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
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		response.BadRequest(c, "Only .png, .jpg, .jpeg, .webp files are allowed", nil)
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
	allowedTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		response.BadRequest(c, "File is not a valid image", nil)
		return
	}

	// Reset file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.InternalError(c, "Failed to process file")
		return
	}

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		response.InternalError(c, "Failed to decode image: "+err.Error())
		return
	}

	uploadDir := filepath.Join("frontend", "public", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.InternalError(c, "Failed to create upload directory: "+err.Error())
		return
	}

	// Always save as .webp
	newFilename := fmt.Sprintf("%s.webp", uuid.New().String())
	dst := filepath.Join(uploadDir, newFilename)

	out, err := os.Create(dst)
	if err != nil {
		response.InternalError(c, "Failed to create destination file")
		return
	}
	defer out.Close()

	// Encode to WebP with 80% quality (balanced quality vs size)
	// Using pure Go deepteams/webp for better portability (no CGO required)
	if err := webp.Encode(out, img, webp.OptionsForPreset(webp.PresetDefault, 80.0)); err != nil {
		response.InternalError(c, "Failed to encode image to WebP: "+err.Error())
		return
	}

	// Return the relative URL starting from /uploads/...
	imageURL := fmt.Sprintf("/uploads/%s", newFilename)

	response.OK(c, "Image uploaded and optimized successfully", map[string]string{
		"url": imageURL,
	})
}
