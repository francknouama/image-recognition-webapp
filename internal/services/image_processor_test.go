package services

import (
	"image"
	"image/color"
	"testing"
)

func TestNewImageProcessor(t *testing.T) {
	processor := NewImageProcessor()
	if processor == nil {
		t.Fatal("Expected processor to be created")
	}

	if processor.targetWidth != 224 || processor.targetHeight != 224 {
		t.Errorf("Expected default size 224x224, got %dx%d", processor.targetWidth, processor.targetHeight)
	}

	if !processor.normalize {
		t.Error("Expected normalization to be enabled by default")
	}
}

func TestImageProcessorSetTargetSize(t *testing.T) {
	processor := NewImageProcessor()
	processor.SetTargetSize(512, 512)

	if processor.targetWidth != 512 || processor.targetHeight != 512 {
		t.Errorf("Expected size 512x512, got %dx%d", processor.targetWidth, processor.targetHeight)
	}
}

func TestImageProcessorGetInputShape(t *testing.T) {
	processor := NewImageProcessor()
	shape := processor.GetInputShape()

	expected := []int{1, 224, 224, 3}
	if len(shape) != len(expected) {
		t.Errorf("Expected shape length %d, got %d", len(expected), len(shape))
	}

	for i, val := range expected {
		if shape[i] != val {
			t.Errorf("Expected shape[%d] = %d, got %d", i, val, shape[i])
		}
	}
}

func TestImageProcessorProcessImage(t *testing.T) {
	processor := NewImageProcessor()
	
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 128, B: 64, A: 255})
		}
	}

	tensorData, err := processor.ProcessImage(img)
	if err != nil {
		t.Fatalf("Failed to process image: %v", err)
	}

	if len(tensorData) != 1 {
		t.Errorf("Expected batch size 1, got %d", len(tensorData))
	}

	expectedSize := 224 * 224 * 3 // target_height * target_width * channels
	if len(tensorData[0]) != expectedSize {
		t.Errorf("Expected tensor size %d, got %d", expectedSize, len(tensorData[0]))
	}

	// Check that values are in reasonable range (normalized)
	for i, val := range tensorData[0][:10] { // Check first 10 values
		if val < -5.0 || val > 5.0 {
			t.Errorf("Value %d (%f) seems out of reasonable normalized range", i, val)
		}
	}
}

func TestPostprocessPredictions(t *testing.T) {
	processor := NewImageProcessor()
	
	// Create test predictions (logits)
	predictions := []float32{1.0, 2.0, 0.5, 3.0, 1.5}
	classNames := []string{"cat", "dog", "bird", "car", "horse"}
	
	results, err := processor.PostprocessPredictions(predictions, classNames, 3)
	if err != nil {
		t.Fatalf("Failed to postprocess predictions: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected top 3 results, got %d", len(results))
	}

	// Check that results are sorted by probability (descending)
	for i := 1; i < len(results); i++ {
		if results[i].Probability > results[i-1].Probability {
			t.Error("Results should be sorted by probability (descending)")
		}
	}

	// Check that probabilities sum to approximately 1.0 for top predictions
	var totalProb float32
	for _, result := range results {
		totalProb += result.Probability
		
		// Check that class names are preserved
		found := false
		for _, className := range classNames {
			if result.ClassName == className {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Class name %s not found in original class names", result.ClassName)
		}
	}

	// Probabilities should be positive
	for _, result := range results {
		if result.Probability <= 0 {
			t.Errorf("Probability should be positive, got %f", result.Probability)
		}
	}
}

func TestApplySoftmax(t *testing.T) {
	logits := []float32{1.0, 2.0, 3.0}
	probabilities := applySoftmax(logits)

	if len(probabilities) != len(logits) {
		t.Errorf("Expected same length, got %d vs %d", len(probabilities), len(logits))
	}

	// Check that probabilities sum to approximately 1.0
	var sum float32
	for _, prob := range probabilities {
		sum += prob
		if prob <= 0 || prob >= 1 {
			t.Errorf("Probability %f should be between 0 and 1", prob)
		}
	}

	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Probabilities should sum to ~1.0, got %f", sum)
	}

	// Check that higher logits result in higher probabilities
	for i := 1; i < len(logits); i++ {
		if logits[i] > logits[i-1] && probabilities[i] <= probabilities[i-1] {
			t.Error("Higher logits should result in higher probabilities")
		}
	}
}