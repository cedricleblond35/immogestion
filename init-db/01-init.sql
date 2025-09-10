-- =====================================
-- SCRIPT D'INITIALISATION CORRIGÉ POUR IMMOBILIER_PROD
-- 1. Exécuter EN TANT QUE SUPERUSER depuis base 'postgres' (une seule fois) :
--    CREATE DATABASE immobilier_prod WITH OWNER = immobilier_user ENCODING = 'UTF8' LC_COLLATE = 'fr_FR.UTF-8' LC_CTYPE = 'fr_FR.UTF-8' TEMPLATE = template0;
--    GRANT ALL PRIVILEGES ON DATABASE immobilier_prod TO immobilier_user;
-- 2. Puis : \c immobilier_prod; et exécuter ce script.
-- =====================================

-- 1. Création des extensions nécessaires (dans la base immobilier_prod)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- Pour UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- Pour email insensible à la casse
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
-- Pour recherche textuelle (trigramme)
CREATE EXTENSION IF NOT EXISTS "btree_gin";
-- Pour index GIN (arrays, full-text)

-- Vérifier si la DB existe déjà
SELECT 1
FROM pg_catalog.pg_database
WHERE
    datname = 'immobilier_prod';

CREATE DATABASE immobilier_prod
WITH
    OWNER = immobilier_user -- Propriétaire (optionnel, sinon c'est l'utilisateur actuel)
    ENCODING = 'UTF8' -- Encodage standard (recommandé)
    LC_COLLATE = 'fr_FR.UTF-8' -- Collation pour tri (français, optionnel)
    LC_CTYPE = 'fr_FR.UTF-8' -- Type de caractères
    TEMPLATE = template0;
-- Modèle vide (évite de copier des objets de template1)

GRANT ALL PRIVILEGES ON DATABASE immobilier_prod TO immobilier_user;

-- 2. Création du schéma auth (si pas déjà fait)
CREATE SCHEMA IF NOT EXISTS auth;

-- Définir search_path pour ce script (optionnel, mais aide)
SET search_path TO auth, public;

-- 3. Fonction pour auto-mettre à jour updated_at (dans schéma auth)
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 4. Table users (email en VARCHAR)
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    company VARCHAR(100) NOT NULL,
    lastname VARCHAR(100) NOT NULL,
    firstname VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE, -- VARCHAR au lieu de CITEXT ; ajoutez LOWER(email) pour insensibilité si needed
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (
        role IN ('user', 'admin', 'moderator')
    ),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_users_email ON auth.users (LOWER(email));
-- Index insensibilisé (optionnel)

DROP TRIGGER IF EXISTS update_auth_users_updated_at ON auth.users;

CREATE TRIGGER update_auth_users_updated_at
    BEFORE UPDATE ON auth.users
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- 5. Tables properties, tenants, contracts
CREATE TABLE IF NOT EXISTS auth.properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    owner_id UUID NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (
        type IN (
            'appartement',
            'maison',
            'bureau',
            'terrain'
        )
    ),
    surface DECIMAL(10, 2),
    rooms INTEGER CHECK (rooms > 0),
    rent_amount DECIMAL(10, 2) CHECK (rent_amount >= 0),
    charges_amount DECIMAL(10, 2) CHECK (charges_amount >= 0),
    deposit_amount DECIMAL(10, 2) CHECK (deposit_amount >= 0),
    available_from DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'available' CHECK (
        status IN (
            'available',
            'rented',
            'maintenance'
        )
    ),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth.tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL, -- VARCHAR au lieu de CITEXT
    phone VARCHAR(20),
    birth_date DATE CHECK (birth_date < CURRENT_DATE),
    identity_number VARCHAR(50) UNIQUE,
    employer VARCHAR(255),
    monthly_income DECIMAL(10, 2) CHECK (monthly_income >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth.contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    property_id UUID NOT NULL REFERENCES auth.properties (id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES auth.tenants (id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE CHECK (
        end_date > start_date
        OR end_date IS NULL
    ),
    rent_amount DECIMAL(10, 2) NOT NULL CHECK (rent_amount >= 0),
    charges_amount DECIMAL(10, 2) CHECK (charges_amount >= 0),
    deposit_amount DECIMAL(10, 2) CHECK (deposit_amount >= 0),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (
        status IN (
            'active',
            'terminated',
            'expired'
        )
    ),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (property_id, tenant_id)
);

-- Triggers pour les autres tables
DROP TRIGGER IF EXISTS update_auth_properties_updated_at ON auth.properties;

CREATE TRIGGER update_auth_properties_updated_at
    BEFORE UPDATE ON auth.properties
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

DROP TRIGGER IF EXISTS update_auth_tenants_updated_at ON auth.tenants;

CREATE TRIGGER update_auth_tenants_updated_at
    BEFORE UPDATE ON auth.tenants
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

DROP TRIGGER IF EXISTS update_auth_contracts_updated_at ON auth.contracts;

CREATE TRIGGER update_auth_contracts_updated_at
    BEFORE UPDATE ON auth.contracts
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- 6. Index
CREATE INDEX IF NOT EXISTS idx_auth_properties_owner_status ON auth.properties (owner_id, status);

CREATE INDEX IF NOT EXISTS idx_auth_properties_city_type ON auth.properties (city, type);

CREATE INDEX IF NOT EXISTS idx_auth_properties_description_trgm ON auth.properties USING GIN (description gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_auth_contracts_property_status ON auth.contracts (property_id, status);

CREATE INDEX IF NOT EXISTS idx_auth_contracts_tenant ON auth.contracts (tenant_id);

-- 7. Insertion admin (changez 'admin123' en prod !)
INSERT INTO
    auth.users (
        company,
        firstname,
        lastname,
        email,
        password_hash,
        role
    )
VALUES (
        'Immobilier System',
        'Admin',
        'System',
        'admin@immobilier.local',
        crypt ('admin123', gen_salt ('bf')),
        'admin'
    ) ON CONFLICT (email) DO NOTHING;

-- 8. Privilèges
GRANT USAGE ON SCHEMA auth TO immobilier_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth TO immobilier_user;

GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA auth TO immobilier_user;

GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA auth TO immobilier_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth
GRANT ALL ON TABLES TO immobilier_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth
GRANT ALL ON SEQUENCES TO immobilier_user;

-- Fin. Vérifiez : \dt auth.* ; SELECT * FROM auth.users;