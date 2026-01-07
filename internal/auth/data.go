package auth

import (
	"database/sql"
)

// Repository provides access to auth-related database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying database connection
func (r *Repository) DB() *sql.DB {
	return r.db
}

// EnableWAL enables Write-Ahead Logging mode for better concurrent performance
func (r *Repository) EnableWAL() error {
	_, err := r.db.Exec("PRAGMA journal_mode=WAL")
	return err
}

// --- Group Operations ---

// GetAllGroups returns all groups
func (r *Repository) GetAllGroups() ([]Group, error) {
	rows, err := r.db.Query(`
		SELECT id, name, default_rpm, description, created_at 
		FROM groups 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		var desc sql.NullString
		if err := rows.Scan(&g.ID, &g.Name, &g.DefaultRPM, &desc, &g.CreatedAt); err != nil {
			return nil, err
		}
		g.Description = ScanNullableString(desc)
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// GetGroupByID returns a group by ID
func (r *Repository) GetGroupByID(id int64) (*Group, error) {
	var g Group
	var desc sql.NullString
	err := r.db.QueryRow(`
		SELECT id, name, default_rpm, description, created_at 
		FROM groups WHERE id = ?
	`, id).Scan(&g.ID, &g.Name, &g.DefaultRPM, &desc, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g.Description = ScanNullableString(desc)
	return &g, nil
}

// GetGroupByName returns a group by name
func (r *Repository) GetGroupByName(name string) (*Group, error) {
	var g Group
	var desc sql.NullString
	err := r.db.QueryRow(`
		SELECT id, name, default_rpm, description, created_at 
		FROM groups WHERE name = ?
	`, name).Scan(&g.ID, &g.Name, &g.DefaultRPM, &desc, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g.Description = ScanNullableString(desc)
	return &g, nil
}

// CreateGroup creates a new group
func (r *Repository) CreateGroup(name string, defaultRPM int, description *string) (*Group, error) {
	result, err := r.db.Exec(`
		INSERT INTO groups (name, default_rpm, description) VALUES (?, ?, ?)
	`, name, defaultRPM, description)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.GetGroupByID(id)
}

// UpdateGroup updates a group
func (r *Repository) UpdateGroup(id int64, name *string, defaultRPM *int, description *string) error {
	if name != nil {
		if _, err := r.db.Exec("UPDATE groups SET name = ? WHERE id = ?", *name, id); err != nil {
			return err
		}
	}
	if defaultRPM != nil {
		if _, err := r.db.Exec("UPDATE groups SET default_rpm = ? WHERE id = ?", *defaultRPM, id); err != nil {
			return err
		}
	}
	if description != nil {
		if _, err := r.db.Exec("UPDATE groups SET description = ? WHERE id = ?", *description, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteGroup deletes a group by ID
func (r *Repository) DeleteGroup(id int64) error {
	_, err := r.db.Exec("DELETE FROM groups WHERE id = ?", id)
	return err
}

// --- Academic Domain Operations ---

// GetAllAcademicDomains returns all academic domains
func (r *Repository) GetAllAcademicDomains() ([]string, error) {
	rows, err := r.db.Query("SELECT domain FROM academic_domains ORDER BY domain")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, rows.Err()
}

// IsAcademicDomain checks if a domain grants academic status
func (r *Repository) IsAcademicDomain(domain string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM academic_domains WHERE domain = ?", domain).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddAcademicDomain adds a new academic domain
func (r *Repository) AddAcademicDomain(domain string) error {
	_, err := r.db.Exec("INSERT OR IGNORE INTO academic_domains (domain) VALUES (?)", domain)
	return err
}

// RemoveAcademicDomain removes an academic domain
func (r *Repository) RemoveAcademicDomain(domain string) error {
	_, err := r.db.Exec("DELETE FROM academic_domains WHERE domain = ?", domain)
	return err
}

// --- User Operations ---

// GetUserByID returns a user by ID with group info
func (r *Repository) GetUserByID(id int64) (*User, error) {
	var u User
	var g Group
	var groupDesc sql.NullString
	err := r.db.QueryRow(`
		SELECT u.id, u.email, u.display_name, u.role, u.status, u.group_id, u.max_tokens, u.created_at,
		       g.id, g.name, g.default_rpm, g.description, g.created_at
		FROM users u
		JOIN groups g ON u.group_id = g.id
		WHERE u.id = ?
	`, id).Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.Status, &u.GroupID, &u.MaxTokens, &u.CreatedAt,
		&g.ID, &g.Name, &g.DefaultRPM, &groupDesc, &g.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g.Description = ScanNullableString(groupDesc)
	u.Group = &g
	return &u, nil
}

// GetUserByEmail returns a user by email
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var u User
	err := r.db.QueryRow(`
		SELECT id, email, display_name, role, status, group_id, max_tokens, created_at
		FROM users WHERE email = ?
	`, email).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.Status, &u.GroupID, &u.MaxTokens, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAllUsers returns all users with pagination
func (r *Repository) GetAllUsers(limit, offset int) ([]User, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.email, u.display_name, u.role, u.status, u.group_id, u.max_tokens, u.created_at,
		       g.id, g.name, g.default_rpm, g.description, g.created_at
		FROM users u
		JOIN groups g ON u.group_id = g.id
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var g Group
		var groupDesc sql.NullString
		if err := rows.Scan(
			&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.Status, &u.GroupID, &u.MaxTokens, &u.CreatedAt,
			&g.ID, &g.Name, &g.DefaultRPM, &groupDesc, &g.CreatedAt,
		); err != nil {
			return nil, err
		}
		g.Description = ScanNullableString(groupDesc)
		u.Group = &g
		users = append(users, u)
	}
	return users, rows.Err()
}

// CreateUser creates a new user
func (r *Repository) CreateUser(email, displayName string, groupID int64) (*User, error) {
	result, err := r.db.Exec(`
		INSERT INTO users (email, display_name, group_id) VALUES (?, ?, ?)
	`, email, displayName, groupID)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.GetUserByID(id)
}

// UpdateUser updates user fields
func (r *Repository) UpdateUser(id int64, role *Role, status *Status, groupID *int64, maxTokens *int) error {
	if role != nil {
		if _, err := r.db.Exec("UPDATE users SET role = ? WHERE id = ?", *role, id); err != nil {
			return err
		}
	}
	if status != nil {
		if _, err := r.db.Exec("UPDATE users SET status = ? WHERE id = ?", *status, id); err != nil {
			return err
		}
	}
	if groupID != nil {
		if _, err := r.db.Exec("UPDATE users SET group_id = ? WHERE id = ?", *groupID, id); err != nil {
			return err
		}
	}
	if maxTokens != nil {
		if _, err := r.db.Exec("UPDATE users SET max_tokens = ? WHERE id = ?", *maxTokens, id); err != nil {
			return err
		}
	}
	return nil
}

// GetUserTokenCount returns the number of active tokens for a user
func (r *Repository) GetUserTokenCount(userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM tokens 
		WHERE user_id = ? AND revoked_at IS NULL
	`, userID).Scan(&count)
	return count, err
}

