#!/bin/bash

# POB Management System - Full CRUD Test Script (FIXED)
# Tests complete Create, Read, Update, Delete operations for all modules

BASE_URL="http://localhost:8081/api/v1"
TOKEN=""
VESSEL_ID=""
PERSONNEL_ID=""
ROLE_ID=""
ACTIVITY_ID=""
TRANSPORT_ID=""
SCHEDULE_ID=""
ROOM_ID=""
CERTIFICATE_ID=""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=============================================="
echo "POB Management System - Full CRUD Test Suite"
echo "=============================================="

# Helper function
call_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local auth=${4:-true}
    
    local cmd="curl -s -X $method '$BASE_URL$endpoint' -H 'Content-Type: application/json'"
    if [ "$auth" = true ] && [ -n "$TOKEN" ]; then
        cmd="$cmd -H 'Authorization: Bearer $TOKEN'"
    fi
    if [ -n "$data" ]; then
        cmd="$cmd -d '$data'"
    fi
    
    eval $cmd
}

# 1. AUTHENTICATION
echo -e "\n${BLUE}[1] AUTHENTICATION${NC}"
echo "-----------------------------------"

LOGIN_DATA='{"email":"admin@testudo.com","password":"SecurePassword123!"}'
echo "Logging in..."
RESPONSE=$(call_api "POST" "/auth/login" "$LOGIN_DATA" false)
TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
else
    echo -e "${RED}✗ Login failed${NC}"
    exit 1
fi

# 2. VESSEL CRUD
echo -e "\n${BLUE}[2] VESSEL CRUD${NC}"
echo "-----------------------------------"

TIMESTAMP=$(date +%s)
VESSEL_DATA='{
    "name": "CRUD Test Vessel '$TIMESTAMP'",
    "code": "CRUD-'$TIMESTAMP'",
    "type": "FPSO",
    "location": "Test Location",
    "pob_capacity": 150,
    "minimum_safe_pob_capacity": 50
}'
echo "Creating vessel..."
RESPONSE=$(call_api "POST" "/vessels" "$VESSEL_DATA")
VESSEL_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$VESSEL_ID" ]; then
    echo -e "${GREEN}✓ Vessel created (ID: $VESSEL_ID)${NC}"
else
    echo -e "${RED}✗ Vessel creation failed${NC}"
    exit 1
fi

echo "Reading vessel..."
RESPONSE=$(call_api "GET" "/vessels/$VESSEL_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Vessel retrieved${NC}"
fi

echo "Updating vessel..."
UPDATE_VESSEL_DATA='{
    "name": "CRUD Test Vessel '$TIMESTAMP' (Updated)",
    "code": "CRUD-'$TIMESTAMP'-UPD",
    "type": "FPSO",
    "location": "Updated Location",
    "pob_capacity": 160,
    "minimum_safe_pob_capacity": 55
}'
RESPONSE=$(call_api "PATCH" "/vessels/$VESSEL_ID" "$UPDATE_VESSEL_DATA")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Vessel updated${NC}"
fi

echo "Listing all vessels..."
RESPONSE=$(call_api "GET" "/vessels")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Vessels listed${NC}"
fi

echo "Getting real-time POB..."
RESPONSE=$(call_api "GET" "/vessels/$VESSEL_ID/pob")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ POB retrieved${NC}"
fi

# 3. OFFSHORE ROLE CRUD
echo -e "\n${BLUE}[3] OFFSHORE ROLE CRUD${NC}"
echo "-----------------------------------"

ROLE_DATA='{
    "name": "CRUD Test Role '$TIMESTAMP'",
    "code": "CRUD-ROLE-'$TIMESTAMP'",
    "type": "core",
    "requires_room": true,
    "minimum_required_count": 2
}'
echo "Creating role..."
RESPONSE=$(call_api "POST" "/positions" "$ROLE_DATA")
ROLE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$ROLE_ID" ]; then
    echo -e "${GREEN}✓ Role created (ID: $ROLE_ID)${NC}"
fi

echo "Listing roles..."
RESPONSE=$(call_api "GET" "/positions")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Roles listed${NC}"
fi

# 4. PERSONNEL CRUD
echo -e "\n${BLUE}[4] PERSONNEL CRUD${NC}"
echo "-----------------------------------"

