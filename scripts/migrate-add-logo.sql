-- ============================================================================
-- Migration: Add logo fields to company table
-- Run this on existing tenant databases to add logo support
-- ============================================================================

-- Add logo columns if they don't exist
ALTER TABLE company ADD COLUMN IF NOT EXISTS logo_type VARCHAR(10);
ALTER TABLE company ADD COLUMN IF NOT EXISTS logo_value TEXT;
ALTER TABLE company ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Create trigger for updated_at if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_company_ts') THEN
        CREATE TRIGGER trg_company_ts BEFORE UPDATE ON company
        FOR EACH ROW EXECUTE FUNCTION update_timestamp();
    END IF;
END$$;
