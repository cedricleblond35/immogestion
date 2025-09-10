package repository

import (
	model "api/services/auth/internal/models"
	"context"
)

// UserRepository définit les opérations pour gérer les utilisateurs
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error                 // Créer un utilisateur
	FindByEmail(ctx context.Context, email string) (*model.User, error) // Trouver par email (pour login)
	Update(ctx context.Context, user *model.User) error                 // Mettre à jour (e.g., last_login_at)
	Delete(ctx context.Context, id uint) error                          // Supprimer (soft delete)
	FindByID(ctx context.Context, id uint) (*model.User, error)         // Trouver par ID
	List(ctx context.Context, limit, offset int) ([]*model.User, error) // Lister avec pagination
	Ping(ctx context.Context) error                                     // Vérifier la connexion (pour healthcheck)
}
