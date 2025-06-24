CC = gcc
CFLAGS = -Wall -Wextra -O2 -fPIC
LDFLAGS = -shared
LIBS = -lpam -lcurl -lcjson

PAM_MODULE = pam_yubiapp.so
PAM_INSTALL_DIR = /usr/lib64/security

# Default target
all: $(PAM_MODULE)

# Build the PAM module
$(PAM_MODULE): pam_yubiapp.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $< $(LIBS)

# Install the PAM module
install: $(PAM_MODULE)
	sudo cp $(PAM_MODULE) $(PAM_INSTALL_DIR)/
	sudo chmod 755 $(PAM_INSTALL_DIR)/$(PAM_MODULE)

# Clean build artifacts
clean:
	rm -f $(PAM_MODULE)

# Install dependencies (Fedora/RHEL)
install-deps:
	sudo dnf install -y gcc make libcurl-devel cjson-devel pam-devel

# Install dependencies (Ubuntu/Debian)
install-deps-ubuntu:
	sudo apt-get update
	sudo apt-get install -y gcc make libcurl4-openssl-dev libcjson-dev libpam0g-dev

.PHONY: all install clean install-deps install-deps-ubuntu 