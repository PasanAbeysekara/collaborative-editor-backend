# Collaborative Real-time Document Editor (Backend)

![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#) 
This project is the backend for a cloud-native, real-time collaborative document editor, similar to Google Docs. It is built using a microservice architecture in Go and deployed on Kubernetes.

## Features

-   **User Authentication:** Secure user registration and login via JWT.
-   **Document Management:** Create documents and share them with other users.
-   **Real-Time Collaboration:** WebSocket-based, operation-driven synchronization for multi-user editing.
-   **Undo Functionality:** Session-based undo support using a Redis-backed operation stack.
-   **Microservice Architecture:** Services are independently deployable, scalable, and fault-tolerant.
-   **Cloud-Native Deployment:** Runs on Kubernetes with a production-ready setup.
-   **Observability:** Centralized logging and metrics with Prometheus, Loki, and Grafana.

## Architecture Overview

The system is composed of three core microservices, a database, a cache, and an API gateway, all running on Kubernetes.

-   **User Service:** Manages user identity, registration, and authentication.
-   **Document Service:** Manages document metadata, content persistence, and sharing permissions.
-   **Real-time Service:** Handles WebSocket connections, processes collaborative operations in real-time, and manages the live session state in Redis.
-   **PostgreSQL:** The primary database for persistent storage of users and documents.
-   **Redis:** High-speed cache for "hot" document sessions and the operation stack for undo functionality.
-   **NGINX Ingress:** Acts as the API Gateway, routing all incoming traffic to the appropriate microservice.

### Architecture Diagram
<img width="5007" height="1998" alt="image (9)" src="https://github.com/user-attachments/assets/31074e3f-1713-4652-b70d-6e7de1ab800a" />

### Sequence Diagram
<img width="5267" height="4086" alt="image (7)" src="https://github.com/user-attachments/assets/68647ddc-d59d-46d3-ad6b-c515848794f2" />

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You must have the following tools installed on your system:

-   **Docker:** To build and run containers.
-   **Kubernetes CLI (`kubectl`):** To interact with the Kubernetes cluster.
-   **Minikube:** To run a local Kubernetes cluster.
-   **Helm:** The package manager for Kubernetes, used to deploy the observability stack.

### Setup and Deployment

Follow these steps to deploy the entire application stack to your local Minikube cluster.

**1. Start Minikube**

Start your local Kubernetes cluster and enable the necessary addons.
```sh
minikube start
minikube addons enable ingress
```

**2. Create the `.env` file**

This project uses an `.env` file for local configuration. Create a file named `.env` in the project root by copying the example.
```sh
cp .env.local .env
```
Review the `.env` file and ensure the values are suitable for your local setup. The defaults should work fine.

**3. Build and Load Docker Images**

Point your Docker client to Minikube's internal registry and build the service images.
```sh
# For Linux/macOS
eval $(minikube -p minikube docker-env)

# For Windows PowerShell
# & minikube -p minikube docker-env | Invoke-Expression

# Build the images
docker build -t user-service:v1 -f ./cmd/user-service/Dockerfile .
docker build -t document-service:v1 -f ./cmd/document-service/Dockerfile .
docker build -t realtime-service:v1 -f ./cmd/realtime-service/Dockerfile .
```

**4. Deploy Application to Kubernetes**

Apply the Kubernetes manifests to create all application resources. The secret manifest uses variables from your `.env` file.
```sh
# Create secrets and deployments
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/
```
Verify that all application pods are running:
```sh
kubectl get pods
```
Wait until `user-service`, `document-service`, `realtime-service`, and `redis` pods are in the `Running` state.

**5. Deploy the Observability Stack**

Use Helm to deploy Prometheus (metrics), Loki (logs), and Grafana (dashboards).
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
Wait for all the new observability pods to enter the `Running` state.

---

## Testing the System

Once all services are deployed and running, you can test the application.

**1. Get the Application IP Address**

Find the entry point IP for your application from the Minikube Ingress.
```sh
export MINIKUBE_IP=$(minikube ip)
echo "Application is accessible at: http://$MINIKUBE_IP"
```
All subsequent API calls will be made to this IP address.

**2. Run the Test Flow (using `curl` and `wscat`)**

You will need at least two terminals to simulate two different users.

*   **Setup Users and Document:**
    ```sh
    # Register User A (Owner)
    curl -X POST -H "Content-Type: application/json" -d '{"email":"owner@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/register

    # Login User A and get TOKEN_A
    TOKEN_A=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"owner@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/login | jq -r .token)
    echo "User A Token: $TOKEN_A"

    # Register User B (Collaborator)
    curl -X POST -H "Content-Type: application/json" -d '{"email":"collab@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/register

    # Login User B and get TOKEN_B
    TOKEN_B=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"collab@example.com", "password":"password123"}' http://$MINIKUBE_IP/auth/login | jq -r .token)
    echo "User B Token: $TOKEN_B"

    # User A creates a document and get DOCUMENT_ID
    DOCUMENT_ID=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN_A" -d '{"title":"Live Test Doc"}' http://$MINIKUBE_IP/documents | jq -r .ID)
    echo "Document ID: $DOCUMENT_ID"

    # User A shares the document with User B
    curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN_A" -d '{"email":"collab@example.com", "role":"editor"}' http://$MINIKUBE_IP/documents/$DOCUMENT_ID/share
    ```

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
    Now, open your browser and navigate to `http://localhost:3000`. Log in with username `admin` and the password you just retrieved.

3.  **Configure Data Sources:**
    *   Navigate to **Connections -> Data Sources**.
    *   **Add Prometheus:** URL: `http://prometheus-server:80`
    *   **Add Loki:** URL: `http://loki-stack:3100`

You can now use the **Explore** tab to view logs from all services (via Loki) and performance metrics (via Prometheus).

## Cleaning Up

To stop and delete all the resources created, run the following commands:
```sh
# Delete all Kubernetes resources from your manifests
kubectl delete -f k8s/

# Uninstall the Helm charts
helm uninstall loki-stack grafana prometheus

# Stop the Minikube cluster
minikube stop
```
