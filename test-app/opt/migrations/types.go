package migrations

import "time"

// Migration represents a single migration file
type Migration struct {
	Version     string // "001", "002", etc.
	Name        string // "add_logo", "add_feature"
	Filename    string // "001_add_logo.sql"
	SQL         string // Content of the file
	Checksum    string // SHA256 of SQL content
	Description string // From comment header
}

// AppliedMigration represents a migration that has been applied
type AppliedMigration struct {
	ID           int64     `bun:"id,pk,autoincrement"`
	Version      string    `bun:"version,notnull,unique"`
	Name         string    `bun:"name,notnull"`
	AppliedAt    time.Time `bun:"applied_at,nullzero,default:current_timestamp"`
	SnapshotPath string    `bun:"snapshot_path"`
	Checksum     string    `bun:"checksum"`
}

// MigrationResult holds the result of a migration run
type MigrationResult struct {
	Database        string
	MigrationsRun   int
	SnapshotPath    string
	Success         bool
	Error           error
	AppliedVersions []string
}

// Config holds migration configuration
type Config struct {
	MigrationsDir     string // Base directory for migrations
	SnapshotDir       string // Where to store snapshots
	SnapshotRetention int    // How many snapshots to keep per DB
	CloudSync         bool   // Whether to sync to cloud
	RcloneRemote      string // rclone remote path (e.g., "gdrive:bro-backups")
	MigrateOnStartup  bool   // Whether to run migrations on startup
	DatabaseURI       string // Admin database URI for pg_dump
	PilotTenant       string // If set, migrate this tenant first as canary test
	AutoInitBaseline  bool   // If true, auto-init baseline for DBs without _migrations
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MigrationsDir:     "/app/migrations",
		SnapshotDir:       "/opt/bro/snapshots",
		SnapshotRetention: 5,
		CloudSync:         false,
		MigrateOnStartup:  true,
	}
}

// DBType represents the type of database
type DBType string

const (
	DBTypePlatform DBType = "platform" // bro_db
	DBTypeTenant   DBType = "tenant"   // gym_*
)
