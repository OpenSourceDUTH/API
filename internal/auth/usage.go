package auth

import (
	"context"
	"sync"
	"time"
)

const (
	// UsageBufferSize is the size of the usage log buffer
	UsageBufferSize = 1000

	// UsageFlushInterval is how often to flush buffered usage logs
	UsageFlushInterval = 2 * time.Second

	// UsageCleanupInterval is how often to clean up old usage logs
	UsageCleanupInterval = 30 * time.Second

	// UsageRetentionPeriod is how long to keep usage logs (60 seconds for RPM)
	UsageRetentionPeriod = 60 * time.Second
)

// UsageEntry represents a single API request for buffered logging
type UsageEntry struct {
	UserID    int64
	FeatureID int64
	Timestamp time.Time
}

// UsageTracker tracks API usage for rate limiting with buffered writes
type UsageTracker struct {
	repo         *Repository
	buffer       chan UsageEntry
	stopCh       chan struct{}
	wg           sync.WaitGroup
	stateStore   *OAuthStateStore
	sessionStore *SessionStore
}

// NewUsageTracker creates a new usage tracker
func NewUsageTracker(repo *Repository, stateStore *OAuthStateStore, sessionStore *SessionStore) *UsageTracker {
	return &UsageTracker{
		repo:         repo,
		buffer:       make(chan UsageEntry, UsageBufferSize),
		stopCh:       make(chan struct{}),
		stateStore:   stateStore,
		sessionStore: sessionStore,
	}
}

// RecordRequest records an API request (non-blocking)
func (t *UsageTracker) RecordRequest(userID int64, featureID int64) {
	entry := UsageEntry{
		UserID:    userID,
		FeatureID: featureID,
		Timestamp: time.Now(),
	}

	// Non-blocking send - if buffer is full, drop the entry
	// This prevents blocking the API request
	select {
	case t.buffer <- entry:
	default:
		// Buffer full, silently drop
		// In production, you might want to log this
	}
}

// GetFeatureRPM returns the current requests per minute for a user on a feature
func (t *UsageTracker) GetFeatureRPM(userID int64, featureID int64) (int, error) {
	cutoff := time.Now().Add(-UsageRetentionPeriod)
	var count int
	err := t.repo.db.QueryRow(`
		SELECT COUNT(*) FROM usage_log
		WHERE user_id = ? AND feature_id = ? AND timestamp > ?
	`, userID, featureID, cutoff).Scan(&count)
	return count, err
}

// GetUserTotalRPM returns the total requests per minute for a user across all features
func (t *UsageTracker) GetUserTotalRPM(userID int64) (int, error) {
	cutoff := time.Now().Add(-UsageRetentionPeriod)
	var count int
	err := t.repo.db.QueryRow(`
		SELECT COUNT(*) FROM usage_log
		WHERE user_id = ? AND timestamp > ?
	`, userID, cutoff).Scan(&count)
	return count, err
}

// Start begins the background goroutines for flushing and cleanup
func (t *UsageTracker) Start(ctx context.Context) {
	t.wg.Add(2)

	// Usage writer goroutine
	go func() {
		defer t.wg.Done()
		t.usageWriter(ctx)
	}()

	// Cleanup goroutine
	go func() {
		defer t.wg.Done()
		t.cleanupTicker(ctx)
	}()
}

// Stop gracefully stops the usage tracker
func (t *UsageTracker) Stop() {
	close(t.stopCh)
	t.wg.Wait()
}

func (t *UsageTracker) usageWriter(ctx context.Context) {
	ticker := time.NewTicker(UsageFlushInterval)
	defer ticker.Stop()

	var batch []UsageEntry

	for {
		select {
		case <-ctx.Done():
			// Flush remaining entries before stopping
			t.flushBatch(batch)
			t.drainAndFlush()
			return
		case <-t.stopCh:
			t.flushBatch(batch)
			t.drainAndFlush()
			return
		case entry := <-t.buffer:
			batch = append(batch, entry)
			// Flush if batch is large enough
			if len(batch) >= 100 {
				t.flushBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			// Periodic flush
			if len(batch) > 0 {
				t.flushBatch(batch)
				batch = nil
			}
		}
	}
}

func (t *UsageTracker) drainAndFlush() {
	var batch []UsageEntry
	for {
		select {
		case entry := <-t.buffer:
			batch = append(batch, entry)
		default:
			if len(batch) > 0 {
				t.flushBatch(batch)
			}
			return
		}
	}
}

func (t *UsageTracker) flushBatch(batch []UsageEntry) {
	if len(batch) == 0 {
		return
	}

	tx, err := t.repo.db.Begin()
	if err != nil {
		return // Silently fail - in production, log this
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO usage_log (user_id, feature_id, timestamp) VALUES (?, ?, ?)
	`)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, entry := range batch {
		stmt.Exec(entry.UserID, entry.FeatureID, entry.Timestamp)
	}

	tx.Commit()
}

func (t *UsageTracker) cleanupTicker(ctx context.Context) {
	ticker := time.NewTicker(UsageCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.cleanup()
		}
	}
}

func (t *UsageTracker) cleanup() {
	cutoff := time.Now().Add(-UsageRetentionPeriod)

	// Clean up old usage logs
	t.repo.db.Exec("DELETE FROM usage_log WHERE timestamp <= ?", cutoff)

	// Clean up expired sessions
	if t.sessionStore != nil {
		t.sessionStore.CleanupExpiredSessions()
	}

	// Clean up expired OAuth states
	if t.stateStore != nil {
		t.stateStore.CleanupExpiredStates()
	}
}

// GetUsageStats returns usage statistics for a user
func (t *UsageTracker) GetUsageStats(userID int64) (map[int64]int, error) {
	cutoff := time.Now().Add(-UsageRetentionPeriod)
	rows, err := t.repo.db.Query(`
		SELECT feature_id, COUNT(*) as count
		FROM usage_log
		WHERE user_id = ? AND timestamp > ?
		GROUP BY feature_id
	`, userID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[int64]int)
	for rows.Next() {
		var featureID int64
		var count int
		if err := rows.Scan(&featureID, &count); err != nil {
			return nil, err
		}
		stats[featureID] = count
	}
	return stats, rows.Err()
}
