// Package ctxdb provides utilities for storing and retrieving database
// connections from context.Context for multi-tenant applications.
package ctxdb

import (
	"context"

	"github.com/uptrace/bun"
)

// contextKey is the key type for storing DB in context
type contextKey struct{}
type txKey struct{}

// dbKey is the actual key instance
var dbKey = contextKey{}

// txOverrideKey stores a bun.IDB (usually a Tx) that overrides the default DB
var txOverrideKey = txKey{}

// WithDB returns a new context with the database connection stored in it.
func WithDB(ctx context.Context, db *bun.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

// FromContext extracts the database connection from the context.
// Returns nil if no database is stored in the context.
func FromContext(ctx context.Context) *bun.DB {
	db, _ := ctx.Value(dbKey).(*bun.DB)
	return db
}

// MustFromContext extracts the database connection from the context.
// Panics if no database is stored in the context.
func MustFromContext(ctx context.Context) *bun.DB {
	db := FromContext(ctx)
	if db == nil {
		panic("ctxdb: no database in context")
	}
	return db
}

// WithTxOverride returns a new context with a transaction override.
// Repos that call IDBFromContext will use this tx instead of the default DB.
func WithTxOverride(ctx context.Context, tx bun.IDB) context.Context {
	return context.WithValue(ctx, txOverrideKey, tx)
}

// IDBFromContext returns the tx override if present, otherwise the default DB.
// Use this in repos instead of MustFromContext when transaction support is needed.
func IDBFromContext(ctx context.Context) bun.IDB {
	if tx, ok := ctx.Value(txOverrideKey).(bun.IDB); ok && tx != nil {
		return tx
	}
	return MustFromContext(ctx)
}
