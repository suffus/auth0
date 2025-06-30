# YubiApp Session Management

This document describes the session management functionality implemented in YubiApp according to the specification in `YubiappSessions.md`.

## Overview

The session system allows users to authenticate once with their device (YubiKey, TOTP, etc.) and then use JWT-based access tokens for subsequent API calls, eliminating the need to authenticate with a device for every request.

## Architecture

### Components

1. **Session Service** (`internal/services/session_service.go`)
   - Manages session creation, storage, and token generation
   - Uses Redis for in-memory session storage
   - Handles JWT token generation and validation

2. **Unified Authentication Middleware** (`internal/server/middleware.go`)
   - `authMiddlewareRead`: Accepts both device and session authentication for read operations
   - `authMiddlewareWrite`: Accepts only device authentication for write operations
   - Provides seamless integration with existing endpoints

3. **Configuration** (`internal/config/config.go`)
   - Redis connection settings
   - Session and token expiry times

### Session Storage

Sessions are stored in Redis with the following structure:
- **Key**: `session:{session_id}`
- **Value**: JSON-encoded session data
- **TTL**: Configurable session expiry time

### Session Data Structure

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "device_id": "uuid", 
  "access_count": 0,
  "refresh_count": 0,
  "created_at": "timestamp",
  "expires_at": "timestamp",
  "is_valid": true
}
```

## API Endpoints

### Create Session
```
POST /api/v1/auth/session
```

**Request Body:**
```json
{
  "device_type": "yubikey",
  "auth_code": "otp_code",
  "permission": "yubiapp:read",
  "nonce": "optional_nonce"
}
```

**Response:**
```json
{
  "authenticated": true,
  "session_id": "uuid",
  "access_token": "jwt_token",
  "refresh_token": "jwt_token",
  "user": { ... },
  "device": { ... }
}
```

### Refresh Session
```
POST /api/v1/auth/session/refresh/{session_id}
```

**Request Body:**
```json
{
  "refresh_token": "jwt_refresh_token"
}
```

**Response:**
```json
{
  "session_id": "uuid",
  "access_token": "new_jwt_token",
  "refresh_token": "new_jwt_token"
}
```

## Unified Authentication

### HTTP Method-Based Authentication

The API now uses a unified approach where **all endpoints support both authentication methods** based on the HTTP method:

#### Read Operations (GET methods)
- **Device Authentication**: `Authorization: yubikey:<otp>`
- **Session Authentication**: `Authorization: Bearer <access_token>`

#### Write Operations (POST, PUT, DELETE methods)
- **Device Authentication Only**: `Authorization: yubikey:<otp>`
- **Session Authentication**: Not allowed (returns 403 Forbidden)

### Example Usage

#### Reading Data (GET methods)
```bash
# Using device authentication
curl -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
     http://localhost:8080/api/v1/users

# Using session authentication
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8080/api/v1/users
```

#### Writing Data (POST, PUT, DELETE methods)
```bash
# Using device authentication (works)
curl -X POST -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","username":"test"}' \
     http://localhost:8080/api/v1/users

# Using session authentication (fails with 403)
curl -X POST -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","username":"test"}' \
     http://localhost:8080/api/v1/users
```

### Available Endpoints

All existing endpoints now support unified authentication:

- `GET /api/v1/users` - List users (both auth methods)
- `POST /api/v1/users` - Create user (device auth only)
- `GET /api/v1/users/{id}` - Get user (both auth methods)
- `PUT /api/v1/users/{id}` - Update user (device auth only)
- `DELETE /api/v1/users/{id}` - Delete user (device auth only)
- `GET /api/v1/roles` - List roles (both auth methods)
- `POST /api/v1/roles` - Create role (device auth only)
- `GET /api/v1/resources` - List resources (both auth methods)
- `POST /api/v1/resources` - Create resource (device auth only)
- `GET /api/v1/devices` - List devices (both auth methods)
- `POST /api/v1/devices` - Create device (device auth only)
- `GET /api/v1/actions` - List actions (both auth methods)
- `POST /api/v1/actions` - Create action (device auth only)

## Token Security Features

### Access Tokens
- **Expiry**: 15 minutes (configurable)
- **Claims**: Session ID, user ID, device ID, access count, refresh count
- **Usage**: Single-use (access count increments with each use)
- **Purpose**: API authentication for read operations

### Refresh Tokens
- **Expiry**: 24 hours (configurable)
- **Claims**: Session ID, user ID, device ID, refresh count
- **Usage**: Single-use (refresh count increments with each refresh)
- **Purpose**: Generate new access and refresh tokens

### Security Measures

1. **Token Reuse Prevention**: Access and refresh tokens are invalidated after use
2. **Count Validation**: Tokens include counters that must match session state
3. **Short-lived Access Tokens**: 15-minute expiry limits exposure
4. **Session Invalidation**: Sessions can be marked as invalid
5. **Redis TTL**: Automatic session cleanup
6. **HTTP Method Restrictions**: Session tokens only work for read operations

## Configuration

### Redis Settings
```yaml
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
```

### Session Settings
```yaml
auth:
  access_token_expiry: 15m  # Access token lifetime
  session_expiry: 24h       # Session lifetime
  jwt_secret: "your-secret" # JWT signing secret
```

## Usage Workflow

1. **Initial Authentication**: User authenticates with device at `/auth/session`
2. **Receive Tokens**: Get access token and refresh token
3. **API Access**: Use access token for read operations on any endpoint
4. **Write Operations**: Use device authentication for POST/PUT/DELETE
5. **Token Refresh**: When access token expires, use refresh token at `/auth/session/refresh/{session_id}`
6. **Continue**: Use new tokens for continued access

## Testing

Use the provided test script to verify session functionality:

```bash
./test_session.sh
```

**Prerequisites:**
- Redis server running
- YubiApp server running
- Valid YubiKey OTP (update the script with a real OTP)

## Implementation Notes

### Dependencies Added
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/golang-jwt/jwt/v5` - JWT token handling

### Database Changes
- Added `Session` model (stored in Redis, not PostgreSQL)
- Added JWT token claim structures

### Security Considerations
- Access tokens are short-lived (15 minutes)
- Refresh tokens can only be used once
- Session state is validated on each request
- All tokens include session ID for tracking
- Write operations require device authentication for security

### Advantages of Unified Approach

1. **Single Source of Truth**: Each resource has one canonical endpoint
2. **RESTful Compliance**: Follows REST principles more closely
3. **Reduced Complexity**: No duplicate endpoint definitions
4. **Cleaner Documentation**: OpenAPI spec is much simpler
5. **Intuitive API**: Developers expect `/users/{id}` to work with any valid auth method
6. **Consistent**: Same endpoint structure regardless of auth method
7. **Future-Proof**: Easy to add new auth methods (OAuth, API keys)

### Future Enhancements
- Permission checking for session-based auth
- Session revocation endpoints
- Session activity logging
- Rate limiting for session creation 