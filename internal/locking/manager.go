package locking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LockType represents different types of locks
type LockType string

const (
	LockTypeRead      LockType = "read"
	LockTypeWrite     LockType = "write"
	LockTypeExclusive LockType = "exclusive"
)

// ResourceType represents different types of resources
type ResourceType string

const (
	ResourceTypeIndex      ResourceType = "index"
	ResourceTypeRepository ResourceType = "repository"
	ResourceTypeFile       ResourceType = "file"
	ResourceTypeSession    ResourceType = "session"
)

// Lock represents a resource lock
type Lock struct {
	ID           string       `json:"id"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   string       `json:"resource_id"`
	LockType     LockType     `json:"lock_type"`
	OwnerID      string       `json:"owner_id"` // Connection or session ID
	AcquiredAt   time.Time    `json:"acquired_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Context      context.Context `json:"-"`
	Cancel       context.CancelFunc `json:"-"`
}

// ResourceLock manages locks for a specific resource
type ResourceLock struct {
	ResourceID   string
	ResourceType ResourceType
	ReadLocks    map[string]*Lock // Multiple read locks allowed
	WriteLock    *Lock            // Only one write lock allowed
	ExclusiveLock *Lock           // Exclusive lock blocks everything
	WaitQueue    []*LockRequest   // Queue of waiting lock requests
	mutex        sync.RWMutex
}

// LockRequest represents a pending lock request
type LockRequest struct {
	ID           string
	ResourceType ResourceType
	ResourceID   string
	LockType     LockType
	OwnerID      string
	Timeout      time.Duration
	RequestedAt  time.Time
	Context      context.Context
	ResultChan   chan *LockResult
}

// LockResult represents the result of a lock request
type LockResult struct {
	Lock  *Lock
	Error error
}

// Manager manages resource locks across the system
type Manager struct {
	resources       map[string]*ResourceLock // resourceType:resourceID -> ResourceLock
	locks          map[string]*Lock         // lockID -> Lock
	config         *LockConfig
	logger         *zap.Logger
	mutex          sync.RWMutex
	
	// Cleanup and monitoring
	cleanupInterval time.Duration
	shutdown        chan struct{}
	wg              sync.WaitGroup
}

// LockConfig contains locking configuration
type LockConfig struct {
	DefaultTimeout      time.Duration
	MaxLockDuration     time.Duration
	CleanupInterval     time.Duration
	EnableDeadlockCheck bool
	MaxWaitQueueSize    int
}

// NewManager creates a new lock manager
func NewManager(config *LockConfig, logger *zap.Logger) *Manager {
	if config == nil {
		config = &LockConfig{
			DefaultTimeout:      30 * time.Second,
			MaxLockDuration:     5 * time.Minute,
			CleanupInterval:     1 * time.Minute,
			EnableDeadlockCheck: true,
			MaxWaitQueueSize:    100,
		}
	}

	manager := &Manager{
		resources:       make(map[string]*ResourceLock),
		locks:          make(map[string]*Lock),
		config:         config,
		logger:         logger,
		cleanupInterval: config.CleanupInterval,
		shutdown:        make(chan struct{}),
	}

	// Start cleanup goroutine
	manager.wg.Add(1)
	go manager.cleanupLoop()

	return manager
}

// AcquireLock attempts to acquire a lock on a resource
func (m *Manager) AcquireLock(ctx context.Context, resourceType ResourceType, resourceID string, lockType LockType, ownerID string, timeout time.Duration) (*Lock, error) {
	if timeout == 0 {
		timeout = m.config.DefaultTimeout
	}

	// Create lock request
	request := &LockRequest{
		ID:           fmt.Sprintf("%s-%s-%s-%d", ownerID, resourceType, resourceID, time.Now().UnixNano()),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		LockType:     lockType,
		OwnerID:      ownerID,
		Timeout:      timeout,
		RequestedAt:  time.Now(),
		Context:      ctx,
		ResultChan:   make(chan *LockResult, 1),
	}

	// Get or create resource lock
	resourceKey := string(resourceType) + ":" + resourceID
	resourceLock := m.getOrCreateResourceLock(resourceKey, resourceType, resourceID)

	// Try to acquire lock immediately or queue the request
	if lock := m.tryAcquireLock(resourceLock, request); lock != nil {
		return lock, nil
	}

	// Add to wait queue
	if err := m.addToWaitQueue(resourceLock, request); err != nil {
		return nil, err
	}

	// Wait for lock or timeout
	select {
	case result := <-request.ResultChan:
		return result.Lock, result.Error
	case <-time.After(timeout):
		m.removeFromWaitQueue(resourceLock, request.ID)
		return nil, fmt.Errorf("lock acquisition timeout after %v", timeout)
	case <-ctx.Done():
		m.removeFromWaitQueue(resourceLock, request.ID)
		return nil, ctx.Err()
	}
}

