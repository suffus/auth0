#!/usr/bin/env python3
"""
Test script for user actions functionality.
Tests the /auth/action/{action} endpoint with ActivityType 'user'.
"""

import requests
import json
import sys
import time
from typing import Dict, Any, Optional

# Configuration
API_BASE = "http://localhost:8080/api/v1"

class UserActionTester:
    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json'
        })
    
    def get_yubikey_otp(self, prompt: str = "Enter YubiKey OTP: ") -> str:
        """Get YubiKey OTP from user input."""
        return input(prompt).strip()
    
    def authenticate(self, device_type: str, auth_code: str) -> Dict[str, Any]:
        """Authenticate using device-based authentication."""
        url = f"{API_BASE}/auth/device"
        data = {
            "device_type": device_type,
            "auth_code": auth_code
        }
        
        response = self.session.post(url, json=data)
        if response.status_code != 200:
            print(f"❌ Authentication failed: {response.status_code}")
            print(response.text)
            sys.exit(1)
        
        result = response.json()
        print(f"✅ Authenticated as: {result['user']['email']}")
        return result
    
    def set_auth_header(self, device_type: str, auth_code: str):
        """Set Authorization header for subsequent requests."""
        self.session.headers['Authorization'] = f"{device_type}:{auth_code}"
    
    def set_auth_header_with_fresh_otp(self, device_type: str, prompt: str = None):
        """Get a fresh OTP and set it in the Authorization header."""
        if prompt is None:
            prompt = f"Enter YubiKey OTP for {device_type} operation: "
        
        auth_code = self.get_yubikey_otp(prompt)
        self.set_auth_header(device_type, auth_code)
    
    def create_location(self, name: str, description: str = "", address: str = "", location_type: str = "other") -> Optional[Dict[str, Any]]:
        """Create a location if it doesn't exist."""
        # First check if location exists
        url = f"{API_BASE}/locations"
        response = self.session.get(url)
        
        if response.status_code == 200:
            locations = response.json().get('data', [])
            for location in locations:
                if location['name'] == name:
                    print(f"✅ Location '{name}' already exists")
                    return location
        
        # Create location
        data = {
            "name": name,
            "description": description,
            "address": address,
            "type": location_type,
            "active": True
        }
        
        response = self.session.post(url, json=data)
        if response.status_code == 201:
            location = response.json().get('data', {})
            print(f"✅ Created location: {name}")
            return location
        else:
            print(f"❌ Failed to create location '{name}': {response.status_code}")
            print(response.text)
            return None
    
    def create_user_status(self, name: str, description: str = "", status_type: str = "other") -> Optional[Dict[str, Any]]:
        """Create a user status if it doesn't exist."""
        # First check if status exists
        url = f"{API_BASE}/user-statuses"
        response = self.session.get(url)
        
        if response.status_code == 200:
            statuses = response.json().get('data', [])
            for status in statuses:
                if status['name'] == name:
                    print(f"✅ User status '{name}' already exists")
                    return status
        
        # Create status
        data = {
            "name": name,
            "description": description,
            "type": status_type,
            "active": True
        }
        
        response = self.session.post(url, json=data)
        if response.status_code == 201:
            status = response.json().get('data', {})
            print(f"✅ Created user status: {name}")
            return status
        else:
            print(f"❌ Failed to create user status '{name}': {response.status_code}")
            print(response.text)
            return None
    
    def create_action(self, name: str, activity_type: str, details: Dict[str, Any] = None, required_permissions: list = None) -> Optional[Dict[str, Any]]:
        """Create an action if it doesn't exist."""
        # First check if action exists
        url = f"{API_BASE}/actions"
        response = self.session.get(url)
        
        if response.status_code == 200:
            actions = response.json().get('data', [])
            for action in actions:
                if action['name'] == name:
                    print(f"✅ Action '{name}' already exists")
                    return action
        
        # Create action
        data = {
            "name": name,
            "activity_type": activity_type,
            "details": details or {},
            "required_permissions": required_permissions or [],
            "active": True
        }
        
        response = self.session.post(url, json=data)
        if response.status_code == 201:
            action = response.json().get('data', {})
            print(f"✅ Created action: {name}")
            return action
        else:
            print(f"❌ Failed to create action '{name}': {response.status_code}")
            print(response.text)
            return None
    
    def execute_action(self, action_name: str, location: str = None, status: str = None, details: Dict[str, Any] = None) -> Optional[Dict[str, Any]]:
        """Execute an action and return the response."""
        url = f"{API_BASE}/auth/action/{action_name}"
        
        data = {}
        if location:
            data["location"] = location
        if status:
            data["status"] = status
        if details:
            data["details"] = details
        
        response = self.session.post(url, json=data)
        if response.status_code == 200:
            result = response.json()
            print(f"✅ Executed action '{action_name}' successfully")
            return result
        else:
            print(f"❌ Failed to execute action '{action_name}': {response.status_code}")
            print(response.text)
            return None
    
    def get_user_activity(self, limit: int = 10) -> Optional[Dict[str, Any]]:
        """Get recent user activity history."""
        url = f"{API_BASE}/user-activity?limit={limit}"
        
        response = self.session.get(url)
        if response.status_code == 200:
            return response.json()
        else:
            print(f"❌ Failed to get user activity: {response.status_code}")
            print(response.text)
            return None
    
    def print_activity_summary(self, activities: list):
        """Print a summary of user activities."""
        print("\n📊 User Activity Summary:")
        print("=" * 80)
        
        for i, activity in enumerate(activities, 1):
            print(f"{i}. Action: {activity.get('action', {}).get('name', 'N/A')}")
            print(f"   Time: {activity.get('from_datetime', 'N/A')}")
            print(f"   Status: {activity.get('status', {}).get('name', 'N/A')}")
            print(f"   Location: {activity.get('location', {}).get('name', 'N/A')}")
            print(f"   Details: {json.dumps(activity.get('details', {}), indent=2)}")
            print("-" * 40)

