package entity

import "time"

type User struct {
	ID        string    `gorm:"type:uuid;primary_key"`
	Email     string    `gorm:"type:varchar(255);unique;not null"`
	Role      string    `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	DeletedAt time.Time `gorm:"not null"`
}

func (User) TableName() string {
	return "users"
}
