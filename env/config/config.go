package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port              int
	Env               string // development, production
	LogLevel          string
	DatabaseURI       string
	JWTSecret         string
	StartupTimeoutSec int

	// Sigil (service-to-service authentication)
	SigilSecret     string // Shared HMAC secret for signing requests
	AllowedServices string // Comma-separated list of allowed service IDs (e.g., "gateway")

	// NATS messaging
	NatsURL string // NATS server URL (e.g., nats://localhost:4222)

	// Tenant provisioning configuration
	AdminDatabaseURI          string // DSN with CREATE DATABASE permissions
	TenantDatabaseURITemplate string // Template: postgres://.../{slug}?...
	TenantSchemaPath          string // Path to schema-tenant.sql

	// Connection pool settings
	MaxPoolConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime int // seconds

	// Frontend URL
	AppURL string

	// CORS: comma-separated allowed origins (e.g., "https://brocrm.com,https://stg.brocrm.com")
	AllowedOrigins string

	// Email configuration (platform-wide)
	// Resend (preferred)
	ResendAPIKey string
	FromEmail    string

	// SMTP configuration (legacy fallback)
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	// Platform DB (read-only access for pie service to resolve plan prices)
	PlatformDatabaseURI string

	// Stripe configuration
	StripeSecretKey     string
	StripeWebhookSecret string

	// bro-pie (payment service) URL for subscription status checks
	PieURL string

	// LLM configuration (for agent service).
	// LLMProvider selects between "qwen" (Alibaba DashScope, OpenAI-compatible
	// remote API) and "ollama" (self-hosted local inference). When set to
	// "qwen" the agent calls LLMAPIURL with LLMAPIKey as Bearer token; when
	// set to "ollama" or empty it falls back to OllamaURL/OllamaModel.
	LLMProvider string
	LLMAPIURL   string
	LLMAPIKey   string
	LLMModel    string

	// Ollama configuration (legacy / fallback for agent service)
	OllamaURL   string
	OllamaModel string

	// Trello configuration (for agent service)
	TrelloAPIKey   string
	TrelloAPIToken string
	TrelloBoardID  string
	TrelloListID   string

	// WhatsApp (Meta Cloud API)
	WhatsAppPhoneNumberID string // Meta phone number ID
	WhatsAppAccessToken   string // System-user permanent token

	// Contact form rate limiting (agent service). Both windows are fixed
	// (1 minute for IP, 1 hour for email) — only the counts are tunable.
	ContactRateLimitPerIP    int // max POST /contact per minute per remote IP
	ContactRateLimitPerEmail int // max POST /contact per hour per email address

	// Field encryption (AES-256-GCM + HMAC for blind indexing)
	EncryptionKey string // 32 bytes hex-encoded for AES-256
	HMACKey       string // 32+ bytes hex-encoded for blind index

	// Migrations
	MigrationsDir     string // Path to migrations directory
	SnapshotDir       string // Path to snapshots directory
	SnapshotRetention int    // Number of snapshots to keep per DB
	MigrateOnStartup  bool   // Whether to run migrations on startup
	SnapshotCloudSync bool   // Whether to sync snapshots to cloud
	RcloneRemote      string // rclone remote path (e.g., "gdrive:bro-backups")
	MigratePilotSlug  string // Pilot tenant slug for canary migrations
	AutoInitBaseline  bool   // Auto-init baseline for existing DBs without _migrations

	// Cloudflare R2 (S3-compatible object storage for product images)
	R2Endpoint        string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string
}

