// Package service
package service

import (
	c "github.com/half-nothing/simple-fsd/internal/config"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
)

type FieldValidator struct {
	Min, Max          int
	ErrShort, ErrLong *ApiStatus
}

func (v *FieldValidator) CheckString(value string) *ApiStatus {
	length := len(value)
	if length > v.Max {
		return v.ErrLong
	}
	if length < v.Min {
		return v.ErrShort
	}
	return nil
}

func (v *FieldValidator) CheckInt(value int) *ApiStatus {
	if value > v.Max {
		return v.ErrLong
	}
	if value < v.Min {
		return v.ErrShort
	}
	return nil
}

var (
	usernameValidator *FieldValidator
	passwordValidator *FieldValidator
	emailValidator    *FieldValidator
	cidValidator      *FieldValidator
)

func InitValidator(config *c.HttpServerLimit) {
	usernameValidator = &FieldValidator{
		Min:      config.UsernameLengthMin,
		Max:      config.UsernameLengthMax,
		ErrShort: &ApiStatus{StatusName: "USERNAME_TOO_SHORT", Description: "用户名过短", HttpCode: BadRequest},
		ErrLong:  &ApiStatus{StatusName: "USERNAME_TOO_LONG", Description: "用户名过长", HttpCode: BadRequest},
	}
	passwordValidator = &FieldValidator{
		Min:      config.PasswordLengthMin,
		Max:      config.PasswordLengthMax,
		ErrShort: &ApiStatus{StatusName: "PASSWORD_TOO_SHORT", Description: "密码长度过短", HttpCode: BadRequest},
		ErrLong:  &ApiStatus{StatusName: "PASSWORD_TOO_LONG", Description: "密码长度过长", HttpCode: BadRequest},
	}
	emailValidator = &FieldValidator{
		Min:      config.EmailLengthMin,
		Max:      config.EmailLengthMax,
		ErrShort: &ApiStatus{StatusName: "EMAIL_TOO_SHORT", Description: "邮箱过短", HttpCode: BadRequest},
		ErrLong:  &ApiStatus{StatusName: "EMAIL_TOO_LONG", Description: "邮箱过长", HttpCode: BadRequest},
	}
	cidValidator = &FieldValidator{
		Min:      config.CidMin,
		Max:      config.CidMax,
		ErrShort: &ApiStatus{StatusName: "CID_TOO_SHORT", Description: "cid过短", HttpCode: BadRequest},
		ErrLong:  &ApiStatus{StatusName: "CID_TOO_LONG", Description: "cid过长", HttpCode: BadRequest},
	}
}