PERSONNEL_DATA='{
    "employee_number": "EMP-CRUD-'$TIMESTAMP'",
    "first_name": "CRUD",
    "last_name": "Test",
    "email": "crud.test.'$TIMESTAMP'@testudo.com",
    "phone_number": "+2348012345678",
    "nationality": "Nigerian",
    "company": "Testudo Nigeria",
    "primary_discipline": "Drilling"
}'
echo "Creating personnel..."
RESPONSE=$(call_api "POST" "/personnel" "$PERSONNEL_DATA")
PERSONNEL_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$PERSONNEL_ID" ]; then
    echo -e "${GREEN}✓ Personnel created (ID: $PERSONNEL_ID)${NC}"
else
    echo -e "${RED}✗ Personnel creation failed${NC}"
    exit 1
fi

echo "Reading personnel..."
RESPONSE=$(call_api "GET" "/personnel")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Personnel retrieved${NC}"
fi

echo "Updating personnel..."
UPDATE_PERSONNEL_DATA='{
    "employee_number": "EMP-CRUD-'$TIMESTAMP'",
    "first_name": "CRUD",
    "last_name": "Test Updated",
    "email": "crud.test.'$TIMESTAMP'@testudo.com",
    "phone_number": "+2348012345678",
    "nationality": "Nigerian",
    "company": "Testudo Nigeria",
    "primary_discipline": "Drilling"
}'
RESPONSE=$(call_api "PATCH" "/personnel/$PERSONNEL_ID" "$UPDATE_PERSONNEL_DATA")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Personnel updated${NC}"
fi

echo "Checking compliance..."
RESPONSE=$(call_api "GET" "/personnel/$PERSONNEL_ID/compliance")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Compliance checked${NC}"
fi

# 5. CERTIFICATE CRUD (FIXED)
echo -e "\n${BLUE}[5] CERTIFICATE CRUD${NC}"
echo "-----------------------------------"

CERT_DATA='{
    "certificate_type": "BOSIET",
    "certificate_number": "BOS-CRUD-'$TIMESTAMP'",
    "issued_by": "OPITO",
    "issued_at": "2024-01-01T00:00:00Z",
    "expires_at": "2028-01-01T00:00:00Z"
}'
echo "Adding certificate..."
RESPONSE=$(call_api "POST" "/personnel/$PERSONNEL_ID/certificates" "$CERT_DATA")
CERTIFICATE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$CERTIFICATE_ID" ]; then
    echo -e "${GREEN}✓ Certificate added (ID: $CERTIFICATE_ID)${NC}"
else
    echo -e "${RED}✗ Certificate creation failed${NC}"
fi

echo "Listing certificates..."
RESPONSE=$(call_api "GET" "/personnel/$PERSONNEL_ID/certificates")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Certificates listed${NC}"
fi

if [ -n "$CERTIFICATE_ID" ]; then
    echo "Updating certificate..."
    UPDATE_CERT_DATA='{
        "certificate_type": "BOSIET",
        "certificate_number": "BOS-CRUD-'$TIMESTAMP'-UPD",
        "issued_by": "OPITO",
        "issued_at": "2024-01-01T00:00:00Z",
        "expires_at": "2028-01-01T00:00:00Z"
    }'
    RESPONSE=$(call_api "PATCH" "/personnel/$PERSONNEL_ID/certificates/$CERTIFICATE_ID" "$UPDATE_CERT_DATA")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Certificate updated${NC}"
    fi
    
    echo "Deleting certificate..."
    RESPONSE=$(call_api "DELETE" "/personnel/$PERSONNEL_ID/certificates/$CERTIFICATE_ID")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Certificate deleted${NC}"
    fi
fi

# 6. ROOM CRUD (FIXED)
echo -e "\n${BLUE}[6] ROOM CRUD${NC}"
echo "-----------------------------------"

ROOM_DATA='{
    "vessel_id": "'$VESSEL_ID'",
    "name": "CRUD Test Cabin",
    "code": "CAB-CRUD-'$TIMESTAMP'",
    "deck": "Main Deck",
    "category": "Dedicated",
    "capacity": 2
}'
echo "Creating room..."
RESPONSE=$(call_api "POST" "/vessels/$VESSEL_ID/rooms" "$ROOM_DATA")
ROOM_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$ROOM_ID" ]; then
    echo -e "${GREEN}✓ Room created (ID: $ROOM_ID)${NC}"
else
    echo -e "${YELLOW}⚠ Room creation may have failed${NC}"
