package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YubiApp/internal/config"
	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/golang-jwt/jwt/v5"
)

type SessionService struct {
	redisClient *redis.Client
	config      *config.Config
}

func NewSessionService(config *config.Config) *SessionService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
		PoolSize: config.Redis.PoolSize,
	})

	return &SessionService{
		redisClient: rdb,
		config:      config,
	}
}

// CreateSession creates a new session for a user and device
func (s *SessionService) CreateSession(userID, deviceID uuid.UUID) (*database.Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(s.config.Auth.SessionExpiry)

	session := &database.Session{
		ID:           sessionID,
		UserID:       userID,
		DeviceID:     deviceID,
		AccessCount:  0,
		RefreshCount: 0,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		IsValid:      true,
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	ctx := context.Background()
	err = s.redisClient.Set(ctx, sessionKey, sessionData, s.config.Auth.SessionExpiry).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store session in Redis: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session from Redis
func (s *SessionService) GetSession(sessionID string) (*database.Session, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	
	ctx := context.Background()
	sessionData, err := s.redisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var session database.Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if !session.IsValid {
		return nil, fmt.Errorf("session is invalid")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	return &session, nil
}

// UpdateSession updates a session in Redis
func (s *SessionService) UpdateSession(session *database.Session) error {
	sessionKey := fmt.Sprintf("session:%s", session.ID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ctx := context.Background()
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session has expired")
	}

	err = s.redisClient.Set(ctx, sessionKey, sessionData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update session in Redis: %w", err)
	}

	return nil
}

// InvalidateSession marks a session as invalid
func (s *SessionService) InvalidateSession(sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.IsValid = false
	return s.UpdateSession(session)
}

// GenerateAccessToken generates a JWT access token for a session
func (s *SessionService) GenerateAccessToken(session *database.Session) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.Auth.AccessTokenExpiry)

	claims := database.SessionToken{
		SessionID:    session.ID,
		UserID:       session.UserID.String(),
		DeviceID:     session.DeviceID.String(),
		AccessCount:  session.AccessCount,
		RefreshCount: session.RefreshCount,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "yubiapp",
			Subject:   session.UserID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Auth.JWTSecret))
}

// GenerateRefreshToken generates a JWT refresh token for a session
func (s *SessionService) GenerateRefreshToken(session *database.Session) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.Auth.SessionExpiry)

	claims := database.RefreshToken{
		SessionID:    session.ID,
		UserID:       session.UserID.String(),
		DeviceID:     session.DeviceID.String(),
		RefreshCount: session.RefreshCount,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "yubiapp",
			Subject:   session.UserID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Auth.JWTSecret))
}

// ValidateAccessToken validates and parses an access token
func (s *SessionService) ValidateAccessToken(tokenString string) (*database.SessionToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &database.SessionToken{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*database.SessionToken); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates and parses a refresh token
func (s *SessionService) ValidateRefreshToken(tokenString string) (*database.RefreshToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &database.RefreshToken{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*database.RefreshToken); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshSession creates new access and refresh tokens for an existing session
func (s *SessionService) RefreshSession(refreshTokenString string) (*database.Session, string, string, error) {
	// Validate refresh token
	refreshClaims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get session from Redis
	session, err := s.GetSession(refreshClaims.SessionID)
	if err != nil {
		return nil, "", "", fmt.Errorf("session not found: %w", err)
	}

	// Verify refresh count matches
	if session.RefreshCount != refreshClaims.RefreshCount {
		return nil, "", "", fmt.Errorf("refresh token is invalid (count mismatch)")
	}

	// Increment refresh count and update session
	session.RefreshCount++
	err = s.UpdateSession(session)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to update session: %w", err)
	}

	// Generate new tokens
	accessToken, err := s.GenerateAccessToken(session)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.GenerateRefreshToken(session)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return session, accessToken, newRefreshToken, nil
}

// Close closes the Redis connection
func (s *SessionService) Close() error {
	return s.redisClient.Close()
} 