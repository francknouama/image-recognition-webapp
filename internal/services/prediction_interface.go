package services

import (
	"github.com/francknouama/image-recognition-webapp/internal/models"
)

// PredictionServiceInterface defines the interface for prediction services
type PredictionServiceInterface interface {
	// PredictImage performs image classification
	PredictImage(imageData []byte, metadata *models.ImageMetadata, modelID string) (*models.PredictionResult, error)
	
	// GetResult retrieves a prediction result by ID
	GetResult(resultID string) (*models.PredictionResult, error)
	
	// ListModels returns available models
	ListModels() []models.ModelInfo
}

// Ensure both services implement the interface
var _ PredictionServiceInterface = (*PredictionService)(nil)
var _ PredictionServiceInterface = (*EnhancedPredictionService)(nil)