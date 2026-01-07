package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

const (
	// OAuthStateExpiry is how long an OAuth state is valid
	OAuthStateExpiry = 10 * time.Minute
)

// OAuthStateStore manages OAuth CSRF state tokens
type OAuthStateStore struct {
	repo *Repository
}

// NewOAuthStateStore creates a new OAuth state store
func NewOAuthStateStore(repo *Repository) *OAuthStateStore {
	return &OAuthStateStore{repo: repo}
}

// CreateState generates a new random state token for CSRF protection
func (s *OAuthStateStore) CreateState() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to URL-safe base64
	state := base64.URLEncoding.EncodeToString(bytes)
	expiresAt := time.Now().Add(OAuthStateExpiry)

	// Store in database
	_, err := s.repo.db.Exec(`
		INSERT INTO oauth_states (state, expires_at) VALUES (?, ?)
	`, state, expiresAt)
	if err != nil {
		return "", err
	}

	return state, nil
}

// ValidateState checks if a state token is valid and not expired.
// The token is deleted after validation (single-use).
func (s *OAuthStateStore) ValidateState(state string) (bool, error) {
	// Try to delete the state and check if it existed and wasn't expired
	result, err := s.repo.db.Exec(`
		DELETE FROM oauth_states 
		WHERE state = ? AND expires_at > ?
	`, state, time.Now())
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// CleanupExpiredStates removes all expired state tokens
func (s *OAuthStateStore) CleanupExpiredStates() error {
	_, err := s.repo.db.Exec(`
		DELETE FROM oauth_states WHERE expires_at <= ?
	`, time.Now())
	return err
}
