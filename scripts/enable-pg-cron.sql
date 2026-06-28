-- ============================================================================
-- ENABLE pg_cron EXTENSION
-- ============================================================================
-- This script enables pg_cron for automatic session cleanup.
-- Requirements:
--   1. PostgreSQL 10+ with pg_cron extension installed
--   2. Superuser privileges (or rds_superuser on AWS RDS)
--   3. pg_cron in shared_preload_libraries (postgresql.conf)
--
-- For AWS RDS:
--   - Add pg_cron to parameter group
--   - Reboot instance
--   - Run this script
--
-- For local Docker:
--   - Use image: postgres:15-bullseye (not alpine)
--   - Mount custom postgresql.conf with shared_preload_libraries = 'pg_cron'
--
-- Usage: psql -U postgres -d bro_db -f enable-pg-cron.sql
-- ============================================================================

-- Enable extension (requires superuser)
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Grant usage to bro user
GRANT USAGE ON SCHEMA cron TO bro;

-- Schedule cleanup job: every 6 hours
SELECT cron.schedule(
    'cleanup-expired-sessions',
    '0 */6 * * *',
    $$SELECT run_scheduled_cleanup()$$
);

-- Verify job was created
SELECT * FROM cron.job;

-- To manually run cleanup:
-- SELECT run_scheduled_cleanup();

-- To view job history:
-- SELECT * FROM cron.job_run_details ORDER BY start_time DESC LIMIT 10;

-- To unschedule:
-- SELECT cron.unschedule('cleanup-expired-sessions');
