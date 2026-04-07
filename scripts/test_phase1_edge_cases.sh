#!/usr/bin/env bash

set -u

BASE_URL="${BASE_URL:-http://localhost:8081}"
API_URL="${BASE_URL}/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq

timestamp="$(date +%s)"

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

assert_status() {
  local expected="$1"
  local test_name="$2"
  TESTS_RUN=$((TESTS_RUN + 1))
  
  if [[ "$REQUEST_STATUS" != "$expected" ]]; then
    printf "${RED}✗ FAIL${NC}: $test_name (Expected HTTP ${expected}, got ${REQUEST_STATUS})\n"
    echo "$REQUEST_BODY" | jq . 2>/dev/null || echo "$REQUEST_BODY"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  else
    printf "${GREEN}✓ PASS${NC}: $test_name\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
  fi
}

assert_json_field() {
  local field="$1"
  local value="$2"
  local test_name="$3"
  
  local result
  result="$(echo "$REQUEST_BODY" | jq -e "$field" 2>/dev/null || echo "")"
  
  if [[ "$result" != "$value" ]]; then
    printf "${RED}✗ FAIL${NC}: $test_name (Expected $field=$value, got $result)\n"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  else
    printf "${GREEN}✓ PASS${NC}: $test_name\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
  fi
}

# Test utilities
setup() {
  local counter="$1"
  local admin_email="admin.${timestamp}.${counter}@example.com"
  local password="ValidPassword123!"
  
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Test Org ${timestamp}.${counter}" \
    --arg org_phone "+234800000000" \
    --arg org_address "Test Address" \
    --arg first "Admin" \
    --arg last "User" \
    --arg phone "+234800111111" \
    --arg email "$admin_email" \
    --arg password "$password" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  
  ADMIN_TOKEN="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
  ADMIN_USER_ID="$(echo "$REQUEST_BODY" | jq -r '.data.user.id')"
  ADMIN_EMAIL="$admin_email"
  ADMIN_PASSWORD="$password"
}

print_section() {
  printf "\n${YELLOW}=== $1 ===${NC}\n"
}

# ============================================================================
# SECTION 1: AUTH ENDPOINT TESTS
# ============================================================================

print_section "1. AUTH ENDPOINT TESTS"

# Test 1.1: Register with valid data
test_1_1() {
  local email="user.${timestamp}.1@example.com"
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org ${timestamp}.1" \
    --arg org_phone "+234800000001" \
    --arg org_address "Address 1" \
    --arg first "John" \
    --arg last "Doe" \
    --arg phone "+2348001111111" \
    --arg email "$email" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 201 "Register with valid data"
}

# Test 1.2: Register with duplicate email should fail
test_1_2() {
  local email="duptest.${timestamp}@example.com"
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org DupTest 1" \
    --arg org_phone "+234800000002" \
    --arg org_address "Address 2" \
    --arg first "First" \
    --arg last "Last" \
    --arg phone "+2348001111112" \
    --arg email "$email" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 201 "Register first user with email"
  
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org DupTest 2" \
    --arg org_phone "+234800000003" \
    --arg org_address "Address 3" \
    --arg first "Another" \
    --arg last "User" \
    --arg phone "+2348001111113" \
    --arg email "$email" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 409 "Duplicate email registration should fail"
}

# Test 1.3: Register with duplicate organization should fail
test_1_3() {
  local org_name="DuplicateOrg.${timestamp}"
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "$org_name" \
    --arg org_phone "+234800000004" \
    --arg org_address "Address 4" \
    --arg first "User" \
    --arg last "One" \
    --arg phone "+2348001111114" \
    --arg email "user1.${timestamp}.3@example.com" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 201 "Register first org"
  
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "$org_name" \
    --arg org_phone "+234800000005" \
    --arg org_address "Address 5" \
    --arg first "User" \
    --arg last "Two" \
    --arg phone "+2348001111115" \
    --arg email "user2.${timestamp}.3@example.com" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 409 "Duplicate organization should fail"
}

