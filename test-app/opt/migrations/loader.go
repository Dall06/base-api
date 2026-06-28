package migrations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Loader loads migration files from disk
type Loader struct {
	baseDir string
}

// NewLoader creates a new migration loader
func NewLoader(baseDir string) *Loader {
	return &Loader{baseDir: baseDir}
}

// LoadMigrations loads all migrations for a specific DB type
func (l *Loader) LoadMigrations(dbType DBType) ([]Migration, error) {
	dir := filepath.Join(l.baseDir, string(dbType))

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory not found: %s", dir)
	}

	// Find all .sql files
	pattern := filepath.Join(dir, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %w", err)
	}

	migrations := make([]Migration, 0, len(files))

	for _, file := range files {
		migration, err := l.loadMigrationFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load migration %s: %w", file, err)
		}
		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// loadMigrationFile loads a single migration file
func (l *Loader) loadMigrationFile(path string) (Migration, error) {
	content, err := os.ReadFile(path) // #nosec G304 -- migration path from config, not user input
	if err != nil {
		return Migration{}, err
	}

	filename := filepath.Base(path)
	version, name := parseFilename(filename)

	// Calculate checksum
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])

	// Extract description from comment header
	description := extractDescription(string(content))

	return Migration{
		Version:     version,
		Name:        name,
		Filename:    filename,
		SQL:         string(content),
		Checksum:    checksum,
		Description: description,
	}, nil
}

// parseFilename extracts version and name from filename
// Format: 001_add_logo.sql -> version="001", name="add_logo"
func parseFilename(filename string) (version, name string) {
	// Remove .sql extension
	base := strings.TrimSuffix(filename, ".sql")

	// Split by first underscore
	parts := strings.SplitN(base, "_", 2)

	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	// Fallback if no underscore
	return base, base
}

// extractDescription extracts description from SQL comment header
// Looks for: -- Description: Some description here
func extractDescription(sql string) string {
	re := regexp.MustCompile(`(?i)--\s*Description:\s*(.+)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// GetMigration returns a specific migration by version
func (l *Loader) GetMigration(dbType DBType, version string) (*Migration, error) {
	migrations, err := l.LoadMigrations(dbType)
	if err != nil {
		return nil, err
	}

	for _, m := range migrations {
		if m.Version == version {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("migration %s not found", version)
}

// ValidateMigrations checks that migration files are valid
func (l *Loader) ValidateMigrations(dbType DBType) error {
	migrations, err := l.LoadMigrations(dbType)
	if err != nil {
		return err
	}

	seen := make(map[string]bool)
	for _, m := range migrations {
		// Check for duplicate versions
		if seen[m.Version] {
			return fmt.Errorf("duplicate migration version: %s", m.Version)
		}
		seen[m.Version] = true

		// Check for empty SQL
		if strings.TrimSpace(m.SQL) == "" {
			return fmt.Errorf("migration %s has empty SQL", m.Filename)
		}
	}

	return nil
}
