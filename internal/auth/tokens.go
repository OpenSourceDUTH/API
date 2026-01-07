package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/mr-tron/base58"
)

const (
	// TokenPrefix is the prefix for all generated tokens
	TokenPrefix = "osduth_"
)

// TokenStore manages API token operations
type TokenStore struct {
	repo     *Repository
	features *FeatureRegistry
}

// NewTokenStore creates a new token store
func NewTokenStore(repo *Repository, features *FeatureRegistry) *TokenStore {
	return &TokenStore{
		repo:     repo,
		features: features,
	}
}

// GenerateToken creates a new random token with the osduth_ prefix
// Format: osduth_ + Base58(SHA256(random_bytes))
func (s *TokenStore) GenerateToken() (rawToken string, tokenHash string, err error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	// SHA256 the random bytes
	hash := sha256.Sum256(randomBytes)

	// Base58 encode the hash
	encoded := base58.Encode(hash[:])

	// Create raw token with prefix
	rawToken = TokenPrefix + encoded

	// Hash the raw token for storage
	tokenHash = hashToken(rawToken)

	return rawToken, tokenHash, nil
}

// hashToken creates a SHA256 hash of a token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateUserToken creates a token for a user with the given parameters
// This enforces max_tokens limit and rejects admin-only features
func (s *TokenStore) CreateUserToken(userID int64, label string, featureSlugs []string, allowedIPs []string, expiresAt *time.Time) (*TokenWithRaw, error) {
	// Validate label
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("token label is required")
	}

	// Check token limit
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	count, err := s.repo.GetUserTokenCount(userID)
	if err != nil {
		return nil, err
	}
	if count >= user.MaxTokens {
		return nil, fmt.Errorf("maximum token limit (%d) reached", user.MaxTokens)
	}

	// Validate features exist and are not admin-only
	features, err := s.features.GetFeaturesBySlugs(featureSlugs)
	if err != nil {
		return nil, err
	}
	if len(features) == 0 {
		return nil, fmt.Errorf("at least one valid feature is required")
	}
	if len(features) != len(featureSlugs) {
		return nil, fmt.Errorf("one or more features not found")
	}

	// Check for admin-only features
	for _, f := range features {
		if f.AdminOnly {
			return nil, fmt.Errorf("feature '%s' is admin-only and cannot be assigned by users", f.Slug)
		}
	}

	// Canonicalize IPs
	canonicalIPs, err := CanonicalizeIPs(allowedIPs)
	if err != nil {
		return nil, err
	}

	// Generate token
	rawToken, tokenHash, err := s.GenerateToken()
	if err != nil {
		return nil, err
	}

	// Create token in database
	return s.createToken(userID, tokenHash, label, false, expiresAt, features, canonicalIPs, rawToken)
}

// CreateAdminToken creates a token without restrictions (admin use)
func (s *TokenStore) CreateAdminToken(userID int64, label string, featureSlugs []string, allowedIPs []string, expiresAt *time.Time) (*TokenWithRaw, error) {
	// Validate label
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("token label is required")
	}

	// Validate features exist
	features, err := s.features.GetFeaturesBySlugs(featureSlugs)
	if err != nil {
		return nil, err
	}
	if len(features) == 0 {
		return nil, fmt.Errorf("at least one valid feature is required")
	}
	if len(features) != len(featureSlugs) {
		return nil, fmt.Errorf("one or more features not found")
	}

	// Canonicalize IPs
	canonicalIPs, err := CanonicalizeIPs(allowedIPs)
	if err != nil {
		return nil, err
	}

	// Generate token
	rawToken, tokenHash, err := s.GenerateToken()
	if err != nil {
		return nil, err
	}

	// Create token in database
	return s.createToken(userID, tokenHash, label, true, expiresAt, features, canonicalIPs, rawToken)
}

