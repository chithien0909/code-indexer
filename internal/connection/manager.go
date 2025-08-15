package connection

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/my-mcp/code-indexer/internal/config"
	"github.com/my-mcp/code-indexer/internal/session"
)

// ConnectionType represents the type of connection
type ConnectionType string

const (
	ConnectionTypeHTTP      ConnectionType = "http"
	ConnectionTypeWebSocket ConnectionType = "websocket"
	ConnectionTypeStdio     ConnectionType = "stdio"
)

// Connection represents a single IDE connection
type Connection struct {
	ID          string         `json:"id"`
	Type        ConnectionType `json:"type"`
	RemoteAddr  string         `json:"remote_addr"`
	UserAgent   string         `json:"user_agent"`
	SessionID   string         `json:"session_id"`
	CreatedAt   time.Time      `json:"created_at"`
	LastActive  time.Time      `json:"last_active"`
	Active      bool           `json:"active"`
	Context     context.Context
	Cancel      context.CancelFunc
	WSConn      *websocket.Conn `json:"-"` // For WebSocket connections
	HTTPWriter  http.ResponseWriter `json:"-"` // For HTTP connections
	mutex       sync.RWMutex
}

// Manager manages multiple IDE connections
type Manager struct {
	connections    map[string]*Connection
	sessionManager *session.Manager
	config         *config.Config
	logger         *zap.Logger
	upgrader       websocket.Upgrader
	mutex          sync.RWMutex
	
	// Connection limits and timeouts
	maxConnections    int
	connectionTimeout time.Duration
	cleanupInterval   time.Duration
	
	// Shutdown handling
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewManager creates a new connection manager
func NewManager(cfg *config.Config, sessionMgr *session.Manager, logger *zap.Logger) *Manager {
	manager := &Manager{
		connections:       make(map[string]*Connection),
		sessionManager:    sessionMgr,
		config:           cfg,
		logger:           logger,
		maxConnections:   cfg.Server.MultiIDE.MaxConnections,
		connectionTimeout: time.Duration(cfg.Server.MultiIDE.ConnectionTimeoutSeconds) * time.Second,
		cleanupInterval:  time.Duration(cfg.Server.MultiIDE.CleanupIntervalMinutes) * time.Minute,
		shutdown:         make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin for now
				// In production, this should be more restrictive
				return true
			},
		},
	}

	// Start cleanup goroutine
	manager.wg.Add(1)
	go manager.cleanupLoop()

	return manager
}

// CreateConnection creates a new connection
func (m *Manager) CreateConnection(connType ConnectionType, remoteAddr, userAgent string) (*Connection, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check connection limits
	if len(m.connections) >= m.maxConnections {
		return nil, fmt.Errorf("maximum connections reached (%d)", m.maxConnections)
	}

	// Create connection
	ctx, cancel := context.WithCancel(context.Background())
	conn := &Connection{
		ID:         uuid.New().String(),
		Type:       connType,
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Active:     true,
		Context:    ctx,
		Cancel:     cancel,
	}

	m.connections[conn.ID] = conn

	m.logger.Info("Created new connection",
		zap.String("connection_id", conn.ID),
		zap.String("type", string(connType)),
		zap.String("remote_addr", remoteAddr),
		zap.String("user_agent", userAgent))

	return conn, nil
}

// GetConnection retrieves a connection by ID
func (m *Manager) GetConnection(connectionID string) (*Connection, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conn, exists := m.connections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// Update last active time
	conn.mutex.Lock()
	conn.LastActive = time.Now()
	conn.mutex.Unlock()

	return conn, nil
}

// AssociateSession associates a connection with a session
func (m *Manager) AssociateSession(connectionID, sessionID string) error {
	conn, err := m.GetConnection(connectionID)
	if err != nil {
		return err
	}

	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	conn.SessionID = sessionID
	conn.LastActive = time.Now()

	m.logger.Debug("Associated connection with session",
		zap.String("connection_id", connectionID),
		zap.String("session_id", sessionID))

	return nil
}

// CloseConnection closes a connection
func (m *Manager) CloseConnection(connectionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conn, exists := m.connections[connectionID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	// Close WebSocket connection if exists
	if conn.WSConn != nil {
		conn.WSConn.Close()
	}

	// Cancel context
	conn.Cancel()

	// Mark as inactive
	conn.mutex.Lock()
	conn.Active = false
	conn.mutex.Unlock()

	// Remove from connections map
	delete(m.connections, connectionID)

	m.logger.Info("Closed connection",
		zap.String("connection_id", connectionID),
		zap.String("type", string(conn.Type)))

	return nil
}

// ListConnections returns all active connections
func (m *Manager) ListConnections() []*Connection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	connections := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}

	return connections
}

// GetConnectionStats returns connection statistics
func (m *Manager) GetConnectionStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(m.connections),
		"max_connections":   m.maxConnections,
		"connection_types":  make(map[string]int),
	}

	// Count by type
	typeCounts := make(map[string]int)
	for _, conn := range m.connections {
		typeCounts[string(conn.Type)]++
	}
	stats["connection_types"] = typeCounts

	return stats
}

// cleanupLoop periodically cleans up inactive connections
func (m *Manager) cleanupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupInactiveConnections()
		case <-m.shutdown:
			return
		}
	}
}

// cleanupInactiveConnections removes connections that have been inactive for too long
func (m *Manager) cleanupInactiveConnections() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for id, conn := range m.connections {
		conn.mutex.RLock()
		inactive := now.Sub(conn.LastActive) > m.connectionTimeout
		conn.mutex.RUnlock()

		if inactive {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		conn := m.connections[id]
		
		// Close WebSocket connection if exists
		if conn.WSConn != nil {
			conn.WSConn.Close()
		}
		
		// Cancel context
		conn.Cancel()
		
		// Remove from map
		delete(m.connections, id)

		m.logger.Info("Cleaned up inactive connection",
			zap.String("connection_id", id),
			zap.String("type", string(conn.Type)))
	}

	if len(toRemove) > 0 {
		m.logger.Info("Connection cleanup completed",
			zap.Int("removed_connections", len(toRemove)),
			zap.Int("active_connections", len(m.connections)))
	}
}

// Close shuts down the connection manager
func (m *Manager) Close() error {
	m.logger.Info("Shutting down connection manager")

	// Signal shutdown
	close(m.shutdown)

	// Close all connections
	m.mutex.Lock()
	for id, conn := range m.connections {
		if conn.WSConn != nil {
			conn.WSConn.Close()
		}
		conn.Cancel()
		delete(m.connections, id)
	}
	m.mutex.Unlock()

	// Wait for cleanup goroutine to finish
	m.wg.Wait()

	m.logger.Info("Connection manager shutdown complete")
	return nil
}
