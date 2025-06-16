package models

import (
	"time"
)

// PredictionRequest represents a request for image prediction
type PredictionRequest struct {
	ImageData []byte `json:"image_data"`
	Filename  string `json:"filename"`
	ModelID   string `json:"model_id,omitempty"`
}

// PredictionResult represents the result of an image prediction
type PredictionResult struct {
	ID          string                 `json:"id"`
	Predictions []ClassificationResult `json:"predictions"`
	Metadata    ImageMetadata          `json:"metadata"`
	ProcessedAt time.Time              `json:"processed_at"`
	ProcessTime float64                `json:"process_time_ms"`
	ModelInfo   ModelInfo              `json:"model_info"`
}

// ClassificationResult represents a single classification prediction
type ClassificationResult struct {
	ClassName   string  `json:"class_name"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Probability float64 `json:"probability"`
}

// ImageMetadata contains metadata about the uploaded image
type ImageMetadata struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Format      string `json:"format"`
	ContentType string `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// ModelInfo contains information about the model used for prediction
type ModelInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	InputShape   []int             `json:"input_shape"`
	OutputShape  []int             `json:"output_shape"`
	Classes      []string          `json:"classes"`
	LoadedAt     time.Time         `json:"loaded_at"`
	Metadata     map[string]string `json:"metadata"`
}

// UploadResponse represents the response after uploading an image
type UploadResponse struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	ResultID   string            `json:"result_id,omitempty"`
	Result     *PredictionResult `json:"result,omitempty"`
	Error      *ErrorResponse    `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthCheck represents the health status of the service
type HealthCheck struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	Version     string            `json:"version"`
	Services    map[string]string `json:"services"`
	ModelStatus ModelStatus       `json:"model_status"`
}

// ModelStatus represents the status of loaded models
type ModelStatus struct {
	LoadedModels int                    `json:"loaded_models"`
	TotalModels  int                    `json:"total_models"`
	Models       map[string]ModelHealth `json:"models"`
}

// ModelHealth represents the health status of a specific model
type ModelHealth struct {
	Status      string    `json:"status"`
	LastUsed    time.Time `json:"last_used"`
	Predictions int64     `json:"predictions"`
	AvgTime     float64   `json:"avg_time_ms"`
	Errors      int64     `json:"errors"`
}

// ModelStats represents statistics about the models and system
type ModelStats struct {
	ModelsLoaded      string `json:"models_loaded"`
	TotalPredictions  string `json:"total_predictions"`
	AverageLatency    string `json:"average_latency"`
	SystemHealth      string `json:"system_health"`
}

// BatchPredictionRequest represents a request for batch image prediction
type BatchPredictionRequest struct {
	Images  []ImageRequest `json:"images"`
	ModelID string         `json:"model_id,omitempty"`
}

// ImageRequest represents a single image in a batch request
type ImageRequest struct {
	ID       string `json:"id"`
	Data     []byte `json:"data"`
	Filename string `json:"filename"`
}

// BatchPredictionResponse represents the response for batch prediction
type BatchPredictionResponse struct {
	Success     bool                       `json:"success"`
	Results     map[string]PredictionResult `json:"results"`
	Errors      map[string]ErrorResponse   `json:"errors"`
	ProcessTime float64                    `json:"total_process_time_ms"`
}

// ModelListResponse represents the response for listing available models
type ModelListResponse struct {
	Models []ModelInfo `json:"models"`
	Total  int         `json:"total"`
}

// StatusCode represents HTTP status codes for responses
type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusUnauthorized        StatusCode = 401
	StatusForbidden           StatusCode = 403
	StatusNotFound            StatusCode = 404
	StatusMethodNotAllowed    StatusCode = 405
	StatusRequestTimeout      StatusCode = 408
	StatusPayloadTooLarge     StatusCode = 413
	StatusUnsupportedMedia    StatusCode = 415
	StatusTooManyRequests     StatusCode = 429
	StatusInternalServerError StatusCode = 500
	StatusBadGateway          StatusCode = 502
	StatusServiceUnavailable  StatusCode = 503
	StatusGatewayTimeout      StatusCode = 504
)

// Error codes for different types of errors
const (
	ErrorCodeInvalidImage      = "INVALID_IMAGE"
	ErrorCodeUnsupportedFormat = "UNSUPPORTED_FORMAT"
	ErrorCodeFileTooLarge      = "FILE_TOO_LARGE"
	ErrorCodeModelNotFound     = "MODEL_NOT_FOUND"
	ErrorCodeModelLoadFailed   = "MODEL_LOAD_FAILED"
	ErrorCodePredictionFailed  = "PREDICTION_FAILED"
	ErrorCodeInternalError     = "INTERNAL_ERROR"
	ErrorCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrorCodeInvalidRequest    = "INVALID_REQUEST"
	ErrorCodeNotFound          = "NOT_FOUND"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// PredictionStatus represents the status of a prediction job
type PredictionStatus string

const (
	StatusPending    PredictionStatus = "pending"
	StatusProcessing PredictionStatus = "processing"
	StatusCompleted  PredictionStatus = "completed"
	StatusFailed     PredictionStatus = "failed"
)

// Job represents an async prediction job
type Job struct {
	ID        string           `json:"id"`
	Status    PredictionStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Result    *PredictionResult `json:"result,omitempty"`
	Error     *ErrorResponse   `json:"error,omitempty"`
	Progress  float64          `json:"progress"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message, details string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ToHTTPStatus converts a status code to HTTP status code
func (s StatusCode) ToHTTPStatus() int {
	return int(s)
}