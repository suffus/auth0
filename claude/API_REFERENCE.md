# YubiApp Device Registration API Reference

## Overview

The Device Registration API allows authorized users to manage device assignments between users with full audit trail capabilities. All endpoints require device-based authentication using YubiKey OTP.

## Authentication

All endpoints use device-based authentication with YubiKey OTP in the Authorization header:

```
Authorization: yubikey:<otp_code>
```

## Endpoints

### 1. Register Device

**POST** `/api/v1/devices/register`

Register a device to a target user.

**Required Permissions:** `yubiapp:register-other`

**Request Body:**
```json
{
  "target_user_id": "john.doe@example.com",
  "device_identifier": "cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj",
  "device_type": "yubikey",
  "notes": "New YubiKey for John"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Device registered successfully",
  "registration": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "device_id": "550e8400-e29b-41d4-a716-446655440001",
    "registrar": {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "email": "admin@example.com"
    },
    "target_user_id": "550e8400-e29b-41d4-a716-446655440003",
    "action_type": "register",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. Deregister Device

**POST** `/api/v1/devices/{device_id}/deregister`

Deregister a device from its current user.

**Required Permissions:** `yubiapp:deregister-other`

**Path Parameters:**
- `device_id` (UUID): ID of the device to deregister

**Request Body:**
```json
{
  "reason": "user_left",
  "notes": "User left the organization"
}
```

**Valid Reasons:**
- `user_left` - User left the organization
- `device_lost` - Device was lost or stolen
- `device_transfer` - Device is being transferred
- `administrative` - Administrative action

**Response (200):**
```json
{
  "success": true,
  "message": "Device deregistered successfully",
  "deregistration": {
    "id": "550e8400-e29b-41d4-a716-446655440004",
    "device_id": "550e8400-e29b-41d4-a716-446655440001",
    "registrar": {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "email": "admin@example.com"
    },
    "action_type": "deregister",
    "reason": "user_left",
    "created_at": "2024-01-15T11:00:00Z"
  }
}
```

### 3. Transfer Device

**POST** `/api/v1/devices/{device_id}/transfer`

Transfer a device from one user to another.

**Required Permissions:** Both `yubiapp:register-other` AND `yubiapp:deregister-other`

**Path Parameters:**
- `device_id` (UUID): ID of the device to transfer

**Request Body:**
```json
{
  "target_user_id": "jane.doe@example.com",
  "notes": "Transferring device to Jane"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Device transferred successfully",
  "transfer": {
    "id": "550e8400-e29b-41d4-a716-446655440005",
    "device_id": "550e8400-e29b-41d4-a716-446655440001",
    "registrar": {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "email": "admin@example.com"
    },
    "target_user_id": "550e8400-e29b-41d4-a716-446655440006",
    "action_type": "register",
    "created_at": "2024-01-15T11:30:00Z"
  }
}
```

### 4. Get Device History

**GET** `/api/v1/devices/{device_id}/history`

Get the complete registration history for a device.

**Required Permissions:** None (any authenticated user can view history)

**Path Parameters:**
- `device_id` (UUID): ID of the device to get history for

**Response (200):**
```json
{
  "device_id": "550e8400-e29b-41d4-a716-446655440001",
  "history": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440005",
      "action_type": "register",
      "registrar": {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "email": "admin@example.com"
      },
      "target_user": {
        "id": "550e8400-e29b-41d4-a716-446655440006",
        "email": "jane.doe@example.com"
      },
      "reason": null,
      "notes": "Transferring device to Jane",
      "ip_address": "192.168.1.100",
      "created_at": "2024-01-15T11:30:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440004",
      "action_type": "deregister",
      "registrar": {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "email": "admin@example.com"
      },
      "target_user": {
        "id": null,
        "email": null
      },
      "reason": "user_left",
      "notes": "User left the organization",
      "ip_address": "192.168.1.100",
      "created_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

## Error Responses

### Common Error Codes

**401 Unauthorized**
```json
{
  "error": "Authentication failed: invalid OTP"
}
```

**403 Forbidden**
```json
{
  "error": "Permission denied: yubiapp:register-other required"
}
```

**404 Not Found**
```json
{
  "error": "Target user not found"
}
```

**409 Conflict**
```json
{
  "error": "Device is already registered to another user"
}
```

**400 Bad Request**
```json
{
  "error": "Invalid reason. Must be one of: user_left, device_lost, device_transfer, administrative"
}
```

## Usage Examples

### cURL Examples

**Register a device:**
```bash
curl -X POST http://localhost:8080/api/v1/devices/register \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "john.doe@example.com",
    "device_identifier": "cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj",
    "device_type": "yubikey",
    "notes": "New YubiKey for John"
  }'
```

**Deregister a device:**
```bash
curl -X POST http://localhost:8080/api/v1/devices/550e8400-e29b-41d4-a716-446655440001/deregister \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "user_left",
    "notes": "User left the organization"
  }'
```

**Transfer a device:**
```bash
curl -X POST http://localhost:8080/api/v1/devices/550e8400-e29b-41d4-a716-446655440001/transfer \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "jane.doe@example.com",
    "notes": "Transferring device to Jane"
  }'
```

**Get device history:**
```bash
curl -X GET http://localhost:8080/api/v1/devices/550e8400-e29b-41d4-a716-446655440001/history \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj"
```

## CLI Examples

**Register a device:**
```bash
./yubiapp device register \
  --registrar-device-code="cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  --target-user="john.doe@example.com" \
  --device-identifier="cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  --device-type="yubikey" \
  --notes="New YubiKey for John"
```

**Deregister a device:**
```bash
./yubiapp device deregister \
  --registrar-device-code="cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  --device-id="550e8400-e29b-41d4-a716-446655440001" \
  --reason="user_left" \
  --notes="User left the organization"
```

**Transfer a device:**
```bash
./yubiapp device transfer \
  --registrar-device-code="cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  --device-id="550e8400-e29b-41d4-a716-446655440001" \
  --target-user="jane.doe@example.com" \
  --notes="Transferring device to Jane"
```

**View device history:**
```bash
./yubiapp device history \
  --device-id="550e8400-e29b-41d4-a716-446655440001"
```

## API Design Notes

### HTTP Method Choices
- **POST for Deregister**: Uses POST instead of DELETE to comply with OpenAPI specification, as DELETE operations should not have request bodies
- **Consistent POST Operations**: All modification operations (register, deregister, transfer) use POST for consistency
- **RESTful Considerations**: While not strictly RESTful, this design prioritizes OpenAPI compliance and practical usability

### Request Body Requirements
- **Deregister**: Requires request body for reason and notes (hence POST method)
- **Register**: Requires request body for target user and device details
- **Transfer**: Requires request body for target user and notes
- **History**: No request body needed (GET method)

## Notes

- All timestamps are in ISO 8601 format (UTC)
- Device identifiers for YubiKeys are the first 12 characters of the OTP
- User identification accepts both UUID and email addresses
- All operations are logged with complete audit information
- Transfer operations require both register and deregister permissions
- Device history is ordered by creation time (newest first)
- API design prioritizes OpenAPI compliance and practical usability over strict RESTful principles 