package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/francknouama/image-recognition-webapp/internal/services"
	"github.com/francknouama/image-recognition-webapp/web/templates"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Config holds handler configuration
type Config struct {
	ImageService      *services.ImageService
	PredictionService services.PredictionServiceInterface
	ModelService      *services.ModelService
	RateLimiter      *rate.Limiter
}

// Handler contains all HTTP handlers
type Handler struct {
	imageService      *services.ImageService
	predictionService services.PredictionServiceInterface
	modelService      *services.ModelService
	rateLimiter      *rate.Limiter
	logger           *logrus.Logger
	startTime        time.Time
}

// New creates a new handler instance
func New(config *Config) *Handler {
	return &Handler{
		imageService:      config.ImageService,
		predictionService: config.PredictionService,
		modelService:      config.ModelService,
		rateLimiter:      config.RateLimiter,
		logger:           logrus.New(),
		startTime:        time.Now(),
	}
}

// Index serves the main homepage
func (h *Handler) Index(c *gin.Context) {
	h.logger.Info("Homepage accessed")
	
	// Get system stats for display
	stats := models.ModelStats{
		ModelsLoaded:      "2",
		TotalPredictions:  "0",
		AverageLatency:    "0",
		SystemHealth:      "healthy",
	}
	
	// Check actual model status
	if h.modelService != nil {
		modelStatus := h.modelService.GetModelStatus()
		stats.ModelsLoaded = fmt.Sprintf("%d", len(modelStatus.Models))
		if len(modelStatus.Models) == 0 {
			stats.SystemHealth = "degraded"
		}
	}
	
	template := templates.Index(stats)
	c.Header("Content-Type", "text/html")
	template.Render(c.Request.Context(), c.Writer)
}

// UploadPage serves the upload page
func (h *Handler) UploadPage(c *gin.Context) {
	template := templates.Upload()
	
	c.Header("Content-Type", "text/html")
	template.Render(c.Request.Context(), c.Writer)
}

// Upload handles image upload and prediction
func (h *Handler) Upload(c *gin.Context) {
	// Check rate limit
	if !h.rateLimiter.Allow() {
		h.respondError(c, http.StatusTooManyRequests, models.ErrorCodeRateLimitExceeded, 
			"Rate limit exceeded", "")
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, models.ErrorCodeInvalidRequest,
			"No image file provided", err.Error())
		return
	}
	defer file.Close()

	// Process image
	metadata, processedData, err := h.imageService.ProcessImage(file, header)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, models.ErrorCodeInvalidImage,
			"Failed to process image", err.Error())
		return
	}

	// Get model ID from form (optional)
	modelID := c.PostForm("model_id")

	// Perform prediction
	result, err := h.predictionService.PredictImage(processedData, metadata, modelID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, models.ErrorCodePredictionFailed,
			"Prediction failed", err.Error())
		return
	}

	// Return HTMX-compatible HTML response
	if h.isHTMXRequest(c) {
		h.renderPredictionResults(c, result)
		return
	}

	// Return JSON response
	response := &models.UploadResponse{
		Success:  true,
		Message:  "Image processed successfully",
		ResultID: result.ID,
		Result:   result,
	}

	c.JSON(http.StatusOK, response)
}

