#!/bin/bash

# YubiApp PAM Module Installation Script

set -e

echo "Installing YubiApp PAM module..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
if command -v dnf &> /dev/null; then
    dnf install -y gcc make libcurl-devel cjson-devel pam-devel
elif command -v apt-get &> /dev/null; then
    apt-get update
    apt-get install -y gcc make libcurl4-openssl-dev libcjson-dev libpam0g-dev
else
    echo "Unsupported package manager. Please install dependencies manually:"
    echo "  - gcc"
    echo "  - make"
    echo "  - libcurl-devel (or libcurl4-openssl-dev)"
    echo "  - cjson-devel (or libcjson-dev)"
    echo "  - pam-devel (or libpam0g-dev)"
    exit 1
fi

# Build the PAM module
echo "Building PAM module..."
make clean
make

# Install the PAM module
echo "Installing PAM module..."
make install

# Backup original SSH config
echo "Backing up SSH configuration..."
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d_%H%M%S)

# Update SSH configuration
echo "Updating SSH configuration..."
cat >> /etc/ssh/sshd_config << EOF

# YubiApp Environment Variables
PermitUserEnvironment yes
AcceptEnv YUBI_USER_NAME YUBI_USER_USERNAME YUBI_USER_EMAIL
EOF

# Backup original PAM SSH config
echo "Backing up PAM SSH configuration..."
cp /etc/pam.d/sshd /etc/pam.d/sshd.backup.$(date +%Y%m%d_%H%M%S)

# Update PAM configuration
echo "Updating PAM configuration..."
# Insert YubiApp module after password-auth
sed -i '/auth.*substack.*password-auth/a auth       required     pam_yubiapp.so permission=user:read' /etc/pam.d/sshd

# Restart SSH service
echo "Restarting SSH service..."
systemctl restart sshd

echo "Installation complete!"
echo ""
echo "The PAM module has been installed and configured."
echo "Users will now be prompted for Yubikey OTP during SSH authentication."
echo "Environment variables will be set:"
echo "  - YUBI_USER_NAME"
echo "  - YUBI_USER_USERNAME"
echo "  - YUBI_USER_EMAIL"
echo ""
echo "To test, try SSHing to this server and you should be prompted for Yubikey OTP." 