package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "osduth_session"

	// DefaultSessionDuration is the default session lifetime
	DefaultSessionDuration = 7 * 24 * time.Hour // 7 days
)

// SessionStore manages server-side sessions
type SessionStore struct {
	repo            *Repository
	sessionDuration time.Duration
	secureCookie    bool
}

// NewSessionStore creates a new session store
func NewSessionStore(repo *Repository, sessionDuration time.Duration, secureCookie bool) *SessionStore {
	if sessionDuration == 0 {
		sessionDuration = DefaultSessionDuration
	}
	return &SessionStore{
		repo:            repo,
		sessionDuration: sessionDuration,
		secureCookie:    secureCookie,
	}
}

// CreateSession creates a new session for a user
func (s *SessionStore) CreateSession(userID int64) (*Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(s.sessionDuration)

	_, err := s.repo.db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)
	`, sessionID, userID, expiresAt)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// GetSession returns a session if it exists and is not expired
func (s *SessionStore) GetSession(sessionID string) (*Session, error) {
	var session Session
	err := s.repo.db.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ? AND expires_at > ?
	`, sessionID, time.Now()).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserFromSession returns the user associated with a session
func (s *SessionStore) GetUserFromSession(sessionID string) (*User, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(session.UserID)
}

// DeleteSession removes a session
func (s *SessionStore) DeleteSession(sessionID string) error {
	_, err := s.repo.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// DeleteUserSessions removes all sessions for a user
func (s *SessionStore) DeleteUserSessions(userID int64) error {
	_, err := s.repo.db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// CleanupExpiredSessions removes all expired sessions
func (s *SessionStore) CleanupExpiredSessions() error {
	_, err := s.repo.db.Exec("DELETE FROM sessions WHERE expires_at <= ?", time.Now())
	return err
}

// SetSessionCookie sets the session cookie on the response
func (s *SessionStore) SetSessionCookie(c *gin.Context, sessionID string) {
	maxAge := int(s.sessionDuration.Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		SessionCookieName,
		sessionID,
		maxAge,
		"/",
		"",
		s.secureCookie,
		true, // httpOnly
	)
}

// ClearSessionCookie removes the session cookie
func (s *SessionStore) ClearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		SessionCookieName,
		"",
		-1,
		"/",
		"",
		s.secureCookie,
		true,
	)
}

// GetSessionFromCookie retrieves the session ID from the request cookie
func (s *SessionStore) GetSessionFromCookie(c *gin.Context) (string, error) {
	return c.Cookie(SessionCookieName)
}

// ExtendSession extends the session expiry time
func (s *SessionStore) ExtendSession(sessionID string) error {
	expiresAt := time.Now().Add(s.sessionDuration)
	_, err := s.repo.db.Exec(`
		UPDATE sessions SET expires_at = ? WHERE id = ?
	`, expiresAt, sessionID)
	return err
}
