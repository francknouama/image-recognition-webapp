# DigitalOcean Kubernetes Setup Guide

This guide will help you set up DigitalOcean Kubernetes clusters for deploying the Image Recognition Web Application.

## Prerequisites

- DigitalOcean account with billing enabled
- `doctl` CLI installed locally
- `kubectl` installed locally

## Step 1: Install and Configure doctl

```bash
# macOS
brew install doctl

# Linux
cd ~
wget https://github.com/digitalocean/doctl/releases/download/v1.104.0/doctl-1.104.0-linux-amd64.tar.gz
tar xf ~/doctl-1.104.0-linux-amd64.tar.gz
sudo mv ~/doctl /usr/local/bin

# Authenticate doctl
doctl auth init
```

## Step 2: Create Kubernetes Clusters

### Create Staging Cluster

```bash
# Create a small staging cluster (costs ~$36/month)
doctl kubernetes cluster create image-recognition-staging \
  --region nyc3 \
  --size s-2vcpu-2gb \
  --count 2 \
  --tag staging,image-recognition \
  --wait

# Save the cluster name
export STAGING_CLUSTER_NAME=$(doctl kubernetes cluster list --format Name --no-header | grep staging)
echo "Staging cluster name: $STAGING_CLUSTER_NAME"
```

### Create Production Cluster (Optional for now)

```bash
# Create a production cluster with autoscaling (costs ~$72-144/month)
doctl kubernetes cluster create image-recognition-production \
  --region nyc3 \
  --size s-2vcpu-4gb \
  --count 3 \
  --tag production,image-recognition \
  --auto-scale \
  --min-nodes 3 \
  --max-nodes 6 \
  --wait

# Save the cluster name
export PRODUCTION_CLUSTER_NAME=$(doctl kubernetes cluster list --format Name --no-header | grep production)
echo "Production cluster name: $PRODUCTION_CLUSTER_NAME"
```

## Step 3: Create Container Registry (if not already created)

```bash
# Create container registry
doctl registry create image-recognition-registry --region nyc3

# Get registry name
export REGISTRY_NAME=$(doctl registry get --format Name --no-header)
echo "Registry name: $REGISTRY_NAME"

# Configure Docker to use the registry
doctl registry login
```

## Step 4: Create DigitalOcean Spaces for Model Storage

```bash
# Create a Space for model storage
doctl spaces create image-recognition-models --region nyc3 --public

# Note the access keys (you'll need to create these in the web console)
echo "Go to: https://cloud.digitalocean.com/account/api/spaces"
echo "Create a new Spaces access key and save the credentials"
```

## Step 5: Create Database and Redis (for staging)

### Option A: Managed Database (Recommended but costs more)

```bash
# Create managed PostgreSQL database
doctl databases create image-recognition-db-staging \
  --engine pg \
  --region nyc3 \
  --size db-s-1vcpu-1gb \
  --version 15

# Create managed Redis
doctl databases create image-recognition-redis-staging \
  --engine redis \
  --region nyc3 \
  --size db-s-1vcpu-1gb \
  --version 7
```

### Option B: In-Cluster Database (Cheaper for testing)

For development/testing, you can deploy PostgreSQL and Redis within the cluster:

```bash
# Apply in-cluster database manifests (we'll create these)
kubectl apply -f k8s/dev/postgres.yaml
kubectl apply -f k8s/dev/redis.yaml
```

## Step 6: Configure GitHub Secrets

Go to your GitHub repository settings → Secrets and variables → Actions, and add:

### Required Secrets

```bash
# DigitalOcean Access
DIGITALOCEAN_ACCESS_TOKEN=<your-do-api-token>
REGISTRY_NAME=<your-registry-name>

# Kubernetes Clusters
DOKS_CLUSTER_NAME_STAGING=<staging-cluster-name>
DOKS_CLUSTER_NAME_PRODUCTION=<production-cluster-name>

# Spaces Configuration
SPACES_ACCESS_KEY=<your-spaces-access-key>
SPACES_SECRET_KEY=<your-spaces-secret-key>
SPACES_ENDPOINT=nyc3.digitaloceanspaces.com
SPACES_BUCKET=image-recognition-models

# Database URLs (staging)
DATABASE_URL_STAGING=postgresql://user:pass@host:5432/dbname
REDIS_URL_STAGING=redis://default:pass@host:6379

# Database URLs (production) - can be same as staging for now
DATABASE_URL_PRODUCTION=postgresql://user:pass@host:5432/dbname
REDIS_URL_PRODUCTION=redis://default:pass@host:6379

# Application Secrets
JWT_SECRET=<generate-a-random-secret>
API_KEY=<generate-a-random-api-key>

# Optional
SLACK_WEBHOOK=<your-slack-webhook-url>
LOAD_BALANCER_TAG_PRODUCTION=<lb-tag-name>
```

### Generate Random Secrets

```bash
# Generate JWT_SECRET
openssl rand -base64 32

# Generate API_KEY
openssl rand -hex 32
```

## Step 7: Test Cluster Connection

```bash
# Connect to staging cluster
doctl kubernetes cluster kubeconfig save <staging-cluster-name>

# Verify connection
kubectl cluster-info
kubectl get nodes
```

## Step 8: Initial Manual Deployment (Optional)

Before relying on CI/CD, you can test deployment manually:

```bash
# Create staging namespace
kubectl create namespace staging

# Apply configurations
kubectl apply -f k8s/webapp-deployment.yaml -n staging
kubectl apply -f k8s/webapp-service.yaml -n staging
kubectl apply -f k8s/webapp-hpa.yaml -n staging

# Check deployment status
kubectl get all -n staging

# Get external IP (may take a few minutes)
kubectl get service webapp-service -n staging
```

## Cost Optimization Tips

1. **Start with staging only** - Skip production cluster until needed
2. **Use smaller node sizes** - s-1vcpu-2gb nodes are sufficient for testing
3. **Use in-cluster databases** for development instead of managed databases
4. **Set up autoscaling** to scale down during low usage
5. **Use spot instances** if available in your region

## Monitoring and Maintenance

```bash
# Monitor cluster resources
doctl kubernetes cluster node-pool list <cluster-id>

# Check cluster costs
doctl billing history list

# Scale node pool manually if needed
doctl kubernetes cluster node-pool update <cluster-id> <pool-id> --count 3
```

## Cleanup (when needed)

```bash
# Delete clusters when not needed
doctl kubernetes cluster delete <cluster-name>

# Delete databases
doctl databases delete <database-id>

# Delete container registry (WARNING: deletes all images)
doctl registry delete
```

## Next Steps

1. Start with creating just the staging cluster
2. Configure the GitHub secrets
3. Push to main branch to trigger automatic deployment
4. Monitor the GitHub Actions workflow for any issues
5. Access your application at the LoadBalancer IP provided by the service

## Troubleshooting

If deployment fails, check:
1. GitHub Actions logs
2. Kubernetes pod logs: `kubectl logs -n staging deployment/webapp-deployment`
3. Service status: `kubectl describe service webapp-service -n staging`
4. Events: `kubectl get events -n staging --sort-by='.lastTimestamp'`