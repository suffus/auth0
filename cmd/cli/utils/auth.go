package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/YubiApp/internal/config"
)

// verifyYubikeyOTP verifies a YubiKey OTP using the YubiCloud API
func VerifyYubikeyOTP(otp string, config config.YubikeyConfig) error {
	if len(otp) < 12 {
		return fmt.Errorf("invalid OTP format")
	}

	// Extract the public ID (first 12 characters) - not used in this implementation
	_ = otp[:12]

	// Build the verification URL
	baseURL := "https://api.yubico.com/wsapi/2.0/verify"
	params := url.Values{}
	params.Set("id", config.ClientID)
	params.Set("otp", otp)
	params.Set("nonce", "yubiapp-cli-nonce")

	// Make the request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return fmt.Errorf("failed to verify OTP: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response := string(body[:n])

	// Parse the response
	lines := strings.Split(response, "\r\n")
	responseMap := make(map[string]string)
	for _, line := range lines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				responseMap[parts[0]] = parts[1]
			}
		}
	}

	// Check if verification was successful
	if status, exists := responseMap["status"]; !exists || status != "OK" {
		return fmt.Errorf("OTP verification failed: %s", responseMap["status"])
	}

	// Verify the signature if a secret key is provided
	if config.SecretKey != "" {
		if err := verifySignature(response, config.SecretKey); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}

	return nil
}

// verifySignature verifies the HMAC signature of the YubiCloud response
func verifySignature(response, secretKey string) error {
	lines := strings.Split(response, "\r\n")
	var hmacParams []string
	var signature string

	for _, line := range lines {
		if strings.HasPrefix(line, "h=") {
			signature = strings.TrimPrefix(line, "h=")
		} else if strings.Contains(line, "=") && !strings.HasPrefix(line, "h=") {
			hmacParams = append(hmacParams, line)
		}
	}

	if signature == "" {
		return fmt.Errorf("no signature found in response")
	}

	// Sort parameters alphabetically
	// Note: This is a simplified version. In production, you'd want proper sorting
	paramString := strings.Join(hmacParams, "&")

	// Calculate HMAC
	key, err := hex.DecodeString(secretKey)
	if err != nil {
		return fmt.Errorf("invalid secret key: %w", err)
	}

	h := hmac.New(sha1.New, key)
	h.Write([]byte(paramString))
	calculatedSignature := hex.EncodeToString(h.Sum(nil))

	if calculatedSignature != signature {
		return fmt.Errorf("signature mismatch")
	}

	return nil
} 