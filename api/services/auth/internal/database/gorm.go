package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"api/services/auth/internal/errors"
	model "api/services/auth/internal/models"

	stderrors "errors"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// GORM implémente UserRepository (le repository pour les utilisateurs).
type GORM struct {
	db *gorm.DB
}

func (g *GORM) DB() (*sql.DB, error) {
	return g.db.DB()
}

// GormOpen ouvre une connexion PostgreSQL et retourne un GORM repository.
// debugSQL active le logging SQL en dev. sugar est utilisé pour logs (toujours non-nil).
func GormOpen(ctx context.Context, debugSQL bool, sugar *zap.SugaredLogger) (*GORM, error) {
	// Construction du DSN (Data Source Name) pour PostgreSQL
	// Utilise des variables d'environnement pour la sécurité
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Paris",
		os.Getenv("POSTGRES_HOST"),     // e.g., "localhost" ou "postgres" en Docker
		os.Getenv("POSTGRES_USER"),     // e.g., "immobilier_user"
		os.Getenv("POSTGRES_PASSWORD"), // e.g., "devpassword" (de .env.dev)
		os.Getenv("POSTGRES_DB"),       // "immobilier_db" depuis config
		os.Getenv("POSTGRES_PORT"),     // e.g., "5432"
	)

	// Ouverture de la connexion GORM avec le driver PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "auth.", // Ajoute "auth." devant chaque nom de table
		},
		PrepareStmt:     true,                                  // Prepared statements pour sécurité et perf
		CreateBatchSize: 1000,                                  // Optimisé pour insertions batch (e.g., users en masse)
		Logger:          logger.Default.LogMode(logger.Silent), // Logging GORM silencieux par défaut (activez avec Debug si besoin)
	})
	if err != nil {
		sugar.Errorf("Failed to open PostgreSQL connection: %v", err)           // Log direct avec sugar
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err) // Wrapping standard Go
	}

	// Récupérer *sql.DB pour configurer le pool de connexions
	sqlDB, err := db.DB()
	if err != nil {
		sugar.Errorf("Failed to get underlying sql.DB: %v", err)
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configuration du pool de connexions (recommandé pour performances en microservices)
	sqlDB.SetMaxIdleConns(10)                 // Max connexions idle
	sqlDB.SetMaxOpenConns(100)                // Max connexions ouvertes
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Durée de vie max d'une connexion

	// Ping pour vérifier la connexion
	if err := sqlDB.PingContext(ctx); err != nil {
		sugar.Errorf("Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	sugar.Infof("Database ping successful (host: %s)", os.Getenv("DB_HOST"))

	// Auto-migrate pour créer/migrer la table users (basé sur model.User)
	// if err := db.AutoMigrate(&model.User{}); err != nil {
	// 	sugar.Errorf("Failed to auto-migrate users table: %v", err)
	// 	return nil, fmt.Errorf("failed to auto-migrate schema: %w", err)
	// }
	// sugar.Info("Users table migrated successfully")

	// Active le debug SQL en dev (logging des queries)
	if debugSQL {
		db = db.Debug()
		sugar.Info("SQL debug mode enabled")
	}

	return &GORM{db: db}, nil
}

// Create implémente Create pour un User (avec gestion de doublons via OnConflict).
func (g *GORM) Create(ctx context.Context, user *model.User, sugar *zap.SugaredLogger) error {
	err := g.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(user).Error
	if err != nil {
		sugar.Errorf("Failed to create user %s: %v", user.Email, err)
		return fmt.Errorf("failed to create user %s: %w", user.Email, err) // Wrapping avec contexte
	}
	sugar.Infof("User created successfully: ID=%d, email=%s", user.ID, user.Email)
	return nil
}

// FindByEmail implémente FindByEmail (avec wrapping et logging).
func (g *GORM) FindByEmail(ctx context.Context, email string, sugar *zap.SugaredLogger) (*model.User, error) {
	var user model.User
	err := g.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) { // Assurez-vous que "errors" est le standard (alias si conflit)
			sugar.Debugf("User not found for email: %s", email) // Debug car cas attendu (e.g., login)
			return nil, errors.NewUserNotFound(email)           // Custom error
		}
		sugar.Errorf("Database error finding user by email %s: %v", email, err)
		return nil, fmt.Errorf("database error finding user by email %s: %w", email, err)
	}
	sugar.Debugf("User found for email: %s (ID: %d)", email, user.ID)
	return &user, nil
}

// FindByID implémente FindByID (similaire à FindByEmail).
func (g *GORM) FindByID(ctx context.Context, id uint, sugar *zap.SugaredLogger) (*model.User, error) {
	var user model.User
	err := g.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			sugar.Debugf("User not found for ID: %d", id)
			return nil, fmt.Errorf("user with ID %d: %w", id, errors.NewUserNotFound("")) // Ajustez si email disponible
		}
		sugar.Errorf("Database error finding user by ID %d: %v", id, err)
		return nil, fmt.Errorf("database error finding user by ID %d: %w", id, err)
	}
	sugar.Debugf("User found for ID: %d (email: %s)", id, user.Email)
	return &user, nil
}

// List implémente List (avec pagination, logging).
func (g *GORM) List(ctx context.Context, limit, offset int, sugar *zap.SugaredLogger) ([]*model.User, error) {
	var users []*model.User
	err := g.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		sugar.Errorf("Failed to list users (limit=%d, offset=%d): %v", limit, offset, err)
		return nil, fmt.Errorf("failed to list users (limit=%d, offset=%d): %w", limit, offset, err)
	}
	sugar.Infof("Listed %d users (offset=%d)", len(users), offset)
	return users, nil
}

// Update implémente Update (e.g., pour last_login_at).
func (g *GORM) Update(ctx context.Context, user *model.User, sugar *zap.SugaredLogger) error {
	err := g.db.WithContext(ctx).Save(user).Error // Save gère upsert et updated_at
	if err != nil {
		sugar.Errorf("Failed to update user %s (ID=%d): %v", user.Email, user.ID, err)
		return fmt.Errorf("failed to update user %s (ID=%d): %w", user.Email, user.ID, err)
	}
	sugar.Infof("User updated successfully: ID=%d, email=%s", user.ID, user.Email)
	return nil
}

// Delete implémente Delete (soft delete avec GORM).
func (g *GORM) Delete(ctx context.Context, id uint, sugar *zap.SugaredLogger) error {
	err := g.db.WithContext(ctx).Delete(&model.User{}, id).Error
	if err != nil {
		sugar.Errorf("Failed to delete user ID=%d: %v", id, err)
		return fmt.Errorf("failed to delete user ID=%d: %w", id, err)
	}
	sugar.Warnf("User deleted: ID=%d") // Warn car suppression
	return nil
}

// Ping implémente Ping (pour healthcheck).
func (g *GORM) Ping(ctx context.Context, sugar *zap.SugaredLogger) error {
	sqlDB, err := g.db.DB()
	if err != nil {
		sugar.Errorf("Failed to get sql.DB for ping: %v", err)
		return err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		sugar.Errorf("Database ping failed: %v", err)
		return fmt.Errorf("database ping failed: %w", err)
	}
	sugar.Debug("Database ping successful")
	return nil
}

// IsNotFound vérifie si l'erreur est "record not found" (GORM-specific).
// Utilisez stderrors.Is si alias (voir notes précédentes pour conflit errors).
func (g *GORM) IsNotFound(err error) bool {
	return stderrors.Is(err, gorm.ErrRecordNotFound)
}
