# Real-Time Collaborative Document Editor (Backend)

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.25+-326CE5.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#) <!-- Placeholder -->

This repository contains the backend for a cloud-native, real-time collaborative document editor, inspired by applications like Google Docs. It is engineered with a **production-grade microservice architecture** using Go, containerized with Docker, and orchestrated with **Kubernetes**. The system provides a robust foundation for building scalable, resilient, and observable distributed systems.

## Core Concepts & Architectural Philosophy

This project serves as a practical, hands-on implementation of modern cloud-native principles.

*   **Microservice Architecture:** The application is decomposed into independent services (`user`, `document`, `realtime`), each with a distinct business capability. This allows for independent development, deployment, and scaling.
*   **Container Orchestration:** We've moved beyond `docker-compose` to **Kubernetes** to manage the application's lifecycle. This provides production-level features like self-healing, automated rollouts, and declarative configuration.
*   **Declarative Infrastructure:** All application components, from the services themselves to the databases and networking rules, are defined as code in Kubernetes YAML manifests.
*   **Observability as a First-Class Citizen:** The system is built with a pre-configured, powerful observability stack (**Prometheus, Loki, Grafana**). This provides immediate, centralized insight into the health, performance, and behavior of all services, which is non-negotiable in a microservice environment.
*   **API Gateway Pattern:** An **NGINX Ingress Controller** acts as the single, unified entry point for all external traffic, handling routing, SSL termination (in a real production setup), and abstracting the internal service topology from the outside world.

## Features

*   **User Authentication**: Secure user registration and login using JWT.
*   **Document Management**: Create and manage documents, with ownership assigned to users.
*   **Secure Document Sharing**: Owners can securely share documents with other registered users via an access control list.
*   **Real-time Collaboration Engine**:
    *   Stateful WebSocket connections managed on a per-document basis.
    *   Operations-based synchronization (`insert`/`delete`/`undo` operations).
    *   Sequential consistency enforced by a document versioning system.
    *   High-speed session state management and operation history using **Redis**.
*   **Cloud-Native Deployment**: Fully orchestrated with **Kubernetes**, including a production-style Ingress for traffic management.
*   **Full Observability Stack**: Centralized logging and metrics out-of-the-box with **Prometheus, Loki, and Grafana**.

## Architecture & Technology Stack

### Architecture Diagram
<img width="5007" height="1998" alt="image (9)" src="https://github.com/user-attachments/assets/31074e3f-1713-4652-b70d-6e7de1ab800a" />

### Technology Table

