package database

import (
	"context"
	"golang.org/x/crypto/bcrypt"
)

type UserId interface {
	GetUser() (*User, error)
}

type IntId int
type StringId string

func (id IntId) GetUser() (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	if err := database.WithContext(ctx).Where("id = ? or cid = ?", id, id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (id StringId) GetUser() (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	if err := database.WithContext(ctx).Where("username = ? or email = ?", id, id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (user *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
