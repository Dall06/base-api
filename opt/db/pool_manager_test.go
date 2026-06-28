package database

import (
	"context"
	"testing"
	"time"

	"base-api/env/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// setupMockPool creates a mock bun.DB for testing
func setupMockPool(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	db := bun.NewDB(mockDB, pgdialect.New())
	return db, mock
}

// TestNewTenantPoolManager tests the creation of a new pool manager
func TestNewTenantPoolManager(t *testing.T) {
	cfg := &config.Config{
		TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
		MaxPoolConnections:        10,
		MaxIdleConnections:        5,
		ConnectionMaxLifetime:     3600,
	}

	manager := NewTenantPoolManager(cfg)

	assert.NotNil(t, manager, "manager should not be nil")
	assert.NotNil(t, manager.pools, "pools map should be initialized")
	assert.Equal(t, cfg, manager.cfg, "config should be stored")
	assert.Equal(t, 0, len(manager.pools), "pools should start empty")
}

// TestCount tests the Count method with mock pools
func TestCount(t *testing.T) {
	tests := []struct {
		name          string
		setupPools    int
		expectedCount int
	}{
		{
			name:          "empty pool",
			setupPools:    0,
			expectedCount: 0,
		},
		{
			name:          "single pool",
			setupPools:    1,
			expectedCount: 1,
		},
		{
			name:          "multiple pools",
			setupPools:    5,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
				MaxPoolConnections:        10,
				MaxIdleConnections:        5,
				ConnectionMaxLifetime:     3600,
			}

			manager := NewTenantPoolManager(cfg)

			// Inject mock pools
			for i := 0; i < tt.setupPools; i++ {
				mockDB, _ := setupMockPool(t)
				manager.pools[string(rune('a'+i))] = mockDB
			}

			count := manager.Count()
			assert.Equal(t, tt.expectedCount, count, "count should match expected")
		})
	}
}

// TestHealth tests the Health method with mock pools
func TestHealth(t *testing.T) {
	tests := []struct {
		name            string
		setupPools      map[string]bool // slug -> ping success
		expectedHealthy map[string]bool
	}{
		{
			name:            "empty pool returns empty health map",
			setupPools:      map[string]bool{},
			expectedHealthy: map[string]bool{},
		},
		{
			name:            "single healthy pool",
			setupPools:      map[string]bool{"tenant-a": true},
			expectedHealthy: map[string]bool{"tenant-a": true},
		},
		{
			name:            "single unhealthy pool",
			setupPools:      map[string]bool{"tenant-b": false},
			expectedHealthy: map[string]bool{"tenant-b": false},
		},
		{
			name:            "mixed healthy and unhealthy pools",
			setupPools:      map[string]bool{"healthy": true, "unhealthy": false},
			expectedHealthy: map[string]bool{"healthy": true, "unhealthy": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
				MaxPoolConnections:        10,
				MaxIdleConnections:        5,
				ConnectionMaxLifetime:     3600,
			}

			manager := NewTenantPoolManager(cfg)
			ctx := context.Background()

			// Inject mock pools with ping expectations
			for slug, shouldSucceed := range tt.setupPools {
				mockDB, mock := setupMockPool(t)
				if shouldSucceed {
					mock.ExpectPing()
				} else {
					mock.ExpectPing().WillReturnError(assert.AnError)
				}
				manager.pools[slug] = mockDB
			}

			health := manager.Health(ctx)

			assert.NotNil(t, health, "health map should not be nil")
			assert.Equal(t, len(tt.expectedHealthy), len(health), "health map should have correct size")
			for slug, expected := range tt.expectedHealthy {
				assert.Equal(t, expected, health[slug], "health status for %s should be %v", slug, expected)
			}
		})
	}
}

// TestBuildTenantURI tests the buildTenantURI method
func TestBuildTenantURI(t *testing.T) {
	tests := []struct {
		name     string
		template string
		slug     string
		expected string
	}{
		{
			name:     "simple slug",
			template: "postgres://user:pass@localhost:5432/{slug}",
			slug:     "my-gym",
			expected: "postgres://user:pass@localhost:5432/gym_my_gym",
		},
		{
			name:     "slug with multiple hyphens",
			template: "postgres://user:pass@localhost:5432/{slug}",
			slug:     "my-super-gym",
			expected: "postgres://user:pass@localhost:5432/gym_my_super_gym",
		},
		{
			name:     "slug with numbers",
			template: "postgres://user:pass@localhost:5432/{slug}",
			slug:     "gym-123",
			expected: "postgres://user:pass@localhost:5432/gym_gym_123",
		},
		{
			name:     "simple slug no hyphens",
			template: "postgres://user:pass@localhost:5432/{slug}",
			slug:     "mygym",
			expected: "postgres://user:pass@localhost:5432/gym_mygym",
		},
		{
			name:     "slug at end of URI",
			template: "postgres://localhost/{slug}",
			slug:     "test-gym",
			expected: "postgres://localhost/gym_test_gym",
		},
		{
			name:     "slug with query params",
			template: "postgres://localhost/{slug}?sslmode=disable",
			slug:     "my-gym",
			expected: "postgres://localhost/gym_my_gym?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TenantDatabaseURITemplate: tt.template,
				MaxPoolConnections:        10,
				MaxIdleConnections:        5,
				ConnectionMaxLifetime:     3600,
			}

			manager := NewTenantPoolManager(cfg)
			uri := manager.buildTenantURI(tt.slug)

			assert.Equal(t, tt.expected, uri, "URI should match expected format")
		})
	}
}

