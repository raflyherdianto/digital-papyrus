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
	"time"

	"github.com/deepteams/webp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/pkg/response"
	"github.com/digitalpapyrus/backend/internal/service"
)

// UploadHandler handles file uploads (images & document drafts) locally.
type UploadHandler struct {
	userService  *service.UserService
	orderService *service.OrderService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(userService *service.UserService, orderService *service.OrderService) *UploadHandler {
	return &UploadHandler{
		userService:  userService,
		orderService: orderService,
	}
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
	if _, err := os.Stat(filepath.Join("..", "frontend", "public")); err == nil {
		uploadDir = filepath.Join("..", "frontend", "public", "uploads")
	}
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

// UploadDraft handles POST /api/v1/upload/draft
func (h *UploadHandler) UploadDraft(c *gin.Context) {
	// Limit size to 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 10 MB
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(c, "File too large. Max size is 10MB", nil)
		return
	}

	file, header, err := c.Request.FormFile("draft")
	if err != nil {
		response.BadRequest(c, "Draft file is required", nil)
		return
	}
	defer file.Close()

	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".doc" && ext != ".docx" {
		response.BadRequest(c, "Only .doc and .docx files are allowed", nil)
		return
	}

	// Sniff content type (first 512 bytes)
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		response.InternalError(c, "Failed to read file")
		return
	}

	contentType := http.DetectContentType(buffer)
	
	// Override content type based on extension because DetectContentType identifies .docx as application/zip
	if ext == ".docx" {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	} else if ext == ".doc" {
		contentType = "application/msword"
	}

	allowedTypes := map[string]bool{
		"application/zip":          true,
		"application/msword":       true,
		"application/octet-stream": true,
		"application/x-ole-storage": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}
	if !allowedTypes[contentType] {
		response.BadRequest(c, "File is not a valid document format (.doc/.docx)", nil)
		return
	}

	// Reset file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.InternalError(c, "Failed to process file")
		return
	}

	orderID := c.Request.FormValue("order_id")
	invoiceNumber := ""
	if orderID != "" {
		if order, err := h.orderService.GetOrderByID(orderID); err == nil && order != nil {
			invoiceNumber = order.Invoice
		}
	}

	// Local draft upload directory
	uploadDir := filepath.Join("frontend", "public", "uploads", "drafts")
	if _, err := os.Stat(filepath.Join("..", "frontend", "public")); err == nil {
		uploadDir = filepath.Join("..", "frontend", "public", "uploads", "drafts")
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.InternalError(c, "Failed to create draft upload directory: "+err.Error())
		return
	}

	// Create safe unique filename
	cleanOriginalName := strings.ReplaceAll(header.Filename, " ", "_")
	var newFilename string
	if invoiceNumber != "" {
		newFilename = fmt.Sprintf("%s_%d_%s", invoiceNumber, time.Now().Unix(), cleanOriginalName)
	} else {
		newFilename = fmt.Sprintf("%s_%d_%s", uuid.New().String()[:8], time.Now().Unix(), cleanOriginalName)
	}

	dst := filepath.Join(uploadDir, newFilename)
	out, err := os.Create(dst)
	if err != nil {
		response.InternalError(c, "Failed to save draft file locally: "+err.Error())
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		response.InternalError(c, "Failed to write draft file: "+err.Error())
		return
	}

	draftURL := fmt.Sprintf("/uploads/drafts/%s", newFilename)

	response.OK(c, "Draft uploaded successfully to local storage", map[string]string{
		"url":      draftURL,
		"filename": header.Filename,
	})
}
