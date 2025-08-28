package database

import (
	"context"
	"errors"
	c "github.com/half-nothing/simple-fsd/internal/config"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type UserOperation struct {
	config       *c.OtherConfig
	db           *gorm.DB
	queryTimeout time.Duration
}

func NewUserOperation(db *gorm.DB, queryTimeout time.Duration, config *c.OtherConfig) *UserOperation {
	return &UserOperation{config: config, db: db, queryTimeout: queryTimeout}
}

func (userOperation *UserOperation) GetUserByUid(uid uint) (user *User, err error) {
	user = &User{}
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	err = userOperation.db.WithContext(ctx).
		Where("id = ?", uid).
		First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrUserNotFound
	}
	return
}

func (userOperation *UserOperation) GetUserByCid(cid int) (user *User, err error) {
	user = &User{}
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	err = userOperation.db.WithContext(ctx).
		Where("cid = ?", cid).
		First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrUserNotFound
	}
	return
}

func (userOperation *UserOperation) GetUserByUsername(username string) (user *User, err error) {
	user = &User{}
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	err = userOperation.db.WithContext(ctx).
		Where("username = ?", username).
		First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrUserNotFound
	}
	return
}

func (userOperation *UserOperation) GetUserByEmail(email string) (user *User, err error) {
	user = &User{}
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	err = userOperation.db.WithContext(ctx).
		Where("email = ?", email).
		First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrUserNotFound
	}
	return user, nil
}

func (userOperation *UserOperation) GetUserByUsernameOrEmail(ident string) (user *User, err error) {
	user = &User{}
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	err = userOperation.db.WithContext(ctx).
		Where("username = ? OR email = ?", ident, ident).
		First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrUserNotFound
	}
	return user, nil
}

func (userOperation *UserOperation) GetUsers(page, pageSize int) (users []*User, total int64, err error) {
	users = make([]*User, 0, pageSize)
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	userOperation.db.WithContext(ctx).Model(&User{}).Select("id").Count(&total)
	err = userOperation.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return
}

func (userOperation *UserOperation) NewUser(username string, email string, cid int, password string) (user *User, err error) {
	encodePassword, err := bcrypt.GenerateFromPassword([]byte(password), userOperation.config.BcryptCost)
	if err != nil {
		return nil, ErrPasswordEncode
	}
	user = &User{
		Username:       username,
		Email:          email,
		Cid:            cid,
		Password:       string(encodePassword),
		AvatarUrl:      "",
		QQ:             0,
		Rating:         0,
		Permission:     0,
		TotalPilotTime: 0,
		TotalAtcTime:   0,
	}
	return
}

func (userOperation *UserOperation) AddUser(user *User) error {
	return userOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).Transaction(func(tx *gorm.DB) error {
		taken, err := userOperation.IsUserIdentifierTaken(tx, user.Cid, user.Username, user.Email)
		if err != nil {
			return ErrIdentifierCheck
		}

		if taken {
			return ErrIdentifierTaken
		}

		ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
		defer cancel()
		return tx.WithContext(ctx).Create(user).Error
	})
}

func (userOperation *UserOperation) UpdateUserAtcTime(user *User, seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Model(user).Update("total_atc_time", gorm.Expr("total_atc_time + ?", seconds)).Error
}

func (userOperation *UserOperation) UpdateUserPilotTime(user *User, seconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Model(user).Update("total_pilot_time", gorm.Expr("total_pilot_time + ?", seconds)).Error
}

func (userOperation *UserOperation) UpdateUserRating(user *User, rating int) error {
	user.Rating = rating
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	return userOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Update("rating", rating).Error
	})
}

func (userOperation *UserOperation) UpdateUserPermission(user *User, permission Permission) error {
	user.Permission = int64(permission)
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	return userOperation.db.Clauses(clause.Locking{Strength: "UPDATE"}).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(user).Update("permission", int64(permission)).Error
	})
}

func (userOperation *UserOperation) UpdateUserInfo(user *User, info map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()
	return userOperation.db.WithContext(ctx).Model(user).Updates(info).Error
}

func (userOperation *UserOperation) UpdateUserPassword(user *User, originalPassword, newPassword string) ([]byte, error) {
	if !userOperation.VerifyUserPassword(user, originalPassword) {
		return nil, ErrOldPassword
	}
	encodePassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), userOperation.config.BcryptCost)
	if err != nil {
		return nil, ErrPasswordEncode
	}
	user.Password = string(encodePassword)
	return encodePassword, nil
}

func (userOperation *UserOperation) SaveUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
	defer cancel()

	err := userOperation.db.WithContext(ctx).Save(user).Error
	return err
}

func (userOperation *UserOperation) VerifyUserPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (userOperation *UserOperation) IsUserIdentifierTaken(tx *gorm.DB, cid int, username, email string) (bool, error) {
	if tx == nil {
		tx = userOperation.db
	}
	ctx, cancel := context.WithTimeout(context.Background(), userOperation.queryTimeout)
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
