#!/bin/bash

# Test script for user action functionality
# This script tests the new /auth/action/{action} endpoint with ActivityType 'user'

set -e

# Configuration
API_BASE="http://localhost:8080/api/v1"
YUBIKEY_OTP="cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj"  # Replace with actual OTP

echo "🧪 Testing User Action Functionality"
echo "=================================="

# Test 1: Create a user action with ActivityType 'user'
echo "📝 Creating a user action..."
curl -X POST "$API_BASE/actions" \
  -H "Authorization: yubikey:$YUBIKEY_OTP" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "work-start",
    "activity_type": "user",
    "details": {
      "default_status": "working",
      "description": "Start of work shift"
    },
    "active": true
  }' | jq '.'

echo -e "\n"

# Test 2: Execute the user action with location and status
echo "🚀 Executing user action with location and status..."
curl -X POST "$API_BASE/auth/action/work-start" \
  -H "Authorization: yubikey:$YUBIKEY_OTP" \
  -H "Content-Type: application/json" \
  -d '{
    "location": "Main Office",
    "status": "working",
    "details": {
      "shift_type": "day",
      "notes": "Starting morning shift"
    }
  }' | jq '.'

echo -e "\n"

# Test 3: Execute the user action with default status from action details
echo "🔄 Executing user action with default status..."
curl -X POST "$API_BASE/auth/action/work-start" \
  -H "Authorization: yubikey:$YUBIKEY_OTP" \
  -H "Content-Type: application/json" \
  -d '{
    "location": "Home Office",
    "details": {
      "shift_type": "remote",
      "notes": "Working from home"
    }
  }' | jq '.'

echo -e "\n"

# Test 4: Create and execute a break action
echo "☕ Creating and executing break action..."
curl -X POST "$API_BASE/actions" \
  -H "Authorization: yubikey:$YUBIKEY_OTP" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "break-start",
    "activity_type": "user",
    "details": {
      "default_status": "break",
      "description": "Start of break period"
    },
    "active": true
  }' | jq '.'

echo -e "\n"

curl -X POST "$API_BASE/auth/action/break-start" \
  -H "Authorization: yubikey:$YUBIKEY_OTP" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "break",
    "details": {
      "break_type": "lunch",
      "duration_minutes": 30
    }
  }' | jq '.'

echo -e "\n"

# Test 5: Check user activity history
echo "📊 Checking user activity history..."
curl -X GET "$API_BASE/user-activity" \
  -H "Authorization: yubikey:$YUBIKEY_OTP" \
  -H "Content-Type: application/json" | jq '.'

echo -e "\n✅ User action functionality test completed!" 