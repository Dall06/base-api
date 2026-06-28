package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Migrator handles database migrations
type Migrator struct {
	platformDB  *bun.DB
	adminURI    string // URI with admin privileges for CREATE DATABASE
	loader      *Loader
	snapshot    *SnapshotManager
	cfg         Config
}

// NewMigrator creates a new migrator
func NewMigrator(platformDB *bun.DB, cfg Config) *Migrator {
	return &Migrator{
		platformDB: platformDB,
		adminURI:   cfg.DatabaseURI,
		loader:     NewLoader(cfg.MigrationsDir),
		snapshot:   NewSnapshotManager(cfg),
		cfg:        cfg,
	}
}

// RunAll runs migrations for platform and all tenants
func (m *Migrator) RunAll(ctx context.Context) ([]MigrationResult, error) {
	results := []MigrationResult{}

	// Auto-init baselines for existing DBs if configured
	if m.cfg.AutoInitBaseline {
		slog.Info("auto-initializing baselines for existing databases...")
		if err := m.InitAllBaselines(ctx); err != nil {
			slog.Warn("failed to auto-init baselines", slog.String("error", err.Error()))
		}
	}

	// Get list of all tenant databases
	tenantDBs, err := m.getTenantDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant databases: %w", err)
	}

	// Check if there are pending migrations
	hasPendingPlatform, err := m.hasPendingMigrations(ctx, m.platformDB, DBTypePlatform)
	if err != nil {
		slog.Warn("failed to check platform migrations", slog.String("error", err.Error()))
	}

	hasPendingTenant := false
	for _, dbName := range tenantDBs {
		db, err := m.connectToDatabase(dbName)
		if err != nil {
			continue
		}
		pending, _ := m.hasPendingMigrations(ctx, db, DBTypeTenant)
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("failed to close tenant DB", "error", closeErr)
		}
		if pending {
			hasPendingTenant = true
			break
		}
	}

	// Create snapshots if there are pending migrations
	if hasPendingPlatform || hasPendingTenant {
		slog.Info("creating snapshots before migrations...")
		_, err := m.snapshot.CreateAll("bro_db", tenantDBs)
		if err != nil {
			return nil, fmt.Errorf("failed to create snapshots: %w", err)
		}
	}

	// Run platform migrations
	if hasPendingPlatform {
		result := m.migrateDatabase(ctx, m.platformDB, "bro_db", DBTypePlatform)
		results = append(results, result)
	}

	// Canary/Pilot tenant pattern: migrate pilot first, then others
	if m.cfg.PilotTenant != "" && hasPendingTenant {
		pilotDBName := fmt.Sprintf("gym_%s", m.cfg.PilotTenant)
		slog.Info("running pilot migration first", slog.String("tenant", m.cfg.PilotTenant))

		db, err := m.connectToDatabase(pilotDBName)
		if err != nil {
			return results, fmt.Errorf("failed to connect to pilot tenant %s: %w", m.cfg.PilotTenant, err)
		}

		result := m.migrateDatabase(ctx, db, pilotDBName, DBTypeTenant)
		results = append(results, result)
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("failed to close tenant DB", "error", closeErr)
		}

		if !result.Success {
			slog.Error("pilot migration failed, stopping rollout",
				slog.String("tenant", m.cfg.PilotTenant),
				slog.String("error", result.Error.Error()),
			)
			return results, fmt.Errorf("pilot migration failed: %w", result.Error)
		}

		slog.Info("pilot migration succeeded, proceeding with remaining tenants",
			slog.String("tenant", m.cfg.PilotTenant),
		)

		// Remove pilot from list to avoid double migration
		filteredDBs := make([]string, 0, len(tenantDBs)-1)
		for _, dbName := range tenantDBs {
			if dbName != pilotDBName {
				filteredDBs = append(filteredDBs, dbName)
			}
		}
		tenantDBs = filteredDBs
	}

	// Run tenant migrations
	for _, dbName := range tenantDBs {
		db, err := m.connectToDatabase(dbName)
		if err != nil {
			results = append(results, MigrationResult{
				Database: dbName,
				Success:  false,
				Error:    err,
			})
			continue
		}

		result := m.migrateDatabase(ctx, db, dbName, DBTypeTenant)
		results = append(results, result)
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("failed to close tenant DB", "error", closeErr)
		}
	}

	return results, nil
}

