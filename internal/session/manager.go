package session

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
)

// Session represents an individual VSCode IDE session
type Session struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	WorkspaceDir string                `json:"workspace_dir"`
	CreatedAt   time.Time              `json:"created_at"`
	LastAccess  time.Time              `json:"last_access"`
	Config      *config.Config         `json:"config"`
	Context     map[string]interface{} `json:"context"`
	Active      bool                   `json:"active"`
	mutex       sync.RWMutex
}

// Manager manages multiple VSCode IDE sessions
type Manager struct {
	sessions    map[string]*Session
	mutex       sync.RWMutex
	logger      *zap.Logger
	baseConfig  *config.Config
	cleanupTicker *time.Ticker
	stopCleanup chan bool
}

// NewManager creates a new session manager
func NewManager(baseConfig *config.Config, logger *zap.Logger) *Manager {
	manager := &Manager{
		sessions:    make(map[string]*Session),
		logger:      logger,
		baseConfig:  baseConfig,
		stopCleanup: make(chan bool),
	}

	// Start cleanup routine for inactive sessions
	manager.startCleanupRoutine()

	return manager
}

// CreateSession creates a new session for a VSCode IDE instance
func (m *Manager) CreateSession(name, workspaceDir string) (*Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := uuid.New().String()
	
	// Create session-specific configuration
	sessionConfig := m.createSessionConfig(sessionID, workspaceDir)

	session := &Session{
		ID:           sessionID,
		Name:         name,
		WorkspaceDir: workspaceDir,
		CreatedAt:    time.Now(),
		LastAccess:   time.Now(),
		Config:       sessionConfig,
		Context:      make(map[string]interface{}),
		Active:       true,
	}

	m.sessions[sessionID] = session

	m.logger.Info("Created new session",
		zap.String("session_id", sessionID),
		zap.String("name", name),
		zap.String("workspace", workspaceDir))

	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Update last access time
	session.mutex.Lock()
	session.LastAccess = time.Now()
	session.mutex.Unlock()

	return session, nil
}

// GetOrCreateSession gets an existing session or creates a new one
func (m *Manager) GetOrCreateSession(sessionID, name, workspaceDir string) (*Session, error) {
	if sessionID != "" {
		if session, err := m.GetSession(sessionID); err == nil {
			return session, nil
		}
	}

	// Create new session if not found or no ID provided
	return m.CreateSession(name, workspaceDir)
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []*Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		if session.Active {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// DeactivateSession marks a session as inactive
func (m *Manager) DeactivateSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.mutex.Lock()
	session.Active = false
	session.mutex.Unlock()

	m.logger.Info("Deactivated session", zap.String("session_id", sessionID))
	return nil
}

// RemoveSession removes a session completely
func (m *Manager) RemoveSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(m.sessions, sessionID)

	m.logger.Info("Removed session", zap.String("session_id", sessionID))
	return nil
}

// UpdateSessionContext updates the context for a session
func (m *Manager) UpdateSessionContext(sessionID string, key string, value interface{}) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.Context[key] = value
	session.LastAccess = time.Now()

	return nil
}

// GetSessionContext retrieves context value for a session
func (m *Manager) GetSessionContext(sessionID string, key string) (interface{}, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	value, exists := session.Context[key]
	if !exists {
		return nil, fmt.Errorf("context key not found: %s", key)
	}

	return value, nil
}

// createSessionConfig creates a session-specific configuration
func (m *Manager) createSessionConfig(sessionID, workspaceDir string) *config.Config {
	// Clone base configuration
	sessionConfig := *m.baseConfig

	// Create session-specific directories
	if workspaceDir != "" {
		// Use workspace-specific index directory
		sessionConfig.Indexer.IndexDir = filepath.Join(m.baseConfig.Indexer.IndexDir, "sessions", sessionID)
		sessionConfig.Indexer.RepoDir = filepath.Join(m.baseConfig.Indexer.RepoDir, "sessions", sessionID)
	}

	return &sessionConfig
}

// startCleanupRoutine starts a background routine to clean up inactive sessions
func (m *Manager) startCleanupRoutine() {
	m.cleanupTicker = time.NewTicker(30 * time.Minute) // Cleanup every 30 minutes

	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				m.cleanupInactiveSessions()
			case <-m.stopCleanup:
				m.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanupInactiveSessions removes sessions that have been inactive for too long
func (m *Manager) cleanupInactiveSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	inactiveThreshold := time.Now().Add(-2 * time.Hour) // 2 hours of inactivity
	var toRemove []string

	for sessionID, session := range m.sessions {
		session.mutex.RLock()
		if !session.Active || session.LastAccess.Before(inactiveThreshold) {
			toRemove = append(toRemove, sessionID)
		}
		session.mutex.RUnlock()
	}

	for _, sessionID := range toRemove {
		delete(m.sessions, sessionID)
		m.logger.Info("Cleaned up inactive session", zap.String("session_id", sessionID))
	}

	if len(toRemove) > 0 {
		m.logger.Info("Session cleanup completed", zap.Int("removed_sessions", len(toRemove)))
	}
}

// Close shuts down the session manager
func (m *Manager) Close() {
	if m.cleanupTicker != nil {
		m.stopCleanup <- true
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Deactivate all sessions
	for _, session := range m.sessions {
		session.mutex.Lock()
		session.Active = false
		session.mutex.Unlock()
	}

	m.logger.Info("Session manager closed")
}

// GetSessionStats returns statistics about sessions
func (m *Manager) GetSessionStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	activeSessions := 0
	totalSessions := len(m.sessions)

	for _, session := range m.sessions {
		session.mutex.RLock()
		if session.Active {
			activeSessions++
		}
		session.mutex.RUnlock()
	}

	return map[string]interface{}{
		"total_sessions":  totalSessions,
		"active_sessions": activeSessions,
		"inactive_sessions": totalSessions - activeSessions,
	}
}
