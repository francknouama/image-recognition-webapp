package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/config"
	"github.com/francknouama/image-recognition-webapp/internal/handlers"
	"github.com/francknouama/image-recognition-webapp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg)

	logrus.Info("Starting image recognition web application...")

	// Initialize services
	imageService := services.NewImageService(cfg)
	modelService := services.NewModelService(cfg)
	predictionService := services.NewPredictionService(modelService, imageService)

	// Initialize handlers
	handlerConfig := &handlers.Config{
		ImageService:      imageService,
		PredictionService: predictionService,
		RateLimiter:      rate.NewLimiter(rate.Limit(cfg.Server.RateLimit), cfg.Server.RateBurst),
	}
	
	h := handlers.New(handlerConfig)

	// Setup router
	router := setupRouter(cfg, h)

	// Create HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start server in goroutine
	go func() {
		logrus.Infof("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func setupLogging(cfg *config.Config) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if cfg.Logging.Output == "file" && cfg.Logging.File != "" {
		file, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Warn("Failed to open log file, using stdout")
		} else {
			logrus.SetOutput(file)
		}
	}
}

func setupRouter(cfg *config.Config, h *handlers.Handler) http.Handler {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	})

	// Static files
	router.Static("/static", "./web/static")
	router.StaticFile("/favicon.ico", "./web/static/images/favicon.ico")

	// Health check
	router.GET("/health", h.HealthCheck)
	router.GET("/api/health", h.APIHealthCheck)

	// Main routes
	router.GET("/", h.Index)
	router.POST("/upload", h.Upload)
	router.GET("/results/:id", h.GetResults)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/predict", h.APIPredictImage)
		api.GET("/models", h.APIListModels)
		api.GET("/results/:id", h.APIGetResults)
	}

	return c.Handler(router)
}