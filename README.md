# Image Recognition Web Application

A modern, high-performance web application for image classification using Go, TEMPL, HTMX, and AlpineJS. This application provides a user-friendly interface for uploading images and getting AI-powered classification results.

## Features

- **Fast Image Processing**: Efficient image handling with validation and preprocessing
- **Multiple Model Support**: Support for various machine learning models
- **Modern Frontend**: Server-side rendering with TEMPL, dynamic updates with HTMX, and client-side interactivity with AlpineJS
- **Responsive Design**: Mobile-first design using PicoCSS
- **RESTful API**: JSON API for programmatic access
- **Production Ready**: Docker support, comprehensive logging, monitoring, and security features
- **Real-time Updates**: Asynchronous image processing with progress tracking

## Tech Stack

- **Backend**: Go 1.24 with Gin web framework
- **Frontend**: TEMPL templates, HTMX, AlpineJS, PicoCSS
- **Image Processing**: Go imaging libraries with format support for JPEG, PNG, WebP
- **Containerization**: Docker with multi-stage builds
- **CI/CD**: GitHub Actions with automated testing and deployment

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker (optional)
- Make (optional, for build automation)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/francknouama/image-recognition-webapp.git
   cd image-recognition-webapp
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Create necessary directories**
   ```bash
   make setup-dirs
   ```

5. **Run the application**
   ```bash
   make run
   ```

The application will be available at `http://localhost:8080`.

## Development

### Using Make

```bash
# Build the application
make build

# Run in development mode with hot reload
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt

# Setup development environment
make setup
```

### Using Docker

```bash
# Build and run with Docker
make docker-build
make docker-run

# Or use Docker Compose for development
docker-compose --profile dev up
```

### Development with Hot Reload

```bash
# Install air for hot reload
go install github.com/air-verse/air@latest

# Run with hot reload
make dev
```

## API Endpoints

### Web Interface

- `GET /` - Main upload interface
- `POST /upload` - File upload and processing (HTMX compatible)
- `GET /results/{id}` - View prediction results

### REST API

- `POST /api/predict` - Image prediction (JSON)
- `GET /api/results/{id}` - Get prediction results (JSON)
- `GET /api/models` - List available models
- `GET /api/health` - Detailed health check

### Health Checks

- `GET /health` - Basic health check
- `GET /api/health` - Detailed health check with model status

## Configuration

Configuration is handled through environment variables. See `.env.example` for all available options.

### Key Configuration Options

```bash
# Server
PORT=8080
ENVIRONMENT=development

# File Upload
MAX_FILE_SIZE=10485760  # 10MB
ALLOWED_TYPES=image/jpeg,image/png,image/webp

# Models
MODEL_PATH=./models
MODEL_VERSION=latest

# Rate Limiting
RATE_LIMIT=10.0
RATE_BURST=20
```

## Usage Examples

### Web Interface

1. Open `http://localhost:8080` in your browser
2. Click or drag and drop an image file
3. Wait for processing to complete
4. View the classification results with confidence scores

### API Usage

**Upload and classify an image:**

```bash
curl -X POST \
  -F "image=@path/to/your/image.jpg" \
  http://localhost:8080/api/predict
```

**Get available models:**

```bash
curl http://localhost:8080/api/models
```

**Check health:**

```bash
curl http://localhost:8080/api/health
```

## Model Integration

The application supports loading machine learning models from the ML pipeline repository. Models should be placed in the `models/` directory with the following structure:

```
models/
├── model-name/
│   ├── saved_model/
│   └── metadata.json
```

### Model Metadata Format

```json
{
  "id": "model-name",
  "name": "Human Readable Model Name",
  "version": "1.0.0",
  "description": "Model description",
  "input_shape": [224, 224, 3],
  "output_shape": [1000],
  "classes": ["class1", "class2", "..."]
}
```

## Deployment

### Docker Deployment

```bash
# Build production image
docker build -t image-recognition-webapp .

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/models:/app/models:ro \
  -e ENVIRONMENT=production \
  image-recognition-webapp
```

### Docker Compose Deployment

```bash
# Production deployment
docker-compose up -d

# With monitoring stack
docker-compose --profile monitoring up -d

# With nginx reverse proxy
docker-compose --profile nginx up -d
```

### Kubernetes Deployment

Kubernetes manifests are available in the `deployments/k8s/` directory:

```bash
kubectl apply -f deployments/k8s/
```

## Monitoring and Observability

### Health Checks

The application provides comprehensive health checks:

- Basic health endpoint for load balancers
- Detailed health with model status and dependencies
- Docker health check configured

### Logging

- Structured JSON logging
- Configurable log levels
- Request/response logging with correlation IDs

### Metrics

- Built-in metrics for monitoring
- Prometheus-compatible endpoints (when enabled)
- Custom application metrics

## Security

### File Upload Security

- File type validation (MIME type detection)
- File size limits
- Secure file handling with temporary storage
- Input sanitization

### Application Security

- Rate limiting
- CORS configuration
- Security headers
- Non-root container execution

### Environment Security

- Secrets management through environment variables
- Container security scanning in CI/CD
- Regular dependency updates

## CI/CD Configuration

### GitHub Actions Setup

The CI/CD pipeline requires the following GitHub Secrets to be configured in your repository settings:

#### Required Secrets for Full Pipeline

- `DIGITALOCEAN_ACCESS_TOKEN`: Your DigitalOcean API token for registry access
- `REGISTRY_NAME`: Your DigitalOcean Container Registry name

#### Pipeline Stages

1. **Test Stage** (runs on all branches):
   - Go linting and formatting checks
   - Unit tests with coverage
   - Security scanning with gosec
   - No secrets required

2. **Docker Stage** (runs on all branches):
   - Builds Docker image
   - Runs Trivy vulnerability scan
   - Tests container health
   - No secrets required

3. **Build and Push Stage** (only on main branch):
   - Requires `DIGITALOCEAN_ACCESS_TOKEN` and `REGISTRY_NAME`
   - Pushes to DigitalOcean Container Registry

### Running Without Registry Access

If you don't have DigitalOcean Container Registry access, the pipeline will still:
- Run all tests and quality checks
- Build and scan Docker images locally
- Only fail on the registry push step (which only runs on main branch)

To disable registry push entirely, comment out the `build-and-push` job in `.github/workflows/ci.yml`.

## Testing

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# Benchmarks
make bench

# Security scan
make security
```

### Test Coverage

The project maintains >80% test coverage. View coverage reports:

```bash
make test-coverage
open coverage.html
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Write tests for new functionality
- Update documentation for API changes
- Run linting and formatting before committing

## Performance

### Benchmarks

- Image processing: ~100ms average for 224x224 images
- API response time: <200ms for typical requests
- Memory usage: <100MB baseline
- Concurrent requests: Supports 1000+ concurrent connections

### Optimization Features

- Efficient image processing pipeline
- Connection pooling and keep-alive
- Gzip compression
- Static file caching
- Graceful shutdown

## Troubleshooting

### Common Issues

**Application won't start:**
- Check port availability
- Verify environment configuration
- Check file permissions for uploads directory

**Image upload fails:**
- Verify file size limits
- Check supported file formats
- Ensure sufficient disk space

**Model loading errors:**
- Verify model file paths
- Check model metadata format
- Review application logs

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
./image-recognition-webapp
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/) for the HTTP router
- [HTMX](https://htmx.org/) for seamless AJAX interactions
- [Alpine.js](https://alpinejs.dev/) for lightweight reactivity
- [PicoCSS](https://picocss.com/) for minimal, semantic styling
- Go community for excellent imaging and HTTP libraries