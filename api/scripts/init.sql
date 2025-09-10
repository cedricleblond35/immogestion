-- =====================================
-- SCRIPT D'INITIALISATION OPTIMISÉ POUR IMMOBILIER_PROD
-- Base créée via POSTGRES_DB dans .env (ex. : immobilier_prod).
-- Toutes les tables dans schéma auth pour cohérence.
-- =====================================

-- 1. Extensions (installées si manquantes)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- UUID v4
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- cryptage (bcrypt)
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
-- recherche textuelle (trigramme)
CREATE EXTENSION IF NOT EXISTS "btree_gin";
-- index GIN

-- Script de création de table PostgreSQL pour la structure GORM User
-- Compatible avec votre structure Go exacte


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
    id SERIAL PRIMARY KEY,
    company VARCHAR(100) NOT NULL,
    lastname VARCHAR(100) NOT NULL,
    firstname VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE, -- VARCHAR au lieu de CITEXT ; ajoutez LOWER(email) pour insensibilité si needed
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (
        role IN ('user', 'admin')
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

-- Insertion d'un utilisateur admin par défaut (si pas déjà présent)
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