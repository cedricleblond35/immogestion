package errors

import (
	"errors"
	"fmt"
)

// UserNotFoundError est une erreur custom indiquant qu'un utilisateur n'a pas été trouvé.
type UserNotFoundError struct {
	Email string
}

// Error implémente l'interface error.
func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user not found with email: %s", e.Email)
}

// NewUserNotFound crée une nouvelle instance d'erreur UserNotFoundError.
func NewUserNotFound(email string) *UserNotFoundError {
	return &UserNotFoundError{Email: email}
}

// Autres erreurs custom potentielles pour l'auth (optionnel)
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrUserInactive = errors.New("user account is inactive")