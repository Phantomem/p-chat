package entity

type ChatRoom struct {
	ID      string `gorm:"type:varchar(255);primary_key"`
	Name    string `gorm:"type:varchar(255);not null"`
	Members string `gorm:"type:jsonb;not null"`
}

func (ChatRoom) TableName() string {
	return "chat_rooms"
}
