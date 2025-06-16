# Required GitHub Secrets Configuration

This document lists all the GitHub secrets that need to be configured for the CI/CD pipelines to work properly.

## Core DigitalOcean Configuration

### `DIGITALOCEAN_ACCESS_TOKEN`
- **Description**: DigitalOcean API token for authentication
- **Usage**: Used by doctl CLI and DigitalOcean actions
- **How to get**: Generate from DigitalOcean Control Panel > API > Tokens
- **Scope**: Read/Write access to all DigitalOcean resources

### `REGISTRY_NAME`
- **Description**: Name of your DigitalOcean Container Registry
- **Usage**: Used to build image paths for Docker push/pull
- **Example**: `my-company-registry`
- **How to get**: Create registry in DigitalOcean Control Panel > Container Registry

## DigitalOcean Spaces Configuration

### `SPACES_ACCESS_KEY`
- **Description**: Access key for DigitalOcean Spaces (S3-compatible storage)
- **Usage**: Used to upload/download ML models and artifacts
- **How to get**: Generate from DigitalOcean Control Panel > API > Spaces Keys

### `SPACES_SECRET_KEY`
- **Description**: Secret key for DigitalOcean Spaces
- **Usage**: Used with access key for Spaces authentication
- **How to get**: Generated together with access key

### `SPACES_ENDPOINT`
- **Description**: DigitalOcean Spaces endpoint URL
- **Usage**: Specifies which region's Spaces to use
- **Example**: `nyc3.digitaloceanspaces.com`
- **Options**: `nyc3.digitaloceanspaces.com`, `ams3.digitaloceanspaces.com`, `sgp1.digitaloceanspaces.com`, `sfo3.digitaloceanspaces.com`

### `SPACES_BUCKET`
- **Description**: Name of the Spaces bucket for storing ML models
- **Usage**: Where trained models are stored and retrieved from
- **Example**: `ml-models-production`
- **Note**: Must be created in DigitalOcean Spaces first

## Database Configuration

### `DATABASE_URL_STAGING`
- **Description**: PostgreSQL connection string for staging environment
- **Usage**: Used by staging deployments to connect to database
- **Format**: `postgresql://username:password@host:port/database?sslmode=require`
- **Example**: `postgresql://user:pass@db-postgresql-nyc3-12345-do-user-123456-0.b.db.ondigitalocean.com:25060/defaultdb?sslmode=require`

### `DATABASE_URL_PRODUCTION`
- **Description**: PostgreSQL connection string for production environment
- **Usage**: Used by production deployments to connect to database
- **Format**: Same as staging but pointing to production database
- **Security**: Should use different credentials than staging

### `REDIS_URL_STAGING`
- **Description**: Redis connection string for staging environment
- **Usage**: Used for caching and session storage in staging
- **Format**: `redis://default:password@host:port`
- **Example**: `redis://default:pass@db-redis-nyc3-67890-do-user-123456-0.b.db.ondigitalocean.com:25061`

### `REDIS_URL_PRODUCTION`
- **Description**: Redis connection string for production environment
- **Usage**: Used for caching and session storage in production
- **Format**: Same as staging but pointing to production Redis
- **Security**: Should use different credentials than staging

## Application Security

### `JWT_SECRET`
- **Description**: Secret key for signing JWT tokens
- **Usage**: Used for user authentication and API tokens
- **Requirements**: Should be a long, random string (at least 32 characters)
- **Example**: Generate with `openssl rand -base64 32`

### `API_KEY`
- **Description**: API key for external service authentication
- **Usage**: Used to authenticate API requests to the application
- **Requirements**: Should be a strong, unique key
- **Example**: Generate with `openssl rand -hex 24`

## Kubernetes Configuration

### `DOKS_CLUSTER_NAME_STAGING`
- **Description**: Name of the DigitalOcean Kubernetes cluster for staging
- **Usage**: Used by doctl to connect to the correct cluster
- **Example**: `image-recognition-staging`
- **Note**: Must match the actual cluster name in DigitalOcean

