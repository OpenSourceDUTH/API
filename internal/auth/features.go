package auth

import (
	"database/sql"
)

// FeatureRegistry manages API features with live database queries
type FeatureRegistry struct {
	repo *Repository
}

// NewFeatureRegistry creates a new feature registry
func NewFeatureRegistry(repo *Repository) *FeatureRegistry {
	return &FeatureRegistry{repo: repo}
}

// GetFeatureBySlug returns a feature by its slug with a live database query
func (r *FeatureRegistry) GetFeatureBySlug(slug string) (*Feature, error) {
	var f Feature
	var parentID sql.NullInt64
	err := r.repo.db.QueryRow(`
		SELECT id, slug, name, parent_id, admin_only, created_at
		FROM features WHERE slug = ?
	`, slug).Scan(&f.ID, &f.Slug, &f.Name, &parentID, &f.AdminOnly, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.ParentID = ScanNullableInt64(parentID)
	return &f, nil
}

// GetFeatureByID returns a feature by its ID
func (r *FeatureRegistry) GetFeatureByID(id int64) (*Feature, error) {
	var f Feature
	var parentID sql.NullInt64
	err := r.repo.db.QueryRow(`
		SELECT id, slug, name, parent_id, admin_only, created_at
		FROM features WHERE id = ?
	`, id).Scan(&f.ID, &f.Slug, &f.Name, &parentID, &f.AdminOnly, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.ParentID = ScanNullableInt64(parentID)
	return &f, nil
}

// IsFeatureAdminOnly checks if a feature is admin-only (live query)
func (r *FeatureRegistry) IsFeatureAdminOnly(featureID int64) (bool, error) {
	var adminOnly bool
	err := r.repo.db.QueryRow(`
		SELECT admin_only FROM features WHERE id = ?
	`, featureID).Scan(&adminOnly)
	if err != nil {
		return false, err
	}
	return adminOnly, nil
}

// IsFeatureSlugAdminOnly checks if a feature slug is admin-only (live query)
func (r *FeatureRegistry) IsFeatureSlugAdminOnly(slug string) (bool, error) {
	var adminOnly bool
	err := r.repo.db.QueryRow(`
		SELECT admin_only FROM features WHERE slug = ?
	`, slug).Scan(&adminOnly)
	if err != nil {
		return false, err
	}
	return adminOnly, nil
}

// GetAllFeatures returns all features (for admins)
func (r *FeatureRegistry) GetAllFeatures() ([]Feature, error) {
	rows, err := r.repo.db.Query(`
		SELECT id, slug, name, parent_id, admin_only, created_at
		FROM features ORDER BY slug
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []Feature
	for rows.Next() {
		var f Feature
		var parentID sql.NullInt64
		if err := rows.Scan(&f.ID, &f.Slug, &f.Name, &parentID, &f.AdminOnly, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.ParentID = ScanNullableInt64(parentID)
		features = append(features, f)
	}
	return features, rows.Err()
}

// GetUserAssignableFeatures returns features that users can assign to their tokens
func (r *FeatureRegistry) GetUserAssignableFeatures() ([]Feature, error) {
	rows, err := r.repo.db.Query(`
		SELECT id, slug, name, parent_id, admin_only, created_at
		FROM features WHERE admin_only = 0 ORDER BY slug
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []Feature
	for rows.Next() {
		var f Feature
		var parentID sql.NullInt64
		if err := rows.Scan(&f.ID, &f.Slug, &f.Name, &parentID, &f.AdminOnly, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.ParentID = ScanNullableInt64(parentID)
		features = append(features, f)
	}
	return features, rows.Err()
}

// GetFeaturesByIDs returns features by their IDs
func (r *FeatureRegistry) GetFeaturesByIDs(ids []int64) ([]Feature, error) {
	if len(ids) == 0 {
		return []Feature{}, nil
	}

	// Build query with placeholders
	query := "SELECT id, slug, name, parent_id, admin_only, created_at FROM features WHERE id IN ("
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ") ORDER BY slug"

	rows, err := r.repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []Feature
	for rows.Next() {
		var f Feature
		var parentID sql.NullInt64
		if err := rows.Scan(&f.ID, &f.Slug, &f.Name, &parentID, &f.AdminOnly, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.ParentID = ScanNullableInt64(parentID)
		features = append(features, f)
	}
	return features, rows.Err()
}

// GetFeaturesBySlugs returns features by their slugs
func (r *FeatureRegistry) GetFeaturesBySlugs(slugs []string) ([]Feature, error) {
	if len(slugs) == 0 {
		return []Feature{}, nil
	}

	query := "SELECT id, slug, name, parent_id, admin_only, created_at FROM features WHERE slug IN ("
	args := make([]interface{}, len(slugs))
	for i, slug := range slugs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = slug
	}
	query += ") ORDER BY slug"

	rows, err := r.repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []Feature
	for rows.Next() {
		var f Feature
		var parentID sql.NullInt64
		if err := rows.Scan(&f.ID, &f.Slug, &f.Name, &parentID, &f.AdminOnly, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.ParentID = ScanNullableInt64(parentID)
		features = append(features, f)
	}
	return features, rows.Err()
}

// GetFeatureAncestors returns a feature and all its ancestors (for quota inheritance)
func (r *FeatureRegistry) GetFeatureAncestors(featureID int64) ([]Feature, error) {
	var ancestors []Feature

	currentID := &featureID
	for currentID != nil {
		feature, err := r.GetFeatureByID(*currentID)
		if err != nil {
			return nil, err
		}
		if feature == nil {
			break
		}
		ancestors = append(ancestors, *feature)
		currentID = feature.ParentID
	}

	return ancestors, nil
}

// CreateFeature creates a new feature
func (r *FeatureRegistry) CreateFeature(slug, name string, parentID *int64, adminOnly bool) (*Feature, error) {
	result, err := r.repo.db.Exec(`
		INSERT INTO features (slug, name, parent_id, admin_only) VALUES (?, ?, ?, ?)
	`, slug, name, parentID, adminOnly)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.GetFeatureByID(id)
}

// UpdateFeature updates a feature
func (r *FeatureRegistry) UpdateFeature(id int64, name *string, parentID *int64, adminOnly *bool) error {
	if name != nil {
		if _, err := r.repo.db.Exec("UPDATE features SET name = ? WHERE id = ?", *name, id); err != nil {
			return err
		}
	}
	if parentID != nil {
		if _, err := r.repo.db.Exec("UPDATE features SET parent_id = ? WHERE id = ?", *parentID, id); err != nil {
			return err
		}
	}
	if adminOnly != nil {
		if _, err := r.repo.db.Exec("UPDATE features SET admin_only = ? WHERE id = ?", *adminOnly, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteFeature deletes a feature
func (r *FeatureRegistry) DeleteFeature(id int64) error {
	_, err := r.repo.db.Exec("DELETE FROM features WHERE id = ?", id)
	return err
}

// HasAdminOnlyFeatures checks if any of the given feature IDs are admin-only
func (r *FeatureRegistry) HasAdminOnlyFeatures(featureIDs []int64) (bool, error) {
	if len(featureIDs) == 0 {
		return false, nil
	}

	query := "SELECT COUNT(*) FROM features WHERE admin_only = 1 AND id IN ("
	args := make([]interface{}, len(featureIDs))
	for i, id := range featureIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	var count int
	err := r.repo.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TokenHasFeatureAccess checks if a token has access to a feature
// This includes checking both direct feature assignment and parent features
func (r *FeatureRegistry) TokenHasFeatureAccess(tokenFeatureIDs []int64, targetFeatureSlug string) (bool, error) {
	// Get the target feature
	targetFeature, err := r.GetFeatureBySlug(targetFeatureSlug)
	if err != nil || targetFeature == nil {
		return false, err
	}

	// Check if the token has direct access to this feature
	for _, id := range tokenFeatureIDs {
		if id == targetFeature.ID {
			return true, nil
		}
	}

	// Check if the token has access to any ancestor of this feature
	// (having access to "maps" grants access to "maps.tiles")
	ancestors, err := r.GetFeatureAncestors(targetFeature.ID)
	if err != nil {
		return false, err
	}

	for _, ancestor := range ancestors {
		for _, tokenFeatureID := range tokenFeatureIDs {
			if tokenFeatureID == ancestor.ID {
				return true, nil
			}
		}
	}

	return false, nil
}
