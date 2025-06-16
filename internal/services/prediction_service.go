package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/sirupsen/logrus"
)

// PredictionService handles model inference operations
type PredictionService struct {
	modelService *ModelService
	imageService *ImageService
	logger       *logrus.Logger
	results      map[string]*models.PredictionResult
}

// NewPredictionService creates a new prediction service
func NewPredictionService(modelService *ModelService, imageService *ImageService) *PredictionService {
	return &PredictionService{
		modelService: modelService,
		imageService: imageService,
		logger:       logrus.New(),
		results:      make(map[string]*models.PredictionResult),
	}
}

// PredictImage performs image classification prediction
func (s *PredictionService) PredictImage(imageData []byte, metadata *models.ImageMetadata, modelID string) (*models.PredictionResult, error) {
	startTime := time.Now()
	
	// Get model
	model, err := s.modelService.GetModel(modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	// Decode image for preprocessing
	img, _, err := s.imageService.decodeImage(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Preprocess image for model input
	processedData, err := s.imageService.preprocessForModel(img)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess image: %w", err)
	}

	// Perform prediction (simulated for now since we don't have actual TensorFlow integration)
	predictions, err := s.performInference(processedData, model)
	if err != nil {
		s.modelService.UpdateModelStats(model.Info.ID, 0, false)
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds
	
	// Update model statistics
	s.modelService.UpdateModelStats(model.Info.ID, processingTime, true)

	// Create result
	resultID := s.generateResultID()
	result := &models.PredictionResult{
		ID:          resultID,
		Predictions: predictions,
		Metadata:    *metadata,
		ProcessedAt: time.Now(),
		ProcessTime: processingTime,
		ModelInfo:   model.Info,
	}

	// Store result for later retrieval
	s.results[resultID] = result

	s.logger.Infof("Prediction completed: %s (%.2fms, model: %s)", 
		resultID, processingTime, model.Info.Name)

	return result, nil
}

// performInference simulates model inference (placeholder for actual TensorFlow integration)
func (s *PredictionService) performInference(imageData []byte, model *LoadedModel) ([]models.ClassificationResult, error) {
	// Simulate processing time
	time.Sleep(time.Millisecond * 100)

	// Generate simulated predictions
	predictions := make([]models.ClassificationResult, 0, 5)
	
	// Use deterministic randomness based on image data for consistent results
	seed := int64(len(imageData))
	for i, class := range model.Info.Classes {
		if i >= 10 { // Limit to top 10 classes for simulation
			break
		}
		
		// Generate pseudo-random confidence based on class index and image data
		confidence := s.generateConfidence(seed, int64(i))
		
		if confidence > 0.01 { // Only include predictions with >1% confidence
			predictions = append(predictions, models.ClassificationResult{
				ClassName:   class,
				Label:       class,
				Description: s.getClassDescription(class),
				Confidence:  confidence,
				Probability: confidence, // For now, confidence and probability are the same
			})
		}
	}

	// Sort by confidence (descending)
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})

	// Normalize probabilities to sum to 1.0
	s.normalizeProbabilities(predictions)

	// Return top 5 predictions
	if len(predictions) > 5 {
		predictions = predictions[:5]
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no valid predictions generated")
	}

	return predictions, nil
}

// generateConfidence generates a pseudo-random confidence score
func (s *PredictionService) generateConfidence(seed, index int64) float64 {
	// Simple pseudo-random generation for consistent results
	x := float64((seed*31+index*17)%1000) / 1000.0
	
	// Use a function that creates a more realistic distribution
	// Higher chance for lower confidences, with occasional high confidence
	confidence := math.Exp(-x*3) * (0.3 + 0.7*math.Sin(x*math.Pi))
	
	// Ensure confidence is between 0 and 1
	if confidence < 0 {
		confidence = -confidence
	}
	if confidence > 1 {
		confidence = 1.0
	}
	
	return confidence
}

// normalizeProbabilities normalizes prediction probabilities to sum to 1.0
func (s *PredictionService) normalizeProbabilities(predictions []models.ClassificationResult) {
	var total float64
	for _, pred := range predictions {
		total += pred.Probability
	}
	
	if total > 0 {
		for i := range predictions {
			predictions[i].Probability /= total
		}
	}
}