// TestBuildTenantURIEdgeCases tests edge cases for buildTenantURI
func TestBuildTenantURIEdgeCases(t *testing.T) {
	t.Run("hyphen to underscore conversion", func(t *testing.T) {
		cfg := &config.Config{
			TenantDatabaseURITemplate: "postgres://localhost/{slug}",
		}
		manager := NewTenantPoolManager(cfg)

		slugs := map[string]string{
			"my-gym":        "gym_my_gym",
			"a-b-c":         "gym_a_b_c",
			"test-gym-2024": "gym_test_gym_2024",
		}

		for slug, expectedDB := range slugs {
			uri := manager.buildTenantURI(slug)
			assert.Contains(t, uri, expectedDB, "URI should contain %s for slug %s", expectedDB, slug)
		}
	})

	t.Run("gym prefix is added", func(t *testing.T) {
		cfg := &config.Config{
			TenantDatabaseURITemplate: "postgres://localhost/{slug}",
		}
		manager := NewTenantPoolManager(cfg)

		uri := manager.buildTenantURI("test")
		assert.Contains(t, uri, "gym_test", "URI should have gym_ prefix")
	})
}

// TestClose tests the Close method with mock pools
func TestClose(t *testing.T) {
	tests := []struct {
		name          string
		setupSlugs    []string
		closeSlug     string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "close non-existent pool returns nil",
			setupSlugs:    []string{},
			closeSlug:     "non-existent",
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "close existing pool",
			setupSlugs:    []string{"tenant-a"},
			closeSlug:     "tenant-a",
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "close one of multiple pools",
			setupSlugs:    []string{"tenant-a", "tenant-b", "tenant-c"},
			closeSlug:     "tenant-b",
			expectError:   false,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
				MaxPoolConnections:        10,
				MaxIdleConnections:        5,
				ConnectionMaxLifetime:     3600,
			}

			manager := NewTenantPoolManager(cfg)

			// Inject mock pools
			for _, slug := range tt.setupSlugs {
				mockDB, mock := setupMockPool(t)
				mock.ExpectClose()
				manager.pools[slug] = mockDB
			}

			err := manager.Close(tt.closeSlug)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCount, manager.Count(), "count should match expected")
		})
	}
}

// TestCloseAll tests the CloseAll method with mock pools
func TestCloseAll(t *testing.T) {
	tests := []struct {
		name       string
		setupSlugs []string
	}{
		{
			name:       "close all on empty pool",
			setupSlugs: []string{},
		},
		{
			name:       "close all with single pool",
			setupSlugs: []string{"tenant-a"},
		},
		{
			name:       "close all with multiple pools",
			setupSlugs: []string{"tenant-a", "tenant-b", "tenant-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
				MaxPoolConnections:        10,
				MaxIdleConnections:        5,
				ConnectionMaxLifetime:     3600,
			}

			manager := NewTenantPoolManager(cfg)

			// Inject mock pools with close expectations
			for _, slug := range tt.setupSlugs {
				mockDB, mock := setupMockPool(t)
				mock.ExpectClose()
				manager.pools[slug] = mockDB
			}

			// Verify initial count
			assert.Equal(t, len(tt.setupSlugs), manager.Count(), "initial count should match setup")

			err := manager.CloseAll()
			assert.NoError(t, err, "CloseAll should not error")

			// Verify all pools are closed
			assert.Equal(t, 0, manager.Count(), "count should be 0 after close all")
		})
	}
}

