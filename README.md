# YubiApp - Multi-factor Authentication Service

A comprehensive authentication and authorization service supporting multiple authentication methods and authorization schemes.

## Features

- Multiple authentication methods:
  - Yubikey OTP
  - TOTP (Google Authenticator, Microsoft Authenticator)
  - SMS OTP
  - Email OTP
- Authorization:
  - Role-based access control (RBAC)
  - Attribute-based access control (ABAC)
- Multi-factor authentication
- Single sign-on (SSO)
- Web management interface
- REST API for authentication and management
- PAM module for SSH integration

## Project Structure

```
.
├── cmd/                    # Application entry points
│   └── api/               # REST API server
├── internal/              # Private application code
│   ├── auth/              # Authentication logic
│   ├── config/            # Configuration handling
│   ├── database/          # Database models and migrations
│   ├── server/            # HTTP server and routing
│   └── service/           # Business logic
├── pkg/                   # Public libraries
│   ├── authz/             # Authorization utilities
│   └── utils/             # Common utilities
├── CCode/                 # C/C++ components
│   ├── pam_yubiapp.c      # PAM module for SSH integration
│   ├── Makefile           # Build configuration for PAM module
│   ├── install_pam.sh     # PAM module installation script
│   ├── sshd_config.patch  # SSH configuration changes
│   ├── pam_sshd_config    # PAM configuration example
│   └── PAM_README.md      # PAM module documentation
├── database/              # Database setup
│   ├── schema.sql         # Database schema
│   └── setup.sh           # Database setup script
├── web/                   # Web interface assets
├── config.example.yaml    # Configuration template
├── go.mod                 # Go module file
└── README.md             # Project documentation
```

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
go run cmd/api/main.go migrate
```

### 5. Start the Server:
```bash
go run cmd/api/main.go serve
```

### 6. Install PAM Module (Optional):
```bash
cd CCode
sudo ./install_pam.sh
```

## API Documentation

API documentation is available at `/swagger/index.html` when running the server.

## PAM Module

The PAM module in the `CCode/` directory provides SSH integration with environment variable injection. See `CCode/PAM_README.md` for detailed documentation.

## License

MIT 