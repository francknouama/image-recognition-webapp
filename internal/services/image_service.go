package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/francknouama/image-recognition-webapp/internal/config"
	"github.com/francknouama/image-recognition-webapp/internal/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/webp"
)

// ImageService handles image processing operations
type ImageService struct {
	config *config.Config
	logger *logrus.Logger
}

// NewImageService creates a new image service
func NewImageService(cfg *config.Config) *ImageService {
	return &ImageService{
		config: cfg,
		logger: logrus.New(),
	}
}

// ValidateImage validates an uploaded image file
func (s *ImageService) ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > s.config.Upload.MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", 
			header.Size, s.config.Upload.MaxFileSize)
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	if !s.isAllowedType(contentType) {
		return fmt.Errorf("unsupported file type: %s", contentType)
	}

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Read first 512 bytes to detect actual file type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Reset file pointer again
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Detect actual MIME type
	detectedType := s.detectMimeType(buffer[:n])
	if !s.isAllowedType(detectedType) {
		return fmt.Errorf("detected file type %s is not allowed", detectedType)
	}

	return nil
}

// ProcessImage processes an uploaded image and returns metadata
func (s *ImageService) ProcessImage(file multipart.File, header *multipart.FileHeader) (*models.ImageMetadata, []byte, error) {
	// Validate the image first
	if err := s.ValidateImage(file, header); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Decode image to get dimensions
	img, format, err := s.decodeImage(bytes.NewReader(fileData))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create metadata
	metadata := &models.ImageMetadata{
		Filename:    header.Filename,
		Size:        header.Size,
		Width:       img.Bounds().Dx(),
		Height:      img.Bounds().Dy(),
		Format:      format,
		ContentType: header.Header.Get("Content-Type"),
		UploadedAt:  time.Now(),
	}

	// Preprocess image for model input
	processedData, err := s.preprocessForModel(img)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to preprocess image: %w", err)
	}

	s.logger.Infof("Processed image: %s (%dx%d, %s, %d bytes)", 
		metadata.Filename, metadata.Width, metadata.Height, metadata.Format, metadata.Size)

	return metadata, processedData, nil
}

// SaveTempFile saves image data to a temporary file
func (s *ImageService) SaveTempFile(data []byte, filename string) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(filename)
	name := fmt.Sprintf("%d_%s%s", time.Now().Unix(), 
		strings.TrimSuffix(filename, ext), ext)
	
	tempPath := filepath.Join(s.config.Upload.TempDir, name)

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(s.config.Upload.TempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	s.logger.Debugf("Saved temp file: %s", tempPath)
	return tempPath, nil
}

// CleanupTempFiles removes old temporary files
func (s *ImageService) CleanupTempFiles() error {
	entries, err := os.ReadDir(s.config.Upload.TempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	now := time.Now()
	cleanupThreshold := time.Duration(s.config.Upload.CleanupAfter) * time.Second

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > cleanupThreshold {
			filePath := filepath.Join(s.config.Upload.TempDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				s.logger.Warnf("Failed to remove temp file %s: %v", filePath, err)
			} else {
				s.logger.Debugf("Cleaned up temp file: %s", filePath)
			}
		}
	}

	return nil
}

// ResizeImage resizes an image to the specified dimensions
func (s *ImageService) ResizeImage(img image.Image, width, height int) image.Image {
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

// ConvertToRGB converts an image to RGB format
func (s *ImageService) ConvertToRGB(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	
	return rgba
}

// isAllowedType checks if the content type is allowed
func (s *ImageService) isAllowedType(contentType string) bool {
	for _, allowedType := range s.config.Upload.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

// detectMimeType detects MIME type from file content
func (s *ImageService) detectMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// JPEG
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg"
	}

	// PNG
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return "image/png"
	}

	// WebP
	if len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && 
		bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}

	return "application/octet-stream"
}

// decodeImage decodes an image from a reader
func (s *ImageService) decodeImage(reader io.Reader) (image.Image, string, error) {
	// Try to decode as different formats
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}

	// Try PNG first
	img, err := png.Decode(bytes.NewReader(data))
	if err == nil {
		return img, "png", nil
	}

	// Try JPEG
	img, err = jpeg.Decode(bytes.NewReader(data))
	if err == nil {
		return img, "jpeg", nil
	}

	// Try WebP
	img, err = webp.Decode(bytes.NewReader(data))
	if err == nil {
		return img, "webp", nil
	}

	return nil, "", fmt.Errorf("unsupported image format")
}

// preprocessForModel preprocesses an image for model input
func (s *ImageService) preprocessForModel(img image.Image) ([]byte, error) {
	// Resize to standard input size (224x224 for most models)
	resized := s.ResizeImage(img, 224, 224)
	
	// Convert to RGB
	rgba := s.ConvertToRGB(resized)
	
	// Convert to JPEG format for consistency
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 95}); err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}
	
	return buf.Bytes(), nil
}

// GetImageThumbnail generates a thumbnail for an image
func (s *ImageService) GetImageThumbnail(img image.Image, size int) ([]byte, error) {
	thumbnail := imaging.Thumbnail(img, size, size, imaging.Lanczos)
	
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}
	
	return buf.Bytes(), nil
}