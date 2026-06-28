-- ============================================================================
-- BRO PLATFORM (bro_db)
-- ============================================================================
-- Central database for SaaS management
-- Tables: saas_plans, plan_features, companies, users, staff, subscriptions, payments
-- User-Staff separation: users have globally unique emails, staff links users to companies
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- TABLES
-- ============================================================================

CREATE TABLE saas_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    monthly_price DECIMAL(10,2) NOT NULL,
    price_3months DECIMAL(10,2),
    price_6months DECIMAL(10,2),
    yearly_price DECIMAL(10,2),
    yearly_cash_price DECIMAL(10,2),
    price_forever DECIMAL(10,2),
    max_members INTEGER,
    max_staff INTEGER,
    max_classes INTEGER,
    max_membership_plans INTEGER,
    max_instructors INTEGER,
    max_products INTEGER,
    max_individual_sessions INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE plan_features (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES saas_plans(id) ON DELETE CASCADE,
    feature_key VARCHAR(50) NOT NULL,
    feature_value VARCHAR(255),
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (plan_id, feature_key)
);

CREATE TABLE plan_payment_methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES saas_plans(id) ON DELETE CASCADE,
    method VARCHAR(20) NOT NULL
        CHECK (method IN ('stripe', 'transfer', 'cash')),
    billing_cycle VARCHAR(20) NOT NULL
        CHECK (billing_cycle IN ('monthly', '3months', '6months', 'yearly', 'yearly_cash', 'forever')),
    duration VARCHAR(20) NOT NULL
        CHECK (duration IN ('monthly', '3months', '6months', 'yearly', 'forever')),
    price DECIMAL(10,2) NOT NULL,
    price_per_month DECIMAL(10,2),
    installments INTEGER,
    per_installment DECIMAL(10,2),
    provider VARCHAR(50) NOT NULL DEFAULT 'stripe',
    label VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'coming_soon', 'disabled')),
    metadata JSONB,
    UNIQUE (plan_id, method, billing_cycle)
);

CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    owner_name VARCHAR(255) NOT NULL,
    owner_email VARCHAR(255) NOT NULL,
    owner_phone VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'suspended', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Users table: central identity management
-- One email = one user, can have multiple staff entries at different companies
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Staff table: links users to companies with roles
-- A user can be staff at multiple companies (like Slack workspaces)
CREATE TABLE staff (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'staff'
        CHECK (role IN ('god', 'owner', 'admin', 'staff')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- A user can only have one staff entry per company
    UNIQUE (user_id, company_id)
);

-- Password resets reference users (not staff) since passwords are per-user
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_resets_token ON password_resets(token_hash);
CREATE INDEX idx_password_resets_user ON password_resets(user_id);

-- User sessions: multiple sessions per user allowed
-- Sessions are per-user (not per-staff) since auth is user-based
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_jti UUID NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_jti ON user_sessions(token_jti);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plan_id UUID REFERENCES saas_plans(id) ON DELETE RESTRICT,
    plan_name VARCHAR(100),
    price DECIMAL(10,2),
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly'
        CHECK (billing_cycle IN ('monthly', '3months', '6months', 'yearly', 'yearly_cash', 'forever')),
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('trial', 'active', 'past_due', 'cancelled', 'suspended', 'pending_payment')),
    invitation_code_id UUID,
    grace_period_end TIMESTAMPTZ,
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    stripe_payment_intent_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'mxn',
    method VARCHAR(20),
    period_start DATE,
    period_end DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'succeeded', 'failed', 'refunded')),
    paid_at TIMESTAMPTZ,
    reference VARCHAR(255),
    notes TEXT,
    stripe_payment_intent_id VARCHAR(255),
    stripe_invoice_id VARCHAR(255),
    stripe_customer_id VARCHAR(255),
    billing_cycle VARCHAR(20),
    plan_name VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_event_id VARCHAR(255) NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload JSONB
);

CREATE TABLE invitation_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code UUID NOT NULL UNIQUE DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES saas_plans(id) ON DELETE RESTRICT,
    billing_cycle VARCHAR(20) NOT NULL
        CHECK (billing_cycle IN ('monthly', '3months', '6months', 'yearly', 'yearly_cash', 'forever')),
    is_used BOOLEAN NOT NULL DEFAULT false,
    used_by_company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add FK from subscriptions to invitation_codes (after both tables exist)
ALTER TABLE subscriptions
    ADD CONSTRAINT fk_subscription_invitation_code
    FOREIGN KEY (invitation_code_id) REFERENCES invitation_codes(id) ON DELETE SET NULL;