// ReleaseLock releases a previously acquired lock
func (m *Manager) ReleaseLock(lockID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lock, exists := m.locks[lockID]
	if !exists {
		return fmt.Errorf("lock not found: %s", lockID)
	}

	// Get resource lock
	resourceKey := string(lock.ResourceType) + ":" + lock.ResourceID
	resourceLock, exists := m.resources[resourceKey]
	if !exists {
		return fmt.Errorf("resource lock not found: %s", resourceKey)
	}

	// Remove lock from resource
	resourceLock.mutex.Lock()
	defer resourceLock.mutex.Unlock()

	switch lock.LockType {
	case LockTypeRead:
		delete(resourceLock.ReadLocks, lockID)
	case LockTypeWrite:
		if resourceLock.WriteLock != nil && resourceLock.WriteLock.ID == lockID {
			resourceLock.WriteLock = nil
		}
	case LockTypeExclusive:
		if resourceLock.ExclusiveLock != nil && resourceLock.ExclusiveLock.ID == lockID {
			resourceLock.ExclusiveLock = nil
		}
	}

	// Cancel lock context
	if lock.Cancel != nil {
		lock.Cancel()
	}

	// Remove from global locks map
	delete(m.locks, lockID)

	m.logger.Debug("Released lock",
		zap.String("lock_id", lockID),
		zap.String("resource_type", string(lock.ResourceType)),
		zap.String("resource_id", lock.ResourceID),
		zap.String("lock_type", string(lock.LockType)),
		zap.String("owner_id", lock.OwnerID))

	// Process wait queue
	m.processWaitQueue(resourceLock)

	return nil
}

// getOrCreateResourceLock gets or creates a resource lock
func (m *Manager) getOrCreateResourceLock(resourceKey string, resourceType ResourceType, resourceID string) *ResourceLock {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	resourceLock, exists := m.resources[resourceKey]
	if !exists {
		resourceLock = &ResourceLock{
			ResourceID:   resourceID,
			ResourceType: resourceType,
			ReadLocks:    make(map[string]*Lock),
			WaitQueue:    make([]*LockRequest, 0),
		}
		m.resources[resourceKey] = resourceLock
	}

	return resourceLock
}

// tryAcquireLock attempts to acquire a lock immediately
func (m *Manager) tryAcquireLock(resourceLock *ResourceLock, request *LockRequest) *Lock {
	resourceLock.mutex.Lock()
	defer resourceLock.mutex.Unlock()

	// Check if lock can be acquired
	if !m.canAcquireLock(resourceLock, request.LockType) {
		return nil
	}

	// Create lock
	ctx, cancel := context.WithTimeout(request.Context, m.config.MaxLockDuration)
	lock := &Lock{
		ID:           request.ID,
		ResourceType: request.ResourceType,
		ResourceID:   request.ResourceID,
		LockType:     request.LockType,
		OwnerID:      request.OwnerID,
		AcquiredAt:   time.Now(),
		ExpiresAt:    time.Now().Add(m.config.MaxLockDuration),
		Context:      ctx,
		Cancel:       cancel,
	}

	// Add lock to resource
	switch request.LockType {
	case LockTypeRead:
		resourceLock.ReadLocks[lock.ID] = lock
	case LockTypeWrite:
		resourceLock.WriteLock = lock
	case LockTypeExclusive:
		resourceLock.ExclusiveLock = lock
	}

	// Add to global locks map
	m.mutex.Lock()
	m.locks[lock.ID] = lock
	m.mutex.Unlock()

	m.logger.Debug("Acquired lock",
		zap.String("lock_id", lock.ID),
		zap.String("resource_type", string(lock.ResourceType)),
		zap.String("resource_id", lock.ResourceID),
		zap.String("lock_type", string(lock.LockType)),
		zap.String("owner_id", lock.OwnerID))

	return lock
}

// canAcquireLock checks if a lock can be acquired
func (m *Manager) canAcquireLock(resourceLock *ResourceLock, lockType LockType) bool {
	// Exclusive lock blocks everything
	if resourceLock.ExclusiveLock != nil {
		return false
	}

	switch lockType {
	case LockTypeRead:
		// Read locks are compatible with other read locks but not write locks
		return resourceLock.WriteLock == nil
	case LockTypeWrite:
		// Write locks require no other locks
		return len(resourceLock.ReadLocks) == 0 && resourceLock.WriteLock == nil
	case LockTypeExclusive:
		// Exclusive locks require no other locks
		return len(resourceLock.ReadLocks) == 0 && resourceLock.WriteLock == nil
	}

	return false
}

// addToWaitQueue adds a lock request to the wait queue
func (m *Manager) addToWaitQueue(resourceLock *ResourceLock, request *LockRequest) error {
	resourceLock.mutex.Lock()
	defer resourceLock.mutex.Unlock()

	if len(resourceLock.WaitQueue) >= m.config.MaxWaitQueueSize {
		return fmt.Errorf("wait queue full for resource %s:%s", request.ResourceType, request.ResourceID)
	}

	resourceLock.WaitQueue = append(resourceLock.WaitQueue, request)
	return nil
}