def main():
    print("🏨 Hotel User Actions Test")
    print("=" * 50)
    
    tester = UserActionTester()
    
    # Step 1: Initial verification (optional)
    print("\n🔐 Step 1: Initial Verification")
    print("-" * 30)
    print("Starting hotel user actions test...")
    
    # Step 2: Create location
    print("\n📍 Step 2: Create Location")
    print("-" * 30)
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for location creation: ")
    hotel_location = tester.create_location(
        name="hotel",
        description="Test hotel location",
        address="123 Test Street",
        location_type="other"
    )
    
    # Step 3: Create user statuses
    print("\n👤 Step 3: Create User Statuses")
    print("-" * 30)
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for user status creation: ")
    in_hotel_status = tester.create_user_status(
        name="in-hotel",
        description="User is currently in the hotel",
        status_type="other"
    )
    
    out_hotel_status = tester.create_user_status(
        name="out-of-hotel",
        description="User is currently outside the hotel",
        status_type="other"
    )
    
    # Step 4: Create actions
    print("\n⚡ Step 4: Create Actions")
    print("-" * 30)
    
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for action creation: ")
    
    # Create enter-hotel action with defaults
    enter_hotel_action = tester.create_action(
        name="enter-hotel",
        activity_type="user",
        details={
            "default_location": "hotel",
            "default_status": "in-hotel",
            "description": "User enters the hotel"
        }
    )
    
    # Create leave-hotel action with default status
    leave_hotel_action = tester.create_action(
        name="leave-hotel",
        activity_type="user",
        details={
            "default_status": "out-of-hotel",
            "description": "User leaves the hotel"
        }
    )
    
    # Step 5: Execute actions and test
    print("\n🚀 Step 5: Execute Actions")
    print("-" * 30)
    
    # Test 1: Enter hotel with explicit location and status
    print("\n📝 Test 1: Enter hotel with explicit location and status")
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for action execution (Test 1): ")
    result1 = tester.execute_action(
        action_name="enter-hotel",
        location="hotel",
        status="in-hotel",
        details={
            "entry_method": "main_door",
            "time_of_day": "morning"
        }
    )
    
    if result1 and 'user_activity' in result1:
        print(f"   Activity ID: {result1['user_activity']['id']}")
        print(f"   From DateTime: {result1['user_activity']['from_datetime']}")
    
    # Wait a moment
    time.sleep(1)
    
    # Test 2: Enter hotel with defaults from action
    print("\n📝 Test 2: Enter hotel with defaults from action")
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for action execution (Test 2): ")
    result2 = tester.execute_action(
        action_name="enter-hotel",
        details={
            "entry_method": "side_door",
            "time_of_day": "afternoon"
        }
    )
    
    if result2 and 'user_activity' in result2:
        print(f"   Activity ID: {result2['user_activity']['id']}")
        print(f"   From DateTime: {result2['user_activity']['from_datetime']}")
    
    # Wait a moment
    time.sleep(1)
    
    # Test 3: Leave hotel
    print("\n📝 Test 3: Leave hotel")
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for action execution (Test 3): ")
    result3 = tester.execute_action(
        action_name="leave-hotel",
        details={
            "exit_method": "main_door",
            "duration_minutes": 120
        }
    )
    
    if result3 and 'user_activity' in result3:
        print(f"   Activity ID: {result3['user_activity']['id']}")
        print(f"   From DateTime: {result3['user_activity']['from_datetime']}")
    
    # Step 6: Verify activity history
    print("\n📊 Step 6: Verify Activity History")
    print("-" * 30)
    
    tester.set_auth_header_with_fresh_otp("yubikey", "Enter YubiKey OTP for activity history retrieval: ")
    activity_data = tester.get_user_activity(limit=10)
    if activity_data:
        activities = activity_data.get('data', [])
        tester.print_activity_summary(activities)
        
        # Verify we have the expected activities
        action_names = [activity.get('action', {}).get('name') for activity in activities]
        print(f"\n✅ Found {len(activities)} activities")
        print(f"   Actions executed: {action_names}")
        
        # Check for specific patterns
        enter_count = action_names.count('enter-hotel')
        leave_count = action_names.count('leave-hotel')
        
        print(f"\n📈 Summary:")
        print(f"   Enter hotel actions: {enter_count}")
        print(f"   Leave hotel actions: {leave_count}")
        
        if enter_count >= 2 and leave_count >= 1:
            print("✅ Test completed successfully!")
        else:
            print("❌ Expected activities not found")
    else:
        print("❌ Failed to retrieve activity history")

if __name__ == "__main__":
    main() 