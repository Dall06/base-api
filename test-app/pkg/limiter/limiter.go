// Package limiter provides utilities for executing operations with plan limit checks.
// It handles the race condition between counting, checking, and creating resources
// by holding an advisory lock for the duration of the entire operation.
package limiter

import (
	"context"
	"hash/fnv"

	"github.com/uptrace/bun"
)

// Operation defines the functions needed to execute a limit-checked operation.
// This pattern ensures the advisory lock is held from count through create.
type Operation struct {
	// LockDB is the database to use for advisory locks (typically platform DB)
	LockDB *bun.DB

	// LockKey is the unique identifier for this resource type + company
	// Example: "members:company-uuid" or "plans:gym-slug"
	LockKey string

	// CountFn returns the current count of resources (typically from tenant DB)
	CountFn func(ctx context.Context) (int, error)

	// CheckFn verifies the count against the limit (typically queries platform DB)
	// Should return an error if limit is exceeded
	CheckFn func(ctx context.Context, currentCount int) error

	// CreateFn creates the resource (typically on tenant DB)
	CreateFn func(ctx context.Context) error
}

// Execute runs a limit-checked operation atomically.
// The advisory lock is held from before count until after create.
// This prevents race conditions where concurrent requests could exceed limits.
func Execute(ctx context.Context, op Operation) error {
	lockKey := GenerateLockKey(op.LockKey)

	if op.LockDB != nil {
		if _, err := op.LockDB.NewSelect().ColumnExpr("pg_advisory_lock(?)", lockKey).Exec(ctx); err != nil {
			return err
		}
		defer func() {
			_, _ = op.LockDB.NewSelect().ColumnExpr("pg_advisory_unlock(?)", lockKey).Exec(context.Background())
		}()
	}

	// Count current resources
	count, err := op.CountFn(ctx)
	if err != nil {
		return err
	}

	// Check against limit
	if err := op.CheckFn(ctx, count); err != nil {
		return err
	}

	// Create the resource (still under lock)
	return op.CreateFn(ctx)
}

// GenerateLockKey creates a unique int64 key from a string using FNV-1a hash.
// Exported for use by other packages that need consistent lock key generation.
func GenerateLockKey(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum64()) // #nosec G115 -- intentional truncation for lock key
}
