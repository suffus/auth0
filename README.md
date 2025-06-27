# YubiApp - Multi-factor Authentication Service

A comprehensive authentication and authorization service supporting multiple authentication methods, authorization schemes, and action-based security controls.

## Features

- **Multiple Authentication Methods:**
  - Yubikey OTP
  - TOTP (Google Authenticator, Microsoft Authenticator)
  - SMS OTP
  - Email OTP
- **Authorization:**
  - Role-based access control (RBAC)
  - Attribute-based access control (ABAC)
  - Action-based security controls
- **Multi-factor Authentication**
- **Single Sign-on (SSO)**
- **Device Registration & Management**
- **Action Logging & Auditing**
- **REST API for authentication and management**
- **CLI for administration and testing**
- **PAM module for SSH integration**

## Project Structure

```
YubiApp/
├── cmd/                           # Application entry points
│   └── cli/                      # Command-line interface
│       ├── main.go               # CLI main application
│       ├── action.sh             # CLI action script
│       ├── yubiapp-cli           # Compiled CLI binary
│       └── README.md             # CLI documentation
├── internal/                      # Private application code
│   ├── auth/                     # Authentication logic
│   │   ├── service.go            # Authentication service
│   │   └── yubikey.go            # Yubikey integration
│   ├── config/                   # Configuration handling
│   │   └── config.go             # Configuration management
│   ├── database/                 # Database models and migrations
│   │   └── models.go             # Database models (Users, Roles, Resources, etc.)
│   ├── server/                   # HTTP server and routing
│   │   ├── server.go             # HTTP server setup
│   │   ├── router.go             # API routing
│   │   ├── middleware.go         # HTTP middleware
│   │   ├── handlers_*.go         # API handlers for different resources
│   │   └── utils.go              # Server utilities
│   └── services/                 # Business logic services
│       ├── user_service.go       # User management
│       ├── role_service.go       # Role management
│       ├── resource_service.go   # Resource management
│       ├── permission_service.go # Permission management
│       ├── auth_service.go       # Authentication service
│       ├── device_service.go     # Device management
│       ├── device_registration_service.go # Device registration
│       └── action_service.go     # Action management
├── CCode/                        # C/C++ components
│   ├── pam_yubiapp.c            # PAM module for SSH integration
│   ├── Makefile                 # Build configuration for PAM module
│   ├── install_pam.sh           # PAM module installation script
│   ├── sshd_config.patch        # SSH configuration changes
│   ├── pam_sshd_config          # PAM configuration example
│   ├── PAM_README.md            # PAM module documentation
│   ├── test_api.sh              # API testing script
│   └── README.md                # CCode documentation
├── database/                     # Database setup
│   ├── schema.sql               # Database schema
│   └── setup.sh                 # Database setup script
├── claude/                       # Documentation and analysis
│   ├── API_DOCUMENTATION.md     # Comprehensive API documentation
│   ├── API_REFERENCE.md         # API reference guide
│   └── DEVICE_REGISTRATION_ANALYSIS.md # Device registration analysis
├── config.example.yaml           # Configuration template
├── go.mod                       # Go module file
├── go.sum                       # Go dependencies checksum
├── openapi.yaml                 # OpenAPI specification
├── appspec.md                   # Application specification
├── useractions.md               # User actions specification
└── README.md                    # Project documentation
```

## Core Components

### Authentication System
- **Multi-factor Authentication**: Support for Yubikey OTP, TOTP, SMS, and Email
- **Device Registration**: Secure device onboarding with QR codes
- **Session Management**: Token-based authentication with refresh capabilities

### Authorization System
- **Role-Based Access Control (RBAC)**: Users, roles, and permissions
- **Resource Management**: Granular resource access control
- **Action-Based Security**: Configurable actions with permission requirements

### API Endpoints
- **Authentication**: `/auth/login`, `/auth/refresh`, `/auth/logout`
- **Device Management**: `/devices`, `/devices/register`
- **User Management**: `/users`, `/roles`, `/resources`
- **Actions**: `/auth/action/{action_name}` - Action-based security controls
- **Audit Logging**: Authentication and action logs

### CLI Interface
- **User Management**: Create, read, update, delete users
- **Role Management**: Manage roles and permissions
- **Resource Management**: Configure access resources
- **Action Management**: Create and manage security actions
- **Testing**: Simulate API calls and test authentication flows

### PAM Integration
- **SSH Integration**: PAM module for SSH authentication
- **Environment Injection**: Secure credential passing to SSH sessions
- **Configuration**: Automated setup scripts and configuration examples

## Setup

### 1. Install Dependencies:
```bash
go mod download
```

### 2. Configure the Application:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

### 3. Set Up Database:
```bash
./database/setup.sh
```

### 4. Run Migrations:
```bash
go run cmd/cli/main.go migrate
```

### 5. Start the Server:
```bash
go run cmd/cli/main.go serve
```

### 6. Install PAM Module (Optional):
```bash
cd CCode
sudo ./install_pam.sh
```

## Usage Examples

### CLI Commands
```bash
# List all users
./cmd/cli/yubiapp-cli users list

# Create a new action
./cmd/cli/yubiapp-cli actions create --name "ssh-login" --description "SSH login action" --permissions "ssh-access"

# Test an action
./cmd/cli/yubiapp-cli log-action --action "ssh-login" --device-id "your-device-id"
```

### API Examples
```bash
# Authenticate a device
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"device_id": "your-device-id", "otp": "your-otp"}'

# Log an action
curl -X POST http://localhost:8080/auth/action/ssh-login \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"details": {"ip": "192.168.1.100", "user": "admin"}}'
```

## API Documentation

- **OpenAPI Specification**: `openapi.yaml`
- **Comprehensive Documentation**: `claude/API_DOCUMENTATION.md`
- **API Reference**: `claude/API_REFERENCE.md`
- **Interactive Documentation**: Available at `/swagger/index.html` when running the server

## Database Schema

The application uses PostgreSQL with the following key tables:
- **Users**: User accounts and profiles
- **Roles**: Role definitions
- **Resources**: Access-controlled resources
- **Permissions**: Permission definitions
- **Devices**: Registered authentication devices
- **Actions**: Configurable security actions
- **AuthenticationLog**: Authentication event logging
- **ActionLog**: Action execution logging

## Security Features

- **Multi-factor Authentication**: Multiple authentication methods
- **Device Registration**: Secure device onboarding
- **Action-Based Security**: Configurable security controls
- **Audit Logging**: Comprehensive event logging
- **Role-Based Access Control**: Granular permission management
- **Token-Based Authentication**: Secure session management

## PAM Module

The PAM module in the `CCode/` directory provides SSH integration with environment variable injection. See `CCode/PAM_README.md` for detailed documentation.

## License

MIT 