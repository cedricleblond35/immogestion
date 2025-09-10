package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Company     string `gorm:"type:varchar(100);not null" json:"company"`
	Lastname    string `gorm:"type:varchar(100);not null" json:"lastname"`
	Firstname   string `gorm:"type:varchar(100);not null" json:"firstname"`
	Email       string `gorm:"type:varchar(100);uniqueIndex:uni_users_email;not null" json:"email"`
	Password    string `gorm:"type:varchar(255);column:password_hash;not null" json:"-"`
	Role        string `gorm:"size:50;default:user;check:role_check"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
	IsActive    bool `gorm:"default:true"`

	changedFields map[string]any `gorm:"-"`
}

// TableName spécifie le nom de la table pour GORM
func (User) TableName() string {
	return "auth.users"
}

func NewUser(ctx context.Context) *User {
	return &User{
		changedFields: make(map[string]any),
	}
}

// BeforeCreate hook pour initialiser les champs avant la création
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return
}

// BeforeUpdate hook pour mettre à jour le champ UpdatedAt avant la mise à jour
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

// RoleCheckConstraint ajoute une contrainte de vérification pour le champ Role
func (User) RoleCheckConstraint(db *gorm.DB) {
	db.Exec("ALTER TABLE auth.users ADD CONSTRAINT role_check CHECK (role IN ('user', 'admin'))")
}

func (User) Create(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}
func (User) FindByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (User) FindByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (User) Update(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}

func (User) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&User{}, id).Error
}

func (User) List(db *gorm.DB, limit int, offset int) ([]User, error) {
	var users []User
	err := db.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
