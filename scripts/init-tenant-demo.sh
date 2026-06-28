#!/bin/sh
# ============================================================================
# Create demo tenant database (gym_demo)
# This script runs after init-db.sql to create the demo tenant database
# Uses /schema-tenant.sql mounted from bro-ops/db/schema-tenant.sql
# ============================================================================

set -e

echo "Creating demo tenant database (gym_demo)..."

# Create the gym_demo database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname=postgres <<-EOSQL
    CREATE DATABASE gym_demo;
EOSQL

# Apply tenant schema from mounted file
echo "Applying tenant schema..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname=gym_demo -f /schema-tenant.sql

# Insert seed data
echo "Inserting seed data..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname=gym_demo <<-'EOSQL'
-- SEED DATA: Company info for demo tenant
INSERT INTO company (id, slug, name) VALUES
('4cb362d6-2018-428d-b758-e024a316f85b', 'demo', 'Demo Gym');

-- SEED DATA: Sample gym plans
INSERT INTO plans (name, description, price, inscription_fee, duration_days, color) VALUES
('Mensualidad Basica', 'Acceso al gimnasio de lunes a viernes', 500.00, 200.00, 30, 'primary'),
('Mensualidad Premium', 'Acceso ilimitado + clases grupales', 800.00, 200.00, 30, 'premium'),
('Pase Diario', 'Acceso por un dia', 80.00, NULL, 1, 'accent');

-- SEED DATA: Sample instructors
INSERT INTO instructors (name, specialty, is_active) VALUES
('Carlos Martinez', 'CrossFit', true),
('Ana Lopez', 'Yoga y Pilates', true),
('Roberto Sanchez', 'Entrenamiento Funcional', true);
EOSQL

echo "Demo tenant database (gym_demo) created successfully!"
