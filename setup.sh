export MINIKUBE_IP=$(minikube ip)

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