### `DOKS_CLUSTER_NAME_PRODUCTION`
- **Description**: Name of the DigitalOcean Kubernetes cluster for production
- **Usage**: Used by doctl to connect to the correct cluster
- **Example**: `image-recognition-production`
- **Note**: Must match the actual cluster name in DigitalOcean

### `LOAD_BALANCER_TAG_STAGING`
- **Description**: Tag used to identify the staging load balancer
- **Usage**: Used to find the load balancer IP for health checks
- **Example**: `image-recognition-lb-staging`
- **Note**: Must match the tag applied to your DigitalOcean Load Balancer

### `LOAD_BALANCER_TAG_PRODUCTION`
- **Description**: Tag used to identify the production load balancer
- **Usage**: Used to find the load balancer IP for health checks
- **Example**: `image-recognition-lb-production`
- **Note**: Must match the tag applied to your DigitalOcean Load Balancer

## Optional Integrations

### `SLACK_WEBHOOK`
- **Description**: Slack webhook URL for deployment notifications
- **Usage**: Used to send success/failure notifications to Slack
- **Format**: `https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK`
- **How to get**: Create a Slack app and incoming webhook

### `GITHUB_TOKEN`
- **Description**: GitHub personal access token (usually auto-provided)
- **Usage**: Used for API calls to GitHub (creating releases, etc.)
- **Note**: Often auto-provided by GitHub Actions as `secrets.GITHUB_TOKEN`
- **Scope**: If custom token needed, requires `repo` scope

## Environment-Specific Secrets

The secrets should be configured at different levels:

1. **Repository Level**: Core secrets like DigitalOcean tokens, Spaces keys
2. **Environment Level**: Database URLs, cluster names (different per environment)

### Repository Secrets
- `DIGITALOCEAN_ACCESS_TOKEN`
- `REGISTRY_NAME`
- `SPACES_ACCESS_KEY`
- `SPACES_SECRET_KEY`
- `SPACES_ENDPOINT`
- `SPACES_BUCKET`
- `JWT_SECRET`
- `API_KEY`
- `SLACK_WEBHOOK` (optional)

### Environment-Specific Secrets

#### Staging Environment
- `DATABASE_URL_STAGING`
- `REDIS_URL_STAGING`
- `DOKS_CLUSTER_NAME_STAGING`
- `LOAD_BALANCER_TAG_STAGING`

#### Production Environment
- `DATABASE_URL_PRODUCTION`
- `REDIS_URL_PRODUCTION`
- `DOKS_CLUSTER_NAME_PRODUCTION`
- `LOAD_BALANCER_TAG_PRODUCTION`

## Security Best Practices

1. **Rotate Secrets Regularly**: Change API keys and tokens periodically
2. **Use Different Credentials**: Staging and production should use separate credentials
3. **Minimum Permissions**: Grant only the required permissions to each token
4. **Environment Separation**: Use GitHub Environment protection rules for production
5. **Monitor Usage**: Review DigitalOcean API usage and database access logs

## Verification Commands

To verify your secrets are working correctly:

```bash
# Test DigitalOcean API access
doctl account get

# Test Spaces access
s3cmd ls s3://your-bucket-name

# Test database connection
psql "$DATABASE_URL" -c "SELECT version();"

# Test Redis connection
redis-cli -u "$REDIS_URL" ping

# Test Kubernetes access
doctl kubernetes cluster kubeconfig save your-cluster-name
kubectl get nodes
```

## Troubleshooting

### Common Issues

1. **Invalid DigitalOcean Token**: Check token is not expired and has correct permissions
2. **Spaces Access Denied**: Verify access key, secret key, and bucket name
3. **Database Connection Failed**: Check firewall rules and connection string format
4. **Kubernetes Access Denied**: Verify cluster name and RBAC permissions
5. **Load Balancer Not Found**: Check tag name matches exactly

### Debug Commands

```bash
# Check if secrets are set (GitHub Actions)
echo "DIGITALOCEAN_ACCESS_TOKEN is set: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN != '' }}"

# Test Spaces connection
s3cmd ls s3:// --host=$SPACES_ENDPOINT

# Test database connectivity
timeout 10 bash -c 'cat < /dev/null > /dev/tcp/hostname/5432'
```