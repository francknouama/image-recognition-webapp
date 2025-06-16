package services

import (
	"testing"

	"github.com/francknouama/image-recognition-webapp/internal/config"
)

func TestNewModelService(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewModelService(cfg)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.config != cfg {
		t.Error("Expected config to be set")
	}

	if len(service.models) == 0 {
		t.Log("No models loaded (expected for test)")
	}
}

func TestModelServiceGetStats(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewModelService(cfg)
	stats := service.GetStats()

	if stats.ModelsLoaded == "" {
		t.Error("Expected ModelsLoaded to be set")
	}

	if stats.TotalPredictions == "" {
		t.Error("Expected TotalPredictions to be set")
	}

	if stats.AverageLatency == "" {
		t.Error("Expected AverageLatency to be set")
	}

	if stats.SystemHealth == "" {
		t.Error("Expected SystemHealth to be set")
	}
}

func TestModelServiceGetModel(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewModelService(cfg)
	
	// Test getting default model (should be dummy)
	model, err := service.GetDefaultModel()
	if err != nil {
		t.Errorf("Expected to get default model, got error: %v", err)
	}

	if model == nil {
		t.Error("Expected model to be returned")
	}

	// Test getting non-existent model
	_, err = service.GetModel("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestModelServiceListModels(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewModelService(cfg)
	models := service.ListModels()

	// Should have at least the dummy model
	if len(models) == 0 {
		t.Error("Expected at least one model (dummy)")
	}
}

func TestModelServiceUpdateStats(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			Path:    "./testdata/models",
			Version: "1.0.0",
		},
	}

	service := NewModelService(cfg)
	
	// Get default model to update its stats
	model, err := service.GetDefaultModel()
	if err != nil {
		t.Fatalf("Failed to get default model: %v", err)
	}

	initialPredictions := model.Predictions
	
	// Update stats
	service.UpdateModelStats(model.Info.ID, 100.0, true)
	
	// Check that stats were updated
	updatedModel, err := service.GetModel(model.Info.ID)
	if err != nil {
		t.Fatalf("Failed to get updated model: %v", err)
	}

	if updatedModel.Predictions != initialPredictions+1 {
		t.Errorf("Expected predictions to be %d, got %d", initialPredictions+1, updatedModel.Predictions)
	}

	if updatedModel.Health.AvgTime == 0 {
		t.Error("Expected average time to be updated")
	}
}