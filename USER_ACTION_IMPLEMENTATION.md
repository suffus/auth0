# User Action Implementation

## Overview

This implementation adds UserActivityHistory tracking for actions with `ActivityType = 'user'` when called via the `/auth/action/{action}` endpoint.

## Changes Made

### 1. Enhanced Action Request Schema

The `/auth/action/{action}` endpoint now accepts an enhanced request body:

```json
{
  "location": "Main Office",           // Optional: Location name
  "status": "working",                 // Optional: User status name
  "start_time": "2025-07-01T10:00:00Z", // Optional: Activity start time
  "end_time": "2025-07-01T18:00:00Z",   // Optional: Activity end time
  "details": {                         // Optional: Additional details
    "shift_type": "day",
    "notes": "Starting morning shift"
  }
}
```

### 2. UserActivityHistory Creation Logic

When an action with `ActivityType = 'user'` is executed:

1. **Location Resolution**: If `location` is provided in the request, it's looked up by name
2. **Status Resolution**: 
   - If `status` is provided in the request, it's used
   - Otherwise, if the action has `default_status` in its details, that's used
   - If neither is available, status remains null
3. **Activity Details**: Combines action details with request details and adds metadata
4. **Previous Activity Closure**: Automatically closes the user's most recent open activity
5. **New Activity Creation**: Creates a new UserActivityHistory entry

### 3. Enhanced Response

For user actions, the response now includes activity information:

```json
{
  "action": "work-start",
  "user_id": "uuid",
  "success": true,
  "message": "Action performed successfully",
  "user_activity": {
    "id": "uuid",
    "from_datetime": "2025-07-01T10:00:00Z",
    "status": { /* UserStatus object */ },
    "location": { /* Location object */ }
  }
}
```

### 4. Authentication Logging

- Authentication middleware still logs to AuthenticationLog
- UserActivityHistory replaces the previous action logging in handlePerformAction
- Both logs are created for comprehensive audit trail

## Implementation Details

### Files Modified

1. **`internal/server/handlers_actions.go`**
   - Added `ActionRequest` struct for new request schema
   - Enhanced `handlePerformAction` with UserActivityHistory creation
   - Added location and status resolution logic
   - Integrated with UserActivityService

2. **`internal/server/router.go`**
   - Updated action handler registration to include new services

3. **`internal/services/auth_service.go`**
   - Fixed UserID pointer assignment in LogAuthentication

4. **`openapi.yaml`**
   - Updated action endpoint documentation
   - Added new request/response schemas

### Services Used

- `UserActivityService`: Creates activity history entries
- `LocationService`: Resolves location names to objects
- `UserStatusService`: Resolves status names to objects
- `ActionService`: Retrieves action details and permissions
- `AuthService`: Handles authentication and logging

## Usage Examples

### Creating a User Action

```bash
curl -X POST "http://localhost:8080/api/v1/actions" \
  -H "Authorization: yubikey:<otp>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "work-start",
    "activity_type": "user",
    "details": {
      "default_status": "working",
      "description": "Start of work shift"
    },
    "active": true
  }'
```

### Executing a User Action

```bash
curl -X POST "http://localhost:8080/api/v1/auth/action/work-start" \
  -H "Authorization: yubikey:<otp>" \
  -H "Content-Type: application/json" \
  -d '{
    "location": "Main Office",
    "status": "working",
    "details": {
      "shift_type": "day",
      "notes": "Starting morning shift"
    }
  }'
```

## Testing

Use the provided test script:

```bash
./test_user_action.sh
```

This script tests:
1. Creating user actions
2. Executing actions with location/status
3. Using default status from action details
4. Creating break actions
5. Checking activity history

## Benefits

1. **Automatic Activity Tracking**: User activities are automatically logged when actions are performed
2. **Flexible Status Management**: Supports both explicit status and default status from action configuration
3. **Location Awareness**: Tracks where actions are performed
4. **Rich Metadata**: Captures device, IP, user agent, and custom details
5. **Activity Continuity**: Automatically closes previous activities when new ones start
6. **Comprehensive Audit Trail**: Maintains both authentication logs and activity history

## Future Enhancements

- Support for activity duration tracking
- Integration with time tracking systems
- Activity analytics and reporting
- Bulk activity operations
- Activity templates and workflows 