#!/bin/bash

# DigitalOcean App Platform Setup Script
# Simple and cost-effective deployment for the Image Recognition Web App

set -e

echo "🚀 DigitalOcean App Platform Setup"
echo "=================================="
echo ""
echo "App Platform is the easiest way to deploy your webapp to DigitalOcean."
echo "It includes automatic builds, managed databases, and costs ~$12-25/month."
echo ""

# Check if doctl is installed
if ! command -v doctl &> /dev/null; then
    echo "❌ doctl CLI is not installed. Please install it first:"
    echo "   brew install doctl (macOS)"
    echo "   or visit: https://docs.digitalocean.com/reference/doctl/how-to/install/"
    exit 1
fi

# Check authentication
echo "🔐 Checking DigitalOcean authentication..."
if ! doctl account get &> /dev/null; then
    echo "❌ Not authenticated. Running 'doctl auth init'..."
    doctl auth init
fi

echo "✅ Authenticated as: $(doctl account get --format Email --no-header)"

# Function to create Spaces bucket
create_spaces_bucket() {
    echo ""
    echo "📦 Setting up DigitalOcean Spaces for model storage..."
    
    # Check if spaces exist
    if doctl spaces list --format Name --no-header | grep -q "image-recognition-models"; then
        echo "✅ Spaces bucket 'image-recognition-models' already exists"
    else
        echo "Creating Spaces bucket for model storage..."
        doctl spaces create image-recognition-models --region nyc3
        echo "✅ Spaces bucket created"
    fi
    
    echo ""
    echo "📝 You need to create Spaces access keys:"
    echo "   1. Go to: https://cloud.digitalocean.com/account/api/spaces"
    echo "   2. Click 'Generate New Key'"
    echo "   3. Save the Access Key ID and Secret Access Key"
    echo ""
    read -p "Press Enter when you have created the Spaces keys..."
}

# Function to create app using App Platform
create_app_platform_app() {
    echo ""
    echo "🏗️  Creating App Platform application..."
    
    # Check if app already exists
    if doctl apps list --format Name --no-header | grep -q "image-recognition-webapp"; then
        echo "✅ App 'image-recognition-webapp' already exists"
        APP_ID=$(doctl apps list --format ID,Name --no-header | grep "image-recognition-webapp" | awk '{print $1}')
        echo "   App ID: $APP_ID"
        return 0
    fi
    
    # Get GitHub repo info
    GITHUB_REPO=$(git config --get remote.origin.url | sed 's/.*github.com[/:]\([^.]*\).*/\1/')
    if [ -z "$GITHUB_REPO" ]; then
        echo "❌ Could not determine GitHub repository. Make sure you're in a git repo with GitHub origin."
        exit 1
    fi
    
    echo "📝 GitHub repository: $GITHUB_REPO"
    
    # Create app spec
    cat > app-spec.yaml << EOF
name: image-recognition-webapp
region: nyc3
services:
- name: webapp
  source_dir: /
  github:
    repo: $GITHUB_REPO
    branch: main
    deploy_on_push: true
  run_command: ./main
  environment_slug: go
  instance_count: 1
  instance_size_slug: basic-xxs
  http_port: 8080
  health_check:
    http_path: /health
    initial_delay_seconds: 60
    period_seconds: 10
    timeout_seconds: 5
    success_threshold: 1
    failure_threshold: 3
  routes:
  - path: /
  envs:
  - key: ENVIRONMENT
    value: production
  - key: PORT
    value: "8080"
  - key: GO_ENV
    value: production
  - key: LOG_LEVEL
    value: info
  - key: UPLOAD_MAX_SIZE
    value: "10485760"
  - key: UPLOAD_ALLOWED_TYPES
    value: "image/jpeg,image/png,image/webp"
  - key: UPLOAD_TEMP_DIR
    value: "/tmp/uploads"
  - key: MODEL_CACHE_PATH
    value: "/tmp/models"

databases:
- engine: PG
  name: webapp-db
  num_nodes: 1
  size: db-s-dev-database
  version: "15"

- engine: REDIS
  name: webapp-redis
  num_nodes: 1
  size: db-s-dev-database
  version: "7"

alerts:
- rule: CPU_UTILIZATION
  disabled: false
- rule: MEM_UTILIZATION
  disabled: false
- rule: RESTART_COUNT
  disabled: false
EOF
    
    echo "Creating App Platform application..."
    APP_ID=$(doctl apps create app-spec.yaml --format ID --no-header)
    
    if [ $? -eq 0 ]; then
        echo "✅ App created successfully!"
        echo "   App ID: $APP_ID"
        
        # Clean up
        rm app-spec.yaml
        
        echo ""
        echo "⏳ Your app is being built and deployed..."
        echo "   This usually takes 3-5 minutes for the first deployment."
        echo ""
        echo "🔗 Monitor progress at:"
        echo "   https://cloud.digitalocean.com/apps/$APP_ID"
        
    else
        echo "❌ Failed to create app"
        rm app-spec.yaml
        exit 1
    fi
}

