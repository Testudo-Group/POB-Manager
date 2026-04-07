#!/bin/bash

BASE_URL="http://localhost:8081"
API_URL="${BASE_URL}/api/v1"

ts="$(date +%s)"

# Register admin
echo "Registering admin..."
response=$(curl -s -X POST -H "Content-Type: application/json" -d "{
  \"organization_name\": \"TestOrg${ts}\",
  \"organization_phone\": \"+234700000001\",
  \"organization_address\": \"Test Address\",
  \"first_name\": \"Admin\",
  \"last_name\": \"User\",
  \"phone_number\": \"+2348001111111\",
  \"email\": \"admin${ts}@test.com\",
  \"password\": \"ValidPass123!\"
}" "${API_URL}/auth/register")

ADMIN_TOKEN=$(echo "$response" | jq -r '.data.tokens.access_token')
echo "Admin token: ${ADMIN_TOKEN:0:20}..."

# Create a user
echo ""
echo "Creating user..."
response=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d "{
  \"first_name\": \"Test\",
  \"last_name\": \"User\",
  \"phone_number\": \"+2348001111112\",
  \"email\": \"testuser${ts}@test.com\",
  \"password\": \"UserPass123!\",
  \"role\": \"personnel\"
}" "${API_URL}/users")

USER_ID=$(echo "$response" | jq -r '.data.id')
echo "User created: $USER_ID"

# Try to update user
echo ""
echo "Updating user..."
echo "Request body: {\"first_name\":\"Updated\",\"last_name\":\"User\"}"
response=$(curl -s -X PATCH -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d '{
  "first_name": "Updated",
  "last_name": "User"
}' "${API_URL}/users/${USER_ID}")

echo "Response:"
echo "$response" | jq .
