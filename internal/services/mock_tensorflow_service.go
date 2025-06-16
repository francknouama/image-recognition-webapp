package services

import (
	"fmt"
	"math"
	"sync"

	"github.com/francknouama/image-recognition-webapp/internal/config"
	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/sirupsen/logrus"
)

// MockTensorFlowService provides TensorFlow interface without requiring C library
// This allows the application to be structured for TensorFlow integration
// while providing a working fallback implementation
type MockTensorFlowService struct {
	config      *config.Config
	logger      *logrus.Logger
	models      map[string]*MockTFModel
	modelsMutex sync.RWMutex
}

// MockTFModel represents a mock TensorFlow model
type MockTFModel struct {
	Info      models.ModelInfo
	LoadedAt  int64
	Available bool
}

// NewTensorFlowService creates a new mock TensorFlow service
func NewTensorFlowService(cfg *config.Config) *MockTensorFlowService {
	service := &MockTensorFlowService{
		config: cfg,
		logger: logrus.New(),
		models: make(map[string]*MockTFModel),
	}
	
	service.logger.Info("Using mock TensorFlow service (C library not available)")
	return service
}

// LoadModel simulates loading a TensorFlow SavedModel
func (s *MockTensorFlowService) LoadModel(modelPath string, modelID string) error {
	s.modelsMutex.Lock()
	defer s.modelsMutex.Unlock()

	s.logger.Infof("Mock: Loading TensorFlow model from %s", modelPath)

	// Create mock model info
	modelInfo := models.ModelInfo{
		ID:          modelID,
		Name:        fmt.Sprintf("Mock TensorFlow Model (%s)", modelID),
		Version:     "1.0.0",
		Description: "Mock TensorFlow SavedModel for development",
		InputShape:  []int{1, 224, 224, 3},
		OutputShape: []int{1, 1000},
		Classes:     s.getImageNetClasses(),
	}

	// Store the mock model
	mockModel := &MockTFModel{
		Info:      modelInfo,
		LoadedAt:  int64(len(s.models)),
		Available: true,
	}

	s.models[modelID] = mockModel
	s.logger.Infof("Mock: Successfully loaded TensorFlow model: %s", modelID)

	return nil
}

// Predict simulates TensorFlow inference
func (s *MockTensorFlowService) Predict(modelID string, imageData [][]float32) ([]float32, error) {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	mockModel, exists := s.models[modelID]
	if !exists {
		return nil, fmt.Errorf("mock model not found: %s", modelID)
	}

	if !mockModel.Available {
		return nil, fmt.Errorf("mock model not available: %s", modelID)
	}

	s.logger.Debugf("Mock: Running inference on model %s", modelID)

	// Generate mock predictions (simulate ImageNet-style output)
	numClasses := len(mockModel.Info.Classes)
	predictions := make([]float32, numClasses)
	
	// Generate pseudo-random but deterministic predictions
	seed := float32(len(imageData[0]) % 1000)
	for i := 0; i < numClasses; i++ {
		// Simple pseudo-random generation
		val := float32(i+1) * seed * 0.001
		predictions[i] = float32(1.0 / (1.0 + math.Exp(-float64(val)))) // Sigmoid-like activation
	}

	return predictions, nil
}

// GetModel returns a mock TensorFlow model
func (s *MockTensorFlowService) GetModel(modelID string) (*MockTFModel, error) {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	model, exists := s.models[modelID]
	if !exists {
		return nil, fmt.Errorf("mock model not found: %s", modelID)
	}

	return model, nil
}

// ListModels returns all mock TensorFlow models
func (s *MockTensorFlowService) ListModels() []models.ModelInfo {
	s.modelsMutex.RLock()
	defer s.modelsMutex.RUnlock()

	var modelList []models.ModelInfo
	for _, model := range s.models {
		modelList = append(modelList, model.Info)
	}

	return modelList
}

// UnloadModel simulates unloading a TensorFlow model
func (s *MockTensorFlowService) UnloadModel(modelID string) error {
	s.modelsMutex.Lock()
	defer s.modelsMutex.Unlock()

	_, exists := s.models[modelID]
	if !exists {
		return fmt.Errorf("mock model not found: %s", modelID)
	}

	delete(s.models, modelID)
	s.logger.Infof("Mock: Unloaded TensorFlow model: %s", modelID)

	return nil
}

// Close simulates closing all models
func (s *MockTensorFlowService) Close() {
	s.modelsMutex.Lock()
	defer s.modelsMutex.Unlock()

	for modelID := range s.models {
		s.logger.Infof("Mock: Closed TensorFlow model: %s", modelID)
	}

	s.models = make(map[string]*MockTFModel)
}

// getImageNetClasses returns a subset of ImageNet classes
func (s *MockTensorFlowService) getImageNetClasses() []string {
	return []string{
		"tench", "goldfish", "great_white_shark", "tiger_shark", "hammerhead",
		"electric_ray", "stingray", "cock", "hen", "ostrich", "brambling",
		"goldfinch", "house_finch", "junco", "indigo_bunting", "robin",
		"bulbul", "jay", "magpie", "chickadee", "water_ouzel", "kite",
		"bald_eagle", "vulture", "great_grey_owl", "cat", "dog", "horse",
		"sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack",
		"umbrella", "handbag", "tie", "suitcase", "frisbee", "skis",
		"snowboard", "sports_ball", "kite", "baseball_bat", "baseball_glove",
		"skateboard", "surfboard", "tennis_racket", "bottle", "wine_glass",
		"cup", "fork", "knife", "spoon", "bowl", "banana", "apple",
		"sandwich", "orange", "broccoli", "carrot", "hot_dog", "pizza",
		"donut", "cake", "chair", "couch", "potted_plant", "bed",
		"dining_table", "toilet", "tv", "laptop", "mouse", "remote",
		"keyboard", "cell_phone", "microwave", "oven", "toaster", "sink",
		"refrigerator", "book", "clock", "vase", "scissors", "teddy_bear",
		"hair_drier", "toothbrush", "person", "bicycle", "car", "motorcycle",
		"airplane", "bus", "train", "truck", "boat", "traffic_light",
		"fire_hydrant", "stop_sign", "parking_meter", "bench", "bird",
	}
}