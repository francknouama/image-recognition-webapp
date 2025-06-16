package services

import (
	"testing"

	"github.com/francknouama/image-recognition-webapp/internal/config"
)

func TestNewTensorFlowService(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewTensorFlowService(cfg)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.config != cfg {
		t.Error("Expected config to be set")
	}

	if len(service.models) != 0 {
		t.Error("Expected no models loaded initially")
	}
}

func TestMockTensorFlowLoadModel(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewTensorFlowService(cfg)
	
	err := service.LoadModel("./testdata/demo_model", "test_model")
	if err != nil {
		t.Errorf("Expected model to load successfully, got error: %v", err)
	}

	// Check that model was added
	models := service.ListModels()
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

	if models[0].ID != "test_model" {
		t.Errorf("Expected model ID 'test_model', got '%s'", models[0].ID)
	}
}

func TestMockTensorFlowGetModel(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewTensorFlowService(cfg)
	
	// Load a model first
	err := service.LoadModel("./testdata/demo_model", "test_model")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// Get the model
	model, err := service.GetModel("test_model")
	if err != nil {
		t.Errorf("Expected to get model, got error: %v", err)
	}

	if model == nil {
		t.Error("Expected model to be returned")
	}

	if model.Info.ID != "test_model" {
		t.Errorf("Expected model ID 'test_model', got '%s'", model.Info.ID)
	}

	// Test getting non-existent model
	_, err = service.GetModel("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestMockTensorFlowPredict(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewTensorFlowService(cfg)
	
	// Load a model first
	err := service.LoadModel("./testdata/demo_model", "test_model")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// Create test input data
	imageData := [][]float32{
		make([]float32, 224*224*3), // Simulate preprocessed image
	}

	// Fill with some test data
	for i := range imageData[0] {
		imageData[0][i] = float32(i%255) / 255.0
	}

	// Run prediction
	predictions, err := service.Predict("test_model", imageData)
	if err != nil {
		t.Errorf("Expected prediction to succeed, got error: %v", err)
	}

	if len(predictions) == 0 {
		t.Error("Expected predictions to be returned")
	}

	// Check that predictions are valid probabilities
	for i, pred := range predictions {
		if pred < 0 || pred > 1 {
			t.Errorf("Prediction %d (%f) should be between 0 and 1", i, pred)
		}
	}

	// Test prediction with non-existent model
	_, err = service.Predict("non_existent", imageData)
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestMockTensorFlowUnloadModel(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewTensorFlowService(cfg)
	
	// Load a model first
	err := service.LoadModel("./testdata/demo_model", "test_model")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// Verify model exists
	models := service.ListModels()
	if len(models) != 1 {
		t.Fatalf("Expected 1 model before unload, got %d", len(models))
	}

	// Unload the model
	err = service.UnloadModel("test_model")
	if err != nil {
		t.Errorf("Expected unload to succeed, got error: %v", err)
	}

	// Verify model is gone
	models = service.ListModels()
	if len(models) != 0 {
		t.Errorf("Expected 0 models after unload, got %d", len(models))
	}

	// Test unloading non-existent model
	err = service.UnloadModel("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestMockTensorFlowClose(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewTensorFlowService(cfg)
	
	// Load multiple models
	service.LoadModel("./testdata/demo_model1", "test_model1")
	service.LoadModel("./testdata/demo_model2", "test_model2")

	// Verify models exist
	models := service.ListModels()
	if len(models) != 2 {
		t.Fatalf("Expected 2 models before close, got %d", len(models))
	}

	// Close all models
	service.Close()

	// Verify all models are gone
	models = service.ListModels()
	if len(models) != 0 {
		t.Errorf("Expected 0 models after close, got %d", len(models))
	}
}