package auth

import (
	"database/sql"
)

const (
	// DefaultSystemRPM is the default RPM when no quota is defined
	DefaultSystemRPM = 60

	// UnlimitedRPM indicates no rate limit
	UnlimitedRPM = -1
)

// QuotaEngine calculates effective rate limits for users
type QuotaEngine struct {
	repo     *Repository
	features *FeatureRegistry
}

// NewQuotaEngine creates a new quota engine
func NewQuotaEngine(repo *Repository, features *FeatureRegistry) *QuotaEngine {
	return &QuotaEngine{
		repo:     repo,
		features: features,
	}
}

// GetEffectiveRPM returns the effective RPM limit for a user on a feature.
// Priority: user override > group quota > parent feature quota > system default
// Returns UnlimitedRPM (-1) if the quota is uncapped (NULL in database)
func (q *QuotaEngine) GetEffectiveRPM(userID int64, featureID int64) (int, error) {
	// 1. Check user override for this feature
	rpm, found, err := q.getUserOverride(userID, featureID)
	if err != nil {
		return 0, err
	}
	if found {
		return rpm, nil
	}

	// 2. Get user's group
	user, err := q.repo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return DefaultSystemRPM, nil
	}

	// 3. Get feature ancestry (including the feature itself)
	ancestors, err := q.features.GetFeatureAncestors(featureID)
	if err != nil {
		return 0, err
	}

	// 4. Check group quota for each feature in the ancestry (starting from most specific)
	for _, feature := range ancestors {
		rpm, found, err := q.getGroupQuota(user.GroupID, feature.ID)
		if err != nil {
			return 0, err
		}
		if found {
			return rpm, nil
		}
	}

	// 5. Fall back to group's default RPM
	if user.Group != nil {
		return user.Group.DefaultRPM, nil
	}

	// 6. Fall back to system default
	return DefaultSystemRPM, nil
}

// GetEffectiveRPMBySlug is a convenience method that looks up the feature by slug
func (q *QuotaEngine) GetEffectiveRPMBySlug(userID int64, featureSlug string) (int, error) {
	feature, err := q.features.GetFeatureBySlug(featureSlug)
	if err != nil {
		return 0, err
	}
	if feature == nil {
		return DefaultSystemRPM, nil
	}
	return q.GetEffectiveRPM(userID, feature.ID)
}

func (q *QuotaEngine) getUserOverride(userID int64, featureID int64) (rpm int, found bool, err error) {
	var rpmLimit sql.NullInt64
	err = q.repo.db.QueryRow(`
		SELECT rpm_limit FROM user_quota_overrides
		WHERE user_id = ? AND feature_id = ?
	`, userID, featureID).Scan(&rpmLimit)

	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	// NULL means uncapped
	if !rpmLimit.Valid {
		return UnlimitedRPM, true, nil
	}
	return int(rpmLimit.Int64), true, nil
}

func (q *QuotaEngine) getGroupQuota(groupID int64, featureID int64) (rpm int, found bool, err error) {
	var rpmLimit sql.NullInt64
	err = q.repo.db.QueryRow(`
		SELECT rpm_limit FROM group_feature_quotas
		WHERE group_id = ? AND feature_id = ?
	`, groupID, featureID).Scan(&rpmLimit)

	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	// NULL means uncapped
	if !rpmLimit.Valid {
		return UnlimitedRPM, true, nil
	}
	return int(rpmLimit.Int64), true, nil
}

// SetUserQuotaOverride sets a quota override for a user on a feature
// Pass nil for rpmLimit to set uncapped (unlimited)
func (q *QuotaEngine) SetUserQuotaOverride(userID int64, featureID int64, rpmLimit *int) error {
	_, err := q.repo.db.Exec(`
		INSERT INTO user_quota_overrides (user_id, feature_id, rpm_limit)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, feature_id) DO UPDATE SET rpm_limit = ?
	`, userID, featureID, rpmLimit, rpmLimit)
	return err
}

