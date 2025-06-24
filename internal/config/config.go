package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Yubikey  YubikeyConfig  `mapstructure:"yubikey"`
	SMS      SMSConfig      `mapstructure:"sms"`
	Email    EmailConfig    `mapstructure:"email"`
	Web      WebConfig      `mapstructure:"web"`
}

type ServerConfig struct {
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
	Debug   bool          `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type AuthConfig struct {
	JWTSecret           string        `mapstructure:"jwt_secret"`
	TokenExpiry         time.Duration `mapstructure:"token_expiry"`
	RefreshTokenExpiry  time.Duration `mapstructure:"refresh_token_expiry"`
}

type YubikeyConfig struct {
	ClientID  string `mapstructure:"client_id"`
	SecretKey string `mapstructure:"secret_key"`
	APIURL    string `mapstructure:"api_url"`
}

type SMSConfig struct {
	Provider    string `mapstructure:"provider"`
	AccountSID  string `mapstructure:"account_sid"`
	AuthToken   string `mapstructure:"auth_token"`
	FromNumber  string `mapstructure:"from_number"`
}

type EmailConfig struct {
	SMTPHost   string `mapstructure:"smtp_host"`
	SMTPPort   int    `mapstructure:"smtp_port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	FromEmail  string `mapstructure:"from_email"`
}

type WebConfig struct {
	SessionSecret string   `mapstructure:"session_secret"`
	CORSOrigins   []string `mapstructure:"cors_origins"`
}

// Load reads the configuration from config.yaml file
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.timeout", "30s")
	viper.SetDefault("server.debug", false)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")

	viper.SetDefault("auth.token_expiry", "24h")
	viper.SetDefault("auth.refresh_token_expiry", "720h")

	viper.SetDefault("yubikey.api_url", "https://api.yubico.com/wsapi/2.0/verify")

	viper.SetDefault("email.smtp_port", 587)
} 