package database

import (
	"context"
	"errors"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/defination"
	. "github.com/half-nothing/fsd-server/internal/server/defination/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserOperation struct {
	config *c.OtherConfig
	db     *gorm.DB
}

func NewUserOperation(config *c.OtherConfig, db *gorm.DB) *UserOperation {
	return &UserOperation{config: config, db: db}
}

func (userOperation *UserOperation) GetUserByUid(uid uint) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := userOperation.db.WithContext(ctx).
		Where("id = ?", uid).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (userOperation *UserOperation) GetUserByCid(cid uint) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := userOperation.db.WithContext(ctx).
		Where("cid = ?", cid).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil

}

func (userOperation *UserOperation) GetUserByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := userOperation.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (userOperation *UserOperation) GetUserByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := userOperation.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (userOperation *UserOperation) GetUserByUsernameOrEmail(ident string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	user := User{}
	err := userOperation.db.WithContext(ctx).
		Where("username = ? OR email = ?", ident, ident).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (userOperation *UserOperation) GetUsers(page, pageSize int) ([]*User, int64, error) {
	var total int64
	users := make([]*User, 0, pageSize)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	userOperation.db.WithContext(ctx).Model(&User{}).Select("id").Count(&total)
	err := userOperation.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error

	return users, total, err
}

func (userOperation *UserOperation) NewUser(username string, email string, cid int, password string) (*User, error) {
	encodePassword, err := bcrypt.GenerateFromPassword([]byte(password), userOperation.config.BcryptCost)
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

func (userOperation *UserOperation) AddUser(user *User) error {
	return userOperation.db.Transaction(func(tx *gorm.DB) error {
		taken, err := userOperation.IsUserIdentifierTaken(tx, user.Cid, user.Username, user.Email)
		if err != nil {
			return ErrIdentifierCheck
		}

		if taken {
			return ErrIdentifierTaken
		}

		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		defer cancel()
		return tx.WithContext(ctx).Create(user).Error
	})
}

func (userOperation *UserOperation) UpdateUserAtcTime(user *User, seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Model(user).Update("total_atc_time", gorm.Expr("total_atc_time + ?", seconds)).Error
}

func (userOperation *UserOperation) UpdateUserPilotTime(user *User, seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Model(user).Update("total_pilot_time", gorm.Expr("total_pilot_time + ?", seconds)).Error
}

func (userOperation *UserOperation) UpdateUserRating(user *User, rating int) error {
	user.Rating = rating
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Update("rating", rating).Error
	})
}

func (userOperation *UserOperation) UpdatePermission(user *User, permission defination.Permission) error {
	user.Permission = int64(permission)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Update("permission", int64(permission)).Error
	})
}

func (userOperation *UserOperation) UpdateUserInfo(user *User, info map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Model(user).Updates(info).Error
}

func (userOperation *UserOperation) UpdateUserPassword(user *User, originalPassword, newPassword string) ([]byte, error) {
	if !userOperation.VerifyUserPassword(user, originalPassword) {
		return nil, ErrOldPassword
	}
	encodePassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), config.Server.General.BcryptCost)
	if err != nil {
		return nil, ErrPasswordEncode
	}
	user.Password = string(encodePassword)
	return encodePassword, nil
}

func (userOperation *UserOperation) SaveUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := userOperation.db.WithContext(ctx).Save(user).Error
	return err
}

func (userOperation *UserOperation) VerifyUserPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (userOperation *UserOperation) IsUserIdentifierTaken(tx *gorm.DB, cid int, username, email string) (bool, error) {
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