// APIPredictImage handles API prediction requests
func (h *Handler) APIPredictImage(c *gin.Context) {
	// Check rate limit
	if !h.rateLimiter.Allow() {
		h.respondError(c, http.StatusTooManyRequests, models.ErrorCodeRateLimitExceeded,
			"Rate limit exceeded", "")
		return
	}

	var request models.PredictionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondError(c, http.StatusBadRequest, models.ErrorCodeInvalidRequest,
			"Invalid request body", err.Error())
		return
	}

	// Create metadata
	metadata := &models.ImageMetadata{
		Filename:   request.Filename,
		Size:       int64(len(request.ImageData)),
		UploadedAt: time.Now(),
	}

	// Perform prediction
	result, err := h.predictionService.PredictImage(request.ImageData, metadata, request.ModelID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, models.ErrorCodePredictionFailed,
			"Prediction failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetResults retrieves prediction results by ID
func (h *Handler) GetResults(c *gin.Context) {
	resultID := c.Param("id")
	if resultID == "" {
		h.respondError(c, http.StatusBadRequest, models.ErrorCodeInvalidRequest,
			"Result ID is required", "")
		return
	}

	result, err := h.predictionService.GetResult(resultID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, models.ErrorCodeNotFound,
			"Result not found", err.Error())
		return
	}

	// Return HTMX-compatible HTML response
	if h.isHTMXRequest(c) {
		h.renderPredictionResults(c, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// APIGetResults handles API result retrieval
func (h *Handler) APIGetResults(c *gin.Context) {
	resultID := c.Param("id")
	if resultID == "" {
		h.respondError(c, http.StatusBadRequest, models.ErrorCodeInvalidRequest,
			"Result ID is required", "")
		return
	}

	result, err := h.predictionService.GetResult(resultID)
	if err != nil {
		h.respondError(c, http.StatusNotFound, models.ErrorCodeNotFound,
			"Result not found", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// APIListModels returns available models
func (h *Handler) APIListModels(c *gin.Context) {
	modelList := h.predictionService.ListModels()
	response := &models.ModelListResponse{
		Models: modelList,
		Total:  len(modelList),
	}

	c.JSON(http.StatusOK, response)
}

// StatusPage serves the status page
func (h *Handler) StatusPage(c *gin.Context) {
	health := h.getHealthStatus()
	template := templates.Status(*health)
	
	c.Header("Content-Type", "text/html")
	template.Render(c.Request.Context(), c.Writer)
}

// HealthCheck provides basic health check
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	})
}

// APIHealthCheck provides detailed health check
func (h *Handler) APIHealthCheck(c *gin.Context) {
	modelStatus := h.modelService.GetModelStatus()
	
	health := &models.HealthCheck{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Uptime:      time.Since(h.startTime).String(),
		Version:     "1.0.0",
		Services: map[string]string{
			"image_service":      "healthy",
			"prediction_service": "healthy",
			"model_service":      "healthy",
		},
		ModelStatus: modelStatus,
	}

	// Check if any models are unhealthy
	for _, modelHealth := range modelStatus.Models {
		if modelHealth.Status != "healthy" {
			health.Status = "degraded"
			health.Services["model_service"] = "degraded"
			break
		}
	}

	c.JSON(http.StatusOK, health)
}

// Helper methods

func (h *Handler) isHTMXRequest(c *gin.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

func (h *Handler) respondError(c *gin.Context, statusCode int, errorCode, message, details string) {
	errorResponse := models.NewErrorResponse(errorCode, message, details)
	
	h.logger.WithFields(logrus.Fields{
		"status_code": statusCode,
		"error_code":  errorCode,
		"message":     message,
		"details":     details,
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
	}).Error("Request error")

	if h.isHTMXRequest(c) {
		// Return HTMX-compatible error response using TEMPL
		template := templates.UploadError(message)
		c.Header("Content-Type", "text/html")
		template.Render(c.Request.Context(), c.Writer)
		return
	}

	c.JSON(statusCode, errorResponse)
}

func (h *Handler) renderPredictionResults(c *gin.Context, result *models.PredictionResult) {
	// Use TEMPL template for results
	template := templates.UploadResults(*result)
	
	c.Header("Content-Type", "text/html")
	template.Render(c.Request.Context(), c.Writer)
}

func (h *Handler) getHealthStatus() *models.HealthCheck {
	modelStatus := h.modelService.GetModelStatus()
	
	health := &models.HealthCheck{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Uptime:      time.Since(h.startTime).String(),
		Version:     "1.0.0",
		Services: map[string]string{
			"image_service":      "healthy",
			"prediction_service": "healthy",
			"model_service":      "healthy",
		},
		ModelStatus: modelStatus,
	}

	// Check if any models are unhealthy
	for _, modelHealth := range modelStatus.Models {
		if modelHealth.Status != "healthy" {
			health.Status = "degraded"
			health.Services["model_service"] = "degraded"
			break
		}
	}

	return health
}