-- Notification settings per company
CREATE TABLE notification_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE UNIQUE,
    whatsapp_enabled BOOLEAN NOT NULL DEFAULT false,
    whatsapp_session_id VARCHAR(100),
    whatsapp_connected BOOLEAN NOT NULL DEFAULT false,
    email_enabled BOOLEAN NOT NULL DEFAULT false,
    smtp_host VARCHAR(255),
    smtp_port INTEGER,
    smtp_user VARCHAR(255),
    smtp_password_encrypted TEXT,
    from_email VARCHAR(255),
    whatsapp_template TEXT,
    email_template TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Notification logs
CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    member_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'email')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('sent', 'failed', 'pending')),
    error_message TEXT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contact requests (support/contact forms)
CREATE TABLE contact_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    folio VARCHAR(20) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL DEFAULT 'other'
        CHECK (type IN ('bug', 'question', 'request', 'billing', 'feedback', 'other')),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'landing'
        CHECK (source IN ('landing', 'app')),
    status VARCHAR(20) NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'in_progress', 'on_hold', 'closed')),
    trello_card_id VARCHAR(50),
    trello_card_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sequence for folio generation (BRO-YYYY-NNNNN)
CREATE SEQUENCE contact_folio_seq START 1;

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_saas_plans_active ON saas_plans(is_active) WHERE is_active = true;
CREATE INDEX idx_plan_features_plan ON plan_features(plan_id);

CREATE INDEX idx_plan_payment_methods_plan ON plan_payment_methods(plan_id);

CREATE INDEX idx_companies_slug ON companies(slug);
CREATE INDEX idx_companies_status ON companies(status);

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_staff_user ON staff(user_id);
CREATE INDEX idx_staff_company ON staff(company_id);
CREATE INDEX idx_staff_user_company ON staff(user_id, company_id);

CREATE INDEX idx_subscriptions_company ON subscriptions(company_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

CREATE INDEX idx_payments_company ON payments(company_id);
CREATE INDEX idx_payments_status ON payments(status);

CREATE INDEX idx_invitation_codes_code ON invitation_codes(code);
CREATE INDEX idx_invitation_codes_unused ON invitation_codes(is_used) WHERE is_used = false;
CREATE INDEX idx_invitation_codes_plan ON invitation_codes(plan_id);

CREATE INDEX idx_notification_settings_company ON notification_settings(company_id);
CREATE INDEX idx_notification_logs_company ON notification_logs(company_id);
CREATE INDEX idx_notification_logs_member ON notification_logs(member_id);

CREATE INDEX idx_contact_requests_folio ON contact_requests(folio);
CREATE INDEX idx_contact_requests_email ON contact_requests(email);
CREATE INDEX idx_contact_requests_type ON contact_requests(type);
CREATE INDEX idx_contact_requests_status ON contact_requests(status);
CREATE INDEX idx_contact_requests_created ON contact_requests(created_at);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_saas_plans_ts BEFORE UPDATE ON saas_plans FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_companies_ts BEFORE UPDATE ON companies FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_users_ts BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_staff_ts BEFORE UPDATE ON staff FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_subscriptions_ts BEFORE UPDATE ON subscriptions FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_contact_requests_ts BEFORE UPDATE ON contact_requests FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Security trigger: suspend company when invitation code is deleted
CREATE OR REPLACE FUNCTION suspend_company_on_code_delete()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE subscriptions
    SET status = 'suspended'
    WHERE invitation_code_id = OLD.id;

    UPDATE companies
    SET status = 'suspended'
    WHERE id = OLD.used_by_company_id;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_suspend_on_code_delete
BEFORE DELETE ON invitation_codes
FOR EACH ROW
WHEN (OLD.is_used = true)
EXECUTE FUNCTION suspend_company_on_code_delete();

-- ============================================================================
-- SCHEDULED CLEANUP (pg_cron)
-- ============================================================================
-- Cleanup function for expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RAISE NOTICE 'Cleaned up % expired sessions', deleted_count;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Cleanup function for expired password reset tokens
CREATE OR REPLACE FUNCTION cleanup_expired_password_resets()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM password_resets WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RAISE NOTICE 'Cleaned up % expired password resets', deleted_count;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Combined cleanup function
CREATE OR REPLACE FUNCTION run_scheduled_cleanup()
RETURNS void AS $$
BEGIN
    PERFORM cleanup_expired_sessions();
    PERFORM cleanup_expired_password_resets();
END;
$$ LANGUAGE plpgsql;

-- pg_cron job (only runs if pg_cron extension is available)
-- Schedule: Every 6 hours
DO $$
BEGIN
    -- Check if pg_cron is available
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        -- Remove existing job if any
        PERFORM cron.unschedule('cleanup-expired-sessions');
        -- Schedule new job: run every 6 hours
        PERFORM cron.schedule(
            'cleanup-expired-sessions',
            '0 */6 * * *',  -- At minute 0 past every 6th hour
            'SELECT run_scheduled_cleanup()'
        );
        RAISE NOTICE 'pg_cron job scheduled for session cleanup';
    ELSE
        RAISE NOTICE 'pg_cron not available - cleanup must be triggered manually or via application';
    END IF;
END $$;

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Pro plan with duration-based pricing (unlimited everything)
INSERT INTO saas_plans (id, name, description, monthly_price, price_3months, price_6months, yearly_price, yearly_cash_price, price_forever, max_members, max_staff, max_classes, max_membership_plans, max_instructors, max_products, max_individual_sessions, is_active) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'pro', 'Plan completo para gimnasios - todas las funcionalidades incluidas, sin limites', 439.00, 1197.00, 2154.00, 3828.00, 3828.00, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, true);

