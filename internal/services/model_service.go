package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/config"
	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/sirupsen/logrus"
)

// ModelService handles machine learning model operations
type ModelService struct {
	config       *config.Config
	logger       *logrus.Logger
	models       map[string]*LoadedModel
	modelsMutex  sync.RWMutex
	defaultModel string
}

// LoadedModel represents a loaded ML model
type LoadedModel struct {
	Info        models.ModelInfo
	Health      models.ModelHealth
	LastUsed    time.Time
	Predictions int64
	Errors      int64
	TotalTime   float64
}

// NewModelService creates a new model service
func NewModelService(cfg *config.Config) *ModelService {
	service := &ModelService{
		config: cfg,
		logger: logrus.New(),
		models: make(map[string]*LoadedModel),
	}

	// Load models on startup
	if err := service.LoadModels(); err != nil {
		service.logger.Errorf("Failed to load models on startup: %v", err)
	}

	return service
}

// LoadModels loads all available models from the model directory
func (s *ModelService) LoadModels() error {
	s.modelsMutex.Lock()
	defer s.modelsMutex.Unlock()

	modelPath := s.config.Model.Path
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		s.logger.Warnf("Model directory does not exist: %s", modelPath)
		// Create a dummy model for development
		s.createDummyModel()
		return nil
	}

	entries, err := os.ReadDir(modelPath)
	if err != nil {
		return fmt.Errorf("failed to read model directory: %w", err)
	}

	loadedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modelID := entry.Name()
		if err := s.loadModel(modelID); err != nil {
			s.logger.Errorf("Failed to load model %s: %v", modelID, err)
			continue
		}
		loadedCount++

		// Set first successfully loaded model as default
		if s.defaultModel == "" {
			s.defaultModel = modelID
		}
	}

	s.logger.Infof("Loaded %d models", loadedCount)

	// If no models were loaded, create a dummy model
	if loadedCount == 0 {
		s.createDummyModel()
	}

	return nil
}

// loadModel loads a specific model by ID
func (s *ModelService) loadModel(modelID string) error {
	modelDir := filepath.Join(s.config.Model.Path, modelID)
	
	// Check if model directory exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		return fmt.Errorf("model directory not found: %s", modelDir)
	}

	// Load model metadata
	metadata, err := s.loadModelMetadata(modelDir)
	if err != nil {
		s.logger.Warnf("Failed to load metadata for model %s, using defaults: %v", modelID, err)
		metadata = s.createDefaultMetadata(modelID)
	}

	// Create loaded model
	loadedModel := &LoadedModel{
		Info: *metadata,
		Health: models.ModelHealth{
			Status:      "healthy",
			LastUsed:    time.Now(),
			Predictions: 0,
			AvgTime:     0,
			Errors:      0,
		},
		LastUsed:    time.Now(),
		Predictions: 0,
		Errors:      0,
		TotalTime:   0,
	}

	s.models[modelID] = loadedModel
	s.logger.Infof("Loaded model: %s (version: %s)", metadata.Name, metadata.Version)

	return nil
}

// loadModelMetadata loads model metadata from a JSON file
func (s *ModelService) loadModelMetadata(modelDir string) (*models.ModelInfo, error) {
	metadataPath := filepath.Join(modelDir, "metadata.json")
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata models.ModelInfo
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	metadata.LoadedAt = time.Now()
	return &metadata, nil
}

// createDefaultMetadata creates default metadata for a model
func (s *ModelService) createDefaultMetadata(modelID string) *models.ModelInfo {
	return &models.ModelInfo{
		ID:          modelID,
		Name:        fmt.Sprintf("Model %s", modelID),
		Version:     s.config.Model.Version,
		Description: "Image classification model",
		InputShape:  []int{224, 224, 3},
		OutputShape: []int{1000},
		Classes:     s.getDefaultClasses(),
		LoadedAt:    time.Now(),
		Metadata:    make(map[string]string),
	}
}

// getDefaultClasses returns default class names for image classification
func (s *ModelService) getDefaultClasses() []string {
	return []string{
		"cat", "dog", "bird", "car", "truck", "airplane", "boat", "train",
		"bicycle", "motorcycle", "person", "horse", "sheep", "cow", "elephant",
		"bear", "zebra", "giraffe", "backpack", "umbrella", "handbag", "tie",
		"suitcase", "frisbee", "skis", "snowboard", "sports ball", "kite",
		"baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket",
		"bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl",
		"banana", "apple", "sandwich", "orange", "broccoli", "carrot",
		"hot dog", "pizza", "donut", "cake", "chair", "couch", "potted plant",
	}
}

