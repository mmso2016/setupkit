package core

import (
	"context"
	"fmt"
	"sync"
)

type contextKey string

func (c contextKey) String() string {
	return "core context key " + string(c)
}

// RollbackManager handles rollback operations
type RollbackManager struct {
	strategy    RollbackStrategy
	checkpoints []RollbackCheckpoint
	mu          sync.Mutex
}

// RollbackCheckpoint represents a rollback checkpoint
type RollbackCheckpoint struct {
	ID       string
	Rollback func(ctx context.Context) error
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(strategy RollbackStrategy) *RollbackManager {
	return &RollbackManager{
		strategy:    strategy,
		checkpoints: []RollbackCheckpoint{},
	}
}

// AddCheckpoint adds a rollback checkpoint
func (r *RollbackManager) AddCheckpoint(id string, rollback func(ctx context.Context) error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rollback == nil {
		return // No rollback function provided
	}

	r.checkpoints = append(r.checkpoints, RollbackCheckpoint{
		ID:       id,
		Rollback: rollback,
	})
}

// Execute performs the rollback based on the strategy
func (r *RollbackManager) Execute(ctx *Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.checkpoints) == 0 {
		return nil // Nothing to rollback
	}

	ctx.Logger.Info("Starting rollback procedure", "strategy", r.strategy)

	// Create a context.Context with necessary values
	rollbackCtx := context.WithValue(context.Background(), contextKey("installer_context"), ctx)
	rollbackCtx = context.WithValue(rollbackCtx, contextKey("logger"), ctx.Logger)
	rollbackCtx = context.WithValue(rollbackCtx, contextKey("config"), ctx.Config)

	var errors []error

	switch r.strategy {
	case RollbackFull:
		// Rollback all checkpoints in reverse order
		for i := len(r.checkpoints) - 1; i >= 0; i-- {
			checkpoint := r.checkpoints[i]
			ctx.Logger.Info("Rolling back component", "id", checkpoint.ID)

			if err := checkpoint.Rollback(rollbackCtx); err != nil {
				ctx.Logger.Error("Rollback failed for component", "id", checkpoint.ID, "error", err)
				errors = append(errors, fmt.Errorf("rollback %s: %w", checkpoint.ID, err))
			}
		}

	case RollbackPartial:
		// Rollback only the last failed component
		if len(r.checkpoints) > 0 {
			checkpoint := r.checkpoints[len(r.checkpoints)-1]
			ctx.Logger.Info("Rolling back last component", "id", checkpoint.ID)

			if err := checkpoint.Rollback(rollbackCtx); err != nil {
				ctx.Logger.Error("Rollback failed", "id", checkpoint.ID, "error", err)
				errors = append(errors, fmt.Errorf("rollback %s: %w", checkpoint.ID, err))
			}
		}

	case RollbackNone:
		// No rollback
		return nil
	}

	// Clear checkpoints after rollback
	r.checkpoints = []RollbackCheckpoint{}

	if len(errors) > 0 {
		return fmt.Errorf("rollback completed with %d errors", len(errors))
	}

	ctx.Logger.Info("Rollback completed successfully")
	return nil
}

// Clear removes all checkpoints
func (r *RollbackManager) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkpoints = []RollbackCheckpoint{}
}

// Count returns the number of checkpoints
func (r *RollbackManager) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.checkpoints)
}
