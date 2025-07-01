# YubiApp CLI Management Tool

A comprehensive command-line tool for managing YubiApp users, roles, permissions, resources, and devices.

## Building the CLI Tool

```bash
go build -o yubiapp-cli cmd/cli/main.go
```

## Configuration

The CLI tool uses the same configuration file as the main application (`config.yaml`). Make sure your database connection details are properly configured.

## Usage

### General Syntax

```bash
./yubiapp-cli [command] [subcommand] [flags] [arguments]
```

### User Management

#### Create a new user

```bash
./yubiapp-cli user create \
  --email "john.doe@example.com" \
  --username "johndoe" \
  --password "securepassword123" \
  --first-name "John" \
  --last-name "Doe" \
  --active true
```

#### List all users

```bash
./yubiapp-cli user list
```

#### List only active users

```bash
./yubiapp-cli user list --active-only
```

#### Delete a user

```bash
# By email
./yubiapp-cli user delete "john.doe@example.com"

# By UUID
./yubiapp-cli user delete "550e8400-e29b-41d4-a716-446655440000"
```

### Resource Management

#### Create a new resource

```bash
./yubiapp-cli resource create \
  --name "web-server-01" \
  --type "server" \
  --location "datacenter-east" \
  --department "IT" \
  --active true
```

#### List all resources

```bash
./yubiapp-cli resource list
```

#### List only active resources

```bash
./yubiapp-cli resource list --active-only
```

#### Delete a resource

```bash
# By name
./yubiapp-cli resource delete "web-server-01"

# By UUID
./yubiapp-cli resource delete "550e8400-e29b-41d4-a716-446655440000"
```

### Role Management

#### Create a new role

```bash
./yubiapp-cli role create \
  --name "admin" \
  --description "Administrator with full access"
```

#### List all roles

```bash
./yubiapp-cli role list
```

#### Delete a role

```bash
# By name
./yubiapp-cli role delete "admin"

# By UUID
./yubiapp-cli role delete "550e8400-e29b-41d4-a716-446655440000"
```

### Permission Management

#### Create a new permission

```bash
./yubiapp-cli permission create \
  --resource-id "550e8400-e29b-41d4-a716-446655440000" \
  --action "read" \
  --effect "allow"
```

#### List all permissions

```bash
./yubiapp-cli permission list
```

#### Delete a permission

```bash
./yubiapp-cli permission delete "550e8400-e29b-41d4-a716-446655440000"
```

### Device Management

#### Create a YubiKey device

```bash
./yubiapp-cli device create \
  --user-id "550e8400-e29b-41d4-a716-446655440000" \
  --type "yubikey" \
  --identifier "ccccccbchvth" \
  --active true
```

#### Create a TOTP device

```bash
./yubiapp-cli device create \
  --user-id "550e8400-e29b-41d4-a716-446655440000" \
  --type "totp" \
  --identifier "john.doe@example.com" \
  --active true
```

#### Create an SMS device

```bash
./yubiapp-cli device create \
  --user-id "550e8400-e29b-41d4-a716-446655440000" \
  --type "sms" \
  --identifier "+1234567890" \
  --active true
```

#### Create an Email device

```bash
./yubiapp-cli device create \
  --user-id "550e8400-e29b-41d4-a716-446655440000" \
  --type "email" \
  --identifier "john.doe@example.com" \
  --active true
```

#### List all devices

```bash
./yubiapp-cli device list
```

#### List only active devices

```bash
./yubiapp-cli device list --active-only
```

#### List devices for a specific user

```bash
./yubiapp-cli device list --user-id "550e8400-e29b-41d4-a716-446655440000"
```

#### List only active devices for a specific user

```bash
./yubiapp-cli device list --user-id "550e8400-e29b-41d4-a716-446655440000" --active-only
```

#### Delete a device

```bash
./yubiapp-cli device delete "550e8400-e29b-41d4-a716-446655440000"
```

### Location Management

#### Create a new location

```bash
./yubiapp-cli location create \
  --name "Main Office" \
  --description "Primary office location" \
  --address "123 Main St, City, State 12345" \
  --type "office" \
  --active true
```

#### List all locations

```bash
./yubiapp-cli location list
```

#### List only active locations

```bash
./yubiapp-cli location list --active-only
```

#### Update a location

```bash
./yubiapp-cli location update "550e8400-e29b-41d4-a716-446655440000" \
  --name "Updated Office Name" \
  --description "Updated description" \
  --address "456 New St, City, State 12345" \
  --type "office" \
  --active false
```

#### Delete a location (soft delete - marks as inactive)

```bash
# By name
./yubiapp-cli location delete "Main Office"

# By UUID
./yubiapp-cli location delete "550e8400-e29b-41d4-a716-446655440000"
```

### Assignment Management

#### Assign a user to a role

```bash
# Using email and role name
./yubiapp-cli assign user-role "john.doe@example.com" "admin"

# Using UUIDs
./yubiapp-cli assign user-role "550e8400-e29b-41d4-a716-446655440000" "550e8400-e29b-41d4-a716-446655440001"
```

#### Remove a user from a role

