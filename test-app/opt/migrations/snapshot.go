package migrations

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SnapshotManager handles database snapshots
type SnapshotManager struct {
	snapshotDir  string
	retention    int
	cloudSync    bool
	rcloneRemote string
	dbURI        string // Base URI for connecting to postgres
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(cfg Config) *SnapshotManager {
	return &SnapshotManager{
		snapshotDir:  cfg.SnapshotDir,
		retention:    cfg.SnapshotRetention,
		cloudSync:    cfg.CloudSync,
		rcloneRemote: cfg.RcloneRemote,
		dbURI:        cfg.DatabaseURI,
	}
}

// CreateAll creates snapshots for all databases (platform + all tenants)
func (s *SnapshotManager) CreateAll(platformDB string, tenantDBs []string) (map[string]string, error) {
	snapshots := make(map[string]string)

	// Ensure snapshot directory exists
	if err := os.MkdirAll(s.snapshotDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create snapshot dir: %w", err)
	}

	// Snapshot platform DB
	path, err := s.Create(platformDB)
	if err != nil {
		return nil, fmt.Errorf("failed to snapshot platform db %s: %w", platformDB, err)
	}
	snapshots[platformDB] = path
	slog.Info("snapshot created", slog.String("database", platformDB), slog.String("path", path))

	// Snapshot all tenant DBs
	for _, db := range tenantDBs {
		path, err := s.Create(db)
		if err != nil {
			slog.Error("failed to snapshot tenant db", slog.String("database", db), slog.String("error", err.Error()))
			continue // Don't fail all, just log and continue
		}
		snapshots[db] = path
		slog.Info("snapshot created", slog.String("database", db), slog.String("path", path))
	}

	return snapshots, nil
}

// Create creates a snapshot of a single database
func (s *SnapshotManager) Create(dbName string) (string, error) {
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	filename := fmt.Sprintf("%s_%s.dump", dbName, timestamp)
	path := filepath.Join(s.snapshotDir, filename)

	// pg_dump -Fc (custom format, compressed)
	// Using the database URI to construct the connection
	cmd := exec.Command("pg_dump",
		"-Fc",                    // Custom format (compressed)
		"-d", s.buildURI(dbName), // Database URI
		"-f", path, // Output file
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump failed: %w, output: %s", err, string(output))
	}

	// Set secure permissions
	if err := os.Chmod(path, 0600); err != nil {
		slog.Warn("failed to set snapshot permissions", slog.String("path", path), slog.String("error", err.Error()))
	}

	// Cloud sync async (don't block)
	if s.cloudSync && s.rcloneRemote != "" {
		go s.syncToCloud(path)
	}

	// Cleanup old snapshots
	s.cleanup(dbName)

	return path, nil
}

// buildURI constructs a database URI for the given database name
func (s *SnapshotManager) buildURI(dbName string) string {
	// Replace the database name in the URI
	// Assumes URI format: postgres://user:pass@host:port/dbname
	uri := s.dbURI

	// Find the last / and replace everything after it
	lastSlash := strings.LastIndex(uri, "/")
	if lastSlash != -1 {
		// Check if there are query params
		queryStart := strings.Index(uri[lastSlash:], "?")
		if queryStart != -1 {
			return uri[:lastSlash+1] + dbName + uri[lastSlash+queryStart:]
		}
		return uri[:lastSlash+1] + dbName
	}

	return uri
}

// syncToCloud syncs a snapshot to cloud storage via rclone
func (s *SnapshotManager) syncToCloud(localPath string) {
	cmd := exec.Command("rclone", "copy", localPath, s.rcloneRemote)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("cloud sync failed", slog.String("path", localPath), slog.String("error", err.Error()), slog.String("output", string(output)))
		return
	}
	slog.Info("snapshot synced to cloud", slog.String("path", localPath), slog.String("remote", s.rcloneRemote))
}

// cleanup removes old snapshots keeping only the most recent ones
func (s *SnapshotManager) cleanup(dbName string) {
	pattern := filepath.Join(s.snapshotDir, dbName+"_*.dump")
	files, err := filepath.Glob(pattern)
	if err != nil {
		slog.Warn("failed to list snapshots for cleanup", slog.String("pattern", pattern), slog.String("error", err.Error()))
		return
	}

	if len(files) <= s.retention {
		return
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Remove old files
	for _, file := range files[s.retention:] {
		if err := os.Remove(file); err != nil {
			slog.Warn("failed to remove old snapshot", slog.String("file", file), slog.String("error", err.Error()))
			continue
		}
		slog.Info("removed old snapshot", slog.String("file", file))
	}
}

// Restore restores a database from a snapshot
func (s *SnapshotManager) Restore(snapshotPath, dbName string) error {
	// pg_restore with clean option (drop objects before recreating)
	cmd := exec.Command("pg_restore",
		"-d", s.buildURI(dbName),
		"-c",          // Clean (drop) database objects before recreating
		"--if-exists", // Don't error if objects don't exist
		snapshotPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore failed: %w, output: %s", err, string(output))
	}

	slog.Info("database restored from snapshot", slog.String("database", dbName), slog.String("snapshot", snapshotPath))
	return nil
}

// ListSnapshots returns all snapshots for a database
func (s *SnapshotManager) ListSnapshots(dbName string) ([]string, error) {
	pattern := filepath.Join(s.snapshotDir, dbName+"_*.dump")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return files, nil
}

// GetLatestSnapshot returns the most recent snapshot for a database
func (s *SnapshotManager) GetLatestSnapshot(dbName string) (string, error) {
	snapshots, err := s.ListSnapshots(dbName)
	if err != nil {
		return "", err
	}
	if len(snapshots) == 0 {
		return "", fmt.Errorf("no snapshots found for database %s", dbName)
	}
	return snapshots[0], nil
}
