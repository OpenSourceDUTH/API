package auth

import (
	"database/sql"
	"time"
)

// Role represents user permission levels
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Status represents user account status
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// Provider represents OAuth providers
type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderGitHub Provider = "github"
)

// Group represents a quota tier
type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	DefaultRPM  int       `json:"defaultRpm"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// User represents an authenticated user
type User struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        Role      `json:"role"`
	Status      Status    `json:"status"`
	GroupID     int64     `json:"groupId"`
	MaxTokens   int       `json:"maxTokens"`
	CreatedAt   time.Time `json:"createdAt"`

	// Joined fields (not always populated)
	Group *Group `json:"group,omitempty"`
}

// OAuthIdentity links a user to an OAuth provider
type OAuthIdentity struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	Provider     Provider  `json:"provider"`
	ProviderID   string    `json:"providerId"`
	AccessToken  *string   `json:"-"` // Never expose in JSON
	RefreshToken *string   `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"createdAt"`
}

// Session represents a server-side user session
type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// OAuthState represents a CSRF protection state
type OAuthState struct {
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Feature represents an API feature (hierarchical)
type Feature struct {
	ID        int64      `json:"id"`
	Slug      string     `json:"slug"`
	Name      string     `json:"name"`
	ParentID  *int64     `json:"parentId,omitempty"`
	AdminOnly bool       `json:"adminOnly"`
	CreatedAt time.Time  `json:"createdAt"`
	Children  []*Feature `json:"children,omitempty"`
}

// GroupFeatureQuota defines the default RPM for a group on a feature
type GroupFeatureQuota struct {
	GroupID   int64 `json:"groupId"`
	FeatureID int64 `json:"featureId"`
	RPMLimit  *int  `json:"rpmLimit"` // NULL = uncapped
}

// UserQuotaOverride defines a per-user RPM override on a feature
type UserQuotaOverride struct {
	UserID    int64 `json:"userId"`
	FeatureID int64 `json:"featureId"`
	RPMLimit  *int  `json:"rpmLimit"` // NULL = uncapped
}

// Token represents an API token
type Token struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"userId"`
	TokenHash    string     `json:"-"` // Never expose
	Label        string     `json:"label"`
	AdminCreated bool       `json:"adminCreated"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
	RevokedAt    *time.Time `json:"revokedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	Features     []Feature  `json:"features,omitempty"`
	AllowedIPs   []string   `json:"allowedIps,omitempty"`
}

// TokenWithRaw includes the raw token value (only returned on creation)
type TokenWithRaw struct {
	Token
	RawToken string `json:"token"`
}

// UsageLogEntry represents a single API request for rate limiting
type UsageLogEntry struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	FeatureID int64     `json:"featureId"`
	Timestamp time.Time `json:"timestamp"`
}

// AcademicDomain represents an email domain that grants academic status
type AcademicDomain struct {
	Domain string `json:"domain"`
}

// TokenCreateRequest represents the request body for creating a token
type TokenCreateRequest struct {
	Label      string     `json:"label" binding:"required"`
	Features   []string   `json:"features" binding:"required,min=1"`
	AllowedIPs []string   `json:"allowedIps"`
	ExpiresAt  *time.Time `json:"expiresAt"`
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
	Role      *Role   `json:"role"`
	Status    *Status `json:"status"`
	GroupID   *int64  `json:"groupId"`
	MaxTokens *int    `json:"maxTokens"`
}

// GroupCreateRequest represents the request body for creating a group
type GroupCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	DefaultRPM  int     `json:"defaultRpm" binding:"required,min=1"`
	Description *string `json:"description"`
}

// GroupUpdateRequest represents the request body for updating a group
type GroupUpdateRequest struct {
	Name        *string `json:"name"`
	DefaultRPM  *int    `json:"defaultRpm"`
	Description *string `json:"description"`
}

// FeatureCreateRequest represents the request body for creating a feature
type FeatureCreateRequest struct {
	Slug      string `json:"slug" binding:"required"`
	Name      string `json:"name" binding:"required"`
	ParentID  *int64 `json:"parentId"`
	AdminOnly bool   `json:"adminOnly"`
}

// FeatureUpdateRequest represents the request body for updating a feature
type FeatureUpdateRequest struct {
	Name      *string `json:"name"`
	ParentID  *int64  `json:"parentId"`
	AdminOnly *bool   `json:"adminOnly"`
}

// QuotaSetRequest represents the request body for setting quotas
type QuotaSetRequest struct {
	Quotas []QuotaEntry `json:"quotas" binding:"required"`
}

// QuotaEntry represents a single quota setting
type QuotaEntry struct {
	FeatureID int64 `json:"featureId" binding:"required"`
	RPMLimit  *int  `json:"rpmLimit"` // NULL = uncapped
}

// ValidatedToken holds the result of token validation
type ValidatedToken struct {
	Token      *Token
	User       *User
	FeatureIDs []int64
	AllowedIPs []string
}

// NullableInt64 helper for scanning nullable int64
func ScanNullableInt64(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

// NullableInt helper for scanning nullable int
func ScanNullableInt(n sql.NullInt64) *int {
	if n.Valid {
		v := int(n.Int64)
		return &v
	}
	return nil
}

// NullableString helper for scanning nullable string
func ScanNullableString(n sql.NullString) *string {
	if n.Valid {
		return &n.String
	}
	return nil
}

// NullableTime helper for scanning nullable time
func ScanNullableTime(n sql.NullTime) *time.Time {
	if n.Valid {
		return &n.Time
	}
	return nil
}