-- FREE plan with limited features (acquisition hook)
INSERT INTO saas_plans (id, name, description, monthly_price, price_3months, price_6months, yearly_price, yearly_cash_price, price_forever, max_members, max_staff, max_classes, max_membership_plans, max_instructors, max_products, max_individual_sessions, is_active) VALUES
('00000000-0000-0000-0000-000000000001', 'free', 'Plan gratuito con funcionalidades limitadas - perfecto para comenzar', 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 12, 2, 4, 3, 2, 20, 8, true);

-- Desktop plan (one-time purchase, coming soon)
INSERT INTO saas_plans (id, name, description, monthly_price, price_3months, price_6months, yearly_price, yearly_cash_price, price_forever, max_members, max_staff, max_classes, max_membership_plans, max_instructors, max_products, max_individual_sessions, is_active) VALUES
('00000000-0000-0000-0000-000000000002', 'desktop', 'Licencia de escritorio - pago unico, sin limite de tiempo', 8199.00, NULL, NULL, NULL, NULL, 8199.00, NULL, NULL, NULL, NULL, NULL, NULL, NULL, false);

-- Pro plan features (all enabled)
INSERT INTO plan_features (plan_id, feature_key, is_enabled) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'qr_verification', true),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'reports', true),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'api_access', true),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'unlimited_members', true),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'unlimited_staff', true);

-- FREE plan features (limited)
INSERT INTO plan_features (plan_id, feature_key, is_enabled) VALUES
('00000000-0000-0000-0000-000000000001', 'qr_verification', true),
('00000000-0000-0000-0000-000000000001', 'reports', false),
('00000000-0000-0000-0000-000000000001', 'api_access', false),
('00000000-0000-0000-0000-000000000001', 'unlimited_members', false),
('00000000-0000-0000-0000-000000000001', 'unlimited_staff', false);

-- Desktop plan features (same as pro)
INSERT INTO plan_features (plan_id, feature_key, is_enabled) VALUES
('00000000-0000-0000-0000-000000000002', 'qr_verification', true),
('00000000-0000-0000-0000-000000000002', 'reports', true),
('00000000-0000-0000-0000-000000000002', 'api_access', true),
('00000000-0000-0000-0000-000000000002', 'unlimited_members', true),
('00000000-0000-0000-0000-000000000002', 'unlimited_staff', true);

-- ============================================================================
-- PLAN PAYMENT METHODS
-- ============================================================================

-- Pro plan: Stripe payment methods
INSERT INTO plan_payment_methods (plan_id, method, billing_cycle, duration, price, price_per_month, provider, label, status) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'stripe', 'monthly',  'monthly',  439.00,  439.00, 'stripe', '1 Mes',    'active'),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'stripe', '3months',  '3months',  1197.00, 399.00, 'stripe', '3 Meses',  'active'),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'stripe', '6months',  '6months',  2154.00, 359.00, 'stripe', '6 Meses',  'active'),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'stripe', 'yearly',   'yearly',   3828.00, 319.00, 'stripe', '12 Meses', 'active');

-- Pro plan: Transfer payment methods with bank metadata
INSERT INTO plan_payment_methods (plan_id, method, billing_cycle, duration, price, price_per_month, provider, label, status, metadata) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'transfer', 'monthly',  'monthly',  439.00,  439.00, 'transfer', '1 Mes',    'active', '{"bank_name":"Banregio","clabe":"058210000148945474","beneficiary":"Diego A. Leon","account_number":"177959680016"}'::jsonb),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'transfer', '3months',  '3months',  1197.00, 399.00, 'transfer', '3 Meses',  'active', '{"bank_name":"Banregio","clabe":"058210000148945474","beneficiary":"Diego A. Leon","account_number":"177959680016"}'::jsonb),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'transfer', '6months',  '6months',  2154.00, 359.00, 'transfer', '6 Meses',  'active', '{"bank_name":"Banregio","clabe":"058210000148945474","beneficiary":"Diego A. Leon","account_number":"177959680016"}'::jsonb),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'transfer', 'yearly',   'yearly',   3828.00, 319.00, 'transfer', '12 Meses', 'active', '{"bank_name":"Banregio","clabe":"058210000148945474","beneficiary":"Diego A. Leon","account_number":"177959680016"}'::jsonb);

