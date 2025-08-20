# Collaborative Real-time Document Editor (Backend)

![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#) <!-- Placeholder -->

This repository contains the backend source code for a real-time collaborative document editor, inspired by applications like Google Docs. It is built in Go and designed with a scalable, microservice-oriented architecture. The core of the application uses WebSockets to provide low-latency, stateful communication for a seamless multi-user editing experience.

## Features

*   **User Authentication**: Secure user registration and login using JWT (JSON Web Tokens).
*   **Document Management**: Create and manage documents, with ownership assigned to users.
*   **Document Sharing**: Owners can securely share documents with other registered users.
*   **Real-time Collaboration Engine**:
    *   Stateful WebSocket connections managed on a per-document basis.
    *   Operations-based synchronization (clients send `insert`/`delete` operations).
    *   Sequential consistency enforced by a document versioning system.
    *   Broadcasting of operations to all connected collaborators.
*   **Persistence**: Document state and version are periodically saved to a PostgreSQL database, ensuring data durability.
*   **Configuration Management**: Secure and flexible configuration loading from environment variables.

## Architecture Overview

The system is designed as a set of logical microservices, communicating over APIs and a shared database.

*   **User Service**: Handles all aspects of user identity, including registration, login, and password management.
*   **Document Service**: Manages document metadata, ownership, and the access control list for sharing.
*   **Real-time Service**: The core of the application. It manages WebSocket connections, orchestrates collaboration "hubs" for each active document, processes editing operations, and ensures clients remain in sync.

All components are containerized with Docker for consistent development and deployment environments.

## Architecture Diagram
<img width="5007" height="1998" alt="image (9)" src="https://github.com/user-attachments/assets/31074e3f-1713-4652-b70d-6e7de1ab800a" />

## Sequence Diagram
<img width="5267" height="4086" alt="image (7)" src="https://github.com/user-attachments/assets/68647ddc-d59d-46d3-ad6b-c515848794f2" />


## Technology Stack

| Category                  | Technology                                     | Purpose                                            |
| ------------------------- | ---------------------------------------------- | -------------------------------------------------- |
| **Language**              | **Go (Golang)**                                | High performance, excellent concurrency support    |
| **Web Framework**         | [Chi](https://github.com/go-chi/chi)           | Lightweight, idiomatic, and powerful HTTP router   |
| **Real-time Communication** | [Gorilla WebSocket](https://github.com/gorilla/websocket) | Robust and widely-used WebSocket library         |
| **Database**              | **PostgreSQL**                                 | Reliable, feature-rich relational database         |
| **Database Driver**       | [pgx](https://github.com/jackc/pgx)            | High-performance and idiomatic PostgreSQL driver   |
| **Containerization**      | **Docker & Docker Compose**                    | For building and running the application stack     |
| **Configuration**         | [envconfig](https://github.com/kelseyhightower/envconfig) | Loading configuration from environment variables   |
| **Authentication**        | **JWT (golang-jwt)**                           | Secure, stateless API authentication             |

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

Follow these instructions to get the backend running on your local machine for development and testing.

### Prerequisites

*   [Go](https://golang.org/doc/install) (version 1.18 or newer)
*   [Docker](https://docs.docker.com/get-docker/)
*   [Docker Compose](https://docs.docker.com/compose/install/)
*   A WebSocket command-line client like [wscat](https://www.npmjs.com/package/wscat) for testing. (`npm install -g wscat`)
*   A tool for making API requests, like `curl` or Postman.

### Running the Application

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/PasanAbeysekara/collaborative-editor.git
    cd collaborative-editor
    ```

2.  **Set up Environment Variables:**
    The application requires environment variables for configuration. Create a `.env` file in the root of the project by copying the example file.
    ```bash
    cp .env.example .env
    ```
    The default values in `.env.example` are suitable for local development with Docker Compose and do not need to be changed. **This `.env` file should never be committed to source control.**

3.  **Run Database Migrations:**
    The database needs its schema initialized. The following commands start the database container, copy the SQL migration files into it, and execute them.

    ```bash
    # Start only the database service in the background
    docker-compose up -d db

    # Copy and apply all migrations
    docker cp migrations/001_initial_schema.sql $(docker-compose ps -q db):/schema_1.sql
    docker-compose exec db psql -U user -d collaborative_editor_db -f /schema_1.sql

    docker cp migrations/002_add_content_to_documents.sql $(docker-compose ps -q db):/schema_2.sql
    docker-compose exec db psql -U user -d collaborative_editor_db -f /schema_2.sql

    docker cp migrations/003_add_version_to_documents.sql $(docker-compose ps -q db):/schema_3.sql
    docker-compose exec db psql -U user -d collaborative_editor_db -f /schema_3.sql

    docker cp migrations/004_add_sharing_table.sql $(docker-compose ps -q db):/schema_4.sql
    docker-compose exec db psql -U user -d collaborative_editor_db -f /schema_4.sql
    ```

4.  **Start the Full Application Stack:**
    This command will build the Go application's Docker image and start both the API and database containers.
    ```bash
    docker-compose up --build
    ```
    The backend API will be available at `http://localhost:8080`.

## API Endpoints

### Authentication

**Register a new user**
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"userA@example.com", "password":"password123"}' \
  http://localhost:8080/register
```

**Login and receive a JWT**
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"userA@example.com", "password":"password123"}' \
  http://localhost:8080/login
```
> **Note:** Save the returned `token` for authenticated requests.

### Documents

*Replace `YOUR_TOKEN_HERE` with the JWT and `YOUR_DOCUMENT_ID` with an actual ID.*

**Create a new document**
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"title":"My First Document"}' \
  http://localhost:8080/documents
```

**Share a document with another user**
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"email": "userB@example.com", "role": "editor"}' \
  http://localhost:8080/documents/YOUR_DOCUMENT_ID/share
```

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

The real-time API is available for authorized users on a per-document basis.

*   **Endpoint:** `ws://localhost:8080/ws/doc/{documentID}`
*   **Authentication:** Requires a valid `Authorization: Bearer YOUR_TOKEN_HERE` header during the initial HTTP upgrade request.

**Example Connection using `wscat`:**
```bash
wscat -c "ws://localhost:8080/ws/doc/YOUR_DOCUMENT_ID" \
  --header "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Server -> Client Messages

The server sends structured JSON messages to the client.

**1. Initial State:** Sent immediately upon successful connection.
```json
{
  "type": "initial_state",
  "content": "This is the current document content."
}
```

**2. Operation:** Broadcast to collaborators when a change is made.
```json
{
  "type": "operation",
  "op": {
    "type": "insert",
    "pos": 5,
    "text": "new ",
    "version": 2
  }
}
```

### Client -> Server Messages

The client sends operation objects to the server.

**Example Insert Operation:**
```json
{
  "type": "insert",
  "pos": 0,
  "text": "Hello ",
  "version": 0
}
```

**Example Delete Operation:**
```json
{
  "type": "delete",
  "pos": 5,
  "len": 6,
  "version": 1
}
```

## Roadmap / Future Work

-   [ ] **Operational Transformation (OT)**: Implement a full OT algorithm to handle concurrent conflicting edits gracefully instead of rejecting them.
-   [ ] **Cursor Presence**: Broadcast user cursor positions and text selections.
-   [ ] **Frontend Application**: Build a web-based frontend using a framework like React or Vue.
-   [ ] **CI/CD Pipeline**: Set up GitHub Actions to automatically build, test, and deploy the application.
-   [ ] **Enhanced Roles**: Utilize the `role` in the `document_permissions` table to enforce viewer/commenter roles.

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/my-new-feature`).
3.  Commit your changes (`git commit -am 'Add some feature'`).
4.  Push to the branch (`git push origin feature/my-new-feature`).
5.  Create a new Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
