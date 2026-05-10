package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to retrieve users", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Users retrieved successfully", "data": users})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input service.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input format"})
		return
	}

	user, err := h.userService.CreateUser(input)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create user", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "User created successfully", "data": user})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var input service.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input format"})
		return
	}

	user, err := h.userService.UpdateUser(id, input)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update user", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User updated successfully", "data": user})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.userService.DeleteUser(id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete user", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully", "data": nil})
}
