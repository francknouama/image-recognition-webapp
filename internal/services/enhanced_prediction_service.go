package services

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/sirupsen/logrus"
)

// EnhancedPredictionService handles ML predictions with both TensorFlow and fallback simulation
type EnhancedPredictionService struct {
	modelService    *ModelService
	imageService    *ImageService
	tfService       *MockTensorFlowService
	imageProcessor  *ImageProcessor
	logger          *logrus.Logger
	results         map[string]*models.PredictionResult
	useTensorFlow   bool
}

// NewEnhancedPredictionService creates a new enhanced prediction service
func NewEnhancedPredictionService(modelService *ModelService, imageService *ImageService, tfService *MockTensorFlowService) *EnhancedPredictionService {
	service := &EnhancedPredictionService{
		modelService:   modelService,
		imageService:   imageService,
		tfService:      tfService,
		imageProcessor: NewImageProcessor(),
		logger:         logrus.New(),
		results:        make(map[string]*models.PredictionResult),
		useTensorFlow:  false,
	}
	
	// Check TensorFlow availability after initialization
	service.useTensorFlow = service.checkTensorFlowAvailability()
	
	return service
}

// checkTensorFlowAvailability checks if TensorFlow models are available
func (s *EnhancedPredictionService) checkTensorFlowAvailability() bool {
	// Check if there are any TensorFlow models loaded
	tfModels := s.tfService.ListModels()
	return len(tfModels) > 0
}

// PredictImage performs image classification using TensorFlow or simulation
func (s *EnhancedPredictionService) PredictImage(imageData []byte, metadata *models.ImageMetadata, modelID string) (*models.PredictionResult, error) {
	startTime := time.Now()
	resultID := s.generateResultID()

	// Get model information
	model, err := s.modelService.GetModel(modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	var predictions []models.ClassificationResult
	
	// Try TensorFlow prediction first
	if s.useTensorFlow {
		predictions, err = s.performTensorFlowInference(imageData, modelID)
		if err != nil {
			s.logger.Warnf("TensorFlow inference failed, falling back to simulation: %v", err)
			predictions, err = s.performSimulatedInference(imageData, model)
		}
	} else {
		// Use simulated inference
		predictions, err = s.performSimulatedInference(imageData, model)
	}

	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	processingTime := time.Since(startTime).Seconds() * 1000

	// Create result
	result := &models.PredictionResult{
		ID:          resultID,
		Predictions: predictions,
		Metadata:    *metadata,
		ProcessedAt: time.Now(),
		ProcessTime: processingTime,
		ModelInfo:   model.Info,
	}

	// Update model statistics
	s.modelService.UpdateModelStats(model.Info.ID, processingTime, err == nil)

	// Store result
	s.results[resultID] = result

	s.logger.Infof("Prediction completed: %s (%.2fms, model: %s, method: %s)", 
		resultID, processingTime, model.Info.Name, s.getInferenceMethod())

	return result, nil
}

// performTensorFlowInference runs actual TensorFlow inference
func (s *EnhancedPredictionService) performTensorFlowInference(imageData []byte, modelID string) ([]models.ClassificationResult, error) {
	// Get TensorFlow model
	tfModel, err := s.tfService.GetModel(modelID)
	if err != nil {
		return nil, fmt.Errorf("TensorFlow model not found: %w", err)
	}

	// Preprocess image
	tensorData, err := s.imageProcessor.ProcessImageBytes(imageData)
	if err != nil {
		return nil, fmt.Errorf("image preprocessing failed: %w", err)
	}

	// Run inference
	rawPredictions, err := s.tfService.Predict(modelID, tensorData)
	if err != nil {
		return nil, fmt.Errorf("TensorFlow prediction failed: %w", err)
	}

	// Postprocess predictions
	classificationPreds, err := s.imageProcessor.PostprocessPredictions(rawPredictions, tfModel.Info.Classes, 5)
	if err != nil {
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	// Convert to models.ClassificationResult format
	var predictions []models.ClassificationResult
	for _, pred := range classificationPreds {
		predictions = append(predictions, models.ClassificationResult{
			ClassName:   pred.ClassName,
			Label:       pred.ClassName,
			Description: s.getClassDescription(pred.ClassName),
			Confidence:  float64(pred.Confidence),
			Probability: float64(pred.Probability),
		})
	}

	return predictions, nil
}

// performSimulatedInference runs simulated inference (fallback)
func (s *EnhancedPredictionService) performSimulatedInference(imageData []byte, model *LoadedModel) ([]models.ClassificationResult, error) {
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

// LoadTensorFlowModel loads a TensorFlow model from disk
func (s *EnhancedPredictionService) LoadTensorFlowModel(modelPath string, modelID string) error {
	if !s.pathExists(modelPath) {
		return fmt.Errorf("model path does not exist: %s", modelPath)
	}

	err := s.tfService.LoadModel(modelPath, modelID)
	if err != nil {
		return fmt.Errorf("failed to load TensorFlow model: %w", err)
	}

	// Update availability status
	s.useTensorFlow = s.checkTensorFlowAvailability()
	
	s.logger.Infof("Successfully loaded TensorFlow model: %s", modelID)
	return nil
}

// GetInferenceMethod returns the current inference method being used
func (s *EnhancedPredictionService) getInferenceMethod() string {
	if s.useTensorFlow {
		return "tensorflow"
	}
	return "simulated"
}

// GetResult retrieves a prediction result by ID
func (s *EnhancedPredictionService) GetResult(resultID string) (*models.PredictionResult, error) {
	result, exists := s.results[resultID]
	if !exists {
		return nil, fmt.Errorf("result not found: %s", resultID)
	}
	return result, nil
}

// ListModels returns available models (both regular and TensorFlow)
func (s *EnhancedPredictionService) ListModels() []models.ModelInfo {
	var allModels []models.ModelInfo
	
	// Add regular models
	regularModels := s.modelService.ListModels()
	allModels = append(allModels, regularModels...)
	
	// Add TensorFlow models
	tfModels := s.tfService.ListModels()
	allModels = append(allModels, tfModels...)
	
	return allModels
}

// Helper methods

func (s *EnhancedPredictionService) generateResultID() string {
	return fmt.Sprintf("pred_%d", time.Now().UnixNano())
}

func (s *EnhancedPredictionService) pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (s *EnhancedPredictionService) generateConfidence(seed, index int64) float64 {
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

func (s *EnhancedPredictionService) normalizeProbabilities(predictions []models.ClassificationResult) {
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

func (s *EnhancedPredictionService) getClassDescription(className string) string {
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
		// ImageNet classes
		"tench":      "A European freshwater fish",
		"goldfish":   "A small golden-colored fish",
		"great_white_shark": "A large predatory shark",
		"tiger_shark": "A large shark with distinctive markings",
		"hammerhead": "A shark with a flattened head",
		"electric_ray": "A cartilaginous fish that can produce electric discharge",
		"stingray":   "A cartilaginous fish with a long tail",
		"cock":       "A male domestic fowl",
		"hen":        "A female domestic fowl",
		"ostrich":    "A large flightless bird",
	}
	
	if desc, exists := descriptions[className]; exists {
		return desc
	}
	
	return fmt.Sprintf("A %s object or entity", className)
}