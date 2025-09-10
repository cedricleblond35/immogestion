package repository

import (
	"context"

	// Chemin vers DBClient
	"api/services/auth/internal/database"
	model "api/services/auth/internal/models"

	"go.uber.org/zap"
)

// UserRepository définit les opérations pour gérer les utilisateurs
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
	Ping(ctx context.Context) error
}

// UserRepositoryImpl implémente UserRepository en utilisant DBClient.
type UserRepositoryImpl struct {
	db     *database.GORM
	logger *zap.SugaredLogger
}

// Create implements UserRepository.
func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	r.logger.Infof("------------ Creating user: %s ----------", user.Email)
	gormDB := r.db
	return gormDB.Create(ctx, user)

}

// Delete implements UserRepository.
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint) error {
	panic("unimplemented")
}

// FindByEmail implements UserRepository.
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	panic("unimplemented")
}

// FindByID implements UserRepository.
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*model.User, error) {
	panic("unimplemented")
}

// List implements UserRepository.
func (r *UserRepositoryImpl) List(ctx context.Context, limit int, offset int) ([]*model.User, error) {
	panic("unimplemented")
}

// Ping implements UserRepository.
func (r *UserRepositoryImpl) Ping(ctx context.Context) error {
	panic("unimplemented")
}

// Update implements UserRepository.
func (r *UserRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	panic("unimplemented")
}

func NewUserRepository(db *database.GORM, logger *zap.SugaredLogger) UserRepository {
	return &UserRepositoryImpl{db: db, logger: logger}
}
