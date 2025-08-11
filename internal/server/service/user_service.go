// Package service
package service

import (
	"errors"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/utils"
)

type RegisterUserData struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Cid       int    `json:"cid"`
	EmailCode int    `json:"email_code"`
}

type RegisterUserResponse struct {
	Username   string `json:"username"`
	Token      string `json:"token"`
	FlushToken string `json:"flush_token"`
}

var (
	ErrDatabaseFail     = CodeStatus{"DATABASE_ERROR", "服务器内部错误", ServerInternalError}
	ErrRegisterFail     = CodeStatus{"REGISTER_FAIL", "注册失败", ServerInternalError}
	ErrIdentifierTaken  = CodeStatus{"USER_EXISTS", "用户已存在", BadRequest}
	ErrEmailNotFound    = CodeStatus{"EMAIL_CODE_NOT_FOUND", "未向该邮箱发送验证码", BadRequest}
	ErrCidNotMatch      = CodeStatus{"CID_NOT_MATCH", "注册cid与验证码发送时的cid不一致", BadRequest}
	ErrEmailExpired     = CodeStatus{"EMAIL_CODE_EXPIRED", "验证码已过期", BadRequest}
	ErrEmailCodeInvalid = CodeStatus{"EMAIL_CODE_INVALID", "邮箱验证码错误", BadRequest}
	RegisterSuccess     = CodeStatus{"REGISTER_SUCCESS", "注册成功", Ok}
)

func (rud *RegisterUserData) RegisterUser() *ApiResponse[RegisterUserResponse] {
	if rud.Username == "" || rud.Email == "" || rud.Password == "" || rud.Cid <= 0 || rud.EmailCode <= 1e5 {
		return NewApiResponse[RegisterUserResponse](&ErrIllegalParam, Unsatisfied, nil)
	}
	err := GetEmailManager().VerifyCode(rud.Email, rud.EmailCode, rud.Cid)
	switch {
	case errors.Is(err, ErrEmailCodeNotFound):
		return NewApiResponse[RegisterUserResponse](&ErrEmailNotFound, Unsatisfied, nil)
	case errors.Is(err, ErrEmailCodeExpired):
		return NewApiResponse[RegisterUserResponse](&ErrEmailExpired, Unsatisfied, nil)
	case errors.Is(err, ErrInvalidEmailCode):
		return NewApiResponse[RegisterUserResponse](&ErrEmailCodeInvalid, Unsatisfied, nil)
	case errors.Is(err, ErrCidMismatch):
		return NewApiResponse[RegisterUserResponse](&ErrCidNotMatch, Unsatisfied, nil)
	default:
	}
	user, err := database.NewUser(rud.Username, rud.Email, rud.Cid, rud.Password)
	if err != nil {
		return NewApiResponse[RegisterUserResponse](&ErrRegisterFail, Unsatisfied, nil)
	}
	err = user.AddUser()
	switch {
	case errors.Is(err, database.ErrIdentifierCheck):
		return NewApiResponse[RegisterUserResponse](&ErrRegisterFail, Unsatisfied, nil)
	case errors.Is(err, database.ErrIdentifierTaken):
		return NewApiResponse[RegisterUserResponse](&ErrIdentifierTaken, Unsatisfied, nil)
	case err == nil:
	default:
		return NewApiResponse[RegisterUserResponse](&ErrDatabaseFail, Unsatisfied, nil)
	}
	token := NewClaims(user, false)
	flushToken := NewClaims(user, true)
	return NewApiResponse(&RegisterSuccess, Unsatisfied, &RegisterUserResponse{
		Username:   rud.Username,
		Token:      token.generateKey(),
		FlushToken: flushToken.generateKey(),
	})
}

type UserLoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Username   string `json:"username"`
	Token      string `json:"token"`
	FlushToken string `json:"flush_token"`
}

var (
	ErrUsernameOrPassword = CodeStatus{"WRONG_USERNAME_OR_PASSWORD", "用户名或密码错误", BadRequest}
	LoginSuccess          = CodeStatus{"LOGIN_SUCCESS", "登陆成功", Ok}
)