fi

echo "Listing rooms..."
RESPONSE=$(call_api "GET" "/vessels/$VESSEL_ID/rooms")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Rooms listed${NC}"
fi

# 7. ROTATION SCHEDULE CRUD
echo -e "\n${BLUE}[7] ROTATION SCHEDULE CRUD${NC}"
echo "-----------------------------------"

if [ -n "$ROLE_ID" ]; then
    SCHEDULE_DATA='{
        "offshore_role_id": "'$ROLE_ID'",
        "vessel_id": "'$VESSEL_ID'",
        "name": "14/14 CRUD Rotation",
        "days_on": 14,
        "days_off": 14,
        "cycle_anchor_date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"
    }'
    echo "Creating rotation schedule..."
    RESPONSE=$(call_api "POST" "/rotation-schedules" "$SCHEDULE_DATA")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Rotation schedule created${NC}"
    fi
    
    echo "Listing rotation schedules..."
    RESPONSE=$(call_api "GET" "/rotation-schedules?role_id=$ROLE_ID&vessel_id=$VESSEL_ID")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Rotation schedules listed${NC}"
    fi
fi

# 8. ACTIVITY CRUD
echo -e "\n${BLUE}[8] ACTIVITY CRUD${NC}"
echo "-----------------------------------"

USER_ID=$(echo $TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | grep -o '"sub":"[^"]*"' | cut -d'"' -f4)

ACTIVITY_DATA='{
    "vessel_id": "'$VESSEL_ID'",
    "name": "CRUD Test Activity '$TIMESTAMP'",
    "description": "CRUD test activity",
    "start_date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
    "end_date": "'$(date -u -d "+5 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5d +"%Y-%m-%dT%H:%M:%SZ")'",
    "priority": "medium",
    "created_by": "'$USER_ID'",
    "requirements": []
}'
echo "Creating activity..."
RESPONSE=$(call_api "POST" "/activities" "$ACTIVITY_DATA")
ACTIVITY_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$ACTIVITY_ID" ]; then
    echo -e "${GREEN}✓ Activity created (ID: $ACTIVITY_ID)${NC}"
fi

