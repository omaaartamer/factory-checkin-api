package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omaaartamer/factory-checkin-api/internal/model"
	"github.com/omaaartamer/factory-checkin-api/internal/service"
)

type Handler struct {
	checkinService *service.CheckinService
}

func NewHandler(checkinService *service.CheckinService) *Handler {
	return &Handler{
		checkinService: checkinService,
	}
}

func (h *Handler) SetupRoutes() *gin.Engine {
	// Set gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(h.corsMiddleware())

	// Health check endpoint
	router.GET("/health", h.healthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/checkin", h.checkin)
		api.GET("/employee/:id/status", h.getEmployeeStatus)
		api.GET("/queue/status", h.getQueueStatus)
	}

	return router
}

func (h *Handler) checkin(c *gin.Context) {
	var req model.CheckinRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate employee ID
	if strings.TrimSpace(req.EmployeeID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Employee ID is required",
		})
		return
	}

	// Process the checkin/checkout
	response, err := h.checkinService.ProcessCheckin(req.EmployeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process checkin",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getEmployeeStatus(c *gin.Context) {
	employeeID := c.Param("id")

	if strings.TrimSpace(employeeID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Employee ID is required",
		})
		return
	}

	session, err := h.checkinService.GetEmployeeStatus(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get employee status",
			"details": err.Error(),
		})
		return
	}

	if session == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  "not_checked_in",
			"message": "Employee is not currently checked in",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  "checked_in",
		"session": session,
	})
}

func (h *Handler) getQueueStatus(c *gin.Context) {
	status := h.checkinService.GetQueueStatus()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   status,
	})
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Factory Check-in API is running",
		"version": "1.0.0",
	})
}

func (h *Handler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