-- Desktop plan: Stripe (coming soon)
INSERT INTO plan_payment_methods (plan_id, method, billing_cycle, duration, price, provider, label, status) VALUES
('00000000-0000-0000-0000-000000000002', 'stripe', 'forever', 'forever', 8199.00, 'stripe', 'Pago unico', 'coming_soon');

-- ============================================================================
-- DEMO COMPANY AND TEST USERS
-- ============================================================================

-- Demo company for testing
INSERT INTO companies (id, slug, name, email, phone, owner_name, owner_email, owner_phone, status) VALUES
('4cb362d6-2018-428d-b758-e024a316f85b', 'demo', 'Demo Gym', 'demo@bro.local', '5551234567', 'Demo Owner', 'owner@bro.local', '5551234567', 'active');

-- ============================================================================
-- TEST USERS (identity - globally unique emails)
-- ============================================================================
-- All passwords: "password" (bcrypt cost 10)

-- God user
INSERT INTO users (id, email, password_hash, name, phone) VALUES
('11111111-aaaa-bbbb-cccc-111111111111', 'god@bro.local', '$2a$10$JXhjHLi17FWZY9im.rAiwejH/5iSHfMbQcMR2hZlO8PmJ7lSA3hEm', 'God User', '5551111111');

-- Owner user
INSERT INTO users (id, email, password_hash, name, phone) VALUES
('22222222-aaaa-bbbb-cccc-222222222222', 'owner@bro.local', '$2a$10$JXhjHLi17FWZY9im.rAiwejH/5iSHfMbQcMR2hZlO8PmJ7lSA3hEm', 'Owner User', '5553333333');

-- Admin user
INSERT INTO users (id, email, password_hash, name, phone) VALUES
('33333333-aaaa-bbbb-cccc-333333333333', 'admin@bro.local', '$2a$10$JXhjHLi17FWZY9im.rAiwejH/5iSHfMbQcMR2hZlO8PmJ7lSA3hEm', 'Admin User', '5552222222');

-- Staff user
INSERT INTO users (id, email, password_hash, name, phone) VALUES
('44444444-aaaa-bbbb-cccc-444444444444', 'staff@bro.local', '$2a$10$JXhjHLi17FWZY9im.rAiwejH/5iSHfMbQcMR2hZlO8PmJ7lSA3hEm', 'Staff User', '5554444444');

-- ============================================================================
-- STAFF ENTRIES (link users to companies with roles)
-- ============================================================================

-- God user -> Demo Gym (god role)
INSERT INTO staff (id, user_id, company_id, name, phone, role, is_active) VALUES
('68188c0c-1e36-4ea0-ba2d-0ead03749aec', '11111111-aaaa-bbbb-cccc-111111111111', '4cb362d6-2018-428d-b758-e024a316f85b', 'God User', '5551111111', 'god', true);

-- Owner user -> Demo Gym (owner role)
INSERT INTO staff (id, user_id, company_id, name, phone, role, is_active) VALUES
('58188c0c-1e36-4ea0-ba2d-0ead03749aed', '22222222-aaaa-bbbb-cccc-222222222222', '4cb362d6-2018-428d-b758-e024a316f85b', 'Owner User', '5553333333', 'owner', true);

-- Admin user -> Demo Gym (admin role)
INSERT INTO staff (id, user_id, company_id, name, phone, role, is_active) VALUES
('78288c1c-2e46-4ea1-ca3d-1ead13850bfd', '33333333-aaaa-bbbb-cccc-333333333333', '4cb362d6-2018-428d-b758-e024a316f85b', 'Admin User', '5552222222', 'admin', true);

-- Staff user -> Demo Gym (staff role)
INSERT INTO staff (id, user_id, company_id, name, phone, role, is_active) VALUES
('88388c2c-3e56-4ea2-da4d-2ead23960cfe', '44444444-aaaa-bbbb-cccc-444444444444', '4cb362d6-2018-428d-b758-e024a316f85b', 'Staff User', '5554444444', 'staff', true);

-- Demo company subscription (pro plan, forever billing)
INSERT INTO subscriptions (id, company_id, plan_id, price, billing_cycle, start_date, end_date, status) VALUES
('11111111-1111-1111-1111-111111111111', '4cb362d6-2018-428d-b758-e024a316f85b', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 0.00, 'forever', CURRENT_DATE, NULL, 'active');