// getClassDescription returns a description for a class name
func (s *PredictionService) getClassDescription(className string) string {
	descriptions := map[string]string{
		"cat":        "A small domestic feline mammal",
		"dog":        "A domestic canine companion animal",
		"bird":       "A feathered, winged, bipedal animal",
		"car":        "A four-wheeled motor vehicle",
		"truck":      "A large motor vehicle for transporting goods",
		"airplane":   "A powered flying vehicle with wings",
		"boat":       "A watercraft designed for travel on water",
		"train":      "A connected series of railway cars",
		"bicycle":    "A two-wheeled vehicle powered by pedaling",
		"motorcycle": "A two-wheeled motor vehicle",
		"person":     "A human being",
		"horse":      "A large domesticated ungulate mammal",
		"sheep":      "A woolly ruminant mammal",
		"cow":        "A large domesticated bovine animal",
		"elephant":   "A large mammal with a trunk",
		"bear":       "A large omnivorous mammal",
		"zebra":      "A black and white striped equine",
		"giraffe":    "A tall African mammal with a long neck",
	}
	
	if desc, exists := descriptions[className]; exists {
		return desc
	}
	
	return fmt.Sprintf("A %s object or entity", className)
}

// BatchPredict performs batch prediction on multiple images
func (s *PredictionService) BatchPredict(requests []models.ImageRequest, modelID string) (*models.BatchPredictionResponse, error) {
	startTime := time.Now()
	
	response := &models.BatchPredictionResponse{
		Results: make(map[string]models.PredictionResult),
		Errors:  make(map[string]models.ErrorResponse),
	}

	for _, req := range requests {
		// Create temporary metadata
		metadata := &models.ImageMetadata{
			Filename:   req.Filename,
			Size:       int64(len(req.Data)),
			UploadedAt: time.Now(),
		}

		// Perform prediction
		result, err := s.PredictImage(req.Data, metadata, modelID)
		if err != nil {
			response.Errors[req.ID] = *models.NewErrorResponse(
				models.ErrorCodePredictionFailed,
				"Prediction failed",
				err.Error(),
			)
			continue
		}

		response.Results[req.ID] = *result
	}

	response.ProcessTime = float64(time.Since(startTime).Nanoseconds()) / 1e6
	response.Success = len(response.Errors) == 0

	return response, nil
}

// GetResult retrieves a prediction result by ID
func (s *PredictionService) GetResult(resultID string) (*models.PredictionResult, error) {
	result, exists := s.results[resultID]
	if !exists {
		return nil, fmt.Errorf("result not found: %s", resultID)
	}

	return result, nil
}

// GetTopPrediction returns the top prediction result
func (s *PredictionService) GetTopPrediction(result *models.PredictionResult) *models.ClassificationResult {
	if len(result.Predictions) == 0 {
		return nil
	}
	return &result.Predictions[0]
}

// GetPredictionsByThreshold returns predictions above a confidence threshold
func (s *PredictionService) GetPredictionsByThreshold(result *models.PredictionResult, threshold float64) []models.ClassificationResult {
	var filtered []models.ClassificationResult
	
	for _, pred := range result.Predictions {
		if pred.Confidence >= threshold {
			filtered = append(filtered, pred)
		}
	}
	
	return filtered
}

// generateResultID generates a unique result ID
func (s *PredictionService) generateResultID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CleanupResults removes old prediction results to prevent memory leaks
func (s *PredictionService) CleanupResults(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	
	for id, result := range s.results {
		if result.ProcessedAt.Before(cutoff) {
			delete(s.results, id)
		}
	}
	
	s.logger.Debugf("Cleaned up old prediction results, current count: %d", len(s.results))
}

// GetResultsCount returns the number of stored results
func (s *PredictionService) GetResultsCount() int {
	return len(s.results)
}

// ListModels returns available models (delegate to model service)
func (s *PredictionService) ListModels() []models.ModelInfo {
	return s.modelService.ListModels()
}

// GetModelStatus returns model status (delegate to model service)
func (s *PredictionService) GetModelStatus() models.ModelStatus {
	return s.modelService.GetModelStatus()
}

// ValidateModelForPrediction checks if a model is suitable for prediction
func (s *PredictionService) ValidateModelForPrediction(modelID string) error {
	model, err := s.modelService.GetModel(modelID)
	if err != nil {
		return err
	}

	if !s.modelService.IsModelHealthy(modelID) {
		return fmt.Errorf("model %s is not healthy", modelID)
	}

	// Check if model has required input/output shapes
	if len(model.Info.InputShape) != 3 {
		return fmt.Errorf("model %s has invalid input shape", modelID)
	}

	if len(model.Info.Classes) == 0 {
		return fmt.Errorf("model %s has no class labels", modelID)
	}

	return nil
}