// TestConcurrentAccess tests thread-safety of Count, Health, Close and CloseAll
func TestConcurrentAccess(t *testing.T) {
	cfg := &config.Config{
		TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
		MaxPoolConnections:        10,
		MaxIdleConnections:        5,
		ConnectionMaxLifetime:     3600,
	}

	manager := NewTenantPoolManager(cfg)
	ctx := context.Background()

	// Add some mock pools for Health to check
	for i := 0; i < 3; i++ {
		mockDB, mock := setupMockPool(t)
		// Expect multiple pings from concurrent Health calls
		for j := 0; j < 15; j++ {
			mock.ExpectPing()
		}
		manager.pools[string(rune('a'+i))] = mockDB
	}

	// Test concurrent reads don't panic
	done := make(chan bool, 30)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.Count()
			done <- true
		}()
		go func() {
			_ = manager.Health(ctx)
			done <- true
		}()
		go func() {
			// Try to close non-existent pool (safe concurrent operation)
			_ = manager.Close("non-existent")
			done <- true
		}()
	}

	// Wait for all goroutines with timeout
	for i := 0; i < 30; i++ {
		select {
		case <-done:
			// success
		case <-time.After(5 * time.Second):
			t.Fatal("test timed out waiting for goroutine")
		}
	}

	assert.Equal(t, 3, manager.Count(), "count should still be 3")
}

// TestPoolManagerConfiguration tests that configuration is properly stored and used
func TestPoolManagerConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "default configuration",
			config: &config.Config{
				TenantDatabaseURITemplate: "postgres://localhost/{slug}",
				MaxPoolConnections:        10,
				MaxIdleConnections:        5,
				ConnectionMaxLifetime:     3600,
			},
		},
		{
			name: "custom configuration",
			config: &config.Config{
				TenantDatabaseURITemplate: "postgres://custom:5433/{slug}?sslmode=require",
				MaxPoolConnections:        20,
				MaxIdleConnections:        10,
				ConnectionMaxLifetime:     7200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTenantPoolManager(tt.config)

			assert.Equal(t, tt.config, manager.cfg, "config should be stored correctly")
			assert.NotNil(t, manager.pools, "pools map should be initialized")
			assert.Equal(t, 0, manager.Count(), "initial pool count should be 0")
		})
	}
}

// TestHealthEmptyPool specifically tests Health behavior with empty pool
func TestHealthEmptyPool(t *testing.T) {
	cfg := &config.Config{
		TenantDatabaseURITemplate: "postgres://localhost/{slug}",
		MaxPoolConnections:        10,
		MaxIdleConnections:        5,
		ConnectionMaxLifetime:     3600,
	}

	manager := NewTenantPoolManager(cfg)
	ctx := context.Background()

	health := manager.Health(ctx)

	assert.NotNil(t, health, "health should not be nil")
	assert.IsType(t, map[string]bool{}, health, "health should be a map[string]bool")
	assert.Empty(t, health, "health should be empty for empty pool")
}

// TestCountEmptyPool specifically tests Count behavior with empty pool
func TestCountEmptyPool(t *testing.T) {
	cfg := &config.Config{
		TenantDatabaseURITemplate: "postgres://localhost/{slug}",
		MaxPoolConnections:        10,
		MaxIdleConnections:        5,
		ConnectionMaxLifetime:     3600,
	}

	manager := NewTenantPoolManager(cfg)

	count := manager.Count()

	assert.Equal(t, 0, count, "count should be 0 for empty pool")
	assert.IsType(t, 0, count, "count should be an int")
}

// =============================================================================
// GetOrCreate Tests (Integration tests - require PostgreSQL)
// =============================================================================

// TestGetOrCreate_NewPool tests that GetOrCreate creates a new pool when it doesn't exist
func TestGetOrCreate_NewPool(t *testing.T) {
	t.Skip("Integration test: requires PostgreSQL database connection")

	// Documentación del test esperado:
	// 1. Crear TenantPoolManager con config válido
	// 2. Llamar GetOrCreate("new-tenant")
	// 3. Verificar que pool fue creado
	// 4. Verificar que Count() == 1
	//
	// Implementación esperada (cuando haya DB disponible):
	//
	// cfg := &config.Config{
	// 	TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
	// 	MaxPoolConnections:        10,
	// 	MaxIdleConnections:        5,
	// 	ConnectionMaxLifetime:     3600,
	// }
	//
	// manager := NewTenantPoolManager(cfg)
	// ctx := context.Background()
	//
	// pool, err := manager.GetOrCreate(ctx, "new-tenant")
	// assert.NoError(t, err, "should create pool without error")
	// assert.NotNil(t, pool, "pool should not be nil")
	// assert.Equal(t, 1, manager.Count(), "should have exactly 1 pool")
	//
	// // Verify pool is functional
	// err = pool.PingContext(ctx)
	// assert.NoError(t, err, "pool should be pingable")
	//
	// // Cleanup
	// manager.CloseAll()
}

