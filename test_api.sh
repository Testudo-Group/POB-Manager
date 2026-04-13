#!/bin/bash
BASE_URL="http://localhost:8081/api/v1"

echo "Ì¥ê Logging in..."
RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@testudo.com", "password": "SecurePassword123!"}')

TOKEN=$(echo $RESPONSE | jq -r '.data.tokens.access_token')
echo "‚úÖ Token obtained"

echo -e "\nÌ±§ Getting current user..."
curl -s -X GET $BASE_URL/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq '.data | {id, email, first_name, last_name, role}'

echo -e "\nÌ≥ã Creating offshore position..."
curl -s -X POST $BASE_URL/positions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Drilling Supervisor",
    "code": "DRL-SUP-01",
    "type": "core",
    "description": "Supervises all drilling operations",
    "rotation_pattern": "14/14"
  }' | jq '.data | {id, title, code, type}'

echo -e "\nÌ±®‚ÄçÌ¥ß Creating personnel..."
curl -s -X POST $BASE_URL/personnel \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Offshore",
    "email": "john.offshore@testudo.com",
    "phone": "+2348012345678",
    "nationality": "Nigerian"
  }' | jq '.data | {id, first_name, last_name, email}'

echo -e "\nÌ∫¢ Creating vessel..."
curl -s -X POST $BASE_URL/vessels \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "FPSO Testudo Alpha",
    "type": "FPSO",
    "imo_number": "IMO1234567",
    "capacity": 120,
    "location": "OML 123 - Offshore Nigeria"
  }' | jq '.data | {id, name, type, pob_capacity}'

echo -e "\nÌ≥ä Listing all vessels..."
curl -s -X GET $BASE_URL/vessels \
  -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, name, type, location}'

echo -e "\n‚úÖ All tests completed!"