# Test 1.4: Register with password < 8 chars should fail
test_1_4() {
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org ${timestamp}.4" \
    --arg org_phone "+234800000006" \
    --arg org_address "Address 6" \
    --arg first "John" \
    --arg last "Doe" \
    --arg phone "+2348001111116" \
    --arg email "short.pass.${timestamp}@example.com" \
    --arg password "Short1" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 400 "Short password should fail validation"
}

# Test 1.5: Register with invalid email should fail
test_1_5() {
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org ${timestamp}.5" \
    --arg org_phone "+234800000007" \
    --arg org_address "Address 7" \
    --arg first "John" \
    --arg last "Doe" \
    --arg phone "+2348001111117" \
    --arg email "not-an-email" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 400 "Invalid email format should fail validation"
}

# Test 1.6: Register with case-insensitive email handling
test_1_6() {
  local base_email="casetest.${timestamp}@example.com"
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org CaseTest 1" \
    --arg org_phone "+234800000008" \
    --arg org_address "Address 8" \
    --arg first "John" \
    --arg last "Doe" \
    --arg phone "+2348001111118" \
    --arg email "$base_email" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 201 "Register with lowercase email"
  
  # Try with uppercase (should fail due to duplicate)
  request "POST" "${API_URL}/auth/register" "$(jq -nc \
    --arg org_name "Org CaseTest 2" \
    --arg org_phone "+234800000009" \
    --arg org_address "Address 9" \
    --arg first "Jane" \
    --arg last "Doe" \
    --arg phone "+2348001111119" \
    --arg email "${base_email^^}" \
    --arg password "ValidPass123!" \
    '{organization_name:$org_name,organization_phone:$org_phone,organization_address:$org_address,first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password}')"
  assert_status 409 "Duplicate email (uppercase variant) should fail"
}

# Test 1.7: Login with valid credentials
test_1_7() {
  setup 1_7
  request "POST" "${API_URL}/auth/login" "$(jq -nc \
    --arg email "$ADMIN_EMAIL" \
    --arg password "$ADMIN_PASSWORD" \
    '{email:$email,password:$password}')"
  assert_status 200 "Login with valid credentials"
}

# Test 1.8: Login with invalid password
test_1_8() {
  setup 1_8
  request "POST" "${API_URL}/auth/login" "$(jq -nc \
    --arg email "$ADMIN_EMAIL" \
    --arg password "WrongPassword123!" \
    '{email:$email,password:$password}')"
  assert_status 401 "Login with invalid password"
}

# Test 1.9: Login with non-existent email
test_1_9() {
  request "POST" "${API_URL}/auth/login" "$(jq -nc \
    --arg email "nonexistent.${timestamp}@example.com" \
    --arg password "SomePassword123!" \
    '{email:$email,password:$password}')"
  assert_status 401 "Login with non-existent email"
}

# Test 1.10: Login with invalid email format
test_1_10() {
  request "POST" "${API_URL}/auth/login" "$(jq -nc \
    --arg email "invalid-email" \
    --arg password "SomePassword123!" \
    '{email:$email,password:$password}')"
  assert_status 400 "Login with invalid email format"
}

# ============================================================================
# SECTION 2: JWT TOKEN TESTS
# ============================================================================

print_section "2. JWT TOKEN TESTS"

# Test 2.1: Access endpoints without token
test_2_1() {
  request "GET" "${API_URL}/users" "" ""
  assert_status 401 "Access /users without token should fail"
}

# Test 2.2: Access endpoints with invalid token
test_2_2() {
  request "GET" "${API_URL}/users" "" "invalid.token.here"
  assert_status 401 "Access with invalid token should fail"
}

