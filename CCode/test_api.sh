#!/bin/bash

# YubiApp API Test Script
# This script tests the authentication endpoint with curl

API_URL="http://localhost:8080/api/v1/auth/device"

echo "Testing YubiApp API Authentication Endpoint"
echo "=========================================="
echo "API URL: $API_URL"
echo ""

# Test 1: Basic authentication without permission
echo "Test 1: Basic authentication (no permission specified)"
echo "-----------------------------------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "yubikey",
    "auth_code": "vvbbccddeerrffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz"
  }' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo ""

# Test 2: Authentication with permission
echo "Test 2: Authentication with permission"
echo "--------------------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "yubikey",
    "auth_code": "vvbbccddeerrffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz",
    "permission": "yubiapp:authenticate"
  }' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo ""

# Test 3: Invalid OTP format
echo "Test 3: Invalid OTP format (too short)"
echo "---------------------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "yubikey",
    "auth_code": "123456"
  }' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo ""

# Test 4: Invalid device type
echo "Test 4: Invalid device type"
echo "---------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "invalid",
    "auth_code": "vvbbccddeerrffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz"
  }' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo ""

# Test 5: Missing required fields
echo "Test 5: Missing required fields"
echo "-------------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "yubikey"
  }' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s
echo ""

echo "Test completed!"
echo ""
echo "Expected responses:"
echo "- 200: Successful authentication with user data"
echo "- 400: Bad request (invalid format, missing fields)"
echo "- 401: Authentication failed (invalid OTP, device not found)"
echo "- 403: Permission denied"
echo "- 500: Server error" 