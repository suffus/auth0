# CCode Directory

This directory contains C/C++ components for the YubiApp project.

## Contents

- **pam_yubiapp.c** - PAM module for SSH integration with YubiApp
- **Makefile** - Build configuration for the PAM module
- **install_pam.sh** - Automated installation script
- **sshd_config.patch** - SSH configuration changes needed
- **pam_sshd_config** - PAM configuration example
- **PAM_README.md** - Detailed documentation for the PAM module

## Quick Start

To build and install the PAM module:

```bash
# Install dependencies
sudo dnf install -y gcc make libcurl-devel cjson-devel pam-devel

# Build and install
make clean && make
sudo make install

# Or use the automated installer
sudo ./install_pam.sh
```

## Documentation

See `PAM_README.md` for complete documentation on the PAM module functionality, configuration, and usage. 