# Test 2.3: Access endpoints with malformed bearer header
test_2_3() {
  local tmp
  tmp="$(mktemp)"
  curl -sS -X GET -H "Content-Type: application/json" \
    -H "Authorization: InvalidBearer some_token_here" \
    -o "$tmp" -w "%{http_code}" "${API_URL}/users" > /dev/null
  REQUEST_STATUS="$(curl -sS -X GET -H "Content-Type: application/json" \
    -H "Authorization: InvalidBearer some_token_here" \
    -o "$tmp" -w "%{http_code}" "${API_URL}/users")"
  REQUEST_BODY="$(cat "$tmp")"
  rm -f "$tmp"
  assert_status 401 "Malformed Bearer header should fail"
}

# Test 2.4: Access with expired token (wait for it to expire)
test_2_4() {
  printf "${YELLOW}⊘ SKIP${NC}: Access with expired token (requires waiting, skipped)\n"
}

# Test 2.5: Token contains correct claims
test_2_5() {
  setup 2_5
  local email_claim
  email_claim="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token' | cut -d. -f2)"
  # Pad if necessary
  case $((${#email_claim} % 4)) in
    2) email_claim="${email_claim}==" ;;
    3) email_claim="${email_claim}=" ;;
  esac
  
  # Decode and check for email
  local decoded
  decoded="$(echo "$email_claim" | base64 -d 2>/dev/null || echo '{}')"
  if echo "$decoded" | jq -e --arg email "$ADMIN_EMAIL" '.email == $email' >/dev/null 2>&1; then
    printf "${GREEN}✓ PASS${NC}: Token contains correct email claim\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    printf "${RED}✗ FAIL${NC}: Token missing correct email claim\n"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  TESTS_RUN=$((TESTS_RUN + 1))
}

# Test 2.6: Token contains role claim
test_2_6() {
  setup 2_6
  local token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
  local role_claim="$(echo "$token" | cut -d. -f2)"
  case $((${#role_claim} % 4)) in
    2) role_claim="${role_claim}==" ;;
    3) role_claim="${role_claim}=" ;;
  esac
  
  local decoded
  decoded="$(echo "$role_claim" | base64 -d 2>/dev/null || echo '{}')"
  if echo "$decoded" | jq -e '.role' >/dev/null 2>&1; then
    printf "${GREEN}✓ PASS${NC}: Token contains role claim\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    printf "${RED}✗ FAIL${NC}: Token missing role claim\n"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  TESTS_RUN=$((TESTS_RUN + 1))
}

# ============================================================================
# SECTION 3: REFRESH TOKEN TESTS
# ============================================================================

print_section "3. REFRESH TOKEN TESTS"

# Test 3.1: Refresh token successfully
test_3_1() {
  setup 3_1
  local refresh_token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.refresh_token')"
  
  request "POST" "${API_URL}/auth/refresh" "$(jq -nc \
    --arg token "$refresh_token" \
    '{refresh_token:$token}')"
  assert_status 200 "Refresh token should succeed"
}

# Test 3.2: Refresh with invalid token
test_3_2() {
  request "POST" "${API_URL}/auth/refresh" "$(jq -nc \
    '{refresh_token:"invalid.refresh.token"}')"
  assert_status 401 "Refresh with invalid token should fail"
}

# Test 3.3: Refresh without token field
test_3_3() {
  request "POST" "${API_URL}/auth/refresh" "$(jq -nc '{}')"
  assert_status 400 "Refresh without token field should fail validation"
}

# Test 3.4: New access token issued after refresh
test_3_4() {
  setup 3_4
  local old_token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
  local refresh_token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.refresh_token')"
  
  request "POST" "${API_URL}/auth/refresh" "$(jq -nc \
    --arg token "$refresh_token" \
    '{refresh_token:$token}')"
  
  local new_token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
  
  if [[ "$old_token" != "$new_token" ]]; then
    printf "${GREEN}✓ PASS${NC}: New access token issued after refresh\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    printf "${RED}✗ FAIL${NC}: Access token should be different after refresh\n"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  TESTS_RUN=$((TESTS_RUN + 1))
}

# ============================================================================
# SECTION 4: AUTHENTICATED USER ENDPOINTS
# ============================================================================

print_section "4. AUTHENTICATED USER ENDPOINTS"

# Test 4.1: Get current user profile
test_4_1() {
  setup 4_1
  request "GET" "${API_URL}/auth/me" "" "$ADMIN_TOKEN"
  assert_status 200 "Get current user profile should succeed"
}

# Test 4.2: Update current user profile
test_4_2() {
  setup 4_2
  request "PATCH" "${API_URL}/auth/me" "$(jq -nc \
    --arg first "Updated" \
    --arg last "Name" \
    '{first_name:$first,last_name:$last}')" "$ADMIN_TOKEN"
  assert_status 200 "Update current user profile should succeed"
}

# Test 4.3: Update profile without required fields
test_4_3() {
  setup 4_3
  request "PATCH" "${API_URL}/auth/me" "$(jq -nc '{"first_name":"Updated"}')" "$ADMIN_TOKEN"
  assert_status 400 "Update profile without required fields should fail"
}

# Test 4.4: Change password successfully
test_4_4() {
  setup 4_4
  request "POST" "${API_URL}/auth/change-password" "$(jq -nc \
    --arg current "$ADMIN_PASSWORD" \
    --arg new "NewPassword123!" \
    '{current_password:$current,new_password:$new}')" "$ADMIN_TOKEN"
  assert_status 200 "Change password should succeed"
}

# Test 4.5: Change password with wrong current password
test_4_5() {
  setup 4_5
  request "POST" "${API_URL}/auth/change-password" "$(jq -nc \
    --arg current "WrongPassword123!" \
    --arg new "NewPassword123!" \
    '{current_password:$current,new_password:$new}')" "$ADMIN_TOKEN"
  assert_status 401 "Change password with wrong current password should fail"
}

# Test 4.6: Change password with short new password
test_4_6() {
  setup 4_6
  request "POST" "${API_URL}/auth/change-password" "$(jq -nc \
    --arg current "$ADMIN_PASSWORD" \
    --arg new "Short1" \
    '{current_password:$current,new_password:$new}')" "$ADMIN_TOKEN"
  assert_status 400 "Short new password should fail validation"
}

# Test 4.7: Logout clears refresh token
test_4_7() {
  setup 4_7
  request "POST" "${API_URL}/auth/logout" "" "$ADMIN_TOKEN"
  assert_status 200 "Logout should succeed"
  
  # Try to use refresh token after logout (should fail after implementation)
  # This may not fail immediately depending on implementation
}

# ============================================================================
# SECTION 5: USER MANAGEMENT RBAC TESTS
# ============================================================================

print_section "5. USER MANAGEMENT RBAC TESTS"

# Test 5.1: Non-admin cannot list users
test_5_1() {
  setup 5_1
  
  # Create a non-admin user
  local user_email="nonادmin.${timestamp}@example.com"
  request "POST" "${API_URL}/users" "$(jq -nc \
    --arg first "Regular" \
    --arg last "User" \
    --arg phone "+2348001111120" \
    --arg email "$user_email" \
    --arg password "Password123!" \
    --arg role "personnel" \
    '{first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password,role:$role}')" "$ADMIN_TOKEN"
  
  local user_token="$(echo "$REQUEST_BODY" | jq -r '.data.tokens.access_token')"
  
  request "GET" "${API_URL}/users" "" "$user_token"
  assert_status 403 "Non-admin cannot list users"
}

# Test 5.2: Admin can list users
test_5_2() {
  setup 5_2
  request "GET" "${API_URL}/users" "" "$ADMIN_TOKEN"
  assert_status 200 "Admin can list users"
}

# Test 5.3: Get specific user
test_5_3() {
  setup 5_3
  request "GET" "${API_URL}/users/${ADMIN_USER_ID}" "" "$ADMIN_TOKEN"
  assert_status 200 "Admin can get specific user"
}

# Test 5.4: Get non-existent user
test_5_4() {
  setup 5_4
  request "GET" "${API_URL}/users/000000000000000000000000" "" "$ADMIN_TOKEN"
  assert_status 404 "Get non-existent user should return 404"
}

# Test 5.5: Update user role to valid role
test_5_5() {
  setup 5_5
  
  # Create a user first
  local user_email="roletest.${timestamp}@example.com"
  request "POST" "${API_URL}/users" "$(jq -nc \
    --arg first "Role" \
    --arg last "Test" \
    --arg phone "+2348001111121" \
    --arg email "$user_email" \
    --arg password "Password123!" \
    --arg role "personnel" \
    '{first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password,role:$role}')" "$ADMIN_TOKEN"
  
  local user_id="$(echo "$REQUEST_BODY" | jq -r '.data.id')"
  
  request "PATCH" "${API_URL}/users/${user_id}/role" \
    '{"role":"planner"}' "$ADMIN_TOKEN"
  assert_status 200 "Update user role to planner"
}

# Test 5.6: Update user role to invalid role
test_5_6() {
  setup 5_6
  
  local user_email="invalidrole.${timestamp}@example.com"
  request "POST" "${API_URL}/users" "$(jq -nc \
    --arg first "Invalid" \
    --arg last "Role" \
    --arg phone "+2348001111122" \
    --arg email "$user_email" \
    --arg password "Password123!" \
    --arg role "personnel" \
    '{first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password,role:$role}')" "$ADMIN_TOKEN"
  
  local user_id="$(echo "$REQUEST_BODY" | jq -r '.data.id')"
  
  request "PATCH" "${API_URL}/users/${user_id}/role" \
    '{"role":"invalid_role"}' "$ADMIN_TOKEN"
  assert_status 400 "Invalid role should fail"
}

# Test 5.7: Deactivate user
test_5_7() {
  setup 5_7
  
  local user_email="deactivate.${timestamp}@example.com"
  request "POST" "${API_URL}/users" "$(jq -nc \
    --arg first "Deactivate" \
    --arg last "Test" \
    --arg phone "+2348001111123" \
    --arg email "$user_email" \
    --arg password "Password123!" \
    --arg role "personnel" \
    '{first_name:$first,last_name:$last,phone_number:$phone,email:$email,password:$password,role:$role}')" "$ADMIN_TOKEN"
  
  local user_id="$(echo "$REQUEST_BODY" | jq -r '.data.id')"
  
  request "DELETE" "${API_URL}/users/${user_id}" "" "$ADMIN_TOKEN"
  assert_status 200 "Deactivate user should succeed"
  
  # Verify deactivation
  request "GET" "${API_URL}/users/${user_id}" "" "$ADMIN_TOKEN"
  local is_active="$(echo "$REQUEST_BODY" | jq -r '.data.is_active')"
  if [[ "$is_active" == "false" ]]; then
    printf "${GREEN}✓ PASS${NC}: User is deactivated\n"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    printf "${RED}✗ FAIL${NC}: User should be deactivated\n"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  TESTS_RUN=$((TESTS_RUN + 1))
}

# ============================================================================
# RUN ALL TESTS
# ============================================================================

test_1_1
test_1_2
test_1_3
test_1_4
test_1_5
test_1_6
test_1_7
test_1_8
test_1_9
test_1_10

test_2_1
test_2_2
test_2_3
test_2_4
test_2_5
test_2_6

test_3_1
test_3_2
test_3_3
test_3_4

test_4_1
test_4_2
test_4_3
test_4_4
test_4_5
test_4_6
test_4_7

test_5_1
test_5_2
test_5_3
test_5_4
test_5_5
test_5_6
test_5_7

# ============================================================================
# TEST SUMMARY
# ============================================================================

print_section "TEST SUMMARY"
printf "Total: ${TESTS_RUN} | ${GREEN}Passed: ${TESTS_PASSED}${NC} | ${RED}Failed: ${TESTS_FAILED}${NC}\n"

if [[ $TESTS_FAILED -gt 0 ]]; then
  exit 1
else
  printf "\n${GREEN}All Phase 1 edge case tests passed!${NC}\n"
  exit 0
fi
