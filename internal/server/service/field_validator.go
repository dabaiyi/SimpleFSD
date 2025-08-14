// Package service
package service

import c "github.com/half-nothing/fsd-server/internal/config"

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

func InitValidator() {
	config, _ := c.GetConfig()
	usernameValidator = &FieldValidator{
		Min:      config.Server.HttpServer.UsernameLengthMin,
		Max:      config.Server.HttpServer.UsernameLengthMax,
		ErrShort: &ApiStatus{"USERNAME_TOO_SHORT", "用户名过短", BadRequest},
		ErrLong:  &ApiStatus{"USERNAME_TOO_LONG", "用户名过长", BadRequest},
	}
	passwordValidator = &FieldValidator{
		Min:      config.Server.HttpServer.PasswordLengthMin,
		Max:      config.Server.HttpServer.PasswordLengthMax,
		ErrShort: &ApiStatus{"PASSWORD_TOO_SHORT", "密码长度过短", BadRequest},
		ErrLong:  &ApiStatus{"PASSWORD_TOO_LONG", "密码长度过长", BadRequest},
	}
	emailValidator = &FieldValidator{
		Min:      config.Server.HttpServer.EmailLengthMin,
		Max:      config.Server.HttpServer.EmailLengthMax,
		ErrShort: &ApiStatus{"EMAIL_TOO_SHORT", "邮箱过短", BadRequest},
		ErrLong:  &ApiStatus{"EMAIL_TOO_LONG", "邮箱过长", BadRequest},
	}
	cidValidator = &FieldValidator{
		Min:      config.Server.HttpServer.CidMin,
		Max:      config.Server.HttpServer.CidMax,
		ErrShort: &ApiStatus{"CID_TOO_SHORT", "cid过短", BadRequest},
		ErrLong:  &ApiStatus{"CID_TOO_LONG", "cid过长", BadRequest},
	}

}