// MigratePlatform runs migrations only for the platform database
func (m *Migrator) MigratePlatform(ctx context.Context) (*MigrationResult, error) {
	hasPending, err := m.hasPendingMigrations(ctx, m.platformDB, DBTypePlatform)
	if err != nil {
		return nil, err
	}

	if hasPending {
		_, err := m.snapshot.Create("bro_db")
		if err != nil {
			return nil, fmt.Errorf("failed to create snapshot: %w", err)
		}
	}

	result := m.migrateDatabase(ctx, m.platformDB, "bro_db", DBTypePlatform)
	return &result, nil
}

// MigrateTenant runs migrations for a specific tenant
func (m *Migrator) MigrateTenant(ctx context.Context, slug string) (*MigrationResult, error) {
	dbName := fmt.Sprintf("gym_%s", slug)

	db, err := m.connectToDatabase(dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	hasPending, err := m.hasPendingMigrations(ctx, db, DBTypeTenant)
	if err != nil {
		return nil, err
	}

	if hasPending {
		_, err := m.snapshot.Create(dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to create snapshot: %w", err)
		}
	}

	result := m.migrateDatabase(ctx, db, dbName, DBTypeTenant)
	return &result, nil
}

// migrateDatabase runs migrations for a single database
func (m *Migrator) migrateDatabase(ctx context.Context, db *bun.DB, dbName string, dbType DBType) MigrationResult {
	result := MigrationResult{
		Database: dbName,
	}

	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx, db); err != nil {
		result.Error = fmt.Errorf("failed to create migrations table: %w", err)
		return result
	}

	// Load available migrations
	migrations, err := m.loader.LoadMigrations(dbType)
	if err != nil {
		result.Error = fmt.Errorf("failed to load migrations: %w", err)
		return result
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx, db)
	if err != nil {
		result.Error = fmt.Errorf("failed to get applied migrations: %w", err)
		return result
	}

	// Find pending migrations
	pending := m.findPendingMigrations(migrations, applied)
	if len(pending) == 0 {
		result.Success = true
		slog.Info("no pending migrations", slog.String("database", dbName))
		return result
	}

	slog.Info("applying migrations", slog.String("database", dbName), slog.Int("count", len(pending)))

	// Apply each migration in a transaction
	for _, migration := range pending {
		if err := m.applyMigration(ctx, db, migration, result.SnapshotPath); err != nil {
			result.Error = fmt.Errorf("migration %s failed: %w", migration.Version, err)
			return result
		}
		result.AppliedVersions = append(result.AppliedVersions, migration.Version)
		result.MigrationsRun++
	}

	result.Success = true
	slog.Info("migrations completed", slog.String("database", dbName), slog.Int("applied", result.MigrationsRun))
	return result
}

