import requests
import sys
import time
import random
import string

BASE_URL = "http://localhost:8080/api/v1"


def print_header(msg):
    print("\n" + "." * 60)
    print(f"{msg}")
    print("." * 60)

def fail(msg):
    print(f"❌ {msg}")
    sys.exit(1)

def get_yubikey_otp(action_desc):
    print(f"🔐 Please touch your YubiKey to generate an OTP for {action_desc}...")
    otp = input("Enter YubiKey OTP: ").strip()
    if not otp or len(otp) < 32:
        fail("Invalid OTP entered.")
    return otp

def generate_random_suffix():
    """Generate a random 6-character suffix to avoid duplicate names"""
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=6))

def create_session(otp):
    """Create a session using YubiKey OTP"""
    print("🔄 Creating session...")
    resp = requests.post(f"{BASE_URL}/auth/session", json={
        "device_type": "yubikey",
        "auth_code": otp
    })
    if resp.status_code != 200:
        fail(f"Failed to create session: {resp.text}")
    result = resp.json()
    session_token = result.get('access_token')
    session_id = result.get('session_id')
    refresh_token = result.get('refresh_token')
    if not session_token:
        fail("No access token in response")
    print("✅ Session created successfully")
    return session_token, session_id, refresh_token

def refresh_session(session_id, refresh_token):
    """Refresh a session using refresh token"""
    url = f"{BASE_URL}/auth/session/refresh/{session_id}"
    headers = {"Content-Type": "application/json"}
    data = {"refresh_token": refresh_token}
    
    print("🔄 Refreshing session...")
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
    
    print("✅ Session refreshed successfully")
    return session_token, new_refresh_token

def main():
    # Create session for read operations
    print_header("🔐 Creating Session")
    session_otp = get_yubikey_otp("creating session")
    session_token, session_id, refresh_token = create_session(session_otp)
    
    # Headers for read operations (using session token)
    read_headers = {"Authorization": f"Bearer {session_token}"}
    
    # 1. Create a user status
    random_suffix = generate_random_suffix()
    status_name = f"Signed In {random_suffix}"
    print_header(f"1️⃣  Creating user status: '{status_name}'")
    otp = get_yubikey_otp("creating user status")
    write_headers = {"Authorization": f"yubikey:{otp}"}
    create_resp = requests.post(f"{BASE_URL}/user-statuses", json={
        "name": status_name,
        "description": "User is currently signed in and working",
        "type": "working",
        "active": True
    }, headers=write_headers)
    if create_resp.status_code != 201:
        fail(f"Failed to create user status: {create_resp.text}")
    status = create_resp.json()
    status_id = status["id"]
    print(f"✅ Created user status: {status['name']} (ID: {status_id})")

    # 2. List all user statuses (read: using session token)
    print_header("2️⃣  Listing all user statuses")
    list_resp = requests.get(f"{BASE_URL}/user-statuses", headers=read_headers)
    if list_resp.status_code != 200:
        fail(f"Failed to list user statuses: {list_resp.text}")
    items = list_resp.json().get("items", [])
    print(f"✅ Found {len(items)} user status(es)")
    for s in items:
        print(f"- {s['name']} (active: {s['active']})")

    # 3. List only active user statuses (read: using session token)
    print_header("3️⃣  Listing only active user statuses")
    active_resp = requests.get(f"{BASE_URL}/user-statuses?active=true", headers=read_headers)
    if active_resp.status_code != 200:
        fail(f"Failed to list active user statuses: {active_resp.text}")
    active_items = active_resp.json().get("items", [])
    print(f"✅ Found {len(active_items)} active user status(es)")
    for s in active_items:
        print(f"- {s['name']} (active: {s['active']})")

    # 4. Update the user status
    updated_status_name = f"On Break {random_suffix}"
    print_header(f"4️⃣  Updating user status to '{updated_status_name}'")
    otp = get_yubikey_otp("updating user status")
    write_headers = {"Authorization": f"yubikey:{otp}"}
    update_resp = requests.put(f"{BASE_URL}/user-statuses/{status_id}", json={
        "name": updated_status_name,
        "description": "User is currently on a break",
        "type": "break",
        "active": True
    }, headers=write_headers)
    if update_resp.status_code != 200:
        fail(f"Failed to update user status: {update_resp.text}")
    updated = update_resp.json()
    print(f"✅ Updated user status: {updated['name']} (type: {updated['type']})")

    # 5. Delete (soft delete) the user status
    print_header("5️⃣  Deleting user status (soft delete)")
    otp = get_yubikey_otp("deleting user status")
    write_headers = {"Authorization": f"yubikey:{otp}"}
    del_resp = requests.delete(f"{BASE_URL}/user-statuses/{status_id}", headers=write_headers)
    if del_resp.status_code != 204:
        fail(f"Failed to delete user status: {del_resp.text}")
    print(f"✅ User status deleted (marked inactive)")

    # 6. List only active user statuses (should not include deleted one)
    print_header("6️⃣  Listing only active user statuses after delete")
    active_resp2 = requests.get(f"{BASE_URL}/user-statuses?active=true", headers=read_headers)
    if active_resp2.status_code != 200:
        fail(f"Failed to list active user statuses: {active_resp2.text}")
    active_items2 = active_resp2.json().get("items", [])
    print(f"✅ Found {len(active_items2)} active user status(es)")
    for s in active_items2:
        print(f"- {s['name']} (active: {s['active']})")
    if any(s["id"] == status_id for s in active_items2):
        fail("Deleted user status is still listed as active!")

    # 7. Try to get the deleted user status by ID (should still exist, but inactive)
    print_header("7️⃣  Get deleted user status by ID (should be inactive)")
    get_resp = requests.get(f"{BASE_URL}/user-statuses/{status_id}", headers=read_headers)
    if get_resp.status_code != 200:
        fail(f"Failed to get user status by ID: {get_resp.text}")
    got = get_resp.json()
    print(f"✅ Got user status: {got['name']} (active: {got['active']})")
    if got["active"]:
        fail("Deleted user status is still marked as active!")
    print("\n🎉 All user status API tests passed!")

if __name__ == "__main__":
    main() 