// createDummyModel creates a dummy model for development/testing
func (s *ModelService) createDummyModel() {
	dummyModel := &LoadedModel{
		Info: models.ModelInfo{
			ID:          "dummy",
			Name:        "Dummy Model",
			Version:     "1.0.0",
			Description: "Development dummy model for testing",
			InputShape:  []int{224, 224, 3},
			OutputShape: []int{50},
			Classes:     s.getDefaultClasses()[:50],
			LoadedAt:    time.Now(),
			Metadata:    map[string]string{"type": "dummy"},
		},
		Health: models.ModelHealth{
			Status:      "healthy",
			LastUsed:    time.Now(),
			Predictions: 0,
			AvgTime:     0,
			Errors:      0,
		},
		LastUsed:    time.Now(),
		Predictions: 0,
		Errors:      0,
		TotalTime:   0,
	}

	s.models["dummy"] = dummyModel
	s.defaultModel = "dummy"
	s.logger.Info("Created dummy model for development")
}

// GetModel returns a model by ID
func (s *ModelService) GetModel(modelID string) (*LoadedModel, error) {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	if modelID == "" {
		modelID = s.defaultModel
	}

	model, exists := s.models[modelID]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	return model, nil
}

// GetDefaultModel returns the default model
func (s *ModelService) GetDefaultModel() (*LoadedModel, error) {
	return s.GetModel(s.defaultModel)
}

// ListModels returns all loaded models
func (s *ModelService) ListModels() []models.ModelInfo {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	var modelList []models.ModelInfo
	for _, model := range s.models {
		modelList = append(modelList, model.Info)
	}

	return modelList
}

// GetModelStatus returns the status of all models
func (s *ModelService) GetModelStatus() models.ModelStatus {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	status := models.ModelStatus{
		LoadedModels: len(s.models),
		TotalModels:  len(s.models), // For now, assume all available models are loaded
		Models:       make(map[string]models.ModelHealth),
	}

	for id, model := range s.models {
		status.Models[id] = model.Health
	}

	return status
}

// UpdateModelStats updates model usage statistics
func (s *ModelService) UpdateModelStats(modelID string, processingTime float64, success bool) {
	s.modelsMutex.Lock()
	defer s.modelsMutex.Unlock()

	model, exists := s.models[modelID]
	if !exists {
		return
	}

	model.LastUsed = time.Now()
	model.Health.LastUsed = time.Now()
	model.Predictions++
	model.Health.Predictions++
	model.TotalTime += processingTime

	if success {
		// Update average time
		model.Health.AvgTime = model.TotalTime / float64(model.Predictions)
	} else {
		model.Errors++
		model.Health.Errors++
	}

	// Update health status based on error rate
	errorRate := float64(model.Errors) / float64(model.Predictions)
	if errorRate > 0.5 {
		model.Health.Status = "unhealthy"
	} else if errorRate > 0.1 {
		model.Health.Status = "degraded"
	} else {
		model.Health.Status = "healthy"
	}
}

// IsModelHealthy checks if a model is healthy
func (s *ModelService) IsModelHealthy(modelID string) bool {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	model, exists := s.models[modelID]
	if !exists {
		return false
	}

	return model.Health.Status == "healthy"
}

// ReloadModel reloads a specific model
func (s *ModelService) ReloadModel(modelID string) error {
	s.modelsMutex.Lock()
	defer s.modelsMutex.Unlock()

	// Remove existing model
	delete(s.models, modelID)

	// Reload the model
	if err := s.loadModel(modelID); err != nil {
		return fmt.Errorf("failed to reload model %s: %w", modelID, err)
	}

	s.logger.Infof("Reloaded model: %s", modelID)
	return nil
}

// GetModelInfo returns model information
func (s *ModelService) GetModelInfo(modelID string) (*models.ModelInfo, error) {
	model, err := s.GetModel(modelID)
	if err != nil {
		return nil, err
	}

	return &model.Info, nil
}