| Category | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | **Go (Golang)** | High performance, excellent concurrency support |
| **Containerization** | **Docker** | Packaging services into portable, immutable artifacts |
| **Orchestration** | **Kubernetes (Minikube)** | Production-grade deployment, scaling, and self-healing |
| **API Gateway** | **NGINX Ingress Controller** | Manages and routes all external HTTP/WebSocket traffic |
| **Web Framework** | [Chi](https://github.com/go-chi/chi) | Lightweight and idiomatic HTTP router |
| **Real-time** | [Gorilla WebSocket](https://github.com/gorilla/websocket) | Robust and battle-tested WebSocket library |
| **Primary Database** | **PostgreSQL (Remote/Cloud)**| Reliable, persistent storage for users & documents |
| **In-Memory Cache** | **Redis** | High-speed session state and operation (undo) stack |
| **Observability** | **Prometheus, Grafana, Loki** | Centralized metrics, dashboards, and logging |
| **Deployment** | **Helm** | Package manager for deploying the observability stack |
| **Configuration** | **Kubernetes Secrets** | Secure, declarative management of sensitive configuration |
| **Authentication** | **JWT (golang-jwt)** | Secure, stateless API authentication |

## Security Considerations

### 🔐 **Authentication & Authorization**
- **JWT-based authentication** for all protected endpoints
- **Bearer token authorization** with secure token validation
- **Role-based access control** for document sharing (owner, editor, viewer)
- **Password hashing** using bcrypt for secure credential storage

### 🛡️ **API Security**
- **Input validation** on all endpoints with proper error handling
- **CORS configuration** for cross-origin resource sharing
- **Rate limiting** considerations for production deployment
- **SQL injection prevention** using parameterized queries with pgx

### 📝 **Documentation Security**
- **Example tokens** in API documentation are clearly marked as fake
- **No real secrets** are committed to the repository
- **GitGuardian configuration** (`.gitguardian.yml`) to prevent false positives
- **Environment-based configuration** keeps sensitive data out of source code

### 🔒 **Production Recommendations**
- Use strong, randomly generated JWT secrets
- Implement proper HTTPS/TLS encryption
- Set up API rate limiting and request throttling
- Configure proper CORS policies for your frontend domain
- Use secrets management systems (Kubernetes secrets, AWS Secrets Manager, etc.)
- Enable audit logging for security monitoring

## Getting Started

Follow these instructions to deploy the entire microservice stack to a local Kubernetes cluster.

### Prerequisites

You must have the following tools installed and configured on your system:

*   **Docker:** To build and run containers.
*   **Kubernetes CLI (`kubectl`):** To interact with the Kubernetes cluster.
*   **Minikube:** To run a local Kubernetes cluster.
*   **Helm:** The package manager for Kubernetes.
*   A WebSocket client like `wscat` (`npm install -g wscat`).
*   A tool for making API requests like `curl`.
*   `jq`: A command-line JSON processor, useful for scripting tests.

### Deployment Instructions

**1. Start the Kubernetes Cluster**

Initialize your local Minikube cluster and enable the Ingress addon, which will act as our API Gateway.
```sh
minikube start
minikube addons enable ingress
```

**2. Configure Environment Variables**

The application's sensitive configuration is managed via a Kubernetes Secret, which is generated from a local `.env` file.

```sh
# Create your local configuration file from the example
cp .env.local .env
```
**IMPORTANT:** Open the newly created `.env` file and:
1.  Replace the placeholder `DATABASE_URL` with your **actual remote PostgreSQL connection string**.
2.  Change `JWT_SECRET` to a long, unique, and random string for security.

**3. Build and Load Docker Images**

This crucial command configures your local Docker client to communicate with the Docker daemon running *inside* the Minikube cluster. This allows you to build images directly into its environment, making them accessible to Kubernetes without needing an external registry.

```sh
# For Linux/macOS
eval $(minikube -p minikube docker-env)

# For Windows PowerShell: & minikube -p minikube docker-env | Invoke-Expression

# Build the images for each microservice
docker build -t user-service:v1 -f ./cmd/user-service/Dockerfile .
docker build -t document-service:v1 -f ./cmd/document-service/Dockerfile .
docker build -t realtime-service:v1 -f ./cmd/realtime-service/Dockerfile .
```

**4. Deploy the Application to Kubernetes**

Apply the Kubernetes manifests to create all application resources.
```sh
# Create the Kubernetes Secret from your .env file
kubectl apply -f k8s/secrets.yaml

# Deploy Redis, our 3 services, and the Ingress routing rules
kubectl apply -f k8s/
```
Verify that all application pods are running successfully:
```sh
kubectl get pods
```

**5. Deploy the Observability Stack**

Use Helm to deploy the Prometheus, Loki, and Grafana stack into your cluster.
```sh
# Add the required Helm repositories
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install the charts
helm install loki-stack grafana/loki-stack
helm install prometheus prometheus-community/prometheus
helm install grafana grafana/grafana
```
Wait for all observability pods (`loki-stack-`, `prometheus-`, `grafana-`) to enter the `Running` state.

## 📚 API Documentation (Swagger/OpenAPI)

This project includes comprehensive API documentation using OpenAPI 3.0.3 specification (Swagger). The documentation provides detailed information about all endpoints, request/response schemas, authentication, and examples.

### 📖 **Viewing the Documentation**

#### **Option 1: Interactive HTML Documentation**
1. Start a local web server in the project root:
   ```bash
   # Using Python (most common)
   python3 -m http.server 8000
   
   # Or using Node.js
   npx http-server -p 8000
   
   # Or using Go
   go run -m http.server :8000
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8000/api-docs.html
   ```

#### **Option 2: Online Swagger Editor**
1. Go to [editor.swagger.io](https://editor.swagger.io)
2. Copy the contents of `swagger.yaml` from this repository
3. Paste it into the editor for interactive documentation

#### **Option 3: Import into API Tools**
- **Postman**: Import `swagger.json` to automatically generate a collection
- **Insomnia**: Import `swagger.yaml` for API testing
- **VS Code**: Use REST Client extension with `requests.http` file

### 📋 **Documentation Files**

| File | Description | Use Case |
|------|-------------|----------|
| `swagger.yaml` | Complete OpenAPI 3.0.3 specification | Human-readable, detailed documentation |
| `swagger.json` | JSON version of the specification | Tool integration, automated processing |
| `api-docs.html` | Interactive HTML documentation | Local viewing with Swagger UI |
| `requests.http` | HTTP request examples | Testing with VS Code REST Client |

### 🚀 **What's Documented**

#### **Authentication Endpoints**
- `POST /auth/register` - User registration with validation
- `POST /auth/login` - User authentication with JWT

#### **Document Management**
- `GET /documents` - List user's documents (owned and shared)
- `POST /documents` - Create new document with required validation
- `GET /documents/{id}` - Get specific document by ID
- `PUT /documents/{id}` - Update document content (internal)
- `POST /documents/{id}/share` - Share document with other users
- `GET /documents/{id}/permissions/{userId}` - Check user permissions

#### **Real-time Collaboration**
- `GET /ws/doc/{documentId}` - WebSocket endpoint for live collaboration

#### **Monitoring & Health**
- `GET /metrics` - Prometheus metrics for monitoring

### 🔐 **Authentication in Documentation**

The Swagger documentation includes:
- **Bearer Token Authentication** setup
- **JWT token examples** and format
- **Authorization headers** for protected endpoints
- **Error responses** for unauthorized access

### 💡 **Key Features**

- **Request/Response Examples**: Real JSON examples for all endpoints
- **Validation Rules**: Field requirements, data types, and constraints
- **Error Scenarios**: Comprehensive error response documentation
- **WebSocket Guide**: Instructions for real-time connection setup
- **Interactive Testing**: Try endpoints directly from the documentation

### 🛠 **Using with Development Tools**

#### **Postman Collection Generation**
```bash
# Import swagger.json into Postman to auto-generate:
# - All API endpoints
# - Request examples
# - Environment variables
# - Authentication setup
```

#### **Client SDK Generation**
```bash
# Generate client SDKs for various languages
npx @openapitools/openapi-generator-cli generate \
  -i swagger.yaml \
  -g javascript \
  -o ./client-sdk
```

#### **API Gateway Integration**
The OpenAPI specification can be used with:
- **AWS API Gateway**
- **Azure API Management**
- **Kong Gateway**
- **Envoy Proxy**

### 📝 **Testing with the Documentation**

1. **Start the application** (see Getting Started section)
2. **Open the interactive documentation** at `http://localhost:8000/api-docs.html`
3. **Try the "Register" endpoint** to create a test user
4. **Use the "Login" endpoint** to get a JWT token
5. **Copy the token** and use it in the "Authorize" button
6. **Test document creation and management** with authenticated requests

The documentation is kept in sync with the actual API implementation and includes all the latest features and validation rules.

## WebSocket API
=======
---

## Testing the System

All interaction with the application now goes through the Minikube Ingress IP.

**1. Get the Application Entry Point**
```sh
export MINIKUBE_IP=$(minikube ip)
echo "Application is accessible at: http://$MINIKUBE_IP"
```

**2. Run the Automated End-to-End Test Script**

The following script will:
1.  Register two users.
2.  Log them in and extract their JWTs.
3.  Create a new document as User A.
4.  Share the document with User B.
5.  Print the exported variables needed for the manual WebSocket test.

```sh
#!/bin/bash
export MINIKUBE_IP=$(minikube ip)

echo "--- 1. Registering Users ---"
curl -X POST -H "Content-Type: application/json" -d '{"email":"owner@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/register
curl -X POST -H "Content-Type: application/json" -d '{"email":"collab@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/register

echo -e "\n--- 2. Logging in Users ---"
export TOKEN_A=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"owner@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/login | jq -r .token)
export TOKEN_B=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"collab@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/login | jq -r .token)

echo "--- 3. Creating Document ---"
export DOCUMENT_ID=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN_A" -d '{"title":"Live Test Doc"}' http://$MINIKUBE_IP/documents | jq -r .ID)

echo "--- 4. Sharing Document ---"
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN_A" -d '{"email":"collab@example.com", "role":"editor"}' http://$MINIKUBE_IP/documents/$DOCUMENT_ID/share

echo -e "\n\n✅ Setup Complete! Use these exported variables for WebSocket testing:"
echo "------------------------------------------------------------------"
echo "export MINIKUBE_IP=$MINIKUBE_IP"
echo "export DOCUMENT_ID=$DOCUMENT_ID"
echo "export TOKEN_A=$TOKEN_A"
echo "export TOKEN_B=$TOKEN_B"
echo "------------------------------------------------------------------"
```

**3. Manual WebSocket Test**

*   **Test Real-Time Editing:**
    Open two terminals. In Terminal 1, connect as User A. In Terminal 2, connect as User B. (Requires `wscat` and `jq` to be installed).

    *   **Terminal 1 (User A):**
        ```sh
        wscat -c "ws://$MINIKUBE_IP/ws/doc/$DOCUMENT_ID" --header "Authorization: Bearer $TOKEN_A"
        ```
    *   **Terminal 2 (User B):**
        ```sh
        wscat -c "ws://$MINIKUBE_IP/ws/doc/$DOCUMENT_ID" --header "Authorization: Bearer $TOKEN_B"
        ```
    *   In Terminal 1, send an `insert` operation. You should see it appear in Terminal 2.
        ```json
        { "type": "insert", "pos": 0, "text": "Hello world", "version": 0 }
        ```
    *   In Terminal 2, send an `undo` operation. Both terminals should receive a `delete` operation from the server.
        ```json
        { "type": "undo" }
        ```

## Accessing the Observability Dashboard (Grafana)

1.  **Get the Grafana Admin Password:**
    ```sh
    kubectl get secret grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
    ```
2.  **Access the UI:**
    Open a new terminal and run this command to forward the port.
    ```sh
    kubectl port-forward svc/grafana 3000:80
    ```
    Open your browser to `http://localhost:3000` and log in with username `admin` and the retrieved password.

3.  **Configure Data Sources:**
    *   Navigate to **Connections -> Data Sources**.
    *   **Add Prometheus:** URL: `http://prometheus-server:80`
    *   **Add Loki:** URL: `http://loki-stack:3100`

You can now use the **Explore** tab to view logs from all services (via Loki) and performance metrics (via Prometheus).

## Roadmap / Future Work

-   [ ] **Phase 3: Distributed Tracing**: Instrument services with OpenTelemetry and deploy Jaeger to visualize request flows across the microservice architecture.
-   [ ] **Phase 4: Resilience Patterns**: Implement Circuit Breakers in the `realtime-service` to gracefully handle failures or slowdowns in the `document-service`.
-   [ ] **Phase 4: Event-Driven Architecture**: Decouple services further by introducing a message broker like NATS for asynchronous communication (e.g., publishing a `DocumentUpdated` event instead of a synchronous save).
-   [ ] **Full OT Algorithm**: Implement a complete Operational Transformation (OT) algorithm to resolve concurrent editing conflicts instead of simple version rejection.

## Cleaning Up

```sh
# Delete all Kubernetes resources from your manifests
kubectl delete -f k8s/

# Uninstall the Helm charts
helm uninstall loki-stack grafana prometheus

# Stop the Minikube cluster
minikube stop
```
