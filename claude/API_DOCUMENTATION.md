# YubiApp API Documentation

## Overview

This document describes the new action-based API endpoints that have been added to the YubiApp system. These endpoints allow for user action authorization and recording, as specified in the `useractions.md` file.

## Database Changes

### New Tables

#### Actions Table
The `actions` table stores all possible actions that users can perform:

```sql
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) UNIQUE NOT NULL,
    required_permissions JSONB DEFAULT '[]'::jsonb
);
```

- `name`: The name of the action (e.g., "ssh-login", "app-install")
- `required_permissions`: JSON array of permission strings that the user must have to perform this action

#### Modified AuthenticationLog Table
The `authentication_logs` table has been updated to include action tracking:

```sql
ALTER TABLE authentication_logs ADD COLUMN action_id UUID REFERENCES actions(id) ON DELETE SET NULL;
ALTER TABLE authentication_logs ADD COLUMN json_detail JSONB;
ALTER TABLE authentication_logs ALTER COLUMN type TYPE VARCHAR(50) CHECK (type IN ('login', 'logout', 'refresh', 'mfa', 'action'));
```

- `action_id`: Reference to the action being performed
- `json_detail`: JSON data specific to the action being performed
- `type`: Now includes 'action' as a valid type

## API Endpoints

### 1. Perform Action - POST `/api/v1/auth/action/{action_name}`

This is the main endpoint for performing user actions with device-based authentication.

#### Request
- **Method**: POST
- **URL**: `/api/v1/auth/action/{action_name}`
- **Headers**:
  - `Authorization`: `yubikey:<device_code>` (required)
  - `Content-Type`: `application/json`

#### Parameters
- `action_name` (path): The name of the action to perform

#### Request Body
JSON object containing action-specific data. The structure varies by action type.

**Example for ssh-login:**
```json
{
  "resource": "aws-cloud-west/server101",
  "login": "support"
}
```

**Example for app-install:**
```json
{
  "app_name": "my-application",
  "version": "1.2.3",
  "target_environment": "production"
}
```

#### Response

**Success (200):**
```json
{
  "action": "ssh-login",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "success": true,
  "message": "Action performed successfully"
}
```

**Authentication Failed (401):**
```json
{
  "error": "Authentication failed: invalid device code"
}
```

**Permission Denied (403):**
```json
{
  "error": "User does not have required permissions for action 'ssh-login'"
}
```

**Action Not Found (404):**
```json
{
  "error": "Action 'invalid-action' not found"
}
```

#### Authentication Flow
1. Extract device code from Authorization header
2. Authenticate user using device code
3. Verify action exists in database
4. Check user has required permissions for the action
5. Log the action in AuthenticationLog table
6. Return success response

### 2. Action Management Endpoints

#### List Actions - GET `/api/v1/actions`
Retrieve all available actions.

**Response:**
```json
{
  "actions": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "ssh-login",
      "required_permissions": ["ssh:login"],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### Get Action - GET `/api/v1/actions/{id}`
Retrieve a specific action by ID.

#### Create Action - POST `/api/v1/actions`
Create a new action.

**Request Body:**
```json
{
  "name": "new-action",
  "required_permissions": ["resource:action1", "resource:action2"]
}
```

#### Update Action - PUT `/api/v1/actions/{id}`
Update an existing action.

#### Delete Action - DELETE `/api/v1/actions/{id}`
Delete an action.

## Permission System

### Permission Format
Permissions are stored as strings in the format `"resource:action"`:

- `ssh:login` - Permission to perform SSH login
- `app:install` - Permission to install applications
- `app:uninstall` - Permission to uninstall applications
- `permission:grant` - Permission to grant permissions to users
- `permission:revoke` - Permission to revoke permissions from users

### Default Actions
The system comes with several pre-configured actions:

1. **ssh-login** - Requires `ssh:login` permission
2. **app-install** - Requires `app:install` permission
3. **app-uninstall** - Requires `app:uninstall` permission
4. **permission-grant** - Requires `permission:grant` permission
5. **permission-revoke** - Requires `permission:revoke` permission
6. **user-signin** - No permissions required
7. **user-signout** - No permissions required

## Usage Examples

### Example 1: SSH Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/action/ssh-login \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "resource": "aws-cloud-west/server101",
    "login": "support"
  }'
```

### Example 2: App Installation
```bash
curl -X POST http://localhost:8080/api/v1/auth/action/app-install \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "app_name": "my-application",
    "version": "1.2.3",
    "target_environment": "production"
  }'
```

### Example 3: Create New Action
```bash
curl -X POST http://localhost:8080/api/v1/actions \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "database-backup",
    "required_permissions": ["database:backup"]
  }'
```

## Error Handling

The API returns appropriate HTTP status codes:

- **200**: Success
- **400**: Bad Request (invalid JSON, missing parameters)
- **401**: Unauthorized (authentication failed)
- **403**: Forbidden (permission denied)
- **404**: Not Found (action doesn't exist)
- **500**: Internal Server Error

## Security Considerations

1. **Authentication**: All action endpoints require device-based authentication
2. **Authorization**: Actions check for required permissions before execution
3. **Logging**: All actions are logged with user, device, and action details
4. **Input Validation**: Request bodies are validated for proper JSON format
5. **Permission Inheritance**: Users inherit permissions through their roles

## Database Migration

To apply the database changes, run the updated `schema.sql` file:

```bash
psql -d yubiapp -f database/schema.sql
```

This will:
1. Create the new `actions` table
2. Add the new columns to `authentication_logs`
3. Insert default actions
4. Create necessary indexes

## Implementation Notes

1. The action system is designed to be extensible - new actions can be added without code changes
2. Action-specific data is stored in the `json_detail` field of the authentication log
3. The system supports both permission-based and permission-free actions
4. All actions are logged for audit purposes
5. The API follows RESTful conventions for action management endpoints 