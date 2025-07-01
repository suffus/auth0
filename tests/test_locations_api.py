#!/usr/bin/env python3
"""
Test script for YubiApp Locations API endpoints
Uses YubiKey authentication for all operations
"""

import requests
import json
import getpass
import sys
from typing import Optional, Dict, Any

class YubiAppLocationTester:
    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.session_token = None
        self.session_id = None
        self.refresh_token = None
        self.user_info = None
        
    def get_yubikey_otp(self, operation: str) -> str:
        """Prompt user for YubiKey OTP"""
        print(f"\n🔑 Please provide YubiKey OTP for {operation}:")
        print("   (Insert your YubiKey and tap the button)")
        otp = getpass.getpass("OTP: ").strip()
        if not otp:
            print("❌ No OTP provided")
            sys.exit(1)
        return otp
    
    def authenticate_with_yubikey(self, otp: str) -> Dict[str, Any]:
        """Authenticate using YubiKey OTP"""
        url = f"{self.base_url}/auth/device"
        data = {
            "device_type": "yubikey",
            "auth_code": otp
        }
        
        print(f"🔐 Authenticating with YubiKey...")
        response = requests.post(url, json=data)
        
        if response.status_code != 200:
            print(f"❌ Authentication failed: {response.status_code}")
            print(f"   Response: {response.text}")
            sys.exit(1)
        
        result = response.json()
        print(f"✅ Authentication successful")
        print(f"   User: {result['user']['email']} ({result['user']['first_name']} {result['user']['last_name']})")
        print(f"   Device: {result['device']['type']} ({result['device']['identifier']})")
        
        return result
    
    def create_session(self, otp: str) -> str:
        """Create a session using YubiKey OTP"""
        url = f"{self.base_url}/auth/session"
        data = {
            "device_type": "yubikey",
            "auth_code": otp
        }
        
        print(f"🔐 Creating session...")
        response = requests.post(url, json=data)
        
        if response.status_code != 200:
            print(f"❌ Session creation failed: {response.status_code}")
            print(f"   Response: {response.text}")
            sys.exit(1)
        
        result = response.json()
        session_token = result.get('access_token')
        session_id = result.get('session_id')
        refresh_token = result.get('refresh_token')
        if not session_token:
            print("❌ No access token in response")
            sys.exit(1)
        
        print(f"✅ Session created successfully")
        return session_token, session_id, refresh_token
    
    def refresh_session(self, session_id: str) -> str:
        """Refresh a session using refresh token"""
        url = f"{self.base_url}/auth/session/refresh/{session_id}"
        
        # Send refresh token in JSON body
        data = {"refresh_token": self.refresh_token} if self.refresh_token else {}
        headers = {"Content-Type": "application/json"}
        
        print(f"🔄 Refreshing session...")
        response = requests.post(url, json=data, headers=headers)
        
        if response.status_code != 200:
            print(f"❌ Session refresh failed: {response.status_code}")
            print(f"   Response: {response.text}")
            return None
        
        result = response.json()
        session_token = result.get('access_token')
        new_refresh_token = result.get('refresh_token')
        if not session_token:
            print("❌ No access token in refresh response")
            return None
        
        # Update refresh token if provided
        if new_refresh_token:
            self.refresh_token = new_refresh_token
        
        print(f"✅ Session refreshed successfully")
        return session_token
    
    def get_auth_headers(self, use_session: bool = True, yubikey_otp: str = None) -> Dict[str, str]:
        """Get authentication headers"""
        if use_session and self.session_token:
            return {"Authorization": f"Bearer {self.session_token}"}
        elif yubikey_otp:
            return {"Authorization": f"yubikey:{yubikey_otp}"}
        else:
            return {}
    
    def list_locations(self) -> Dict[str, Any]:
        """List all locations"""
        url = f"{self.base_url}/locations"
        headers = self.get_auth_headers()
        
        print(f"\n📋 Listing locations...")
        response = requests.get(url, headers=headers)
        
        if response.status_code == 401 and "count mismatch" in response.text:
            print(f"⚠️  Session token expired (count mismatch), refreshing...")
            new_token = self.refresh_session(self.session_id)
            if new_token:
                self.session_token = new_token
                headers = self.get_auth_headers()
                response = requests.get(url, headers=headers)
                if response.status_code != 200:
                    print(f"❌ Failed to list locations after refresh: {response.status_code}")
                    print(f"   Response: {response.text}")
                    return {}
            else:
                print(f"❌ Failed to refresh session")
                return {}
        elif response.status_code != 200:
            print(f"❌ Failed to list locations: {response.status_code}")
            print(f"   Response: {response.text}")
            return {}
        
        result = response.json()
        # Handle different response structures
        locations = result.get('data', result.get('items', []))
        print(f"✅ Found {len(locations)} location(s)")
        
        for i, location in enumerate(locations, 1):
            print(f"   {i}. {location['name']} ({location['id']})")
            print(f"      Type: {location['type']}, Active: {location['active']}")
            print(f"      Address: {location['address']}")
            print(f"      Description: {location['description']}")
            print()
        
        return result
    
    def create_location(self, otp: str) -> Dict[str, Any]:
        """Create a new location"""
        url = f"{self.base_url}/locations"
        headers = self.get_auth_headers(use_session=False, yubikey_otp=otp)
        
        # Get location details from user
        print(f"\n🏢 Creating new location...")
        name = input("Location name: ").strip()
        if not name:
            print("❌ Location name is required")
            return {}
        
        description = input("Description (optional): ").strip()
        address = input("Address (optional): ").strip()
        
        print("Location type options:")
        print("  1. office")
        print("  2. home") 
        print("  3. event")
        print("  4. other")
        type_choice = input("Choose type (1-4, default=1): ").strip()
        
        type_map = {"1": "office", "2": "home", "3": "event", "4": "other"}
        location_type = type_map.get(type_choice, "office")
        
        active = input("Active (y/n, default=y): ").strip().lower() != "n"
        
        data = {
            "name": name,
            "description": description,
            "address": address,
            "type": location_type,
            "active": active
        }
        
        print(f"🔐 Creating location with YubiKey authentication...")
        response = requests.post(url, json=data, headers=headers)
        
        if response.status_code != 201:
            print(f"❌ Failed to create location: {response.status_code}")
            print(f"   Response: {response.text}")
            return {}
        
        result = response.json()
        print(f"✅ Location created successfully")
        
        # Handle different possible response structures
        if 'data' in result:
            location = result['data']
        elif 'item' in result:
            location = result['item']
        else:
            location = result
            
        if 'id' in location:
            print(f"   ID: {location['id']}")
        if 'name' in location:
            print(f"   Name: {location['name']}")
        if 'type' in location:
            print(f"   Type: {location['type']}")
        if 'active' in location:
            print(f"   Active: {location['active']}")
        
        return result
    
    def update_location(self, location_id: str, otp: str) -> Dict[str, Any]:
        """Update a location"""
        url = f"{self.base_url}/locations/{location_id}"
        headers = self.get_auth_headers(use_session=False, yubikey_otp=otp)
        
        print(f"\n✏️  Updating location {location_id}...")
        
        # Get update details from user
        name = input("New name (leave empty to keep current): ").strip()
        description = input("New description (leave empty to keep current): ").strip()
        address = input("New address (leave empty to keep current): ").strip()
        
        print("Location type options:")
        print("  1. office")
        print("  2. home") 
        print("  3. event")
        print("  4. other")
        type_choice = input("New type (1-4, leave empty to keep current): ").strip()
        
        type_map = {"1": "office", "2": "home", "3": "event", "4": "other"}
        location_type = type_map.get(type_choice) if type_choice else None
        
        active_choice = input("Active (y/n, leave empty to keep current): ").strip().lower()
        active = None if not active_choice else (active_choice == "y")
        
        # Build update data
        data = {}
        if name:
            data["name"] = name
        if description:
            data["description"] = description
        if address:
            data["address"] = address
        if location_type:
            data["type"] = location_type
        if active is not None:
            data["active"] = active
        
        if not data:
            print("❌ No updates provided")
            return {}
        
        print(f"🔐 Updating location with YubiKey authentication...")
        response = requests.put(url, json=data, headers=headers)
        
        if response.status_code != 200:
            print(f"❌ Failed to update location: {response.status_code}")
            print(f"   Response: {response.text}")
            return {}
        
        result = response.json()
        print(f"✅ Location updated successfully")
        
        # Handle different possible response structures
        if 'data' in result:
            location = result['data']
        elif 'item' in result:
            location = result['item']
        else:
            location = result
            
        if 'id' in location:
            print(f"   ID: {location['id']}")
        if 'name' in location:
            print(f"   Name: {location['name']}")
        if 'type' in location:
            print(f"   Type: {location['type']}")
        if 'active' in location:
            print(f"   Active: {location['active']}")
        if 'updated_at' in location:
            print(f"   Updated: {location['updated_at']}")
        
        return result
    
    def delete_location(self, location_id: str, otp: str) -> bool:
        """Delete a location"""
        url = f"{self.base_url}/locations/{location_id}"
        headers = self.get_auth_headers(use_session=False, yubikey_otp=otp)
        
        print(f"\n🗑️  Deleting location {location_id}...")
        confirm = input("Are you sure? (y/n): ").strip().lower()
        if confirm != "y":
            print("❌ Deletion cancelled")
            return False
        
        print(f"🔐 Deleting location with YubiKey authentication...")
        response = requests.delete(url, headers=headers)
        
        if response.status_code not in [200, 204]:
            print(f"❌ Failed to delete location: {response.status_code}")
            print(f"   Response: {response.text}")
            return False
        
        print(f"✅ Location deleted successfully (marked as inactive)")
        return True
    
    def run_tests(self):
        """Run the complete test suite"""
        print("🚀 YubiApp Locations API Test Suite")
        print("=" * 50)
        
        # Initial authentication to get session
        print("\n1️⃣  Initial Authentication")
        otp = self.get_yubikey_otp("initial authentication")
        auth_result = self.authenticate_with_yubikey(otp)
        self.user_info = auth_result['user']
        
        # Create session for read operations
        print("\n2️⃣  Creating Session")
        session_otp = self.get_yubikey_otp("creating session")
        self.session_token, self.session_id, self.refresh_token = self.create_session(session_otp)
        
        # List locations (read operation)
        print("\n3️⃣  Testing Read Operations")
        self.list_locations()
        
        # Create location (write operation)
        print("\n4️⃣  Testing Create Operation")
        otp = self.get_yubikey_otp("creating a location")
        create_result = self.create_location(otp)
        
        if not create_result:
            print("❌ Create operation failed, stopping tests")
            return
        
        # Extract location ID from response
        if 'data' in create_result and 'id' in create_result['data']:
            location_id = create_result['data']['id']
        elif 'item' in create_result and 'id' in create_result['item']:
            location_id = create_result['item']['id']
        elif 'id' in create_result:
            location_id = create_result['id']
        else:
            print("❌ Could not find location ID in response, stopping tests")
            print(f"   Response: {create_result}")
            return
        
        # List locations again to see the new one
        print("\n5️⃣  Verifying Create Operation")
        self.list_locations()
        
        # Update location (write operation)
        print("\n6️⃣  Testing Update Operation")
        otp = self.get_yubikey_otp("updating the location")
        update_result = self.update_location(location_id, otp)
        
        if not update_result:
            print("❌ Update operation failed, stopping tests")
            return
        
        # List locations again to see the changes
        print("\n7️⃣  Verifying Update Operation")
        self.list_locations()
        
        # Delete location (write operation)
        print("\n8️⃣  Testing Delete Operation")
        otp = self.get_yubikey_otp("deleting the location")
        delete_success = self.delete_location(location_id, otp)
        
        if not delete_success:
            print("❌ Delete operation failed, stopping tests")
            return
        
        # List locations again to see the deletion
        print("\n9️⃣  Verifying Delete Operation")
        self.list_locations()
        
        print("\n🎉 All tests completed successfully!")
        print("=" * 50)

def main():
    """Main function"""
    # Check for help flag
    if len(sys.argv) > 1 and sys.argv[1] in ['--help', '-h', 'help']:
        print("YubiApp Locations API Test Script")
        print("=" * 40)
        print("Usage:")
        print("  python3 test_locations_api.py [base_url]")
        print("")
        print("Arguments:")
        print("  base_url    API base URL (default: http://localhost:8080/api/v1)")
        print("")
        print("Examples:")
        print("  python3 test_locations_api.py")
        print("  python3 test_locations_api.py http://your-server:8080/api/v1")
        print("")
        print("Prerequisites:")
        print("  - YubiApp API server running")
        print("  - YubiKey configured and registered")
        print("  - Python requests library installed")
        return
    
    if len(sys.argv) > 1:
        base_url = sys.argv[1]
    else:
        base_url = "http://localhost:8080/api/v1"
    
    tester = YubiAppLocationTester(base_url)
    
    try:
        tester.run_tests()
    except KeyboardInterrupt:
        print("\n\n❌ Test interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n\n❌ Test failed with error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main() 