-- Créer un schéma dédié pour le microservice auth (best practice pour isolation)
CREATE SCHEMA IF NOT EXISTS auth;

-- Table users dans le schéma auth
CREATE TABLE IF NOT EXISTS auth.users (
    id BIGSERIAL PRIMARY KEY,  -- Utilisez BIGSERIAL pour scalabilité (au lieu de SERIAL)
    company VARCHAR(255) NOT NULL,
    lastname VARCHAR(255) NOT NULL,
    firstname VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),  -- Validation regex basique pour email
    password_hash TEXT NOT NULL,  -- Hash bcrypt ou Argon2 (longueur variable)
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('user', 'admin')),  -- CHECK au lieu d'ENUM pour simplicité (ENUM est ok mais moins flexible)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    -- Contrainte pour forcer updated_at sur UPDATE
    CONSTRAINT users_email_lowercase CHECK (email = LOWER(email))  -- Optionnel : forcer email en minuscules
);

-- Index pour performance et sécurité
CREATE INDEX CONCURRENTLY idx_users_email ON auth.users (email);  -- Sur email pour lookups rapides
CREATE INDEX CONCURRENTLY idx_users_active ON auth.users (is_active) WHERE is_active = TRUE;  -- Pour queries sur users actifs

-- Trigger pour updated_at (PostgreSQL natif)
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE
    ON auth.users FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Privilèges (best practice : restreindre l'accès)
REVOKE ALL ON SCHEMA public FROM PUBLIC;  -- Sécuriser le schéma public par défaut
GRANT USAGE ON SCHEMA auth TO auth_app_role;  -- Rôle pour l'app (e.g., votre service)
GRANT SELECT, INSERT, UPDATE ON auth.users TO auth_app_role;
GRANT USAGE, SELECT ON SEQUENCE auth.users_id_seq TO auth_app_role;  -- Pour SERIAL/BIGSERIAL

-- Pour un rôle readonly (e.g., pour audits)
GRANT SELECT ON auth.users TO readonly_role;