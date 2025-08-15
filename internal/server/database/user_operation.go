package database

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound    = errors.New("user does not exist")
	ErrIdentifierTaken = errors.New("user identifiers have been used")
	ErrPasswordEncode  = errors.New("password encode error")
	ErrIdentifierCheck = errors.New("identifier check error")
	ErrOldPassword     = errors.New("old password error")
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
	err := database.WithContext(ctx).
		Where("cid = ?", id).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (id StringUserId) GetUser() (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := database.WithContext(ctx).
		Where("username = ? OR email = ?", id, id).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func GetUserById(uid uint) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := database.WithContext(ctx).
		Where("id = ?", uid).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func NewUser(username string, email string, cid int, password string) (*User, error) {
	encodePassword, err := bcrypt.GenerateFromPassword([]byte(password), config.Server.General.BcryptCost)
	if err != nil {
		return nil, ErrPasswordEncode
	}
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
	}, nil
}

func GetUsers(page, pageSize int) ([]*User, int64, error) {
	var total int64
	users := make([]*User, 0, pageSize)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	database.WithContext(ctx).Model(&User{}).Select("id").Count(&total)
	err := database.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error

	return users, total, err
}

func (user *User) UpdateInfo(info map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Model(user).Updates(info).Error
}

func (user *User) UpdatePassword(originalPassword string, newPassword string) ([]byte, error) {
	if !user.VerifyPassword(originalPassword) {
		return nil, ErrOldPassword
	}
	encodePassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), config.Server.General.BcryptCost)
	if err != nil {
		return nil, ErrPasswordEncode
	}
	user.Password = string(encodePassword)
	return encodePassword, nil
}

func (user *User) addUser(tx *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return tx.WithContext(ctx).Create(user).Error
}

func (user *User) AddUser() error {
	return database.Transaction(func(tx *gorm.DB) error {
		taken, err := isUserIdentifierTaken(tx, user.Cid, user.Username, user.Email)
		if err != nil {
			return ErrIdentifierCheck
		}

		if taken {
			return ErrIdentifierTaken
		}

		return user.addUser(tx)
	})
}

func isUserIdentifierTaken(tx *gorm.DB, cid int, username, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var count int64
	err := tx.WithContext(ctx).
		Model(&User{}).
		Where("cid = ? OR username = ? OR email = ?", cid, username, email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsUserIdentifierTaken 检查用户唯一标识是否已存在（cid/username/email任一重复即返回true）
func IsUserIdentifierTaken(cid int, username, email string) (bool, error) {
	return isUserIdentifierTaken(database, cid, username, email)
}

func (user *User) AddAtcTime(seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Model(user).Update("total_atc_time", gorm.Expr("total_atc_time + ?", seconds)).Error
}

func (user *User) AddPilotTime(seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Model(user).Update("total_pilot_time", gorm.Expr("total_pilot_time + ?", seconds)).Error
}

func (user *User) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := database.WithContext(ctx).Save(user).Error
	return err
}

func (user *User) UpdateRating(rating int) error {
	user.Rating = rating
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Update("rating", rating).Error
	})
}

func (user *User) UpdatePermission(permission Permission) error {
	user.Permission = int64(permission)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Update("permission", int64(permission)).Error
	})
}

func (user *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
