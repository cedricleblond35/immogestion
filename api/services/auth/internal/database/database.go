package database
import (
	"api/services/auth/internal/models")

// A Client can interact with the database.
type User interface {
	// Ping verifies a connection to the database is still alive,
	// establishing a connection if necessary.
	Ping() error
	// Load loads the entry from database according to the query.
	Load(m model.Model, query string, args ...any) error
	LoadLast(m model.Model, query string, args ...any) error
	// Load loads the entry from database according to the array query.
	Loads(m any, query string, args ...any) error
	// Load loads the entry from database according to the array query.
	LoadsDESC(m any, query string, args ...any) error
	// Create inserts the entry in database with the given model.
	Create(m model.Model) error
	CreateBatch(m []model.Model) error
	// Save inserts (if no primary key is given) or updates the entry in database with the given model.
	// It will overwrite with all the fields of the model.
	Save(m model.Model) error
	// Update updates the model based on primary key(s).
	Update(m model.Model, fields map[string]any) error
	// Delete deletes the entry in database with the given model.
	Delete(m model.Model) error
	// IsNotFound returns true if err is nil or a not found error.
	IsNotFound(err error) bool
}
