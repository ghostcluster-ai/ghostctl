package auth

import (
	"testing"
)

// TestNewTokenManager tests TokenManager creation
func TestNewTokenManager(t *testing.T) {
	tm, err := NewTokenManager()

	if err != nil {
		t.Fatalf("NewTokenManager() err = %v", err)
	}

	if tm == nil {
		t.Error("NewTokenManager() returned nil")
	}

	if tm.config == nil {
		t.Error("NewTokenManager() config is nil")
	}
}

// TestGenerateToken tests token generation
func TestGenerateToken(t *testing.T) {
	tm, err := NewTokenManager()
	if err != nil {
		t.Fatalf("NewTokenManager() err = %v", err)
	}

	tests := []struct {
		name   string
		length int
	}{
		{"default length", 0},
		{"small length", 16},
		{"large length", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tm.GenerateToken(tt.length)
			if err != nil {
				t.Errorf("GenerateToken() err = %v", err)
			}

			if token == "" {
				t.Error("GenerateToken() returned empty token")
			}

			// Verify token length is at least 16 characters
			if len(token) < 16 {
				t.Errorf("GenerateToken() token too short: %d", len(token))
			}
		})
	}
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	tm, err := NewTokenManager()
	if err != nil {
		t.Fatalf("NewTokenManager() err = %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"valid token", "validtoken12345678", false},
		{"empty token", "", true},
		{"too short", "short", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
