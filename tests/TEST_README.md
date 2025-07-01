# YubiApp Locations API Test Script

This script tests the YubiApp Locations API endpoints using YubiKey authentication.

## Prerequisites

1. **Python 3.6+** installed
2. **YubiApp API server** running on `http://localhost:8080`
3. **YubiKey** configured and registered with a user in the system
4. **Python requests library** installed

## Installation

```bash
# Install required Python packages
pip install -r requirements.txt

# Or install requests directly
pip install requests
```

## Usage

### Basic Usage
```bash
python3 test_locations_api.py
```

### Custom API URL
```bash
python3 test_locations_api.py http://your-api-server:8080/api/v1
```

## What the Script Tests

The script performs a complete test of the Locations API:

1. **Initial Authentication** - Authenticates with YubiKey to identify the user
2. **Session Creation** - Creates a session for read operations
3. **List Locations** - Tests GET `/api/v1/locations` (read operation)
4. **Create Location** - Tests POST `/api/v1/locations` (write operation)
5. **Verify Create** - Lists locations to confirm creation
6. **Update Location** - Tests PUT `/api/v1/locations/{id}` (write operation)
7. **Verify Update** - Lists locations to confirm updates
8. **Delete Location** - Tests DELETE `/api/v1/locations/{id}` (write operation)
9. **Verify Delete** - Lists locations to confirm deletion

## Authentication Flow

- **Read Operations** (GET): Uses session token from initial authentication
- **Write Operations** (POST, PUT, DELETE): Requires fresh YubiKey OTP for each operation

## Interactive Features

The script is fully interactive and will prompt you for:

- **YubiKey OTPs** - Insert your YubiKey and tap the button when prompted
- **Location Details** - Name, description, address, type, and active status
- **Confirmation** - For destructive operations like deletion

## Example Output

```
🚀 YubiApp Locations API Test Suite
==================================================

1️⃣  Initial Authentication

🔑 Please provide YubiKey OTP for initial authentication:
   (Insert your YubiKey and tap the button)
OTP: 
🔐 Authenticating with YubiKey...
✅ Authentication successful
   User: admin@example.com (Admin User)
   Device: yubikey (ccccccefghij)

2️⃣  Creating Session
🔐 Creating session...
✅ Session created successfully

3️⃣  Testing Read Operations

📋 Listing locations...
✅ Found 0 location(s)

4️⃣  Testing Create Operation

🔑 Please provide YubiKey OTP for creating a location:
   (Insert your YubiKey and tap the button)
OTP: 
🏢 Creating new location...
Location name: Test Office
Description (optional): A test office location
Address (optional): 123 Test Street
Location type options:
  1. office
  2. home
  3. event
  4. other
Choose type (1-4, default=1): 1
Active (y/n, default=y): y
🔐 Creating location with YubiKey authentication...
✅ Location created successfully
   ID: 12345678-1234-1234-1234-123456789abc
   Name: Test Office
   Type: office
   Active: true

...
```

## Troubleshooting

### Common Issues

1. **Connection Error**: Make sure the API server is running on the correct URL
2. **Authentication Failed**: Ensure your YubiKey is properly configured and registered
3. **Permission Denied**: Verify your user has the required permissions (`yubiapp:read`, `yubiapp:write`)

### Error Messages

- `❌ Authentication failed`: YubiKey OTP was invalid or user not found
- `❌ Session creation failed`: Could not create session token
- `❌ Failed to list locations`: Read permission denied or API error
- `❌ Failed to create/update/delete location`: Write permission denied or API error

## API Endpoints Tested

- `POST /api/v1/auth/device` - Device authentication
- `POST /api/v1/auth/session` - Session creation
- `GET /api/v1/locations` - List locations
- `POST /api/v1/locations` - Create location
- `PUT /api/v1/locations/{id}` - Update location
- `DELETE /api/v1/locations/{id}` - Delete location

## Security Notes

- The script uses `getpass` to securely prompt for YubiKey OTPs (input is hidden)
- Session tokens are used for read operations to reduce YubiKey usage
- Write operations require fresh YubiKey authentication for security
- All API calls include proper authentication headers 