// removeFromWaitQueue removes a lock request from the wait queue
func (m *Manager) removeFromWaitQueue(resourceLock *ResourceLock, requestID string) {
	resourceLock.mutex.Lock()
	defer resourceLock.mutex.Unlock()

	for i, req := range resourceLock.WaitQueue {
		if req.ID == requestID {
			resourceLock.WaitQueue = append(resourceLock.WaitQueue[:i], resourceLock.WaitQueue[i+1:]...)
			break
		}
	}
}

// processWaitQueue processes pending lock requests in the wait queue
func (m *Manager) processWaitQueue(resourceLock *ResourceLock) {
	resourceLock.mutex.Lock()
	defer resourceLock.mutex.Unlock()

	for i := 0; i < len(resourceLock.WaitQueue); {
		request := resourceLock.WaitQueue[i]
		
		// Check if request has timed out
		if time.Since(request.RequestedAt) > request.Timeout {
			// Remove from queue and send timeout error
			resourceLock.WaitQueue = append(resourceLock.WaitQueue[:i], resourceLock.WaitQueue[i+1:]...)
			select {
			case request.ResultChan <- &LockResult{Error: fmt.Errorf("lock request timeout")}:
			default:
			}
			continue
		}

		// Try to acquire lock
		if lock := m.tryAcquireLockFromQueue(resourceLock, request); lock != nil {
			// Remove from queue and send success
			resourceLock.WaitQueue = append(resourceLock.WaitQueue[:i], resourceLock.WaitQueue[i+1:]...)
			select {
			case request.ResultChan <- &LockResult{Lock: lock}:
			default:
			}
			continue
		}

		i++
	}
}

// tryAcquireLockFromQueue attempts to acquire a lock from the wait queue
func (m *Manager) tryAcquireLockFromQueue(resourceLock *ResourceLock, request *LockRequest) *Lock {
	// This is similar to tryAcquireLock but assumes resourceLock is already locked
	if !m.canAcquireLock(resourceLock, request.LockType) {
		return nil
	}

	// Create lock
	ctx, cancel := context.WithTimeout(request.Context, m.config.MaxLockDuration)
	lock := &Lock{
		ID:           request.ID,
		ResourceType: request.ResourceType,
		ResourceID:   request.ResourceID,
		LockType:     request.LockType,
		OwnerID:      request.OwnerID,
		AcquiredAt:   time.Now(),
		ExpiresAt:    time.Now().Add(m.config.MaxLockDuration),
		Context:      ctx,
		Cancel:       cancel,
	}

	// Add lock to resource
	switch request.LockType {
	case LockTypeRead:
		resourceLock.ReadLocks[lock.ID] = lock
	case LockTypeWrite:
		resourceLock.WriteLock = lock
	case LockTypeExclusive:
		resourceLock.ExclusiveLock = lock
	}

	// Add to global locks map
	m.mutex.Lock()
	m.locks[lock.ID] = lock
	m.mutex.Unlock()

	return lock
}

// cleanupLoop periodically cleans up expired locks
func (m *Manager) cleanupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpiredLocks()
		case <-m.shutdown:
			return
		}
	}
}

// cleanupExpiredLocks removes expired locks
func (m *Manager) cleanupExpiredLocks() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	expiredLocks := make([]string, 0)

	for lockID, lock := range m.locks {
		if now.After(lock.ExpiresAt) {
			expiredLocks = append(expiredLocks, lockID)
		}
	}

	for _, lockID := range expiredLocks {
		m.ReleaseLock(lockID)
	}

	if len(expiredLocks) > 0 {
		m.logger.Info("Cleaned up expired locks", zap.Int("count", len(expiredLocks)))
	}
}

// GetLockStats returns locking statistics
func (m *Manager) GetLockStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_locks":     len(m.locks),
		"total_resources": len(m.resources),
		"lock_types":      make(map[string]int),
		"resource_types":  make(map[string]int),
	}

	// Count by lock type
	lockTypes := make(map[string]int)
	resourceTypes := make(map[string]int)
	
	for _, lock := range m.locks {
		lockTypes[string(lock.LockType)]++
		resourceTypes[string(lock.ResourceType)]++
	}
	
	stats["lock_types"] = lockTypes
	stats["resource_types"] = resourceTypes

	return stats
}

// Close shuts down the lock manager
func (m *Manager) Close() error {
	m.logger.Info("Shutting down lock manager")

	// Signal shutdown
	close(m.shutdown)

	// Release all locks
	m.mutex.Lock()
	for lockID := range m.locks {
		m.ReleaseLock(lockID)
	}
	m.mutex.Unlock()

	// Wait for cleanup goroutine to finish
	m.wg.Wait()

	m.logger.Info("Lock manager shutdown complete")
	return nil
}