func (s *TokenStore) createToken(userID int64, tokenHash, label string, adminCreated bool, expiresAt *time.Time, features []Feature, allowedIPs []string, rawToken string) (*TokenWithRaw, error) {
	tx, err := s.repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert token
	result, err := tx.Exec(`
		INSERT INTO tokens (user_id, token_hash, label, admin_created, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, userID, tokenHash, label, adminCreated, expiresAt)
	if err != nil {
		return nil, err
	}

	tokenID, _ := result.LastInsertId()

	// Insert feature associations
	for _, f := range features {
		if _, err := tx.Exec(`
			INSERT INTO token_features (token_id, feature_id) VALUES (?, ?)
		`, tokenID, f.ID); err != nil {
			return nil, err
		}
	}

	// Insert allowed IPs
	for _, ip := range allowedIPs {
		if _, err := tx.Exec(`
			INSERT INTO token_allowed_ips (token_id, ip_address) VALUES (?, ?)
		`, tokenID, ip); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Build response
	token := &TokenWithRaw{
		Token: Token{
			ID:           tokenID,
			UserID:       userID,
			Label:        label,
			AdminCreated: adminCreated,
			ExpiresAt:    expiresAt,
			CreatedAt:    time.Now(),
			Features:     features,
			AllowedIPs:   allowedIPs,
		},
		RawToken: rawToken,
	}

	return token, nil
}

// ValidateToken validates a raw token and returns the token with user info
func (s *TokenStore) ValidateToken(rawToken string) (*ValidatedToken, error) {
	// Check prefix
	if !strings.HasPrefix(rawToken, TokenPrefix) {
		return nil, fmt.Errorf("invalid token format")
	}

	// Hash the token for lookup
	tokenHash := hashToken(rawToken)

	// Look up token
	var t Token
	var expiresAt, revokedAt sql.NullTime
	err := s.repo.db.QueryRow(`
		SELECT id, user_id, token_hash, label, admin_created, expires_at, revoked_at, created_at
		FROM tokens WHERE token_hash = ?
	`, tokenHash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.Label, &t.AdminCreated, &expiresAt, &revokedAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid token")
	}
	if err != nil {
		return nil, err
	}

	t.ExpiresAt = ScanNullableTime(expiresAt)
	t.RevokedAt = ScanNullableTime(revokedAt)

	// Check if revoked
	if t.RevokedAt != nil {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Check if expired
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Get user
	user, err := s.repo.GetUserByID(t.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check user status
	if user.Status != StatusActive {
		return nil, fmt.Errorf("user account is %s", user.Status)
	}

	// Get feature IDs
	featureIDs, err := s.getTokenFeatureIDs(t.ID)
	if err != nil {
		return nil, err
	}

	// Get allowed IPs
	allowedIPs, err := s.getTokenAllowedIPs(t.ID)
	if err != nil {
		return nil, err
	}

	return &ValidatedToken{
		Token:      &t,
		User:       user,
		FeatureIDs: featureIDs,
		AllowedIPs: allowedIPs,
	}, nil
}

func (s *TokenStore) getTokenFeatureIDs(tokenID int64) ([]int64, error) {
	rows, err := s.repo.db.Query(`
		SELECT feature_id FROM token_features WHERE token_id = ?
	`, tokenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *TokenStore) getTokenAllowedIPs(tokenID int64) ([]string, error) {
	rows, err := s.repo.db.Query(`
		SELECT ip_address FROM token_allowed_ips WHERE token_id = ?
	`, tokenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// ListUserTokens returns all tokens for a user (without raw values)
func (s *TokenStore) ListUserTokens(userID int64) ([]Token, error) {
	rows, err := s.repo.db.Query(`
		SELECT id, user_id, label, admin_created, expires_at, revoked_at, created_at
		FROM tokens WHERE user_id = ? ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		var expiresAt, revokedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.Label, &t.AdminCreated, &expiresAt, &revokedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.ExpiresAt = ScanNullableTime(expiresAt)
		t.RevokedAt = ScanNullableTime(revokedAt)

		// Get features
		featureIDs, err := s.getTokenFeatureIDs(t.ID)
		if err != nil {
			return nil, err
		}
		features, err := s.features.GetFeaturesByIDs(featureIDs)
		if err != nil {
			return nil, err
		}
		t.Features = features

		// Get allowed IPs
		t.AllowedIPs, err = s.getTokenAllowedIPs(t.ID)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// GetTokenByID returns a token by ID
func (s *TokenStore) GetTokenByID(tokenID int64) (*Token, error) {
	var t Token
	var expiresAt, revokedAt sql.NullTime
	err := s.repo.db.QueryRow(`
		SELECT id, user_id, label, admin_created, expires_at, revoked_at, created_at
		FROM tokens WHERE id = ?
	`, tokenID).Scan(&t.ID, &t.UserID, &t.Label, &t.AdminCreated, &expiresAt, &revokedAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.ExpiresAt = ScanNullableTime(expiresAt)
	t.RevokedAt = ScanNullableTime(revokedAt)

	// Get features
	featureIDs, err := s.getTokenFeatureIDs(t.ID)
	if err != nil {
		return nil, err
	}
	features, err := s.features.GetFeaturesByIDs(featureIDs)
	if err != nil {
		return nil, err
	}
	t.Features = features

	// Get allowed IPs
	t.AllowedIPs, err = s.getTokenAllowedIPs(t.ID)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// RevokeToken revokes a token (user can only revoke their own tokens)
func (s *TokenStore) RevokeToken(tokenID int64, userID int64) error {
	result, err := s.repo.db.Exec(`
		UPDATE tokens SET revoked_at = ? 
		WHERE id = ? AND user_id = ? AND revoked_at IS NULL
	`, time.Now(), tokenID, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("token not found or already revoked")
	}
	return nil
}

// AdminRevokeToken revokes any token (admin use)
func (s *TokenStore) AdminRevokeToken(tokenID int64) error {
	result, err := s.repo.db.Exec(`
		UPDATE tokens SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL
	`, time.Now(), tokenID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("token not found or already revoked")
	}
	return nil
}
