# Real-Time Collaborative Document Editor (Backend)

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.25+-326CE5.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
[![Build Status](https://img.shields.io/badge/Build-passing-brightgreen.svg)](https://github.com/your-org/your-repo/actions)  

## WebSocket API

The WebSocket API provides real-time collaborative document editing capabilities. This section covers the complete message format specification for client-server communication.

### Connection Establishment

**Endpoint:** `wss://your-domain/ws/doc/{documentId}` or `ws://localhost/ws/doc/{documentId}`

**Authentication:** Required via `token` query parameter
```bash
wscat -c "wss://your-domain/ws/doc/{documentId}?token={jwt_token}"
```

### Message Format Specification

#### Initial State Message (Server → Client)
When a client first connects, the server sends the initial document state:

```json
{
  "type": "initial_state"
}
```

This message indicates that the connection is established and the client can begin sending operations.

#### Operation Messages

##### Client Input Format (Client → Server)
Clients send operations in the following format:

**Insert Operation:**
```json
{
  "type": "insert",
  "pos": 0,
  "text": "Hello world",
  "version": 0
}
```

**Delete Operation:**
```json
{
  "type": "delete",
  "pos": 0,
  "len": 5,
  "version": 1
}
```

**Undo Operation:**
```json
{
  "type": "undo"
}
```

##### Server Output Format (Server → All Clients)
The server broadcasts operations to all connected clients in this format:

```json
{
  "type": "operation",
  "op": {
    "type": "insert",
    "pos": 0,
    "text": "Hello world",
    "len": 0,
    "version": 1
  }
}
```

**Field Descriptions:**
- `type`: Always "operation" for broadcasted operations
- `op.type`: Operation type (`insert`, `delete`, `undo`)
- `op.pos`: Character position in the document (0-based)
- `op.text`: Text content (for insert operations)
- `op.len`: Length of text affected (for delete operations)
- `op.version`: Document version after this operation

#### Operation Types Details

| Operation | Client Input | Server Output | Description |
|-----------|--------------|---------------|-------------|
| **Insert** | `{"type":"insert","pos":5,"text":"Hello","version":0}` | `{"type":"operation","op":{"type":"insert","pos":5,"text":"Hello","len":0,"version":1}}` | Inserts text at specified position |
| **Delete** | `{"type":"delete","pos":0,"len":5,"version":1}` | `{"type":"operation","op":{"type":"delete","pos":0,"text":"","len":5,"version":2}}` | Deletes specified length of text from position |
| **Undo** | `{"type":"undo"}` | `{"type":"operation","op":{"type":"delete","pos":0,"text":"","len":5,"version":3}}` | Undoes the last operation by the requesting user |

#### Version Control and Conflict Resolution

- Each operation includes a `version` field representing the expected document state
- The server maintains the authoritative document version
- Operations with outdated versions may be rejected
- Clients should track the latest version from server responses

#### Error Handling

**Authentication Error:**
```json
{
  "type": "error",
  "message": "Invalid token",
  "code": "UNAUTHORIZED"
}
```

**Permission Error:**
```json
{
  "type": "error",
  "message": "Access denied",
  "code": "FORBIDDEN"
}
```

**Version Conflict:**
```json
{
  "type": "error",
  "message": "Version mismatch",
  "code": "CONFLICT",
  "current_version": 5
}
```

### WebSocket API Usage Examples

#### Complete Workflow Example

**1. Connect to Document:**
```bash
wscat -c "wss://your-domain/ws/doc/doc-uuid?token=jwt-token"
```

**2. Receive Initial State:**
```
< {"type":"initial_state"}
```

**3. Send Insert Operation:**
```
> {"type": "insert", "pos": 0, "text": "Hello world", "version": 0}
```

**4. Receive Broadcast:**
```
< {"type":"operation","op":{"type":"insert","pos":0,"text":"Hello world","len":0,"version":1}}
```

**5. Send Delete Operation:**
```
> {"type": "delete", "pos": 0, "len": 5, "version": 1}
```

**6. Receive Broadcast:**
```
< {"type":"operation","op":{"type":"delete","pos":0,"text":"","len":5,"version":2}}
```

#### Multi-User Collaboration Example

**User A sends:**
```json
{"type": "insert", "pos": 0, "text": "Hello ", "version": 0}
```

**Both users receive:**
```json
{"type":"operation","op":{"type":"insert","pos":0,"text":"Hello ","len":0,"version":1}}
```

**User B sends:**
```json
{"type": "insert", "pos": 6, "text": "world!", "version": 1}
```

**Both users receive:**
```json
{"type":"operation","op":{"type":"insert","pos":6,"text":"world!","len":0,"version":2}}
```

**User A sends undo:**
```json
{"type": "undo"}
```

**Both users receive (User A's operation is undone):**
```json
{"type":"operation","op":{"type":"delete","pos":0,"text":"","len":6,"version":3}}
```

### Client Implementation Guidelines

#### Connection Management
```javascript
// Example WebSocket client implementation
const ws = new WebSocket(`wss://your-domain/ws/doc/doc-id?token=${jwt_token}`);

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    if (message.type === 'initial_state') {
        console.log('Connected successfully');
    } else if (message.type === 'operation') {
        applyOperation(message.op);
    } else if (message.type === 'error') {
        handleError(message);
    }
};
```

#### Operation Sending
```javascript
// Send insert operation
function insertText(pos, text, version) {
    ws.send(JSON.stringify({
        type: 'insert',
        pos: pos,
        text: text,
        version: version
    }));
}

// Send delete operation
function deleteText(pos, len, version) {
    ws.send(JSON.stringify({
        type: 'delete',
        pos: pos,
        len: len,
        version: version
    }));
}

// Send undo operation
function undoOperation() {
    ws.send(JSON.stringify({
        type: 'undo'
    }));
}
```
### Testing the System  

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
<img width="945" height="551" alt="collaborative-editor drawio (1)" src="https://github.com/user-attachments/assets/6fb5b86f-3364-4eee-93bc-2f2a6e760415" />


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
- **Bearer token authorization** for REST API endpoints with secure token validation
- **Query parameter authentication** for WebSocket connections (`?token=jwt_token`)
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

Follow these instructions to deploy the entire microservice stack to a local Kubernetes cluster or access a remote deployment.

### Prerequisites

You must have the following tools installed and configured on your system:

*   **Docker:** To build and run containers.
*   **Kubernetes CLI (`kubectl`):** To interact with the Kubernetes cluster.
*   **Minikube:** To run a local Kubernetes cluster.
*   **Helm:** The package manager for Kubernetes.
*   A WebSocket client like `wscat` (`npm install -g wscat`).
*   A tool for making API requests like `curl`.
*   `jq`: A command-line JSON processor, useful for scripting tests.

### Local Development Deployment

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
<img width="851" height="371" alt="image" src="https://github.com/user-attachments/assets/da8381aa-d214-4a1a-94c6-98021dbf0b66" />
<img width="979" height="398" alt="image" src="https://github.com/user-attachments/assets/aff5a57e-c1a2-44e1-9df4-ed84c44718b0" />


**5. Deploy the Observability Stack**

Use Helm to deploy the Prometheus, Loki, and Grafana stack into your cluster.
```sh
# Add the required Helm repositories
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install the charts
helm install loki-stack grafana/loki-stack --version 2.9.11
helm install prometheus prometheus-community/prometheus
helm install grafana grafana/grafana --version 6.58.9
```
Wait for all observability pods (`loki-stack-`, `prometheus-`, `grafana-`) to enter the `Running` state.

### Remote/Production Deployment Access

For accessing a remote Kubernetes deployment (e.g., in a cloud environment), you can use port forwarding to access the application through the Ingress controller:

**Port Forward the Ingress Controller**

```sh
# Forward the ingress controller to make the application accessible locally
kubectl port-forward -n ingress-nginx service/ingress-nginx-controller 8080:80 --address=0.0.0.0

# The application will be accessible at: http://localhost:8080
# For GitHub Codespaces or similar environments, use the forwarded URL provided by the platform
```

This command:
- Forwards port 8080 on your local machine to port 80 of the ingress controller
- Uses `--address=0.0.0.0` to bind to all interfaces (required for cloud development environments)
- Allows access to the application as if it were running locally

**Note:** In cloud development environments (like GitHub Codespaces), the forwarded port will be accessible via a generated URL (e.g., `https://miniature-system-vww9j7vqj7ghp5g9-8080.app.github.dev`).

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
<img width="1916" height="932" alt="image" src="https://github.com/user-attachments/assets/35b3e097-f518-4110-b5d4-3fe7314ea73a" />
<img width="1897" height="932" alt="image" src="https://github.com/user-attachments/assets/89b7c7ae-0f9c-4853-b41e-a4d271bf5408" />
<img width="1889" height="890" alt="image" src="https://github.com/user-attachments/assets/e9b3755e-ffcc-4c42-8331-8d5836540957" />


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

*Note: Document endpoints use Bearer token authentication via Authorization header*

#### **Real-time Collaboration**
- `GET /ws/doc/{documentId}` - WebSocket endpoint for live collaboration

*Note: WebSocket endpoint uses token authentication via query parameter (`?token=jwt_token`)*

#### **Monitoring & Health**
- `GET /metrics` - Prometheus metrics for monitoring

### 🔐 **Authentication in Documentation**

The Swagger documentation includes:
- **Bearer Token Authentication** setup for REST API endpoints
- **Query Parameter Authentication** for WebSocket connections
- **JWT token examples** and format
- **Authorization headers** for protected REST endpoints
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

## Testing the System

### Automated Setup Scripts

This repository includes two setup scripts to help you quickly test the application:

#### `testing_locally.sh` - Local Development Testing
Use this script when testing with a local Minikube deployment:

```bash
# Make the script executable and run it
chmod +x testing_locally.sh
./testing_locally.sh
```

This script:
1. Gets the Minikube IP address
2. Logs in two test users (assumes they're already registered)
3. Creates a test document as User A
4. Shares the document with User B
5. Exports variables for WebSocket testing

#### `testing_public.sh` - Remote/Production Testing
Use this script when testing with a remote deployment or port-forwarded environment:

```bash
# Make the script executable and run it
chmod +x testing_public.sh
./testing_public.sh
```

This script:
1. Uses a predefined SOURCE_URL (update the URL in the script for your environment)
2. Performs the same user login and document setup as `testing_locally.sh`
3. Exports variables for WebSocket testing with the remote URL

**Important:** Before running `testing_public.sh`, update the `SOURCE_URL` variable at the top of the script to match your environment:
```bash
# For port-forwarded local access
export SOURCE_URL=localhost:8080

# For GitHub Codespaces or similar cloud environments
export SOURCE_URL=your-forwarded-url-here.app.github.dev
```

### Manual Testing Steps

#### For Local Development (Minikube)

**1. Get the Application Entry Point**
```sh
export MINIKUBE_IP=$(minikube ip)
echo "Application is accessible at: http://$MINIKUBE_IP"
```

**2. Register Test Users**
```sh
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com", "password":"password123"}' \
  http://$MINIKUBE_IP/auth/register

curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"collab@example.com", "password":"password123"}' \
  http://$MINIKUBE_IP/auth/register
```

**3. Run the Setup Script**
```sh
./testing_locally.sh
```

#### For Remote/Port-Forwarded Development

**1. Start Port Forwarding**
```sh
kubectl port-forward -n ingress-nginx service/ingress-nginx-controller 8080:80 --address=0.0.0.0
```

**2. Register Test Users**
```sh
# Replace localhost:8080 with your actual forwarded URL
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com", "password":"password123"}' \
  http://localhost:8080/auth/register

curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"collab@example.com", "password":"password123"}' \
  http://localhost:8080/auth/register
```

**3. Update and Run the Deployed Setup Script**
```sh
# Edit testing_public.sh to set the correct SOURCE_URL
# Then run:
./testing_public.sh
```

### WebSocket Real-Time Collaboration Testing

After running either setup script, you'll have the necessary environment variables exported. Use these for WebSocket testing:

#### Testing with Local Minikube
```sh
# Open two terminals and run these commands after ./testing_locally.sh

# Terminal 1 (User A - Document Owner)
wscat -c "ws://$MINIKUBE_IP/ws/doc/$DOCUMENT_ID?token=$TOKEN_A"

# Terminal 2 (User B - Collaborator)  
wscat -c "ws://$MINIKUBE_IP/ws/doc/$DOCUMENT_ID?token=$TOKEN_B"
```

#### Testing with Remote/Port-Forwarded Environment
```sh
# Open two terminals and run these commands after ./testing_public.sh

# Terminal 1 (User A - Document Owner)
wscat -c "wss://$SOURCE_URL/ws/doc/$DOCUMENT_ID?token=$TOKEN_A"

# Terminal 2 (User B - Collaborator)
wscat -c "wss://$SOURCE_URL/ws/doc/$DOCUMENT_ID?token=$TOKEN_B"
```

**Note:** Use `ws://` for local HTTP connections and `wss://` for HTTPS connections (common in cloud environments).

#### Interactive Real-Time Operations

Once both WebSocket connections are established, try these operations:

**1. Insert Text (Terminal 1):**
```json
{ "type": "insert", "pos": 0, "text": "Hello world!", "version": 0 }
```
You should see this operation appear in Terminal 2.

**2. Insert More Text (Terminal 2):**
```json
{ "type": "insert", "pos": 12, "text": " How are you?", "version": 1 }
```
Both terminals should now show the combined text.

**3. Delete Text (Terminal 1):**
```json
{ "type": "delete", "pos": 0, "len": 5, "version": 2 }
```
This removes "Hello" from the beginning.

**4. Undo Operation (Terminal 2):**
```json
{ "type": "undo" }
```
This should undo the last operation performed by User B.

#### Expected WebSocket Response Format

All operations will be broadcast to connected clients in this format:
```json
{
  "type": "insert|delete|undo",
  "pos": 0,
  "text": "content",
  "len": 5,
  "version": 1,
  "userId": "user-uuid",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Troubleshooting WebSocket Connections

- **Connection Refused:** Verify the application is running and the URL is correct
- **Authentication Failed:** Check that the JWT token is valid and not expired
- **Operation Rejected:** Ensure the document version matches the server state
- **Permission Denied:** Verify the user has access to the document (owner or shared)

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

### Grafana UI

<img width="1604" height="782" alt="image" src="https://github.com/user-attachments/assets/9b15a7f5-b980-43f4-ba60-bc10947f0ac8" />

###  view logs from user service

<img width="1600" height="781" alt="image" src="https://github.com/user-attachments/assets/b1bd391d-0a97-4fe0-894d-c32bfb03ccee" />

<img width="1596" height="780" alt="image" src="https://github.com/user-attachments/assets/200afb72-cce0-4ad3-953c-49492bb2b06a" />

###  view performance metrics of user service

<img width="1601" height="821" alt="image" src="https://github.com/user-attachments/assets/949ca835-acd3-4d1d-8957-59aa86c90186" />

<img width="1596" height="701" alt="image" src="https://github.com/user-attachments/assets/419ebdbd-5bcf-4313-a471-11177c399aa6" />



## Cleaning Up

```sh
# Delete all Kubernetes resources from your manifests
kubectl delete -f k8s/

# Uninstall the Helm charts
helm uninstall loki-stack grafana prometheus

# Stop the Minikube cluster
minikube stop
```
