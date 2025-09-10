package model

// A Model defines an object that can be stored in database.
type Model interface {
	// Database returns the database's name in which the model is stored.
	Database() string
	// TableName returns the table name for this model.
	TableName() string
	
}
