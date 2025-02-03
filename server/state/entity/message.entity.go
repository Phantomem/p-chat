package entity

import "time"

type Message struct {
	ID         string    `gorm:"type:uuid;primary_key"`
	ChatRoomID string    `gorm:"type:varchar(255);not null"`
	AuthorID   string    `gorm:"type:uuid;not null"`
	Text       string    `gorm:"type:text;not null"`
	SeenBy     string    `gorm:"type:jsonb;not null"`
	ReceivedBy string    `gorm:"type:jsonb;not null"`
	SentAt     time.Time `gorm:"not null;default:current_timestamp"`
	DeletedAt  time.Time
}

func (Message) TableName() string {
	return "messages"
}
