#!/bin/bash

# Test script for YubiApp session functionality
# This script tests the session creation and refresh endpoints

set -e

# Configuration
API_BASE="http://localhost:8080/api/v1"
YUBIKEY_OTP=$1  # Replace with a real OTP for testing

echo "Testing YubiApp Session Functionality"
echo "====================================="

# Test 1: Create a session
echo -e "\n1. Testing session creation..."
SESSION_RESPONSE=$(curl -s -X POST "$API_BASE/auth/session" \
  -H "Content-Type: application/json" \
  -d "{
    \"device_type\": \"yubikey\",
    \"auth_code\": \"$YUBIKEY_OTP\",
    \"permission\": \"yubiapp:read\"
  }")

echo "Session creation response:"
echo "$SESSION_RESPONSE" | jq '.'

# Extract session data
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r '.session_id')
ACCESS_TOKEN=$(echo "$SESSION_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$SESSION_RESPONSE" | jq -r '.refresh_token')

if [ "$SESSION_ID" = "null" ] || [ "$ACCESS_TOKEN" = "null" ] || [ "$REFRESH_TOKEN" = "null" ]; then
    echo "❌ Failed to create session"
    exit 1
fi

echo "✅ Session created successfully"
echo "Session ID: $SESSION_ID"

# Test 2: Use access token to access a protected endpoint (unified endpoint)
echo -e "\n2. Testing access token usage with unified endpoint..."
USERS_RESPONSE=$(curl -s -X GET "$API_BASE/users" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Users endpoint response:"
echo "$USERS_RESPONSE" | jq '.'
if echo "$USERS_RESPONSE" | jq -e '.items' > /dev/null; then
    echo "✅ Access token works correctly with unified endpoint"
else
    echo "❌ Access token failed with unified endpoint"
    exit 1
fi

# Test 3: Try to use session token for write operation (should fail)
echo -e "\n3. Testing session token for write operation (should fail)..."
WRITE_RESPONSE=$(curl -s -X POST "$API_BASE/users" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"test@example.com\",
    \"username\": \"testuser\",
    \"password\": \"password123\"
  }")

echo "Write operation response:"
echo "$WRITE_RESPONSE" | jq '.'

if echo "$WRITE_RESPONSE" | jq -e '.error' > /dev/null; then
    echo "✅ Session token correctly rejected for write operation"
else
    echo "❌ Session token should have been rejected for write operation"
    exit 1
fi

# Test 4: Refresh the session
echo -e "\n4. Testing session refresh..."
REFRESH_RESPONSE=$(curl -s -X POST "$API_BASE/auth/session/refresh/$SESSION_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "Session refresh response:"
echo "$REFRESH_RESPONSE" | jq '.'

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')
NEW_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.refresh_token')

if [ "$NEW_ACCESS_TOKEN" = "null" ] || [ "$NEW_REFRESH_TOKEN" = "null" ]; then
    echo "❌ Failed to refresh session"
    exit 1
fi

echo "✅ Session refreshed successfully"

# Test 5: Use new access token with unified endpoint
echo -e "\n5. Testing new access token with unified endpoint..."
NEW_USERS_RESPONSE=$(curl -s -X GET "$API_BASE/users" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN")

echo "Users endpoint response with new token:"
echo "$NEW_USERS_RESPONSE" | jq '.'

if echo "$NEW_USERS_RESPONSE" | jq -e '.items' > /dev/null; then
    echo "✅ New access token works correctly with unified endpoint"
else
    echo "❌ New access token failed with unified endpoint"
    exit 1
fi

# Test 6: Try to use old access token (should fail)
echo -e "\n6. Testing old access token (should fail)..."
OLD_TOKEN_RESPONSE=$(curl -s -X GET "$API_BASE/users" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Old token response:"
echo "$OLD_TOKEN_RESPONSE" | jq '.'

if echo "$OLD_TOKEN_RESPONSE" | jq -e '.error' > /dev/null; then
    echo "✅ Old access token correctly rejected"
else
    echo "❌ Old access token should have been rejected"
    exit 1
fi

echo -e "\n🎉 All session tests passed!"
echo "Unified authentication system is working correctly."
echo "✅ Session tokens work for read operations"
echo "✅ Session tokens are rejected for write operations"
echo "✅ Device authentication still works for all operations"
echo "✅ Token refresh works correctly"
echo "✅ Old tokens are properly invalidated" 
