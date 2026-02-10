package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/config"
)

// TokenManager handles authentication tokens
type TokenManager struct {
	config *config.Config
}

// NewTokenManager creates a new token manager
func NewTokenManager() (*TokenManager, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &TokenManager{
		config: cfg,
	}, nil
}

// GetToken retrieves the current auth token
func (tm *TokenManager) GetToken() (string, error) {
	if tm.config.AuthToken == "" {
		return "", fmt.Errorf("no auth token configured")
	}
	return tm.config.AuthToken, nil
}

// SetToken sets the auth token
func (tm *TokenManager) SetToken(token string) error {
	tm.config.AuthToken = token
	return tm.config.Save()
}

// GenerateToken generates a new authentication token
func (tm *TokenManager) GenerateToken(length int) (string, error) {
	if length < 16 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	token := base64.StdEncoding.EncodeToString(bytes)
	return token, nil
}

// SaveToken saves token to file with restricted permissions
func (tm *TokenManager) SaveToken(token string) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(filepath.Dir(configPath), ".token")

	// Save token with restricted permissions (readable only by owner)
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// LoadToken loads token from file
func (tm *TokenManager) LoadToken() (string, error) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return "", err
	}

	tokenPath := filepath.Join(filepath.Dir(configPath), ".token")

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to load token: %w", err)
	}

	return string(data), nil
}

// ValidateToken validates a token
func (tm *TokenManager) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}

	if len(token) < 16 {
		return fmt.Errorf("token too short")
	}

	// Additional validation could be performed here
	// (e.g., checking token expiration, format, signature)

	return nil
}

// TokenCache represents a cached token with metadata
type TokenCache struct {
	Token     string
	ExpiresAt time.Time
	Metadata  map[string]string
}

// CachedTokenManager extends TokenManager with caching
type CachedTokenManager struct {
	*TokenManager
	cache *TokenCache
}

// NewCachedTokenManager creates a token manager with caching
func NewCachedTokenManager() (*CachedTokenManager, error) {
	tm, err := NewTokenManager()
	if err != nil {
		return nil, err
	}

	return &CachedTokenManager{
		TokenManager: tm,
		cache:        nil,
	}, nil
}

// GetCachedToken gets a token from cache if valid
func (ctm *CachedTokenManager) GetCachedToken() (string, error) {
	if ctm.cache != nil && time.Now().Before(ctm.cache.ExpiresAt) {
		return ctm.cache.Token, nil
	}

	// Cache expired or not set, fetch fresh token
	token, err := ctm.GetToken()
	if err != nil {
		return "", err
	}

	// Cache token for 1 hour
	ctm.cache = &TokenCache{
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Metadata:  make(map[string]string),
	}

	return token, nil
}

// InvalidateCache invalidates the token cache
func (ctm *CachedTokenManager) InvalidateCache() {
	ctm.cache = nil
}
