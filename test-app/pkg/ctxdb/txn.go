package ctxdb

import (
	"context"
	"database/sql"

	"github.com/diegoaleon/test-app/pkg/limiter"

	"github.com/uptrace/bun"
)

// TxFunc is the function signature for transactional operations
type TxFunc func(ctx context.Context) error

// RunInTx executes fn within a database transaction.
// If fn returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func RunInTx(ctx context.Context, db *bun.DB, fn TxFunc) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx)
	})
}

// RunInTxWithOpts executes fn within a database transaction with custom options.
func RunInTxWithOpts(ctx context.Context, db *bun.DB, opts *sql.TxOptions, fn TxFunc) error {
	return db.RunInTx(ctx, opts, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx)
	})
}

// WithAdvisoryLock executes fn while holding a transaction-level advisory lock.
// The lock is automatically released when the transaction ends.
// lockName is hashed to create a unique int64 key.
func WithAdvisoryLock(ctx context.Context, db *bun.DB, lockName string, fn TxFunc) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		lockKey := limiter.GenerateLockKey(lockName)
		if _, err := tx.NewSelect().ColumnExpr("pg_advisory_xact_lock(?)", lockKey).Exec(ctx); err != nil {
			return err
		}
		return fn(ctx)
	})
}

// WithAdvisoryLockForResource executes fn while holding a transaction-level advisory lock
// specific to a resource type and ID. Useful for preventing concurrent operations
// on the same resource.
func WithAdvisoryLockForResource(ctx context.Context, db *bun.DB, resourceType, resourceID string, fn TxFunc) error {
	lockName := resourceType + ":" + resourceID
	return WithAdvisoryLock(ctx, db, lockName, fn)
}
