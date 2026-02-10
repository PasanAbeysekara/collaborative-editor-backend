# GitHub Actions CI/CD Setup

This repository uses GitHub Actions for automated CI/CD workflows. The pipeline follows the architecture:

1. **Developer pushes to GitHub**
2. **CI Workflow** - Builds and pushes Docker images to GitHub Container Registry
3. **CD Workflow** - Deploys to Kubernetes cluster using kubectl

## Workflows

### 1. CI - Build and Push Docker Images (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `dev` branches
- Pull requests to `main` or `dev` branches

**What it does:**
- Builds Docker images for all microservices (user-service, document-service, realtime-service, notification-service)
- Pushes images to GitHub Container Registry (ghcr.io)
- Creates multi-platform images (linux/amd64, linux/arm64)
- Uses caching for faster builds

**Image Naming Convention:**
```
ghcr.io/<owner>/<repo>/<service-name>:<tag>
```

**Tags created:**
- Branch name (e.g., `main`, `dev`)
- Branch + SHA (e.g., `main-abc1234`)
- `latest` (only for default branch)

### 2. CD - Deploy to Kubernetes (`.github/workflows/cd.yml`)

**Triggers:**
- Automatically after successful CI workflow completion
- Manual trigger via workflow_dispatch

**What it does:**
- Updates Kubernetes deployments with new image tags
- Applies infrastructure manifests (Redis, RabbitMQ)
- Creates secrets for the application
- Deploys all microservices
- Deploys NGINX ingress controller
- Waits for rollout completion
- Verifies deployment status

**Environments:**
- `dev` - Deploys when pushing to dev branch
- `production` - Deploys when pushing to main branch

### 3. Pull Request Checks (`.github/workflows/pr-checks.yml`)

**Triggers:**
- Pull requests to `main` or `dev` branches

**What it does:**
- Runs `go vet` and `staticcheck` for code quality
- Executes unit tests with race detection
- Generates code coverage reports
- Builds all services to ensure they compile
- Checks code formatting with `gofmt`
- Tests Docker builds for all services
- Runs security scans (Trivy, gosec)
- Validates Kubernetes YAML manifests

## Required GitHub Secrets

You need to configure the following secrets in your GitHub repository:

### Repository Secrets

1. **KUBE_CONFIG** (Required for CD)
   - Base64-encoded Kubernetes config file
   - To generate:
     ```bash
     cat ~/.kube/config | base64 -w 0
     ```
   - This allows GitHub Actions to connect to your Kubernetes cluster

2. **DATABASE_URL** (Required for deployment)
   - PostgreSQL connection string
   - Example: `postgresql://user:password@postgres-service:5432/dbname?sslmode=disable`

3. **JWT_SECRET** (Required for deployment)
   - Secret key for JWT token signing
   - Generate a secure random string:
     ```bash
     openssl rand -base64 32
     ```

4. **REDIS_URL** (Required for deployment)
   - Redis connection URL
   - Example: `redis://redis-service:6379`

5. **RABBITMQ_URL** (Required for deployment)
   - RabbitMQ connection URL
   - Example: `amqp://guest:guest@rabbitmq-service:5672/`

6. **GITHUB_TOKEN** (Automatic)
   - Automatically provided by GitHub Actions
   - Used for pushing to GitHub Container Registry
   - No manual configuration needed

## Setting Up Secrets

### Via GitHub Web Interface

1. Go to your repository on GitHub
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add each secret with its name and value

### Via GitHub CLI

```bash
# Set KUBE_CONFIG
cat ~/.kube/config | base64 -w 0 | gh secret set KUBE_CONFIG

# Set DATABASE_URL
gh secret set DATABASE_URL -b "postgresql://user:password@postgres-service:5432/dbname"

# Set JWT_SECRET
gh secret set JWT_SECRET -b "$(openssl rand -base64 32)"

# Set REDIS_URL
gh secret set REDIS_URL -b "redis://redis-service:6379"

# Set RABBITMQ_URL
gh secret set RABBITMQ_URL -b "amqp://guest:guest@rabbitmq-service:5672/"
```

## Kubernetes Cluster Setup

### For Minikube (Local Development)

1. **Start Minikube:**
   ```bash
   minikube start --driver=docker
   ```

2. **Enable Ingress:**
   ```bash
   minikube addons enable ingress
   ```

3. **Get Kubeconfig:**
   ```bash
   kubectl config view --flatten --minify
   ```

4. **Expose to GitHub Actions:**
   - For local development, you'll need to expose your cluster
   - Consider using ngrok or similar tools for testing
   - For production, use a cloud-hosted Kubernetes cluster

### For Cloud Kubernetes (Recommended for Production)

- **AWS EKS**, **Google GKE**, or **Azure AKS**
- Get the kubeconfig from your cloud provider
- Ensure the GitHub Actions runner can reach the cluster

## Image Pull Secrets

The CD workflow automatically creates an `imagePullSecret` called `ghcr-secret` to pull images from GitHub Container Registry. This uses the `GITHUB_TOKEN` for authentication.

## Monitoring Deployments

### Check Workflow Status

- Go to **Actions** tab in your GitHub repository
- View running/completed workflows
- Check logs for any errors

### Check Kubernetes Status

```bash
# Check pods
kubectl get pods

# Check services
kubectl get services

# Check ingress
kubectl get ingress

# View logs
kubectl logs -l app=user-service
kubectl logs -l app=document-service
kubectl logs -l app=realtime-service
kubectl logs -l app=notification-service
```

## Manual Deployment

You can manually trigger deployment:

1. Go to **Actions** → **CD - Deploy to Kubernetes**
2. Click **Run workflow**
3. Select branch and environment
4. Click **Run workflow**

## Rollback

To rollback a deployment:

```bash
# Rollback specific deployment
kubectl rollout undo deployment/user-service-deployment

# Rollback to specific revision
kubectl rollout undo deployment/user-service-deployment --to-revision=2

# Check rollout history
kubectl rollout history deployment/user-service-deployment
```

## Architecture Flow

```
Developer → Git Push → GitHub
                         ↓
                  CI Workflow (Build & Push)
                         ↓
                  Container Registry (ghcr.io)
                         ↓
                  CD Workflow (Deploy)
                         ↓
                  Kubernetes Cluster
                         ↓
           Services (User, Document, Realtime, Notification)
```

## Best Practices

1. **Branch Protection:** Enable branch protection rules for `main` branch
2. **Required Checks:** Make PR checks required before merging
3. **Review Process:** Require code reviews before merging to main
4. **Secrets Rotation:** Regularly rotate secrets (JWT_SECRET, DATABASE_URL)
5. **Monitoring:** Set up monitoring and alerting for your Kubernetes cluster
6. **Resource Limits:** Define resource limits in Kubernetes deployments
7. **Health Checks:** Implement readiness and liveness probes

## Troubleshooting

### CI Workflow Issues

- **Build fails:** Check Docker build logs in GitHub Actions
- **Push fails:** Verify GITHUB_TOKEN permissions (should be automatic)

### CD Workflow Issues

- **Connection timeout:** Check KUBE_CONFIG secret
- **Image pull errors:** Verify imagePullSecret is created correctly
- **Pod crashes:** Check pod logs with `kubectl logs`
- **Service unavailable:** Check service and ingress configuration

### Common Commands

```bash
# View all resources
kubectl get all

# Describe resource for details
kubectl describe pod <pod-name>

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp

# Force restart deployment
kubectl rollout restart deployment/user-service-deployment
```

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
