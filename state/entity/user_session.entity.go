package entity

type UserSession struct {
	UserID       string `gorm:"type:uuid;primary_key"`
	Token        string `gorm:"type:text;not null"`
	RefreshToken string `gorm:"type:text;not null"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