// Load reads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %w", err)
	}

	timeoutSec, err := strconv.Atoi(getEnv("STARTUP_TIMEOUT_SEC", "30"))
	if err != nil {
		return nil, fmt.Errorf("invalid STARTUP_TIMEOUT_SEC value: %w", err)
	}

	maxPoolConns, err := strconv.Atoi(getEnv("MAX_POOL_CONNECTIONS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_POOL_CONNECTIONS value: %w", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnv("MAX_IDLE_CONNECTIONS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_IDLE_CONNECTIONS value: %w", err)
	}

	connMaxLifetime, err := strconv.Atoi(getEnv("CONNECTION_MAX_LIFETIME", "3600"))
	if err != nil {
		return nil, fmt.Errorf("invalid CONNECTION_MAX_LIFETIME value: %w", err)
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	snapshotRetention, _ := strconv.Atoi(getEnv("SNAPSHOT_RETENTION", "5"))
	contactRateLimitIP, _ := strconv.Atoi(getEnv("CONTACT_RATE_LIMIT_PER_IP", "10"))
	contactRateLimitEmail, _ := strconv.Atoi(getEnv("CONTACT_RATE_LIMIT_PER_EMAIL", "25"))
	migrateOnStartup := getEnv("MIGRATE_ON_STARTUP", "true") == "true"
	snapshotCloudSync := getEnv("SNAPSHOT_CLOUD_SYNC", "false") == "true"
	autoInitBaseline := getEnv("AUTO_INIT_BASELINE", "true") == "true"

	cfg := &Config{
		Port:              port,
		Env:               getEnv("ENV", "development"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		DatabaseURI:       getEnv("DATABASE_URI", ""),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		StartupTimeoutSec: timeoutSec,

		// Sigil
		SigilSecret:     getEnv("SIGIL_SECRET", ""),
		AllowedServices: getEnv("ALLOWED_SERVICES", "gateway"),

		// NATS
		NatsURL: getEnv("NATS_URL", ""),

		// Tenant provisioning
		AdminDatabaseURI:          getEnv("ADMIN_DATABASE_URI", ""),
		TenantDatabaseURITemplate: getEnv("TENANT_DATABASE_URI_TEMPLATE", ""),
		TenantSchemaPath:          getEnv("TENANT_SCHEMA_PATH", ""),

		// Connection pool
		MaxPoolConnections:    maxPoolConns,
		MaxIdleConnections:    maxIdleConns,
		ConnectionMaxLifetime: connMaxLifetime,

		// Frontend
		AppURL:         getEnv("APP_URL", "http://localhost:5173"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:5173"),

		// Email (Resend preferred, SMTP fallback)
		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		FromEmail:    getEnv("FROM_EMAIL", "BRO <onboarding@resend.dev>"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		// Platform DB (pie reads plan prices from bro_db)
		PlatformDatabaseURI: getEnv("PLATFORM_DATABASE_URI", ""),

		// Stripe
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),

		// bro-pie
		PieURL: getEnv("PIE_URL", ""),

		// LLM (agent service). Default to "ollama" when unset so existing
		// dev setups without a remote API key still work.
		LLMProvider: getEnv("LLM_PROVIDER", "ollama"),
		LLMAPIURL:   getEnv("LLM_API_URL", ""),
		LLMAPIKey:   getEnv("LLM_API_KEY", ""),
		LLMModel:    getEnv("LLM_MODEL", ""),

		// Ollama (legacy / fallback agent service)
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "qwen2.5:3b"),

		// Trello (agent service)
		TrelloAPIKey:   getEnv("TRELLO_API_KEY", ""),
		TrelloAPIToken: getEnv("TRELLO_API_TOKEN", ""),
		TrelloBoardID:  getEnv("TRELLO_BOARD_ID", ""),
		TrelloListID:   getEnv("TRELLO_LIST_ID", ""),

		// WhatsApp (Meta Cloud API)
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppAccessToken:   getEnv("WHATSAPP_ACCESS_TOKEN", ""),

		// Contact form rate limiting (agent service)
		ContactRateLimitPerIP:    contactRateLimitIP,
		ContactRateLimitPerEmail: contactRateLimitEmail,

		// Field encryption
		EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
		HMACKey:       getEnv("HMAC_KEY", ""),

		// Migrations
		MigrationsDir:     getEnv("MIGRATIONS_DIR", "/app/migrations"),
		SnapshotDir:       getEnv("SNAPSHOT_DIR", "/opt/bro/snapshots"),
		SnapshotRetention: snapshotRetention,
		MigrateOnStartup:  migrateOnStartup,
		SnapshotCloudSync: snapshotCloudSync,
		RcloneRemote:      getEnv("RCLONE_REMOTE", ""),
		MigratePilotSlug:  getEnv("MIGRATE_PILOT_SLUG", ""),
		AutoInitBaseline:  autoInitBaseline,

		// Cloudflare R2
		R2Endpoint:        getEnv("R2_ENDPOINT", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", ""),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),
	}

	if cfg.DatabaseURI == "" {
		return nil, fmt.Errorf("DATABASE_URI is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Require SIGIL_SECRET in non-development environments
	if cfg.Env != "development" && cfg.SigilSecret == "" {
		return nil, fmt.Errorf("SIGIL_SECRET is required in non-development environments")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}
