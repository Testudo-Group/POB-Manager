#!/usr/bin/env bash

set -u

BASE_URL="${BASE_URL:-http://localhost:8081}"
API_URL="${BASE_URL}/api/v1"

# Helper functions
request() {
  local method="$1"
  local url="$2"
  local data="${3:-}"
  local token="${4:-}"

  local tmp
  tmp="$(mktemp)"

  local curl_args=(
    -sS
    -X "$method"
    -H "Content-Type: application/json"
    -o "$tmp"
    -w "%{http_code}"
  )

  if [[ -n "$token" ]]; then
    curl_args+=(-H "Authorization: Bearer ${token}")
  fi

  if [[ -n "$data" ]]; then
    curl_args+=(-d "$data")
  fi

  local status
  status="$(curl "${curl_args[@]}" "$url")"
  REQUEST_STATUS="$status"
  REQUEST_BODY="$(cat "$tmp")"
  rm -f "$tmp"
}

timestamp="$(date +%s)"

echo "=========================================="
echo "DEBUG: Testing Refresh Token Flow"
echo "=========================================="

# Step 1: Register
echo ""
echo "Step 1: Registering user..."
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org_name "Debug Org ${timestamp}" \
  --arg org_phone "+234800000000" \
  --arg org_address "Test Address" \
  --arg first "Debug" \
  --arg last "User" \
  --arg phone "+234800111111" \
  --arg email "debug.${timestamp}@example.com" \
  --arg password "ValidPassword123!" \
  '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"

echo "Response status: $REQUEST_STATUS"
if [[ "$REQUEST_STATUS" != "201" ]]; then
  echo "ERROR: Registration failed!"
  echo "$REQUEST_BODY" | jq .
  exit 1
fi

ACCESS_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
REFRESH_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.refresh_token')"
USER_ID="$(echo "$REQUEST_BODY" | jq -r '.data.user.id')"

echo "✓ Registration successful"
echo "  User ID: $USER_ID"
echo "  Access Token: ${ACCESS_TOKEN:0:20}..."
echo "  Refresh Token: ${REFRESH_TOKEN:0:20}..."

# Step 2: Decode tokens to inspect
echo ""
echo "Step 2: Inspecting tokens..."

decode_jwt_payload() {
  local token="$1"
  local payload="$(echo "$token" | cut -d. -f2)"
  # Add padding if needed
  case $((${#payload} % 4)) in
    2) payload="${payload}==" ;;
    3) payload="${payload}=" ;;
  esac
  echo "$payload" | base64 -d 2>/dev/null
}

echo "Access Token Claims:"
decode_jwt_payload "$ACCESS_TOKEN" | jq .

echo ""
echo "Refresh Token Claims:"
decode_jwt_payload "$REFRESH_TOKEN" | jq .

# Step 3: Test refresh endpoint
echo ""
echo "Step 3: Calling refresh endpoint..."
echo "Payload: {'refresh_token': '${REFRESH_TOKEN:0:20}...'}"

request "POST" "${API_URL}/auth/refresh" "$(jq -nc \
  --arg token "$REFRESH_TOKEN" \
  '{refresh_token:$token}')"

echo "Response status: $REQUEST_STATUS"
echo "Response body:"
echo "$REQUEST_BODY" | jq .

if [[ "$REQUEST_STATUS" == "200" ]]; then
  echo "✓ Refresh successful"
  NEW_ACCESS_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
  NEW_REFRESH_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.refresh_token')"
  echo "  New Access Token: ${NEW_ACCESS_TOKEN:0:20}..."
  echo "  New Refresh Token: ${NEW_REFRESH_TOKEN:0:20}..."
else
  echo "✗ Refresh failed with status $REQUEST_STATUS"
  echo ""
  echo "Debugging info:"
  echo "- Checking if access token works..."
  request "GET" "${API_URL}/auth/me" "" "$ACCESS_TOKEN"
  echo "  GET /auth/me status: $REQUEST_STATUS"
  if [[ "$REQUEST_STATUS" == "200" ]]; then
    echo "  ✓ Access token is valid"
  else
    echo "  ✗ Access token is invalid"
  fi
fi

echo ""
echo "=========================================="