// TestGetOrCreate_ExistingPool tests that GetOrCreate returns existing pool without creating a new one
func TestGetOrCreate_ExistingPool(t *testing.T) {
	t.Skip("Integration test: requires PostgreSQL database connection")

	// Documentación del test esperado:
	// 1. Crear pool para "existing-tenant"
	// 2. Llamar GetOrCreate("existing-tenant") segunda vez
	// 3. Verificar que retorna el mismo pool (fast path)
	// 4. Verificar que Count() sigue siendo 1
	//
	// Implementación esperada (cuando haya DB disponible):
	//
	// cfg := &config.Config{
	// 	TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
	// 	MaxPoolConnections:        10,
	// 	MaxIdleConnections:        5,
	// 	ConnectionMaxLifetime:     3600,
	// }
	//
	// manager := NewTenantPoolManager(cfg)
	// ctx := context.Background()
	//
	// // Create first pool
	// pool1, err := manager.GetOrCreate(ctx, "existing-tenant")
	// assert.NoError(t, err, "should create first pool without error")
	// assert.NotNil(t, pool1, "first pool should not be nil")
	//
	// // Get same pool again (should use fast path)
	// pool2, err := manager.GetOrCreate(ctx, "existing-tenant")
	// assert.NoError(t, err, "should get existing pool without error")
	// assert.NotNil(t, pool2, "second pool should not be nil")
	//
	// // Verify it's the same instance (pointer comparison)
	// assert.Equal(t, pool1, pool2, "should return same pool instance")
	// assert.Equal(t, 1, manager.Count(), "should still have exactly 1 pool")
	//
	// // Cleanup
	// manager.CloseAll()
}

// TestGetOrCreate_ConcurrentAccess tests that concurrent access creates only one pool
func TestGetOrCreate_ConcurrentAccess(t *testing.T) {
	t.Skip("Integration test: requires PostgreSQL database connection")

	// Documentación del test esperado:
	// 1. Lanzar 10 goroutines llamando GetOrCreate("same-slug")
	// 2. Esperar con WaitGroup
	// 3. Verificar que Count() == 1 (solo 1 pool creado)
	// 4. Verificar con -race flag
	//
	// Implementación esperada (cuando haya DB disponible):
	//
	// cfg := &config.Config{
	// 	TenantDatabaseURITemplate: "postgres://user:pass@localhost:5432/{slug}",
	// 	MaxPoolConnections:        10,
	// 	MaxIdleConnections:        5,
	// 	ConnectionMaxLifetime:     3600,
	// }
	//
	// manager := NewTenantPoolManager(cfg)
	// ctx := context.Background()
	//
	// const numGoroutines = 10
	// var wg sync.WaitGroup
	// wg.Add(numGoroutines)
	//
	// // Launch concurrent GetOrCreate calls
	// for i := 0; i < numGoroutines; i++ {
	// 	go func() {
	// 		defer wg.Done()
	// 		pool, err := manager.GetOrCreate(ctx, "same-slug")
	// 		assert.NoError(t, err, "concurrent GetOrCreate should not error")
	// 		assert.NotNil(t, pool, "concurrent GetOrCreate should return pool")
	// 	}()
	// }
	//
	// // Wait for all goroutines to complete
	// wg.Wait()
	//
	// // Verify only one pool was created (double-check locking worked)
	// assert.Equal(t, 1, manager.Count(), "should create only 1 pool despite concurrent access")
	//
	// // Cleanup
	// manager.CloseAll()
}

// TestGetOrCreate_InvalidConfig tests error handling when connection fails
func TestGetOrCreate_InvalidConfig(t *testing.T) {
	t.Skip("Integration test: requires PostgreSQL database connection")

	// Documentación del test esperado:
	// 1. Crear TenantPoolManager con config inválido
	// 2. Llamar GetOrCreate con cualquier slug
	// 3. Verificar que retorna error errs.InternalError("connection failed")
	//
	// Implementación esperada (cuando haya DB disponible):
	//
	// cfg := &config.Config{
	// 	TenantDatabaseURITemplate: "postgres://invalid:9999/{slug}",
	// 	MaxPoolConnections:        10,
	// 	MaxIdleConnections:        5,
	// 	ConnectionMaxLifetime:     3600,
	// }
	//
	// manager := NewTenantPoolManager(cfg)
	// ctx := context.Background()
	//
	// pool, err := manager.GetOrCreate(ctx, "test-tenant")
	// assert.Error(t, err, "should return error for invalid connection")
	// assert.Nil(t, pool, "pool should be nil on error")
	// assert.ErrorIs(t, err, errs.InternalError("connection failed"), "should return ErrConnectionFailed")
	// assert.Equal(t, 0, manager.Count(), "failed pool should not be stored")
}
