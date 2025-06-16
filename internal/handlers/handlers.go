package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/francknouama/image-recognition-webapp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Config holds handler configuration
type Config struct {
	ImageService      *services.ImageService
	PredictionService *services.PredictionService
	RateLimiter      *rate.Limiter
}

// Handler contains all HTTP handlers
type Handler struct {
	imageService      *services.ImageService
	predictionService *services.PredictionService
	rateLimiter      *rate.Limiter
	logger           *logrus.Logger
	startTime        time.Time
}

// New creates a new handler instance
func New(config *Config) *Handler {
	return &Handler{
		imageService:      config.ImageService,
		predictionService: config.PredictionService,
		rateLimiter:      config.RateLimiter,
		logger:           logrus.New(),
		startTime:        time.Now(),
	}
}

// Index serves the main upload page
func (h *Handler) Index(c *gin.Context) {
	// For now, return a simple HTML response
	// This will be replaced with TEMPL templates later
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Image Recognition</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body>
    <main class="container">
        <h1>Image Recognition</h1>
        <p>Upload an image to get AI-powered classification results.</p>
        
        <div x-data="imageUpload()">
            <form hx-post="/upload" hx-encoding="multipart/form-data" hx-target="#results" hx-indicator="#loading">
                <input type="file" name="image" accept="image/*" @change="previewImage" required>
                <div x-show="imagePreview" class="image-preview" style="margin: 1rem 0;">
                    <img :src="imagePreview" alt="Preview" style="max-width: 300px; max-height: 300px;">
                </div>
                <button type="submit">Analyze Image</button>
            </form>
            
            <div id="loading" class="htmx-indicator">
                <p>Processing image...</p>
            </div>
            
            <div id="results"></div>
        </div>
    </main>

    <script>
        function imageUpload() {
            return {
                imagePreview: null,
                previewImage(event) {
                    const file = event.target.files[0];
                    if (file) {
                        const reader = new FileReader();
                        reader.onload = (e) => {
                            this.imagePreview = e.target.result;
                        };
                        reader.readAsDataURL(file);
                    }
                }
            }
        }
    </script>
</body>
</html>`
	
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
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
	modelStatus := h.predictionService.GetModelStatus()
	
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
		// Return HTMX-compatible error response
		html := fmt.Sprintf(`
		<div class="error" style="color: red; padding: 1rem; border: 1px solid red; border-radius: 4px;">
			<h3>Error</h3>
			<p><strong>%s</strong></p>
			<p>%s</p>
		</div>`, message, details)
		c.Header("Content-Type", "text/html")
		c.String(statusCode, html)
		return
	}

	c.JSON(statusCode, errorResponse)
}

func (h *Handler) renderPredictionResults(c *gin.Context, result *models.PredictionResult) {
	// Build HTML for prediction results
	var html strings.Builder
	
	html.WriteString(`<div class="results" style="margin-top: 2rem;">`)
	html.WriteString(`<h3>Prediction Results</h3>`)
	
	// Image metadata
	html.WriteString(fmt.Sprintf(`
		<div class="metadata" style="margin-bottom: 1rem;">
			<p><strong>File:</strong> %s</p>
			<p><strong>Size:</strong> %d bytes</p>
			<p><strong>Dimensions:</strong> %dx%d</p>
			<p><strong>Processing Time:</strong> %.2f ms</p>
			<p><strong>Model:</strong> %s (v%s)</p>
		</div>`,
		result.Metadata.Filename,
		result.Metadata.Size,
		result.Metadata.Width,
		result.Metadata.Height,
		result.ProcessTime,
		result.ModelInfo.Name,
		result.ModelInfo.Version,
	))

	// Predictions
	html.WriteString(`<div class="predictions">`)
	html.WriteString(`<h4>Top Predictions:</h4>`)
	html.WriteString(`<table>`)
	html.WriteString(`<thead><tr><th>Class</th><th>Confidence</th><th>Probability</th></tr></thead>`)
	html.WriteString(`<tbody>`)
	
	for _, pred := range result.Predictions {
		confidencePercent := pred.Confidence * 100
		probabilityPercent := pred.Probability * 100
		html.WriteString(fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%.2f%%</td>
				<td>%.2f%%</td>
			</tr>`,
			pred.ClassName,
			confidencePercent,
			probabilityPercent,
		))
	}
	
	html.WriteString(`</tbody></table>`)
	html.WriteString(`</div>`)
	
	// Action buttons
	html.WriteString(`
		<div style="margin-top: 1rem;">
			<button onclick="location.reload()">Analyze Another Image</button>
		</div>
	`)
	
	html.WriteString(`</div>`)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html.String())
}