func (ul *UserLoginData) UserLogin() *ApiResponse[UserLoginResponse] {
	if ul.Username == "" || ul.Password == "" {
		return NewApiResponse[UserLoginResponse](&ErrIllegalParam, Unsatisfied, nil)
	}
	userId := database.StringUserId(ul.Username)
	user, err := userId.GetUser()
	if errors.Is(err, database.ErrUserNotFound) {
		return NewApiResponse[UserLoginResponse](&ErrUsernameOrPassword, Unsatisfied, nil)
	} else if err != nil {
		return NewApiResponse[UserLoginResponse](&ErrDatabaseFail, Unsatisfied, nil)
	}
	pass := user.VerifyPassword(ul.Password)
	token := NewClaims(user, false)
	flushToken := NewClaims(user, true)
	if pass {
		return NewApiResponse(&LoginSuccess, Unsatisfied, &UserLoginResponse{
			Username:   user.Username,
			Token:      token.generateKey(),
			FlushToken: flushToken.generateKey(),
		})
	}
	return NewApiResponse[UserLoginResponse](&ErrUsernameOrPassword, Unsatisfied, nil)
}

type UserAvailabilityData struct {
	Username string `query:"username"`
	Email    string `query:"email"`
	Cid      string `query:"cid"`
}

type UserAvailabilityResponse bool

var (
	NameNotAvailability = CodeStatus{"INFO_NOT_AVAILABILITY", "用户信息不可用", Ok}
	NameAvailability    = CodeStatus{"INFO_AVAILABILITY", "用户信息可用", Ok}
)

func (ua *UserAvailabilityData) CheckAvailability() *ApiResponse[UserAvailabilityResponse] {
	if ua.Username == "" && ua.Email == "" && ua.Cid == "" {
		return NewApiResponse[UserAvailabilityResponse](&ErrIllegalParam, Unsatisfied, nil)
	}
	exist, _ := database.IsUserIdentifierTaken(utils.StrToInt(ua.Cid, 0), ua.Username, ua.Email)
	data := UserAvailabilityResponse(!exist)
	if exist {
		return NewApiResponse(&NameNotAvailability, Unsatisfied, &data)
	}
	return NewApiResponse(&NameAvailability, Unsatisfied, &data)
}

type UserCurrentProfileData struct {
	Username string
}

type UserCurrentProfileResponse struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Cid            int    `json:"cid"`
	QQ             int    `json:"qq"`
	Rating         int    `json:"rating"`
	TotalPilotTime int    `json:"total_pilot_time"`
	TotalAtcTime   int    `json:"total_atc_time"`
	Permission     int64  `json:"permission"`
}

var (
	ErrUserNotFound          = CodeStatus{"USER_NOT_FOUND", "指定用户不存在", NotFound}
	GetCurrentProfileSuccess = CodeStatus{"GET_CURRENT_PROFILE_SUCCESS", "获取当前用户信息成功", Ok}
)

func (up *UserCurrentProfileData) GetCurrentProfile() *ApiResponse[UserCurrentProfileResponse] {
	userId := database.StringUserId(up.Username)
	user, err := userId.GetUser()
	if errors.Is(err, database.ErrUserNotFound) {
		return NewApiResponse[UserCurrentProfileResponse](&ErrUserNotFound, Unsatisfied, nil)
	} else if err != nil {
		return NewApiResponse[UserCurrentProfileResponse](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := UserCurrentProfileResponse{
		Username:       user.Username,
		Email:          user.Email,
		Cid:            user.Cid,
		QQ:             user.QQ,
		Rating:         user.Rating,
		TotalPilotTime: user.TotalPilotTime,
		TotalAtcTime:   user.TotalAtcTime,
		Permission:     user.Permission,
	}
	return NewApiResponse(&GetCurrentProfileSuccess, Unsatisfied, &data)
}
