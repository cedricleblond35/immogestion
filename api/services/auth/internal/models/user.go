package model

import (
	"time"
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
