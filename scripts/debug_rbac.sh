#!/bin/bash

BASE_URL="http://localhost:8081"
API_URL="${BASE_URL}/api/v1"

ts="$(date +%s)"

# Register org 1 with admin user
echo "Registering admin..."
response=$(curl -s -X POST -H "Content-Type: application/json" -d "{
  \"organization_name\": \"OrgAdmin${ts}\",
  \"organization_phone\": \"+234700000001\",
  \"organization_address\": \"Test Address\",
  \"first_name\": \"Admin\",
  \"last_name\": \"User\",
  \"phone_number\": \"+2348001111111\",
  \"email\": \"admin${ts}@test.com\",
  \"password\": \"ValidPass123!\"
}" "${API_URL}/auth/register")

ADMIN_TOKEN=$(echo "$response" | jq -r '.data.tokens.access_token')
ADMIN_ROLE=$(echo "$response" | jq -r '.data.user.role')
echo "Admin Token Role: $ADMIN_ROLE"

# Register org 2 with personnel user
echo ""
echo "Registering personnel as org admin..."
response=$(curl -s -X POST -H "Content-Type: application/json" -d "{
  \"organization_name\": \"OrgPersonnel${ts}\",
  \"organization_phone\": \"+234700000002\",
  \"organization_address\": \"Test Address\",
  \"first_name\": \"Personnel\",
  \"last_name\": \"User\",
  \"phone_number\": \"+2348001111112\",
  \"email\": \"personnel${ts}@test.com\",
  \"password\": \"ValidPass123!\"
}" "${API_URL}/auth/register")

PERSONNEL_TOKEN=$(echo "$response" | jq -r '.data.tokens.access_token')
PERSONNEL_ROLE=$(echo "$response" | jq -r '.data.user.role')
echo "Personnel Token Role: $PERSONNEL_ROLE"

# Test 1: Admin lists users
echo ""
echo "Test 1: Admin tries to list users..."
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" "${API_URL}/users")
status=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | grep -v "HTTP_STATUS")
echo "Response status: $status"
echo "Response body:"
echo "$body" | jq .

# Test 2: Personnel tries to list users
echo ""
echo "Test 2: Personnel tries to list users..."
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $PERSONNEL_TOKEN" "${API_URL}/users")
status=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | grep -v "HTTP_STATUS")
echo "Response status: $status"
echo "Response body:"
echo "$body" | jq .
