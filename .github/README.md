# Collaborative Editor - CI/CD Pipeline

## Quick Start Guide

This guide will help you set up the complete CI/CD pipeline for the collaborative editor project.

### 1. Prerequisites

- GitHub repository with admin access
- Kubernetes cluster (Minikube for local, or cloud provider for production)
- `kubectl` installed and configured locally
- GitHub CLI (optional, for easier secret management)

### 2. Configure GitHub Secrets

Set up the following secrets in your GitHub repository:

```bash
# Navigate to: Settings → Secrets and variables → Actions → New repository secret

Required Secrets:
- KUBE_CONFIG       # Base64-encoded kubeconfig file
- DATABASE_URL      # PostgreSQL connection string
- JWT_SECRET        # JWT signing key
- REDIS_URL         # Redis connection URL
- RABBITMQ_URL      # RabbitMQ connection URL
```

#### Generate Secrets

```bash
# KUBE_CONFIG
cat ~/.kube/config | base64 -w 0

# JWT_SECRET
openssl rand -base64 32

# DATABASE_URL (example)
postgresql://user:password@postgres-host:5432/collaborative_editor

# REDIS_URL (example)
redis://redis-service:6379

# RABBITMQ_URL (example)
amqp://guest:guest@rabbitmq-service:5672/
```

### 3. Workflow Overview

```
┌─────────────┐
│   Developer │
│  git push   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  CI Workflow (Automatic)            │
│  - Build Docker images              │
│  - Push to ghcr.io                  │
│  - Run tests & security scans       │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  CD Workflow (Automatic)            │
│  - Update K8s deployments           │
│  - Apply infrastructure             │
│  - Deploy services                  │
│  - Verify deployment                │
└─────────────────────────────────────┘
```

### 4. Available Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to main/dev | Build and push Docker images |
| `cd.yml` | After successful CI | Deploy to Kubernetes |
| `pr-checks.yml` | Pull requests | Run tests, linting, security scans |
| `cleanup.yml` | Weekly schedule | Clean old container images |
| `db-migration.yml` | Manual | Run database migrations |

### 5. Branch Strategy

- **`main`** → Production environment
- **`dev`** → Development environment
- **Feature branches** → Create PRs to dev

### 6. First-Time Setup

#### Step 1: Enable GitHub Container Registry

1. Go to repository **Settings**
2. Scroll to **Packages**
3. Ensure visibility is set appropriately

#### Step 2: Set Up Branch Protection

1. Go to **Settings** → **Branches**
2. Add rule for `main` branch:
   - ✅ Require pull request reviews
   - ✅ Require status checks to pass
   - ✅ Require branches to be up to date
   - Select required checks: `lint-and-test`, `docker-build-test`

#### Step 3: Configure Kubernetes Cluster

For **Minikube** (local):
```bash
minikube start --driver=docker --cpus=4 --memory=8192
minikube addons enable ingress
```

For **Cloud** (production):
- Use EKS, GKE, or AKS
- Ensure cluster is accessible from GitHub Actions
- Get kubeconfig file

#### Step 4: Add GitHub Secrets

Using GitHub CLI:
```bash
gh secret set KUBE_CONFIG < ~/.kube/config | base64 -w 0
gh secret set JWT_SECRET -b "$(openssl rand -base64 32)"
gh secret set DATABASE_URL -b "your-database-url"
gh secret set REDIS_URL -b "redis://redis-service:6379"
gh secret set RABBITMQ_URL -b "amqp://guest:guest@rabbitmq-service:5672/"
```

Or via web interface: **Settings** → **Secrets and variables** → **Actions**

### 7. Deployment Flow

#### Automatic Deployment

1. **Make changes** to your code
2. **Commit and push** to `dev` or `main`
3. **CI workflow** automatically builds images
4. **CD workflow** automatically deploys to Kubernetes
5. **Monitor** deployment in Actions tab

#### Manual Deployment

1. Go to **Actions** tab
2. Select **CD - Deploy to Kubernetes**
3. Click **Run workflow**
4. Choose environment and branch
5. Click **Run workflow** button

### 8. Monitoring

#### Check Workflow Status

```bash
# Via GitHub CLI
gh run list
gh run view <run-id>
gh run watch
```

#### Check Kubernetes Status

```bash
# Pods
kubectl get pods
kubectl describe pod <pod-name>
kubectl logs -f <pod-name>

# Services
kubectl get svc
kubectl get ingress

# Events
kubectl get events --sort-by=.metadata.creationTimestamp
```

### 9. Common Operations

#### View Application Logs

```bash
# User Service
kubectl logs -l app=user-service --tail=100 -f

# Document Service
kubectl logs -l app=document-service --tail=100 -f

# Realtime Service
kubectl logs -l app=realtime-service --tail=100 -f

# Notification Service
kubectl logs -l app=notification-service --tail=100 -f
```

#### Scale Services

```bash
# Scale up
kubectl scale deployment/user-service-deployment --replicas=3

# Scale down
kubectl scale deployment/user-service-deployment --replicas=1
```

#### Rollback Deployment

```bash
# View history
kubectl rollout history deployment/user-service-deployment

# Rollback to previous
kubectl rollout undo deployment/user-service-deployment

# Rollback to specific revision
kubectl rollout undo deployment/user-service-deployment --to-revision=2
```

#### Run Database Migration

1. Go to **Actions** → **Database Migration**
2. Click **Run workflow**
3. Select environment (dev/staging/production)
4. Select direction (up/down)
5. Click **Run workflow**

### 10. Troubleshooting

#### Image Pull Errors

```bash
# Check image pull secret
kubectl get secret ghcr-secret

# Recreate if needed
kubectl delete secret ghcr-secret
# CD workflow will recreate it automatically on next run
```

#### Pod Not Starting

```bash
# Check pod status
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>

# Check events
kubectl get events --field-selector involvedObject.name=<pod-name>
```

#### Service Unreachable

```bash
# Check service endpoints
kubectl get endpoints

# Check ingress
kubectl describe ingress

# Port forward for testing
kubectl port-forward svc/user-service 8080:8080
```

### 11. Security Best Practices

- ✅ Rotate secrets regularly
- ✅ Use least privilege for service accounts
- ✅ Enable vulnerability scanning
- ✅ Keep dependencies updated
- ✅ Review security scan results
- ✅ Use network policies in Kubernetes
- ✅ Enable audit logging

### 12. Performance Optimization

- Configure resource requests and limits
- Use horizontal pod autoscaling
- Implement proper health checks
- Use connection pooling
- Enable caching with Redis
- Monitor with Prometheus/Grafana

### 13. Next Steps

- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure alerting
- [ ] Add integration tests
- [ ] Set up staging environment
- [ ] Implement blue-green deployments
- [ ] Add performance testing
- [ ] Configure backup strategies

### 14. Support

For issues or questions:
- Check workflow logs in GitHub Actions
- Review Kubernetes events and logs
- Consult [CICD_SETUP.md](.github/CICD_SETUP.md) for detailed documentation

---

**Made with ❤️ for Collaborative Editor Project**