// DeleteUserQuotaOverride removes a quota override
func (q *QuotaEngine) DeleteUserQuotaOverride(userID int64, featureID int64) error {
	_, err := q.repo.db.Exec(`
		DELETE FROM user_quota_overrides WHERE user_id = ? AND feature_id = ?
	`, userID, featureID)
	return err
}

// GetUserQuotaOverrides returns all quota overrides for a user
func (q *QuotaEngine) GetUserQuotaOverrides(userID int64) ([]UserQuotaOverride, error) {
	rows, err := q.repo.db.Query(`
		SELECT user_id, feature_id, rpm_limit
		FROM user_quota_overrides WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []UserQuotaOverride
	for rows.Next() {
		var o UserQuotaOverride
		var rpmLimit sql.NullInt64
		if err := rows.Scan(&o.UserID, &o.FeatureID, &rpmLimit); err != nil {
			return nil, err
		}
		o.RPMLimit = ScanNullableInt(rpmLimit)
		overrides = append(overrides, o)
	}
	return overrides, rows.Err()
}

// SetGroupFeatureQuota sets a quota for a group on a feature
func (q *QuotaEngine) SetGroupFeatureQuota(groupID int64, featureID int64, rpmLimit *int) error {
	_, err := q.repo.db.Exec(`
		INSERT INTO group_feature_quotas (group_id, feature_id, rpm_limit)
		VALUES (?, ?, ?)
		ON CONFLICT (group_id, feature_id) DO UPDATE SET rpm_limit = ?
	`, groupID, featureID, rpmLimit, rpmLimit)
	return err
}

// DeleteGroupFeatureQuota removes a quota for a group on a feature
func (q *QuotaEngine) DeleteGroupFeatureQuota(groupID int64, featureID int64) error {
	_, err := q.repo.db.Exec(`
		DELETE FROM group_feature_quotas WHERE group_id = ? AND feature_id = ?
	`, groupID, featureID)
	return err
}

// GetGroupFeatureQuotas returns all quotas for a group
func (q *QuotaEngine) GetGroupFeatureQuotas(groupID int64) ([]GroupFeatureQuota, error) {
	rows, err := q.repo.db.Query(`
		SELECT group_id, feature_id, rpm_limit
		FROM group_feature_quotas WHERE group_id = ?
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []GroupFeatureQuota
	for rows.Next() {
		var gq GroupFeatureQuota
		var rpmLimit sql.NullInt64
		if err := rows.Scan(&gq.GroupID, &gq.FeatureID, &rpmLimit); err != nil {
			return nil, err
		}
		gq.RPMLimit = ScanNullableInt(rpmLimit)
		quotas = append(quotas, gq)
	}
	return quotas, rows.Err()
}

// BulkSetGroupFeatureQuotas sets multiple quotas for a group at once
func (q *QuotaEngine) BulkSetGroupFeatureQuotas(groupID int64, quotas []QuotaEntry) error {
	tx, err := q.repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, entry := range quotas {
		_, err := tx.Exec(`
			INSERT INTO group_feature_quotas (group_id, feature_id, rpm_limit)
			VALUES (?, ?, ?)
			ON CONFLICT (group_id, feature_id) DO UPDATE SET rpm_limit = ?
		`, groupID, entry.FeatureID, entry.RPMLimit, entry.RPMLimit)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BulkSetUserQuotaOverrides sets multiple quota overrides for a user at once
func (q *QuotaEngine) BulkSetUserQuotaOverrides(userID int64, quotas []QuotaEntry) error {
	tx, err := q.repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, entry := range quotas {
		_, err := tx.Exec(`
			INSERT INTO user_quota_overrides (user_id, feature_id, rpm_limit)
			VALUES (?, ?, ?)
			ON CONFLICT (user_id, feature_id) DO UPDATE SET rpm_limit = ?
		`, userID, entry.FeatureID, entry.RPMLimit, entry.RPMLimit)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
