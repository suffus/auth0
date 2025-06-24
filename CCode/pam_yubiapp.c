/*
 * pam_yubiapp.c - PAM module for YubiApp authentication
 * 
 * This module authenticates users using Yubikey OTP and sets environment
 * variables with user information from the JSON response.
 */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <curl/curl.h>
#include <cjson/cJSON.h>
#include <security/pam_modules.h>
#include <security/pam_ext.h>
#include <syslog.h>

#define YUBIAPP_URL "http://localhost:8080/api/v1/auth/device"
#define MAX_OTP_LENGTH 64
#define MAX_RESPONSE_SIZE 4096

// Structure to hold response data for libcurl
struct MemoryStruct {
    char *memory;
    size_t size;
};

// Callback function for libcurl to write response data
static size_t WriteMemoryCallback(void *contents, size_t size, size_t nmemb, void *userp) {
    size_t realsize = size * nmemb;
    struct MemoryStruct *mem = (struct MemoryStruct *)userp;

    char *ptr = realloc(mem->memory, mem->size + realsize + 1);
    if (!ptr) {
        printf("not enough memory (realloc returned NULL)\n");
        return 0;
    }

    mem->memory = ptr;
    memcpy(&(mem->memory[mem->size]), contents, realsize);
    mem->size += realsize;
    mem->memory[mem->size] = 0;

    return realsize;
}

// Function to set environment variable
static void set_env_var(pam_handle_t *pamh, const char *name, const char *value) {
    if (name && value) {
        char *env_var = malloc(strlen(name) + strlen(value) + 2);
        if (env_var) {
            sprintf(env_var, "%s=%s", name, value);
            pam_putenv(pamh, env_var);
            free(env_var);
        }
    }
}

// Function to parse JSON response and set environment variables
static int parse_response_and_set_env(pam_handle_t *pamh, const char *json_response) {
    cJSON *json = cJSON_Parse(json_response);
    if (!json) {
        pam_syslog(pamh, LOG_ERR, "Failed to parse JSON response: %s", json_response);
        return PAM_SYSTEM_ERR;
    }

    // Check if authentication was successful
    cJSON *authenticated = cJSON_GetObjectItem(json, "authenticated");
    if (!authenticated || !cJSON_IsTrue(authenticated)) {
        cJSON *error = cJSON_GetObjectItem(json, "error");
        if (error && cJSON_IsString(error)) {
            pam_syslog(pamh, LOG_ERR, "Authentication failed: %s", error->valuestring);
        } else {
            pam_syslog(pamh, LOG_ERR, "Authentication failed - authenticated field is false");
        }
        cJSON_Delete(json);
        return PAM_AUTH_ERR;
    }

    // Get user information from the response
    cJSON *user = cJSON_GetObjectItem(json, "user");
    if (user && cJSON_IsObject(user)) {
        cJSON *firstName = cJSON_GetObjectItem(user, "first_name");
        cJSON *lastName = cJSON_GetObjectItem(user, "last_name");
        cJSON *email = cJSON_GetObjectItem(user, "email");
        cJSON *username = cJSON_GetObjectItem(user, "username");

        // Combine first and last name for YUBI_USER_NAME
        if (firstName && cJSON_IsString(firstName) && lastName && cJSON_IsString(lastName)) {
            char *fullName = malloc(strlen(firstName->valuestring) + strlen(lastName->valuestring) + 2);
            if (fullName) {
                sprintf(fullName, "%s %s", firstName->valuestring, lastName->valuestring);
                set_env_var(pamh, "YUBI_USER_NAME", fullName);
                pam_syslog(pamh, LOG_INFO, "Set YUBI_USER_NAME=%s", fullName);
                free(fullName);
            }
        } else if (firstName && cJSON_IsString(firstName)) {
            set_env_var(pamh, "YUBI_USER_NAME", firstName->valuestring);
            pam_syslog(pamh, LOG_INFO, "Set YUBI_USER_NAME=%s", firstName->valuestring);
        } else {
            pam_syslog(pamh, LOG_WARNING, "User name not found in response");
        }
        
        if (email && cJSON_IsString(email)) {
            set_env_var(pamh, "YUBI_USER_EMAIL", email->valuestring);
            pam_syslog(pamh, LOG_INFO, "Set YUBI_USER_EMAIL=%s", email->valuestring);
        } else {
            pam_syslog(pamh, LOG_WARNING, "User email not found in response");
        }
        
        if (username && cJSON_IsString(username)) {
            set_env_var(pamh, "YUBI_USER_USERNAME", username->valuestring);
            pam_syslog(pamh, LOG_INFO, "Set YUBI_USER_USERNAME=%s", username->valuestring);
        } else {
            pam_syslog(pamh, LOG_WARNING, "User username not found in response");
        }
    } else {
        pam_syslog(pamh, LOG_WARNING, "User object not found in response");
    }

    cJSON_Delete(json);
    return PAM_SUCCESS;
}