// ensureMigrationsTable creates the _migrations table if it doesn't exist.
// RAW SQL: DDL (CREATE TABLE) cannot be expressed through Bun's query builder.
func (m *Migrator) ensureMigrationsTable(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(10) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			snapshot_path VARCHAR(500),
			checksum VARCHAR(64)
		)
	`)
	return err
}

// getAppliedMigrations returns all migrations that have been applied
func (m *Migrator) getAppliedMigrations(ctx context.Context, db *bun.DB) (map[string]AppliedMigration, error) {
	var applied []AppliedMigration
	err := db.NewSelect().
		Model(&applied).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]AppliedMigration)
	for _, a := range applied {
		result[a.Version] = a
	}
	return result, nil
}

// findPendingMigrations returns migrations that haven't been applied yet
func (m *Migrator) findPendingMigrations(all []Migration, applied map[string]AppliedMigration) []Migration {
	pending := []Migration{}
	for _, migration := range all {
		if _, exists := applied[migration.Version]; !exists {
			pending = append(pending, migration)
		}
	}
	return pending
}

// hasPendingMigrations checks if there are any pending migrations
func (m *Migrator) hasPendingMigrations(ctx context.Context, db *bun.DB, dbType DBType) (bool, error) {
	if err := m.ensureMigrationsTable(ctx, db); err != nil {
		return false, err
	}

	migrations, err := m.loader.LoadMigrations(dbType)
	if err != nil {
		return false, err
	}

	applied, err := m.getAppliedMigrations(ctx, db)
	if err != nil {
		return false, err
	}

	pending := m.findPendingMigrations(migrations, applied)
	return len(pending) > 0, nil
}

// applyMigration applies a single migration within a transaction
func (m *Migrator) applyMigration(ctx context.Context, db *bun.DB, migration Migration, snapshotPath string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// RAW SQL: migration files contain arbitrary DDL/DML that must be executed as-is.
	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("SQL execution failed: %w", err)
	}

	// Record migration
	record := &AppliedMigration{
		Version:      migration.Version,
		Name:         migration.Name,
		AppliedAt:    time.Now(),
		SnapshotPath: snapshotPath,
		Checksum:     migration.Checksum,
	}

	if _, err := tx.NewInsert().Model(record).Exec(ctx); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// getTenantDatabases returns all tenant database names
func (m *Migrator) getTenantDatabases(ctx context.Context) ([]string, error) {
	// Query companies to get all slugs
	var slugs []string
	err := m.platformDB.NewSelect().
		Table("companies").
		Column("slug").
		Where("deleted_at IS NULL").
		Scan(ctx, &slugs)
	if err != nil {
		return nil, err
	}

	// Convert slugs to database names
	dbNames := make([]string, len(slugs))
	for i, slug := range slugs {
		dbNames[i] = fmt.Sprintf("gym_%s", slug)
	}
	return dbNames, nil
}

// connectToDatabase creates a connection to a specific database
func (m *Migrator) connectToDatabase(dbName string) (*bun.DB, error) {
	uri := m.snapshot.buildURI(dbName)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(uri)))

	// Quick ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		return nil, fmt.Errorf("failed to connect to %s: %w", dbName, err)
	}

	return bun.NewDB(sqldb, pgdialect.New()), nil
}

// InitBaseline initializes the baseline migration for an existing database
// This is used for databases that were created before the migration system existed
func (m *Migrator) InitBaseline(ctx context.Context, dbName string, dbType DBType) error {
	var db *bun.DB
	var err error

	if dbName == "bro_db" {
		db = m.platformDB
	}
	if dbName != "bro_db" {
		db, err = m.connectToDatabase(dbName)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()
	}

	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Check if baseline already exists
	applied, err := m.getAppliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if _, exists := applied["000"]; exists {
		slog.Info("baseline already initialized", slog.String("database", dbName))
		return nil
	}

	// Insert baseline record
	record := &AppliedMigration{
		Version:   "000",
		Name:      "baseline",
		AppliedAt: time.Now(),
		Checksum:  "existing_schema",
	}

	if _, err := db.NewInsert().Model(record).Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert baseline record: %w", err)
	}

	slog.Info("baseline initialized", slog.String("database", dbName))
	return nil
}

// InitAllBaselines initializes baselines for all existing databases
func (m *Migrator) InitAllBaselines(ctx context.Context) error {
	// Platform
	if err := m.InitBaseline(ctx, "bro_db", DBTypePlatform); err != nil {
		return fmt.Errorf("failed to init platform baseline: %w", err)
	}

	// Tenants
	tenantDBs, err := m.getTenantDatabases(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant databases: %w", err)
	}

	for _, dbName := range tenantDBs {
		if err := m.InitBaseline(ctx, dbName, DBTypeTenant); err != nil {
			slog.Error("failed to init tenant baseline",
				slog.String("database", dbName),
				slog.String("error", err.Error()),
			)
			// Continue with other tenants
		}
	}

	return nil
}

// GetStatus returns migration status for all databases
func (m *Migrator) GetStatus(ctx context.Context) (map[string][]AppliedMigration, error) {
	status := make(map[string][]AppliedMigration)

	// Platform
	applied, err := m.getAppliedMigrations(ctx, m.platformDB)
	if err == nil {
		list := make([]AppliedMigration, 0, len(applied))
		for _, a := range applied {
			list = append(list, a)
		}
		status["bro_db"] = list
	}

	// Tenants
	tenantDBs, _ := m.getTenantDatabases(ctx)
	for _, dbName := range tenantDBs {
		db, err := m.connectToDatabase(dbName)
		if err != nil {
			continue
		}
		applied, err := m.getAppliedMigrations(ctx, db)
		if closeErr := db.Close(); closeErr != nil {
			slog.Warn("failed to close tenant DB", "error", closeErr)
		}
		if err == nil {
			list := make([]AppliedMigration, 0, len(applied))
			for _, a := range applied {
				list = append(list, a)
			}
			status[dbName] = list
		}
	}

	return status, nil
}
