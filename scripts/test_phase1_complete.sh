#!/usr/bin/env bash

set -u

BASE_URL="${BASE_URL:-http://localhost:8081}"
API_URL="${BASE_URL}/api/v1"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

require_cmd curl jq

request() {
  local method="$1" url="$2" data="${3:-}" token="${4:-}"
  local tmp="$(mktemp)"
  local curl_args=(-sS -X "$method" -H "Content-Type: application/json" -o "$tmp" -w "%{http_code}")
  [[ -n "$token" ]] && curl_args+=(-H "Authorization: Bearer ${token}")
  [[ -n "$data" ]] && curl_args+=(-d "$data")
  local status="$(curl "${curl_args[@]}" "$url")"
  REQUEST_STATUS="$status"
  REQUEST_BODY="$(cat "$tmp")"
  rm -f "$tmp"
}

assert_status() {
  local expected="$1" test_name="$2"
  TESTS_RUN=$((TESTS_RUN + 1))
  if [[ "$REQUEST_STATUS" == "$expected" ]]; then
    printf "${GREEN}✓${NC} $test_name\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    printf "${RED}✗${NC} $test_name (Expected $expected, got $REQUEST_STATUS)\n"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
}

ts="$(date +%s)"

printf "${YELLOW}Phase 1 - Foundation Tests${NC}\n\n"

# Test 1: Register with valid data
printf "${YELLOW}1. Authentication${NC}\n"
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org "Org${ts}.1" --arg first "Admin" --arg last "User" --arg phone "+2348001111111" \
  --arg email "admin${ts}.1@test.com" --arg password "ValidPass123!" \
  '{organization_name:$org,organization_phone:"+234700000001",organization_address:"Test",first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
assert_status 201 "Register system admin"
TOKEN1="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
USER1_ID="$(echo "$REQUEST_BODY" | jq -r '.data.user.id')"
REFRESH_TOKEN1="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.refresh_token')"

# Test 2: Duplicate email fails
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org "Org${ts}.2" --arg email "admin${ts}.1@test.com" \
  '{organization_name:$org,organization_phone:"+234700000002",organization_address:"Test",first_name:"User",last_name:"Two",phone_number:"+2348001111112",email:$email,password:"ValidPass123!"}')"
assert_status 409 "Duplicate email rejected"

# Test 3: Login with valid credentials
request "POST" "${API_URL}/auth/login" "$(jq -nc --arg email "admin${ts}.1@test.com" \
  '{email:$email,password:"ValidPass123!"}')"
assert_status 200 "Login successful"

# Test 4: Login with invalid password
request "POST" "${API_URL}/auth/login" "$(jq -nc --arg email "admin${ts}.1@test.com" \
  '{email:$email,password:"WrongPassword!"}')"
assert_status 401 "Invalid password rejected"

# Test 5: Refresh token works
request "POST" "${API_URL}/auth/refresh" "$(jq -nc --arg token "$REFRESH_TOKEN1" \
  '{refresh_token:$token}')"
assert_status 200 "Refresh token succeeds"
NEW_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"

# Test 6: Get current user
request "GET" "${API_URL}/auth/me" "" "$TOKEN1"

assert_status 200 "Get current user profile"

# Test 7: Update profile
request "PATCH" "${API_URL}/auth/me" '{"first_name":"Updated","last_name":"Name"}' "$TOKEN1"
assert_status 200 "Update user profile"

# Test 8: Change password
request "POST" "${API_URL}/auth/change-password" \
  '{"current_password":"ValidPass123!","new_password":"NewPassword123!"}' "$TOKEN1"
assert_status 200 "Change password"

# Test 9: Logout
request "POST" "${API_URL}/auth/logout" "" "$TOKEN1"
assert_status 200 "Logout"

printf "\n${YELLOW}2. User Management${NC}\n"

# Create new admin for user management tests
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org "Org${ts}.3" --arg email "admin${ts}.2@test.com" \
  '{organization_name:$org,organization_phone:"+234700000003",organization_address:"Test",first_name:"Admin",last_name:"Two",phone_number:"+2348001111113",email:$email,password:"ValidPass123!"}')"
ADMIN_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"

# Test 10: Create user
request "POST" "${API_URL}/users" "$(jq -nc \
  --arg email "user${ts}@test.com" \
  '{first_name:"Regular",last_name:"User",phone_number:"+2348001111114",email:$email,password:"UserPass123!",role:"personnel"}')" "$ADMIN_TOKEN"
assert_status 201 "Create user"
USER2_ID="$(echo "$REQUEST_BODY" | jq -r '.data.id')"

# Test 11: List users
request "GET" "${API_URL}/users" "" "$ADMIN_TOKEN"
assert_status 200 "List users"

# Test 12: Get user by ID
request "GET" "${API_URL}/users/${USER2_ID}" "" "$ADMIN_TOKEN"
assert_status 200 "Get user by ID"

# Test 13: Update user
request "PATCH" "${API_URL}/users/${USER2_ID}" \
  '{"first_name":"Updated","last_name":"User","phone_number":"+2348001111114","email":"user'${ts}'@test.com"}' "$ADMIN_TOKEN"
assert_status 200 "Update user"

# Test 14: Change user role
request "PATCH" "${API_URL}/users/${USER2_ID}/role" '{"role":"planner"}' "$ADMIN_TOKEN"
assert_status 200 "Change user role"

# Test 15: Deactivate user
request "DELETE" "${API_URL}/users/${USER2_ID}" "" "$ADMIN_TOKEN"
assert_status 200 "Deactivate user"

printf "\n${YELLOW}3. JWT & RBAC${NC}\n"

# Test 16: Access without token fails
request "GET" "${API_URL}/users" "" ""
assert_status 401 "Access without token denied"

# Test 17: Access with invalid token fails
request "GET" "${API_URL}/users" "" "invalid.token.here"
assert_status 401 "Invalid token rejected"

# Test 18: Non-admin cannot list users
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org "Org${ts}.4" --arg email "user${ts}.2@test.com" \
  '{organization_name:$org,organization_phone:"+234700000004",organization_address:"Test",first_name:"Regular",last_name:"User",phone_number:"+2348001111115",email:$email,password:"ValidPass123!"}')"
USER_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"

request "GET" "${API_URL}/users" "" "$USER_TOKEN"
assert_status 403 "Permission denied for non-admin"

printf "\n${YELLOW}4. Input Validation${NC}\n"

# Test 19: Short password rejected
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org "Org${ts}.5" --arg email "short${ts}@test.com" \
  '{organization_name:$org,organization_phone:"+234700000005",organization_address:"Test",first_name:"Test",last_name:"User",phone_number:"+2348001111116",email:$email,password:"Short1"}')"
assert_status 400 "Short password rejected"

# Test 20: Invalid email rejected
request "POST" "${API_URL}/auth/register" "$(jq -nc \
  --arg org "Org${ts}.6" \
  '{organization_name:$org,organization_phone:"+234700000006",organization_address:"Test",first_name:"Test",last_name:"User",phone_number:"+2348001111117",email:"invalid-email",password:"ValidPass123!"}')"
assert_status 400 "Invalid email rejected"

printf "\n${YELLOW}Summary${NC}\n"
printf "Total: ${TESTS_RUN} | ${GREEN}Passed: ${TESTS_PASSED}${NC} | ${RED}Failed: ${TESTS_FAILED}${NC}\n"

if [[ $TESTS_FAILED -eq 0 ]]; then
  printf "${GREEN}✓ All Phase 1 tests passed!${NC}\n"
  exit 0
else
  printf "${RED}✗ Some tests failed${NC}\n"
  exit 1
fi