```bash
# Using email and role name
./yubiapp-cli assign remove-user-role "john.doe@example.com" "admin"

# Using UUIDs
./yubiapp-cli assign remove-user-role "550e8400-e29b-41d4-a716-446655440000" "550e8400-e29b-41d4-a716-446655440001"
```

#### Assign a permission to a role

```bash
# Using permission UUID and role name
./yubiapp-cli assign permission-role "550e8400-e29b-41d4-a716-446655440000" "admin"

# Using permission UUID and role UUID
./yubiapp-cli assign permission-role "550e8400-e29b-41d4-a716-446655440000" "550e8400-e29b-41d4-a716-446655440001"
```

#### Remove a permission from a role

```bash
# Using permission UUID and role name
./yubiapp-cli assign remove-permission-role "550e8400-e29b-41d4-a716-446655440000" "admin"

# Using permission UUID and role UUID
./yubiapp-cli assign remove-permission-role "550e8400-e29b-41d4-a716-446655440000" "550e8400-e29b-41d4-a716-446655440001"
```

## Complete Example Workflow

Here's a complete example of setting up a user with roles, resources, permissions, and devices:

```bash
# 1. Create a user
./yubiapp-cli user create \
  --email "admin@example.com" \
  --username "admin" \
  --password "adminpass123" \
  --first-name "Admin" \
  --last-name "User" \
  --active true

# 2. Create locations
./yubiapp-cli location create \
  --name "Main Office" \
  --description "Primary office location" \
  --address "123 Main St, City, State 12345" \
  --type "office" \
  --active true

./yubiapp-cli location create \
  --name "Data Center East" \
  --description "Primary data center" \
  --address "456 Tech Blvd, City, State 12345" \
  --type "office" \
  --active true

# 3. Create resources
./yubiapp-cli resource create \
  --name "web-server-01" \
  --type "server" \
  --location "datacenter-east" \
  --department "IT"

./yubiapp-cli resource create \
  --name "user-database" \
  --type "database" \
  --location "datacenter-west" \
  --department "Engineering"

# 4. Create admin role
./yubiapp-cli role create \
  --name "admin" \
  --description "Administrator with full access"

# 5. Create permissions for resources
./yubiapp-cli permission create \
  --resource-id "WEB_SERVER_RESOURCE_UUID" \
  --action "read" \
  --effect "allow"

./yubiapp-cli permission create \
  --resource-id "WEB_SERVER_RESOURCE_UUID" \
  --action "write" \
  --effect "allow"

./yubiapp-cli permission create \
  --resource-id "DATABASE_RESOURCE_UUID" \
  --action "read" \
  --effect "allow"

./yubiapp-cli permission create \
  --resource-id "DATABASE_RESOURCE_UUID" \
  --action "write" \
  --effect "allow"

# 6. Assign permissions to admin role
./yubiapp-cli assign permission-role "PERMISSION_UUID_1" "admin"
./yubiapp-cli assign permission-role "PERMISSION_UUID_2" "admin"
./yubiapp-cli assign permission-role "PERMISSION_UUID_3" "admin"
./yubiapp-cli assign permission-role "PERMISSION_UUID_4" "admin"

# 7. Assign user to admin role
./yubiapp-cli assign user-role "admin@example.com" "admin"

# 8. Add a YubiKey device for the user
./yubiapp-cli device create \
  --user-id "USER_UUID" \
  --type "yubikey" \
  --identifier "ccccccbchvth" \
  --active true

# 9. Verify the setup
./yubiapp-cli user list
./yubiapp-cli location list
./yubiapp-cli resource list
./yubiapp-cli role list
./yubiapp-cli device list --user-id "USER_UUID"
```

## Resource Types

The CLI supports the following resource types:

- **server**: Physical or virtual servers
- **service**: Application services or microservices
- **database**: Database instances
- **application**: Applications or software systems

## Location Types

The CLI supports the following location types:

- **office**: Office locations
- **home**: Home/remote work locations
- **event**: Event or temporary locations
- **other**: Other location types

## Device Types

The CLI supports the following device types:

- **yubikey**: YubiKey OTP devices
- **totp**: Time-based One-Time Password devices (Google Authenticator, etc.)
- **sms**: SMS-based authentication
- **email**: Email-based authentication

## Notes

- For TOTP devices, if no secret is provided, a random 32-byte secret will be automatically generated
- User passwords are automatically hashed using bcrypt
- All UUIDs are automatically generated for new entities
- The tool validates device types, resource types, location types, and permission effects
- Duplicate assignments are prevented (users can't be assigned to the same role twice)
- Resources are now properly separated from permissions, allowing for better resource management
- The `--active-only` flag can be used with list commands to show only active entities (users, resources, devices, locations)
- Location deletion is a soft delete that marks the location as inactive rather than removing it from the database
- The tool provides detailed error messages for common issues

## Error Handling

The CLI tool provides comprehensive error handling:

- Database connection errors
- Invalid UUID formats
- Missing required fields
- Duplicate entries
- Non-existent entities
- Invalid device types, resource types, or permission effects

All errors are displayed with clear messages to help troubleshoot issues. 