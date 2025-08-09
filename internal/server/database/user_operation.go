package database

import (
	"context"
	"golang.org/x/crypto/bcrypt"
)

type UserId interface {
	GetUser() (*User, error)
}

type IntUserId int
type StringUserId string

func (id IntUserId) GetUser() (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	if err := database.WithContext(ctx).Where("cid = ?", id, id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (id StringUserId) GetUser() (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	if err := database.WithContext(ctx).Where("username = ? or email = ?", id, id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func NewUser(username string, email string, cid int, password string) *User {
	encodePassword, _ := bcrypt.GenerateFromPassword([]byte(password), config.Server.General.BcryptCost)
	return &User{
		Username:       username,
		Email:          email,
		Cid:            cid,
		Password:       string(encodePassword),
		QQ:             0,
		Rating:         0,
		Permission:     0,
		TotalPilotTime: 0,
		TotalAtcTime:   0,
	}
}

func AddUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Create(user).Error
}

func (user *User) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(&user).Error
	return err
}

func (user *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
