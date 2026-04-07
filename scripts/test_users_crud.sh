#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8081}"
API_URL="${BASE_URL}/api/v1"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq

timestamp="$(date +%s)"
admin_email="sysadmin.${timestamp}@example.com"
user_email="user.${timestamp}@example.com"
password="Password123!"
updated_email="updated.${timestamp}@example.com"
organization_name="Org ${timestamp}"
organization_phone="+234800000${timestamp: -4}"
organization_address="12 Marina Road, Lagos"

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

assert_status() {
  local expected="$1"
  if [[ "$REQUEST_STATUS" != "$expected" ]]; then
    echo "Expected HTTP ${expected}, got ${REQUEST_STATUS}" >&2
    echo "$REQUEST_BODY" | jq . >&2 || echo "$REQUEST_BODY" >&2
    exit 1
  fi
}

echo "Registering sys admin user..."
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org_name "$organization_name" \
  --arg org_phone "$organization_phone" \
  --arg org_address "$organization_address" \
  --arg first "System" \
  --arg last "Admin" \
  --arg phone "+2348011111111" \
  --arg email "$admin_email" \
  --arg password "$password" \
  '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
assert_status 201
admin_token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
echo "Creating regular user..."
request "POST" "${API_URL}/users" "$(jq -nc \
  --arg first "Regular" \
  --arg last "User" \
  --arg phone "+2348022222222" \
  --arg email "$user_email" \
  --arg password "$password" \
  --arg role "personnel" \
  '{first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password,role:$role}')" "$admin_token"
assert_status 201
user_id="$(echo "$REQUEST_BODY" | jq -r '.data.id')"

echo "Listing users..."
request "GET" "${API_URL}/users" "" "$admin_token"
assert_status 200
echo "$REQUEST_BODY" | jq -e --arg id "$user_id" '.data[] | select(.id == $id)' >/dev/null

echo "Fetching created user..."
request "GET" "${API_URL}/users/${user_id}" "" "$admin_token"
assert_status 200
echo "$REQUEST_BODY" | jq -e --arg id "$user_id" '.data.id == $id' >/dev/null

echo "Updating user profile..."
request "PATCH" "${API_URL}/users/${user_id}" "$(jq -nc \
  --arg first "Updated" \
  --arg last "Person" \
  --arg phone "+2348033333333" \
  --arg email "$updated_email" \
  '{first_name:$first,last_name:$last,phone_number:$phone,email:$email}')" "$admin_token"
assert_status 200
echo "$REQUEST_BODY" | jq -e --arg email "$updated_email" '.data.email == $email' >/dev/null

echo "Updating user role..."
request "PATCH" "${API_URL}/users/${user_id}/role" '{"role":"planner"}' "$admin_token"
assert_status 200
echo "$REQUEST_BODY" | jq -e '.data.role == "planner"' >/dev/null

echo "Deactivating user..."
request "DELETE" "${API_URL}/users/${user_id}" "" "$admin_token"
assert_status 200

echo "Verifying deactivation..."
request "GET" "${API_URL}/users/${user_id}" "" "$admin_token"
assert_status 200
echo "$REQUEST_BODY" | jq -e '.data.is_active == false' >/dev/null

echo "User CRUD flow verified successfully."
