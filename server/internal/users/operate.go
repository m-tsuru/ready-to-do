package users

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, Mail string, Password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Id:        uuid.New().String(),
		Mail:      Mail,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByMail(db *gorm.DB, Mail string) (*User, error) {
	var user User
	if err := db.Where("mail = ?", Mail).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserById(db *gorm.DB, Id string) (*User, error) {
	var user User
	if err := db.Where("id = ?", Id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u User) CheckPassword(Password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u User) UpdatePassword(db *gorm.DB, Password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	u.UpdatedAt = time.Now()

	return db.Save(&u).Error
}

func (u User) CreateUserSession(db *gorm.DB) (*UserSession, error) {
	// ランダムなトークンを生成
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	session := &UserSession{
		Id:        uuid.New().String(),
		UserId:    u.Id,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour * 30), // 30日間有効
	}

	if err := db.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

func GetUserSessionByToken(db *gorm.DB, Token string) (*User, *UserSession, error) {
	var session UserSession
	if err := db.Where("token = ? AND expires_at > ?", Token, time.Now()).First(&session).Error; err != nil {
		return nil, nil, err
	}

	var user User
	if err := db.Where("id = ?", session.UserId).First(&user).Error; err != nil {
		return nil, nil, err
	}

	return &user, &session, nil
}