// Function to authenticate with YubiApp API
static int authenticate_with_yubiapp(pam_handle_t *pamh, const char *otp, const char *permission) {
    CURL *curl;
    CURLcode res;
    struct MemoryStruct chunk;
    int result = PAM_AUTH_ERR;

    chunk.memory = malloc(1);
    chunk.size = 0;

    curl = curl_easy_init();
    if (!curl) {
        pam_syslog(pamh, LOG_ERR, "Failed to initialize libcurl");
        free(chunk.memory);
        return PAM_SYSTEM_ERR;
    }

    // Prepare JSON request - use the format expected by the Go API
    char json_request[512];
    if (permission && strlen(permission) > 0) {
        snprintf(json_request, sizeof(json_request), 
                 "{\"device_type\":\"yubikey\",\"auth_code\":\"%s\",\"permission\":\"%s\"}", otp, permission);
    } else {
        snprintf(json_request, sizeof(json_request), 
                 "{\"device_type\":\"yubikey\",\"auth_code\":\"%s\"}", otp);
    }

    pam_syslog(pamh, LOG_INFO, "Sending request to YubiApp API: %s", json_request);

    // Set up curl options
    curl_easy_setopt(curl, CURLOPT_URL, YUBIAPP_URL);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_request);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteMemoryCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)&chunk);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);
    curl_easy_setopt(curl, CURLOPT_CONNECTTIMEOUT, 5L);

    // Set headers
    struct curl_slist *headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

    // Perform the request
    res = curl_easy_perform(curl);

    if (res == CURLE_OK) {
        long response_code;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &response_code);

        pam_syslog(pamh, LOG_INFO, "Received response from YubiApp API (HTTP %ld): %s", 
                   response_code, chunk.memory);

        if (response_code == 200) {
            // Parse response and set environment variables
            result = parse_response_and_set_env(pamh, chunk.memory);
        } else {
            pam_syslog(pamh, LOG_ERR, "HTTP error: %ld", response_code);
            result = PAM_AUTH_ERR;
        }
    } else {
        pam_syslog(pamh, LOG_ERR, "curl_easy_perform() failed: %s", curl_easy_strerror(res));
        result = PAM_SYSTEM_ERR;
    }

    // Cleanup
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);
    free(chunk.memory);

    return result;
}

// PAM authentication function
PAM_EXTERN int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    char *otp = NULL;  // Changed from const char* to char* for pam_prompt
    const char *permission = "yubiapp:authenticate";  // Default permission
    int retval = PAM_AUTH_ERR;

    (void)flags;  // Suppress unused parameter warning

    // Parse module arguments
    for (int i = 0; i < argc; i++) {
        if (strncmp(argv[i], "permission=", 11) == 0) {
            permission = argv[i] + 11;
        }
    }

    pam_syslog(pamh, LOG_INFO, "YubiApp PAM module starting authentication with permission: %s", permission);

    // Get OTP from user
    retval = pam_prompt(pamh, PAM_PROMPT_ECHO_OFF, &otp, "Yubikey OTP: ");
    if (retval != PAM_SUCCESS || !otp) {
        pam_syslog(pamh, LOG_ERR, "Failed to get OTP from user");
        return PAM_AUTH_ERR;
    }

    // Validate OTP format (basic check)
    if (strlen(otp) < 12) {
        pam_syslog(pamh, LOG_ERR, "Invalid OTP format (too short)");
        return PAM_AUTH_ERR;
    }

    // Authenticate with YubiApp API
    retval = authenticate_with_yubiapp(pamh, otp, permission);

    if (retval == PAM_SUCCESS) {
        pam_syslog(pamh, LOG_INFO, "YubiApp authentication successful");
    } else {
        pam_syslog(pamh, LOG_ERR, "YubiApp authentication failed");
    }

    return retval;
}

// PAM account management function
PAM_EXTERN int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    (void)pamh;   // Suppress unused parameter warning
    (void)flags;  // Suppress unused parameter warning
    (void)argc;   // Suppress unused parameter warning
    (void)argv;   // Suppress unused parameter warning
    return PAM_SUCCESS;
}

// PAM session management functions
PAM_EXTERN int pam_sm_open_session(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    (void)pamh;   // Suppress unused parameter warning
    (void)flags;  // Suppress unused parameter warning
    (void)argc;   // Suppress unused parameter warning
    (void)argv;   // Suppress unused parameter warning
    return PAM_SUCCESS;
}

PAM_EXTERN int pam_sm_close_session(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    (void)pamh;   // Suppress unused parameter warning
    (void)flags;  // Suppress unused parameter warning
    (void)argc;   // Suppress unused parameter warning
    (void)argv;   // Suppress unused parameter warning
    return PAM_SUCCESS;
}

// PAM password management function
PAM_EXTERN int pam_sm_chauthtok(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    (void)pamh;   // Suppress unused parameter warning
    (void)flags;  // Suppress unused parameter warning
    (void)argc;   // Suppress unused parameter warning
    (void)argv;   // Suppress unused parameter warning
    return PAM_SUCCESS;
}

// PAM setcred function
PAM_EXTERN int pam_sm_setcred(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    (void)pamh;   // Suppress unused parameter warning
    (void)flags;  // Suppress unused parameter warning
    (void)argc;   // Suppress unused parameter warning
    (void)argv;   // Suppress unused parameter warning
    return PAM_SUCCESS;
} 