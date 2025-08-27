minikube start
minikube addons enable ingress

eval $(minikube -p minikube docker-env)

docker build -t user-service:v1 -f ./cmd/user-service/Dockerfile .
docker build -t document-service:v1 -f ./cmd/document-service/Dockerfile .
docker build -t realtime-service:v1 -f ./cmd/realtime-service/Dockerfile .
docker build -t notification-service:v1 -f ./cmd/notification-service/Dockerfile .

kubectl apply -f k8s/secrets.yaml

kubectl apply -f k8s/

kubectl get pods
