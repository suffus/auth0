# YubiApp PAM Module

This PAM module provides Yubikey OTP authentication for SSH sessions using the YubiApp API. It also sets environment variables with user information from the authentication response.

## Features

- Yubikey OTP authentication via YubiApp API
- Environment variable injection for SSH sessions
- Configurable permission requirements
- JSON response parsing using cJSON
- HTTP communication using libcurl

## Environment Variables Set

Upon successful authentication, the following environment variables are set:

- `YUBI_USER_NAME` - User's full name (first_name + last_name)
- `YUBI_USER_USERNAME` - User's username
- `YUBI_USER_EMAIL` - User's email address

## Installation

### Quick Install

```bash
sudo ./install_pam.sh
```

### Manual Installation

1. **Install Dependencies:**

   **Fedora/RHEL:**
   ```bash
   sudo dnf install -y gcc make libcurl-devel cjson-devel pam-devel
   ```

   **Ubuntu/Debian:**
   ```bash
   sudo apt-get update
   sudo apt-get install -y gcc make libcurl4-openssl-dev libcjson-dev libpam0g-dev
   ```

2. **Build and Install:**
   ```bash
   make clean
   make
   sudo make install
   ```

3. **Configure SSH:**

   Add to `/etc/ssh/sshd_config`:
   ```
   PermitUserEnvironment yes
   AcceptEnv YUBI_USER_NAME YUBI_USER_USERNAME YUBI_USER_EMAIL
   ```

4. **Configure PAM:**

   Add to `/etc/pam.d/sshd` after the password-auth line:
   ```
   auth       required     pam_yubiapp.so permission=user:read
   ```

5. **Restart SSH:**
   ```bash
   sudo systemctl restart sshd
   ```

## Configuration

### Module Options

The PAM module accepts the following options:

- `permission=<resource>:<action>` - Required permission for authentication
  - Default: `user:read`
  - Example: `permission=admin:write`

### Example PAM Configuration

```
# Basic authentication
auth required pam_yubiapp.so

# With specific permission
auth required pam_yubiapp.so permission=admin:read

# Multiple permissions (use multiple lines)
auth required pam_yubiapp.so permission=user:read
auth required pam_yubiapp.so permission=admin:write
```

## Usage

1. **SSH Connection:**
   ```bash
   ssh user@server
   ```

2. **Authentication Prompt:**
   ```
   Password: [enter password]
   Yubikey OTP: [enter Yubikey OTP]
   ```

3. **Environment Variables:**
   Once authenticated, the environment variables will be available in the SSH session:
   ```bash
   echo $YUBI_USER_NAME
   echo $YUBI_USER_USERNAME
   echo $YUBI_USER_EMAIL
   ```

## API Requirements

The PAM module expects the YubiApp API to be running and accessible at:
```
http://localhost:8080/api/v1/auth/device
```

The API should accept POST requests with JSON:
```json
{
  "device_type": "yubikey",
  "auth_code": "yubikey_otp_here",
  "permission": "resource:action"
}
```

And return JSON:
```json
{
  "authenticated": true,
  "user": {
    "id": "uuid",
    "email": "john.doe@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "active": true,
    "roles": [...]
  }
}
```

## Troubleshooting

### Check PAM Module Loading
```bash
sudo pam_tally2 --user username
```

### Check SSH Logs
```bash
sudo journalctl -u sshd -f
```

### Test PAM Module
```bash
sudo pam_test pam_yubiapp.so
```

### Verify Environment Variables
```bash
ssh user@server 'env | grep YUBI'
```

### Check API Response
The PAM module logs the full API request and response. Check the system logs:
```bash
sudo journalctl -f | grep pam_yubiapp
```

## Security Considerations

1. **Network Security:** Ensure the YubiApp API is only accessible from trusted networks
2. **API Security:** Use HTTPS for the YubiApp API in production
3. **Permission Granularity:** Use specific permissions rather than broad ones
4. **Logging:** Monitor authentication logs for suspicious activity
5. **Backup:** Keep backups of original SSH and PAM configurations

## Files

- `pam_yubiapp.c` - PAM module source code
- `Makefile` - Build configuration
- `install_pam.sh` - Automated installation script
- `sshd_config.patch` - SSH configuration changes
- `pam_sshd_config` - PAM configuration example

## Dependencies

- libpam (PAM development libraries)
- libcurl (HTTP client library)
- cjson (JSON parsing library)
- gcc (C compiler)
- make (Build system) 