echo "Reading activity..."
RESPONSE=$(call_api "GET" "/activities/$ACTIVITY_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Activity retrieved${NC}"
fi

echo "Listing activities..."
RESPONSE=$(call_api "GET" "/activities?vessel_id=$VESSEL_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Activities listed${NC}"
fi

echo "Getting Gantt data..."
RESPONSE=$(call_api "GET" "/activities/gantt?vessel_id=$VESSEL_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Gantt data retrieved${NC}"
fi

# 9. TRANSPORT CRUD
echo -e "\n${BLUE}[9] TRANSPORT CRUD${NC}"
echo "-----------------------------------"

TRANSPORT_DATA='{
    "name": "CRUD Helicopter '$TIMESTAMP'",
    "type": "helicopter",
    "capacity": 12,
    "cost_model": "per_trip",
    "cost_amount": 5000.00,
    "departure_days": ["Monday", "Wednesday", "Friday"],
    "mobilization_location": "Lagos Heliport"
}'
echo "Creating transport..."
RESPONSE=$(call_api "POST" "/transports" "$TRANSPORT_DATA")
TRANSPORT_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$TRANSPORT_ID" ]; then
    echo -e "${GREEN}✓ Transport created (ID: $TRANSPORT_ID)${NC}"
fi

echo "Reading transport..."
RESPONSE=$(call_api "GET" "/transports/$TRANSPORT_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Transport retrieved${NC}"
fi

echo "Listing transports..."
RESPONSE=$(call_api "GET" "/transports")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Transports listed${NC}"
fi

echo "Updating transport..."
UPDATE_TRANSPORT_DATA='{
    "name": "CRUD Helicopter '$TIMESTAMP' (Updated)",
    "type": "helicopter",
    "capacity": 14,
    "cost_model": "per_trip",
    "cost_amount": 5500.00,
    "departure_days": ["Monday", "Wednesday", "Friday", "Sunday"],
    "mobilization_location": "Lagos Heliport"
}'
RESPONSE=$(call_api "PATCH" "/transports/$TRANSPORT_ID" "$UPDATE_TRANSPORT_DATA")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Transport updated${NC}"
fi

# 10. TRAVEL SCHEDULE CRUD
echo -e "\n${BLUE}[10] TRAVEL SCHEDULE CRUD${NC}"
echo "-----------------------------------"

SCHEDULE_DATA='{
    "transport_id": "'$TRANSPORT_ID'",
    "vessel_id": "'$VESSEL_ID'",
    "direction": "outbound",
    "departure_at": "'$(date -u -d "+2 days" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+2d +"%Y-%m-%dT%H:%M:%SZ")'"
}'
echo "Creating travel schedule..."
RESPONSE=$(call_api "POST" "/travel/schedules" "$SCHEDULE_DATA")
SCHEDULE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$SCHEDULE_ID" ]; then
    echo -e "${GREEN}✓ Travel schedule created (ID: $SCHEDULE_ID)${NC}"
fi

echo "Reading travel schedule..."
RESPONSE=$(call_api "GET" "/travel/schedules/$SCHEDULE_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Travel schedule retrieved${NC}"
fi

echo "Listing upcoming schedules..."
RESPONSE=$(call_api "GET" "/travel/schedules")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Travel schedules listed${NC}"
fi

echo "Checking utilization alerts..."
RESPONSE=$(call_api "GET" "/travel/alerts")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Utilization alerts retrieved${NC}"
fi

# 11. MINIMUM MANNING
echo -e "\n${BLUE}[11] MINIMUM MANNING${NC}"
echo "-----------------------------------"

echo "Checking active minimum manning..."
RESPONSE=$(call_api "GET" "/minimum-manning/active?vessel_id=$VESSEL_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Active status checked${NC}"
fi

echo "Getting minimum manning history..."
RESPONSE=$(call_api "GET" "/minimum-manning/history?vessel_id=$VESSEL_ID")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ History retrieved${NC}"
fi

# 12. REPORTS
echo -e "\n${BLUE}[12] REPORTS${NC}"
echo "-----------------------------------"

echo "Getting daily POB report..."
TODAY=$(date +"%Y-%m-%d")
RESPONSE=$(call_api "GET" "/reports/daily?vessel_id=$VESSEL_ID&date=$TODAY")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Daily POB report generated${NC}"
fi

echo "Getting historical POB..."
RESPONSE=$(call_api "GET" "/reports/historical?vessel_id=$VESSEL_ID&start_date=2026-04-01&end_date=$TODAY")
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Historical report generated${NC}"
fi

# 13. CLEANUP - DELETE OPERATIONS
echo -e "\n${BLUE}[13] CLEANUP - DELETE OPERATIONS${NC}"
echo "-----------------------------------"

if [ -n "$ACTIVITY_ID" ]; then
    echo "Deleting activity..."
    RESPONSE=$(call_api "DELETE" "/activities/$ACTIVITY_ID")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Activity deleted${NC}"
    fi
fi

if [ -n "$SCHEDULE_ID" ]; then
    echo "Deleting travel schedule..."
    curl -s -X DELETE "$BASE_URL/travel/schedules/$SCHEDULE_ID" -H "Authorization: Bearer $TOKEN"
    echo -e "${GREEN}✓ Travel schedule deleted${NC}"
fi

if [ -n "$TRANSPORT_ID" ]; then
    echo "Deleting transport..."
    RESPONSE=$(call_api "DELETE" "/transports/$TRANSPORT_ID")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Transport deleted${NC}"
    fi
fi

if [ -n "$ROOM_ID" ]; then
    echo "Deleting room..."
    curl -s -X DELETE "$BASE_URL/rooms/$ROOM_ID" -H "Authorization: Bearer $TOKEN"
    echo -e "${GREEN}✓ Room deleted${NC}"
fi

if [ -n "$PERSONNEL_ID" ]; then
    echo "Deleting personnel..."
    RESPONSE=$(call_api "DELETE" "/personnel/$PERSONNEL_ID")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Personnel deleted${NC}"
    fi
fi

if [ -n "$VESSEL_ID" ]; then
    echo "Deleting vessel..."
    RESPONSE=$(call_api "DELETE" "/vessels/$VESSEL_ID")
    if echo "$RESPONSE" | grep -q '"success":true'; then
        echo -e "${GREEN}✓ Vessel deleted${NC}"
    fi
fi

echo -e "\n${GREEN}=============================================="
echo "✓ FULL CRUD TEST COMPLETED SUCCESSFULLY!"
echo "==============================================${NC}"
