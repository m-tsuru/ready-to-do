package users

import "time"

type User struct {
	Id        string    `gorm:"column:id;primaryKey" db:"id"`
	Mail      string    `gorm:"column:mail;unique;not null" db:"mail"`
	Password  string    `gorm:"column:password;not null" db:"password"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type UserSession struct {
	Id        string    `gorm:"column:id;primaryKey" db:"id"`
	UserId    string    `gorm:"column:user_id;not null;index" db:"user_id"`
	Token     string    `gorm:"column:token;unique;not null;index" db:"token"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null;index" db:"expires_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