# Function to display GitHub secrets
display_github_secrets() {
    echo ""
    echo "📋 GitHub Secrets to Configure"
    echo "=============================="
    echo "Go to: https://github.com/$GITHUB_REPO/settings/secrets/actions"
    echo ""
    echo "Add these secrets:"
    echo ""
    
    # Get DO token
    DO_TOKEN=$(doctl auth list --format AccessToken --no-header | head -1)
    
    echo "# Required for deployment"
    echo "DIGITALOCEAN_ACCESS_TOKEN=$DO_TOKEN"
    echo ""
    echo "# Spaces configuration (enter your keys from the console)"
    echo "SPACES_ACCESS_KEY=<your-spaces-access-key>"
    echo "SPACES_SECRET_KEY=<your-spaces-secret-key>"
    echo "SPACES_ENDPOINT=nyc3.digitaloceanspaces.com"
    echo "SPACES_BUCKET=image-recognition-models"
    echo ""
    echo "# Optional: Slack notifications"
    echo "SLACK_WEBHOOK=<your-slack-webhook-url>"
    echo ""
}

# Function to show next steps
show_next_steps() {
    GITHUB_REPO=$(git config --get remote.origin.url | sed 's/.*github.com[/:]\([^.]*\).*/\1/')
    APP_ID=$(doctl apps list --format ID,Name --no-header | grep "image-recognition-webapp" | awk '{print $1}' 2>/dev/null || echo "")
    
    echo ""
    echo "🎯 Next Steps"
    echo "============"
    echo ""
    echo "1. 📋 Configure GitHub Secrets (see above)"
    echo "2. 🔑 Add Spaces environment variables to your app:"
    echo "   - Go to: https://cloud.digitalocean.com/apps/$APP_ID/settings"
    echo "   - Add the Spaces environment variables as app-level secrets"
    echo ""
    echo "3. 🚀 Deploy your app:"
    echo "   - Push to main branch, or"
    echo "   - Use GitHub Actions: https://github.com/$GITHUB_REPO/actions"
    echo ""
    echo "4. 📊 Monitor your app:"
    echo "   - App Platform Console: https://cloud.digitalocean.com/apps/$APP_ID"
    echo "   - GitHub Actions: https://github.com/$GITHUB_REPO/actions"
    echo ""
    echo "💰 Estimated monthly cost: \$12-25 (much cheaper than Kubernetes!)"
    echo ""
    echo "🔧 Useful commands:"
    echo "   doctl apps list"
    echo "   doctl apps get $APP_ID"
    echo "   doctl apps logs $APP_ID webapp"
    echo "   doctl apps spec get $APP_ID"
}

# Main execution
echo "What would you like to do?"
echo ""
echo "1. Create Spaces bucket for model storage"
echo "2. Create App Platform application"
echo "3. Show GitHub secrets to configure"
echo "4. Do everything (recommended)"
echo "5. Show next steps"
echo "6. Exit"
echo ""

read -p "Enter your choice (1-6): " choice

case $choice in
    1)
        create_spaces_bucket
        ;;
    2)
        create_app_platform_app
        ;;
    3)
        GITHUB_REPO=$(git config --get remote.origin.url | sed 's/.*github.com[/:]\([^.]*\).*/\1/')
        display_github_secrets
        ;;
    4)
        create_spaces_bucket
        create_app_platform_app
        display_github_secrets
        show_next_steps
        ;;
    5)
        show_next_steps
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
echo "✨ App Platform setup complete!"
echo ""
echo "💡 Tip: App Platform automatically builds and deploys when you push to main branch!"
echo "    No need to manage containers, Kubernetes, or infrastructure."