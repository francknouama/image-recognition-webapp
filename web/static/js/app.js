// Image Recognition Web App JavaScript

// Alpine.js components and utilities
document.addEventListener('alpine:init', () => {
    // Image upload component
    Alpine.data('imageUpload', () => ({
        selectedFile: null,
        previewUrl: null,
        loading: false,
        dragOver: false,
        
        init() {
            console.log('Image upload component initialized');
        },
        
        setupDragAndDrop() {
            const uploadArea = this.$refs.uploadArea;
            
            uploadArea.addEventListener('dragover', (e) => {
                e.preventDefault();
                this.isDragging = true;
            });
            
            uploadArea.addEventListener('dragleave', (e) => {
                e.preventDefault();
                this.isDragging = false;
            });
            
            uploadArea.addEventListener('drop', (e) => {
                e.preventDefault();
                this.isDragging = false;
                
                const files = e.dataTransfer.files;
                if (files.length > 0) {
                    this.handleFileSelect(files[0]);
                }
            });
        },
        
        previewImage(event) {
            const file = event.target.files[0];
            if (file) {
                this.handleFileSelect(file);
            }
        },
        
        handleFileSelect(event) {
            const file = event.target.files?.[0];
            if (!file) return;
            
            // Validate file type
            if (!file.type.startsWith('image/')) {
                alert('Please select a valid image file.');
                return;
            }
            
            // Validate file size (10MB limit)
            if (file.size > 10 * 1024 * 1024) {
                alert('File size must be less than 10MB.');
                return;
            }
            
            this.selectedFile = file;
            
            // Create preview
            const reader = new FileReader();
            reader.onload = (e) => {
                this.previewUrl = e.target.result;
            };
            reader.readAsDataURL(file);
        },
        
        clearSelection() {
            this.selectedFile = null;
            this.previewUrl = null;
            this.$refs.fileInput.value = '';
        },
        
        formatFileSize(bytes) {
            if (!bytes) return '0 Bytes';
            
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },
        
        showError(message) {
            // Create and show error notification
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error-message fade-in';
            errorDiv.innerHTML = `
                <h3>Error</h3>
                <p>${message}</p>
            `;
            
            this.$refs.uploadArea.appendChild(errorDiv);
            
            setTimeout(() => {
                errorDiv.remove();
            }, 5000);
        }
    }));
    
    // Results component
    Alpine.data('predictionResults', () => ({
        results: null,
        isVisible: false,
        
        show(data) {
            this.results = data;
            this.isVisible = true;
        },
        
        hide() {
            this.isVisible = false;
            this.results = null;
        },
        
        getConfidenceColor(confidence) {
            if (confidence >= 0.8) return '#10b981'; // Green
            if (confidence >= 0.6) return '#f59e0b'; // Yellow
            return '#ef4444'; // Red
        }
    }));
});

// HTMX configuration and event handlers
document.addEventListener('DOMContentLoaded', function() {
    // Configure HTMX
    htmx.config.defaultSwapStyle = 'innerHTML';
    htmx.config.defaultSwapDelay = 100;
    htmx.config.defaultSettleDelay = 100;
    
    // Upload progress tracking
    document.body.addEventListener('htmx:xhr:progress', function(evt) {
        const progressBar = document.querySelector('.progress-fill');
        if (progressBar && evt.detail.lengthComputable) {
            const percentComplete = (evt.detail.loaded / evt.detail.total) * 100;
            progressBar.style.width = percentComplete + '%';
        }
    });
    
    // Request start
    document.body.addEventListener('htmx:beforeRequest', function(evt) {
        const uploadArea = document.querySelector('.upload-area');
        if (uploadArea) {
            uploadArea.classList.add('processing');
        }
        
        // Show progress bar
        const progressContainer = document.querySelector('.progress-container');
        if (progressContainer) {
            progressContainer.style.display = 'block';
        }
    });
    
    // Request complete
    document.body.addEventListener('htmx:afterRequest', function(evt) {
        const uploadArea = document.querySelector('.upload-area');
        if (uploadArea) {
            uploadArea.classList.remove('processing');
        }
        
        // Hide progress bar
        const progressContainer = document.querySelector('.progress-container');
        if (progressContainer) {
            setTimeout(() => {
                progressContainer.style.display = 'none';
            }, 1000);
        }
        
        // Animate results
        const results = document.querySelector('#results');
        if (results && results.children.length > 0) {
            results.classList.add('fade-in');
        }
    });
    
    // Error handling
    document.body.addEventListener('htmx:responseError', function(evt) {
        console.error('HTMX Response Error:', evt.detail);
        showNotification('An error occurred while processing your request.', 'error');
    });
    
    document.body.addEventListener('htmx:sendError', function(evt) {
        console.error('HTMX Send Error:', evt.detail);
        showNotification('Network error. Please check your connection.', 'error');
    });
});

// Utility functions
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type} fade-in`;
    notification.innerHTML = `
        <div class="notification-content">
            <span>${message}</span>
            <button onclick="this.parentElement.parentElement.remove()" class="notification-close">&times;</button>
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDuration(milliseconds) {
    if (milliseconds < 1000) {
        return milliseconds.toFixed(1) + ' ms';
    } else {
        return (milliseconds / 1000).toFixed(2) + ' s';
    }
}

// Prediction confidence visualization
function updateConfidenceBars() {
    const confidenceBars = document.querySelectorAll('.confidence-fill');
    confidenceBars.forEach((bar, index) => {
        const confidence = parseFloat(bar.dataset.confidence) || 0;
        setTimeout(() => {
            bar.style.width = (confidence * 100) + '%';
        }, index * 100);
    });
}

// Initialize confidence bars when results are loaded
const resultsObserver = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
        if (mutation.type === 'childList') {
            const addedNodes = Array.from(mutation.addedNodes);
            const hasResults = addedNodes.some(node => 
                node.nodeType === Node.ELEMENT_NODE && 
                (node.classList?.contains('results') || node.querySelector?.('.confidence-fill'))
            );
            
            if (hasResults) {
                setTimeout(updateConfidenceBars, 100);
            }
        }
    });
});

// Start observing results container
document.addEventListener('DOMContentLoaded', () => {
    const resultsContainer = document.querySelector('#results');
    if (resultsContainer) {
        resultsObserver.observe(resultsContainer, {
            childList: true,
            subtree: true
        });
    }
});

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
    // Ctrl/Cmd + U: Focus upload input
    if ((e.ctrlKey || e.metaKey) && e.key === 'u') {
        e.preventDefault();
        const fileInput = document.querySelector('input[type="file"]');
        if (fileInput) {
            fileInput.click();
        }
    }
    
    // Escape: Clear preview/results
    if (e.key === 'Escape') {
        const clearButton = document.querySelector('[x-on\\:click="clearPreview()"]');
        if (clearButton) {
            clearButton.click();
        }
    }
});

// Service worker registration for offline support (optional)
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/static/sw.js')
            .then((registration) => {
                console.log('SW registered: ', registration);
            })
            .catch((registrationError) => {
                console.log('SW registration failed: ', registrationError);
            });
    });
}