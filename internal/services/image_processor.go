package services

import (
	"fmt"
	"image"
	"image/color"
	"bytes"
	"math"

	"github.com/disintegration/imaging"
)

// ImageProcessor handles image preprocessing for TensorFlow models
type ImageProcessor struct {
	targetWidth  int
	targetHeight int
	normalize    bool
	meanValues   []float32
	stdValues    []float32
}

// NewImageProcessor creates a new image processor
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		targetWidth:  224,
		targetHeight: 224,
		normalize:    true,
		// ImageNet normalization values
		meanValues: []float32{0.485, 0.456, 0.406},
		stdValues:  []float32{0.229, 0.224, 0.225},
	}
}

// SetTargetSize sets the target dimensions for preprocessing
func (p *ImageProcessor) SetTargetSize(width, height int) {
	p.targetWidth = width
	p.targetHeight = height
}

// SetNormalization configures normalization parameters
func (p *ImageProcessor) SetNormalization(normalize bool, meanValues, stdValues []float32) {
	p.normalize = normalize
	if meanValues != nil {
		p.meanValues = meanValues
	}
	if stdValues != nil {
		p.stdValues = stdValues
	}
}

// ProcessImageBytes converts image bytes to TensorFlow-ready tensor data
func (p *ImageProcessor) ProcessImageBytes(imageData []byte) ([][]float32, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return p.ProcessImage(img)
}

// ProcessImage converts an image.Image to TensorFlow-ready tensor data
func (p *ImageProcessor) ProcessImage(img image.Image) ([][]float32, error) {
	// Resize image to target dimensions
	resized := imaging.Resize(img, p.targetWidth, p.targetHeight, imaging.Lanczos)

	// Convert to RGB if necessary
	rgbImg := imaging.Clone(resized)

	// Convert to tensor format [1, height, width, channels]
	bounds := rgbImg.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Create tensor data
	tensorData := make([]float32, height*width*3)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := color.RGBAModel.Convert(rgbImg.At(x, y)).(color.RGBA)
			
			// Calculate index for HWC format
			baseIdx := (y*width + x) * 3
			
			// Convert to float32 and normalize to [0, 1]
			r := float32(pixel.R) / 255.0
			g := float32(pixel.G) / 255.0
			b := float32(pixel.B) / 255.0
			
			// Apply normalization if enabled
			if p.normalize {
				r = (r - p.meanValues[0]) / p.stdValues[0]
				g = (g - p.meanValues[1]) / p.stdValues[1]
				b = (b - p.meanValues[2]) / p.stdValues[2]
			}
			
			tensorData[baseIdx] = r
			tensorData[baseIdx+1] = g
			tensorData[baseIdx+2] = b
		}
	}

	// Reshape to batch format [1, height, width, channels]
	batchData := make([][]float32, 1)
	batchData[0] = tensorData

	return batchData, nil
}

// ProcessImageForBatch converts multiple images to tensor format
func (p *ImageProcessor) ProcessImageForBatch(images []image.Image) ([][][]float32, error) {
	batchSize := len(images)
	if batchSize == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	// Process each image
	var batchData [][][]float32
	for _, img := range images {
		tensorData, err := p.ProcessImage(img)
		if err != nil {
			return nil, fmt.Errorf("failed to process image: %w", err)
		}
		batchData = append(batchData, tensorData)
	}

	return batchData, nil
}

// GetInputShape returns the expected input shape for the processor
func (p *ImageProcessor) GetInputShape() []int {
	return []int{1, p.targetHeight, p.targetWidth, 3}
}

// PostprocessPredictions converts raw model outputs to classification results
func (p *ImageProcessor) PostprocessPredictions(predictions []float32, classNames []string, topK int) ([]ClassificationPrediction, error) {
	if len(predictions) != len(classNames) {
		return nil, fmt.Errorf("predictions length (%d) does not match class names length (%d)", 
			len(predictions), len(classNames))
	}

	if topK <= 0 {
		topK = 5
	}

	// Apply softmax to get probabilities
	softmaxPreds := applySoftmax(predictions)

	// Create prediction structs
	var results []ClassificationPrediction
	for i, prob := range softmaxPreds {
		if i < len(classNames) {
			results = append(results, ClassificationPrediction{
				ClassIndex:  i,
				ClassName:   classNames[i],
				Probability: prob,
				Confidence:  prob, // Using probability as confidence for now
			})
		}
	}

	// Sort by probability (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Probability > results[i].Probability {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Return top K results
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// ClassificationPrediction represents a single classification result
type ClassificationPrediction struct {
	ClassIndex  int     `json:"class_index"`
	ClassName   string  `json:"class_name"`
	Probability float32 `json:"probability"`
	Confidence  float32 `json:"confidence"`
}

// applySoftmax applies softmax activation to convert logits to probabilities
func applySoftmax(logits []float32) []float32 {
	// Find max value for numerical stability
	maxVal := logits[0]
	for _, val := range logits {
		if val > maxVal {
			maxVal = val
		}
	}

	// Calculate exp(x - max) and sum
	var expSum float32
	expVals := make([]float32, len(logits))
	
	for i, val := range logits {
		expVals[i] = float32(fastExp(float64(val - maxVal)))
		expSum += expVals[i]
	}

	// Normalize
	probabilities := make([]float32, len(logits))
	for i, expVal := range expVals {
		probabilities[i] = expVal / expSum
	}

	return probabilities
}

// fastExp is a fast approximation of exp function
func fastExp(x float64) float64 {
	if x < -700 {
		return 0
	}
	if x > 700 {
		return 1e300
	}
	
	return math.Exp(x)
}