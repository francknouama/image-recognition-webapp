#!/bin/bash

# DigitalOcean Kubernetes Quick Setup Script
# This script helps set up a minimal staging environment on DigitalOcean

set -e

echo "🚀 DigitalOcean Kubernetes Setup for Image Recognition Web App"
echo "============================================================"

# Check if doctl is installed
if ! command -v doctl &> /dev/null; then
    echo "❌ doctl CLI is not installed. Please install it first:"
    echo "   brew install doctl (macOS)"
    echo "   or visit: https://docs.digitalocean.com/reference/doctl/how-to/install/"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install it first:"
    echo "   brew install kubectl (macOS)"
    exit 1
fi

# Check authentication
echo "🔐 Checking DigitalOcean authentication..."
if ! doctl account get &> /dev/null; then
    echo "❌ Not authenticated. Running 'doctl auth init'..."
    doctl auth init
fi

echo "✅ Authenticated as: $(doctl account get --format Email --no-header)"

# Function to create staging cluster
create_staging_cluster() {
    echo ""
    echo "📦 Creating Kubernetes cluster for staging..."
    echo "This will create a small 2-node cluster (~$36/month)"
    read -p "Continue? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        doctl kubernetes cluster create image-recognition-staging \
            --region nyc3 \
            --size s-2vcpu-2gb \
            --count 2 \
            --tag staging,image-recognition \
            --wait
        
        echo "✅ Staging cluster created successfully!"
    else
        echo "⏭️  Skipping cluster creation"
    fi
}

# Function to create container registry
create_container_registry() {
    echo ""
    echo "📦 Creating Container Registry..."
    
    # Check if registry exists
    if doctl registry get &> /dev/null; then
        echo "✅ Container registry already exists"
        REGISTRY_NAME=$(doctl registry get --format Name --no-header)
    else
        read -p "Create container registry? (y/n) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            doctl registry create image-recognition-registry --region nyc3
            REGISTRY_NAME=$(doctl registry get --format Name --no-header)
            echo "✅ Container registry created: $REGISTRY_NAME"
        fi
    fi
    
    echo "📝 Registry name: $REGISTRY_NAME"
}

# Function to setup in-cluster databases
setup_databases() {
    echo ""
    echo "🗄️  Setting up in-cluster databases (PostgreSQL and Redis)..."
    
    # Create namespace
    kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply database manifests
    kubectl apply -f k8s/dev/postgres.yaml
    kubectl apply -f k8s/dev/redis.yaml
    
    echo "✅ Databases deployed to cluster"
    echo ""
    echo "📝 Database connection strings:"
    echo "   PostgreSQL: postgresql://webapp:devpassword123@postgres:5432/imagerecognition"
    echo "   Redis: redis://redis:6379"
}

# Function to display GitHub secrets
display_github_secrets() {
    echo ""
    echo "📋 GitHub Secrets to Configure"
    echo "=============================="
    echo "Go to: https://github.com/YOUR_USERNAME/image-recognition-webapp/settings/secrets/actions"
    echo ""
    echo "Add these secrets:"
    echo ""
    
    # Get values
    DO_TOKEN=$(doctl auth list --format AccessToken --no-header | head -1)
    CLUSTER_NAME=$(doctl kubernetes cluster list --format Name --no-header | grep staging | head -1)
    REGISTRY_NAME=$(doctl registry get --format Name --no-header 2>/dev/null || echo "")
    
    echo "DIGITALOCEAN_ACCESS_TOKEN=$DO_TOKEN"
    echo "REGISTRY_NAME=$REGISTRY_NAME"
    echo "DOKS_CLUSTER_NAME_STAGING=$CLUSTER_NAME"
    echo ""
    echo "# Spaces (create these in DO console)"
    echo "SPACES_ACCESS_KEY=<create-in-console>"
    echo "SPACES_SECRET_KEY=<create-in-console>"
    echo "SPACES_ENDPOINT=nyc3.digitaloceanspaces.com"
    echo "SPACES_BUCKET=image-recognition-models"
    echo ""
    echo "# Database URLs (for in-cluster databases)"
    echo "DATABASE_URL_STAGING=postgresql://webapp:devpassword123@postgres:5432/imagerecognition"
    echo "REDIS_URL_STAGING=redis://redis:6379"
    echo "DATABASE_URL_PRODUCTION=postgresql://webapp:devpassword123@postgres:5432/imagerecognition"
    echo "REDIS_URL_PRODUCTION=redis://redis:6379"
    echo ""
    echo "# Application secrets (generate these)"
    echo "JWT_SECRET=$(openssl rand -base64 32)"
    echo "API_KEY=$(openssl rand -hex 32)"
    echo ""
}

# Main menu
echo ""
echo "Select what you want to set up:"
echo "1. Create staging Kubernetes cluster"
echo "2. Create container registry"
echo "3. Setup in-cluster databases"
echo "4. Display GitHub secrets to configure"
echo "5. Do everything (recommended for first time)"
echo "6. Exit"
echo ""

read -p "Enter your choice (1-6): " choice

case $choice in
    1)
        create_staging_cluster
        ;;
    2)
        create_container_registry
        ;;
    3)
        # Make sure we're connected to cluster
        CLUSTER_NAME=$(doctl kubernetes cluster list --format Name --no-header | grep staging | head -1)
        if [ -n "$CLUSTER_NAME" ]; then
            doctl kubernetes cluster kubeconfig save "$CLUSTER_NAME"
            setup_databases
        else
            echo "❌ No staging cluster found. Please create one first."
        fi
        ;;
    4)
        display_github_secrets
        ;;
    5)
        create_staging_cluster
        create_container_registry
        
        # Connect to the new cluster
        CLUSTER_NAME=$(doctl kubernetes cluster list --format Name --no-header | grep staging | head -1)
        if [ -n "$CLUSTER_NAME" ]; then
            echo "🔗 Connecting to cluster..."
            doctl kubernetes cluster kubeconfig save "$CLUSTER_NAME"
            setup_databases
        fi
        
        display_github_secrets
        
        echo ""
        echo "✅ Setup complete!"
        echo ""
        echo "Next steps:"
        echo "1. Configure the GitHub secrets shown above"
        echo "2. Push to main branch to trigger deployment"
        echo "3. Monitor the deployment in GitHub Actions"
        echo ""
        ;;
    6)
        echo "👋 Exiting..."
        exit 0
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "🎯 Quick Commands:"
echo "  Check cluster: kubectl get nodes"
echo "  Check namespace: kubectl get all -n staging"
echo "  Get service IP: kubectl get svc webapp-service -n staging"
echo "  View logs: kubectl logs -n staging deployment/webapp-deployment"