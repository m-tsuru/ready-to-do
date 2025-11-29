package users

import "time"

type User struct {
	Id        string    `db:"id"`
	Mail      string    `db:"mail"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserSession struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
