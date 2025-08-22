export SOURCE_URL=miniature-system-vww9j7vqj7ghp5g9-8080.app.github.dev

echo -e "\n--- 2. Logging in Users ---"
export TOKEN_A=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"owner@example.com", "password":"password123"}' https://$SOURCE_URL/auth/login | jq -r .token)
export TOKEN_B=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"collab@example.com", "password":"password123"}' https://$SOURCE_URL/auth/login | jq -r .token)

echo "--- 3. Creating Document ---"
export DOCUMENT_ID=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN_A" -d '{"title":"Live Test Doc"}' https://$SOURCE_URL/documents | jq -r .ID)

echo "--- 4. Sharing Document ---"
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN_A" -d '{"email":"collab@example.com", "role":"editor"}' https://$SOURCE_URL/documents/$DOCUMENT_ID/share

echo -e "\n\n✅ Setup Complete! Use these exported variables for WebSocket testing:"
echo "------------------------------------------------------------------"
echo "export SOURCE_URL=$SOURCE_URL"
echo "export DOCUMENT_ID=$DOCUMENT_ID"
echo "export TOKEN_A=$TOKEN_A"
echo "export TOKEN_B=$TOKEN_B"
echo "------------------------------------------------------------------"
