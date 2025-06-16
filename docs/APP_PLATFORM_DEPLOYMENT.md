# DigitalOcean App Platform Deployment Guide

This guide shows how to deploy the Image Recognition Web Application using DigitalOcean App Platform - the easiest and most cost-effective option.

## Why App Platform?

- ✅ **Simple**: No containers or Kubernetes to manage
- ✅ **Affordable**: ~$12-25/month vs $60-100+ for Kubernetes
- ✅ **Managed**: Automatic builds, deployments, SSL, scaling
- ✅ **Integrated**: Built-in databases, monitoring, and logs
- ✅ **Git-based**: Auto-deploy on push to main branch

## Quick Start

### Option 1: Automated Setup (Recommended)

```bash
# Run the setup script
./scripts/setup-app-platform.sh

# Choose option 4 (Do everything)
```

### Option 2: Manual Setup

1. **Create the app manually in the console**:
   - Go to [DigitalOcean App Platform](https://cloud.digitalocean.com/apps)
   - Click "Create App"
   - Connect your GitHub repository
   - Use the provided app spec (`.do/app.yaml`)

## What Gets Created

### Application Components

1. **Web Service** (`webapp`)
   - Go application serving the web interface
   - Auto-scaling based on traffic
   - Health checks and zero-downtime deployments
   - Custom domain support

2. **Databases**
   - **PostgreSQL**: Application data storage
   - **Redis**: Caching and session storage
   - Managed backups and monitoring

3. **Static Assets** (optional)
   - Served from CDN for better performance

### Estimated Costs

| Component | Size | Monthly Cost |
|-----------|------|--------------|
| Web Service (Basic XXS) | 0.5 vCPU, 0.5GB RAM | $5 |
| PostgreSQL (Dev) | 1 vCPU, 1GB RAM | $7 |
| Redis (Dev) | 1 vCPU, 1GB RAM | $7 |
| **Total** | | **~$19/month** |

*Prices may vary by region and are subject to change*

## Configuration

### Required GitHub Secrets

Add these to your repository secrets (`Settings → Secrets and variables → Actions`):

```bash
# DigitalOcean API access
DIGITALOCEAN_ACCESS_TOKEN=dop_v1_...

# Spaces configuration for model storage
SPACES_ACCESS_KEY=your_spaces_key
SPACES_SECRET_KEY=your_spaces_secret
SPACES_ENDPOINT=nyc3.digitaloceanspaces.com
SPACES_BUCKET=image-recognition-models

# Optional: Slack notifications
SLACK_WEBHOOK=https://hooks.slack.com/...
```

### Environment Variables

The app automatically configures these environment variables:

```bash
# Application settings
ENVIRONMENT=production
PORT=8080
GO_ENV=production
LOG_LEVEL=info

# Upload configuration
UPLOAD_MAX_SIZE=10485760        # 10MB
UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,image/webp
UPLOAD_TEMP_DIR=/tmp/uploads
MODEL_CACHE_PATH=/tmp/models

# Database connections (auto-injected)
DATABASE_URL=${webapp-db.DATABASE_URL}
REDIS_URL=${webapp-redis.DATABASE_URL}

# Spaces configuration (from secrets)
SPACES_ENDPOINT=${SPACES_ENDPOINT}
SPACES_BUCKET=${SPACES_BUCKET}
SPACES_ACCESS_KEY=${SPACES_ACCESS_KEY}
SPACES_SECRET_KEY=${SPACES_SECRET_KEY}
```

## Deployment Process

### Automatic Deployment

1. **Push to main branch**
2. **GitHub Actions** triggers the deployment workflow
3. **App Platform** automatically builds and deploys
4. **Health checks** verify the deployment
5. **Traffic** switches to the new version

### Manual Deployment

```bash
# Using doctl CLI
doctl apps create-deployment <app-id>

# Or update the app spec
doctl apps update <app-id> --spec .do/app.yaml
```

## Monitoring and Management

### App Platform Console

Access your app dashboard at:
```
https://cloud.digitalocean.com/apps/<app-id>
```

Features:
- Real-time logs
- Performance metrics
- Deployment history
- Environment variable management
- Database management

### CLI Commands

```bash
# List all apps
doctl apps list

# Get app details
doctl apps get <app-id>

# View logs
doctl apps logs <app-id> webapp

# View deployments
doctl apps list-deployments <app-id>

# Get app spec
doctl apps spec get <app-id>
```

### Health Checks

The app includes several health endpoints:

- `/health` - Basic application health
- `/api/health` - Detailed health with dependencies
- `/api/models` - Available models status

## Scaling and Performance

### Automatic Scaling

App Platform automatically scales based on:
- CPU usage
- Memory usage
- Request queue length

### Manual Scaling

```bash
# Scale web service to 3 instances
doctl apps update <app-id> --spec modified-spec.yaml
```

Or update the spec file:
```yaml
services:
- name: webapp
  instance_count: 3  # Scale to 3 instances
  instance_size_slug: basic-xs  # Upgrade to more resources
```

## Custom Domains

### Add Custom Domain

1. Go to your app settings in the console
2. Click "Settings" → "Domains"
3. Add your domain (e.g., `myapp.example.com`)
4. Update your DNS to point to the provided CNAME

### SSL Certificates

- **Automatic**: App Platform provides free SSL certificates
- **Custom**: Upload your own certificates if needed

## Database Management

### Access Databases

```bash
# Get connection details
doctl databases get <db-id>

# Connect to PostgreSQL
doctl databases db get <db-id> --format ConnURIPrivate --no-header

# Access Redis
doctl databases db get <redis-id> --format ConnURIPrivate --no-header
```

### Backups

- **Automatic daily backups** for PostgreSQL
- **Point-in-time recovery** available
- **Manual backups** can be triggered

### Database Migrations

For schema changes, you can:

1. **Use migration scripts** in your application
2. **Connect directly** to run SQL commands
3. **Use the console** database interface

## Troubleshooting

### Common Issues

1. **Build Failures**
   ```bash
   # Check build logs
   doctl apps logs <app-id> webapp --type build
   ```

2. **Runtime Errors**
   ```bash
   # Check runtime logs
   doctl apps logs <app-id> webapp --type run
   ```

3. **Health Check Failures**
   - Verify `/health` endpoint works locally
   - Check if application starts within 60 seconds
   - Ensure port 8080 is used

4. **Database Connection Issues**
   - Verify environment variables are set
   - Check database status in console
   - Review connection string format

### Debug Commands

```bash
# Get app status
doctl apps get <app-id> --format Phase,LiveURL

# View recent deployments
doctl apps list-deployments <app-id> --format ID,Phase,CreatedAt

# Get detailed app info
doctl apps get <app-id> --format . --output yaml
```

## Migration from Other Platforms

### From Kubernetes

1. **Export environment variables** from Kubernetes secrets
2. **Update connection strings** for managed databases
3. **Remove Kubernetes-specific configurations**
4. **Test the deployment** in App Platform

### From Docker/Containers

1. **Ensure your Dockerfile** builds correctly
2. **Update environment variables** as needed
3. **Configure health checks** properly
4. **Test locally** before deploying

## Cost Optimization

### Development/Staging

- Use **Basic XXS** instances ($5/month)
- Use **development database** sizes ($7/month each)
- **Scale down** when not in use

### Production

- Use **Basic XS** or **Basic S** instances
- Use **production database** sizes with backups
- Enable **auto-scaling** for traffic spikes

### Cost Monitoring

```bash
# Check current usage
doctl billing history list

# Monitor app metrics
doctl apps get <app-id> --format Phase,InstanceCount
```

## Security Best Practices

1. **Environment Variables**: Store secrets as encrypted environment variables
2. **Database Access**: Use private networking for database connections
3. **SSL/TLS**: Enforce HTTPS for all traffic
4. **Updates**: Keep dependencies updated automatically
5. **Monitoring**: Set up alerts for failures and unusual activity

## Backup and Recovery

### Application Backup

- **Code**: Stored in Git repository
- **Container Images**: Automatically built and stored
- **Deployment History**: Available in App Platform console

### Database Backup

```bash
# Create manual backup
doctl databases backups create <db-id>

# List backups
doctl databases backups list <db-id>

# Restore from backup
doctl databases restore <db-id> <backup-id>
```

## Next Steps

1. ✅ **Deploy to staging** using the setup script
2. ✅ **Test all functionality** thoroughly
3. ✅ **Configure monitoring** and alerts
4. ✅ **Set up custom domain** (optional)
5. ✅ **Scale for production** traffic when ready

## Support

- [DigitalOcean App Platform Docs](https://docs.digitalocean.com/products/app-platform/)
- [Community Forums](https://www.digitalocean.com/community/questions)
- [Support Tickets](https://cloud.digitalocean.com/support) (for paying customers)