// --- OAuth Identity Operations ---

// GetOAuthIdentity returns an OAuth identity by provider and provider ID
func (r *Repository) GetOAuthIdentity(provider Provider, providerID string) (*OAuthIdentity, error) {
	var o OAuthIdentity
	var accessToken, refreshToken sql.NullString
	err := r.db.QueryRow(`
		SELECT id, user_id, provider, provider_id, access_token, refresh_token, created_at
		FROM oauth_identities
		WHERE provider = ? AND provider_id = ?
	`, provider, providerID).Scan(&o.ID, &o.UserID, &o.Provider, &o.ProviderID, &accessToken, &refreshToken, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.AccessToken = ScanNullableString(accessToken)
	o.RefreshToken = ScanNullableString(refreshToken)
	return &o, nil
}

// CreateOAuthIdentity creates a new OAuth identity
func (r *Repository) CreateOAuthIdentity(userID int64, provider Provider, providerID, accessToken, refreshToken string) (*OAuthIdentity, error) {
	result, err := r.db.Exec(`
		INSERT INTO oauth_identities (user_id, provider, provider_id, access_token, refresh_token)
		VALUES (?, ?, ?, ?, ?)
	`, userID, provider, providerID, accessToken, refreshToken)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()

	var o OAuthIdentity
	var at, rt sql.NullString
	err = r.db.QueryRow(`
		SELECT id, user_id, provider, provider_id, access_token, refresh_token, created_at
		FROM oauth_identities WHERE id = ?
	`, id).Scan(&o.ID, &o.UserID, &o.Provider, &o.ProviderID, &at, &rt, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	o.AccessToken = ScanNullableString(at)
	o.RefreshToken = ScanNullableString(rt)
	return &o, nil
}

// UpdateOAuthIdentityTokens updates the tokens for an OAuth identity
func (r *Repository) UpdateOAuthIdentityTokens(id int64, accessToken, refreshToken string) error {
	_, err := r.db.Exec(`
		UPDATE oauth_identities SET access_token = ?, refresh_token = ? WHERE id = ?
	`, accessToken, refreshToken, id)
	return err
}
