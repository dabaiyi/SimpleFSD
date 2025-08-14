// Package service
package service

import (
	"errors"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/utils"
)

type RequestRegisterUser struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Cid       int    `json:"cid"`
	EmailCode int    `json:"email_code"`
}

type ResponseRegisterUser struct {
	Username   string `json:"username"`
	Token      string `json:"token"`
	FlushToken string `json:"flush_token"`
}

var (
	ErrRegisterFail     = ApiStatus{"REGISTER_FAIL", "注册失败", ServerInternalError}
	ErrIdentifierTaken  = ApiStatus{"USER_EXISTS", "用户已存在", BadRequest}
	ErrEmailNotFound    = ApiStatus{"EMAIL_CODE_NOT_FOUND", "未向该邮箱发送验证码", BadRequest}
	ErrCidNotMatch      = ApiStatus{"CID_NOT_MATCH", "注册cid与验证码发送时的cid不一致", BadRequest}
	ErrEmailExpired     = ApiStatus{"EMAIL_CODE_EXPIRED", "验证码已过期", BadRequest}
	ErrEmailCodeInvalid = ApiStatus{"EMAIL_CODE_INVALID", "邮箱验证码错误", BadRequest}
	SuccessRegister     = ApiStatus{"REGISTER_SUCCESS", "注册成功", Ok}
)

func (rud *RequestRegisterUser) RegisterUser() *ApiResponse[ResponseRegisterUser] {
	if rud.Username == "" || rud.Email == "" || rud.Password == "" || rud.Cid <= 0 || rud.EmailCode < 1e5 {
		return NewApiResponse[ResponseRegisterUser](&ErrIllegalParam, Unsatisfied, nil)
	}
	err := GetEmailManager().VerifyCode(rud.Email, rud.EmailCode, rud.Cid)
	switch {
	case errors.Is(err, ErrEmailCodeNotFound):
		return NewApiResponse[ResponseRegisterUser](&ErrEmailNotFound, Unsatisfied, nil)
	case errors.Is(err, ErrEmailCodeExpired):
		return NewApiResponse[ResponseRegisterUser](&ErrEmailExpired, Unsatisfied, nil)
	case errors.Is(err, ErrInvalidEmailCode):
		return NewApiResponse[ResponseRegisterUser](&ErrEmailCodeInvalid, Unsatisfied, nil)
	case errors.Is(err, ErrCidMismatch):
		return NewApiResponse[ResponseRegisterUser](&ErrCidNotMatch, Unsatisfied, nil)
	default:
	}
	if err := usernameValidator.CheckString(rud.Username); err != nil {
		return NewApiResponse[ResponseRegisterUser](err, Unsatisfied, nil)
	}
	if err := emailValidator.CheckString(rud.Email); err != nil {
		return NewApiResponse[ResponseRegisterUser](err, Unsatisfied, nil)
	}
	if err := passwordValidator.CheckString(rud.Password); err != nil {
		return NewApiResponse[ResponseRegisterUser](err, Unsatisfied, nil)
	}
	if err := cidValidator.CheckInt(rud.Cid); err != nil {
		return NewApiResponse[ResponseRegisterUser](err, Unsatisfied, nil)
	}
	user, err := database.NewUser(rud.Username, rud.Email, rud.Cid, rud.Password)
	if err != nil {
		return NewApiResponse[ResponseRegisterUser](&ErrRegisterFail, Unsatisfied, nil)
	}
	err = user.AddUser()
	switch {
	case errors.Is(err, database.ErrIdentifierCheck):
		return NewApiResponse[ResponseRegisterUser](&ErrRegisterFail, Unsatisfied, nil)
	case errors.Is(err, database.ErrIdentifierTaken):
		return NewApiResponse[ResponseRegisterUser](&ErrIdentifierTaken, Unsatisfied, nil)
	case err == nil:
	default:
		return NewApiResponse[ResponseRegisterUser](&ErrDatabaseFail, Unsatisfied, nil)
	}
	token := NewClaims(user, false)
	flushToken := NewClaims(user, true)
	return NewApiResponse(&SuccessRegister, Unsatisfied, &ResponseRegisterUser{
		Username:   rud.Username,
		Token:      token.generateKey(),
		FlushToken: flushToken.generateKey(),
	})
}

type RequestUserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseUserLogin struct {
	Username   string `json:"username"`
	Token      string `json:"token"`
	FlushToken string `json:"flush_token"`
}

var (
	ErrUsernameOrPassword = ApiStatus{"WRONG_USERNAME_OR_PASSWORD", "用户名或密码错误", BadRequest}
	SuccessLogin          = ApiStatus{"LOGIN_SUCCESS", "登陆成功", Ok}
)

func (ul *RequestUserLogin) UserLogin() *ApiResponse[ResponseUserLogin] {
	if ul.Username == "" || ul.Password == "" {
		return NewApiResponse[ResponseUserLogin](&ErrIllegalParam, Unsatisfied, nil)
	}
	userId := database.StringUserId(ul.Username)
	user, err := userId.GetUser()
	if errors.Is(err, database.ErrUserNotFound) {
		return NewApiResponse[ResponseUserLogin](&ErrUsernameOrPassword, Unsatisfied, nil)
	} else if err != nil {
		return NewApiResponse[ResponseUserLogin](&ErrDatabaseFail, Unsatisfied, nil)
	}
	pass := user.VerifyPassword(ul.Password)
	token := NewClaims(user, false)
	flushToken := NewClaims(user, true)
	if pass {
		return NewApiResponse(&SuccessLogin, Unsatisfied, &ResponseUserLogin{
			Username:   user.Username,
			Token:      token.generateKey(),
			FlushToken: flushToken.generateKey(),
		})
	}
	return NewApiResponse[ResponseUserLogin](&ErrUsernameOrPassword, Unsatisfied, nil)
}

type RequestUserAvailability struct {
	Username string `query:"username"`
	Email    string `query:"email"`
	Cid      string `query:"cid"`
}

type ResponseUserAvailability bool

var (
	NameNotAvailability = ApiStatus{"INFO_NOT_AVAILABILITY", "用户信息不可用", Ok}
	NameAvailability    = ApiStatus{"INFO_AVAILABILITY", "用户信息可用", Ok}
)

func (ua *RequestUserAvailability) CheckAvailability() *ApiResponse[ResponseUserAvailability] {
	if ua.Username == "" && ua.Email == "" && ua.Cid == "" {
		return NewApiResponse[ResponseUserAvailability](&ErrIllegalParam, Unsatisfied, nil)
	}
	exist, _ := database.IsUserIdentifierTaken(utils.StrToInt(ua.Cid, 0), ua.Username, ua.Email)
	data := ResponseUserAvailability(!exist)
	if exist {
		return NewApiResponse(&NameNotAvailability, Unsatisfied, &data)
	}
	return NewApiResponse(&NameAvailability, Unsatisfied, &data)
}

type RequestUserCurrentProfile struct {
	Uid uint
}

type ResponseUserCurrentProfile struct {
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
	ErrUserNotFound          = ApiStatus{"USER_NOT_FOUND", "指定用户不存在", NotFound}
	SuccessGetCurrentProfile = ApiStatus{"GET_CURRENT_PROFILE_SUCCESS", "获取当前用户信息成功", Ok}
)

func (up *RequestUserCurrentProfile) GetCurrentProfile() *ApiResponse[ResponseUserCurrentProfile] {
	user, err := database.GetUserById(up.Uid)
	if errors.Is(err, database.ErrUserNotFound) {
		return NewApiResponse[ResponseUserCurrentProfile](&ErrUserNotFound, Unsatisfied, nil)
	} else if err != nil {
		return NewApiResponse[ResponseUserCurrentProfile](&ErrDatabaseFail, Unsatisfied, nil)
	}
	data := ResponseUserCurrentProfile{
		Username:       user.Username,
		Email:          user.Email,
		Cid:            user.Cid,
		QQ:             user.QQ,
		Rating:         user.Rating,
		TotalPilotTime: user.TotalPilotTime,
		TotalAtcTime:   user.TotalAtcTime,
		Permission:     user.Permission,
	}
	return NewApiResponse(&SuccessGetCurrentProfile, Unsatisfied, &data)
}

type RequestUserEditCurrentProfile struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	QQ             int    `json:"qq"`
	OriginPassword string `json:"origin_password"`
	NewPassword    string `json:"new_password"`
}

type ResponseUserEditCurrentProfile struct {
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
	ErrOriginPasswordRequired = ApiStatus{"ORIGIN_PASSWORD_REQUIRED", "请输入原始密码", BadRequest}
	ErrNewPasswordRequired    = ApiStatus{"NEW_PASSWORD_REQUIRED", "请输入新密码", BadRequest}
	ErrOriginPassword         = ApiStatus{"ORIGIN_PASSWORD_ERROR", "原始密码不正确", BadRequest}
	ErrQQInvalid              = ApiStatus{"QQ_INVALID", "qq号不正确", BadRequest}
	SuccessEditCurrentProfile = ApiStatus{"SUCCESS_EDIT_CURRENT_PROFILE", "编辑用户信息成功", Ok}
)

func checkQQ(qq int) *ApiStatus {
	if 1e4 <= qq && qq < 1e11 {
		return nil
	}
	return &ErrQQInvalid
}

func (ue *RequestUserEditCurrentProfile) EditCurrentProfile() *ApiResponse[ResponseUserEditCurrentProfile] {
	if ue.Username == "" && ue.Email == "" && ue.QQ <= 0 && ue.OriginPassword == "" && ue.NewPassword == "" {
		return NewApiResponse[ResponseUserEditCurrentProfile](&ErrIllegalParam, Unsatisfied, nil)
	}
	if ue.OriginPassword != "" && ue.NewPassword != "" {
		if err := passwordValidator.CheckString(ue.NewPassword); err != nil {
			return NewApiResponse[ResponseUserEditCurrentProfile](err, Unsatisfied, nil)
		}
	} else if ue.OriginPassword != "" && ue.NewPassword == "" {
		return NewApiResponse[ResponseUserEditCurrentProfile](&ErrNewPasswordRequired, Unsatisfied, nil)
	} else if ue.OriginPassword == "" && ue.NewPassword != "" {
		return NewApiResponse[ResponseUserEditCurrentProfile](&ErrOriginPasswordRequired, Unsatisfied, nil)
	}
	if ue.Username != "" {
		if err := usernameValidator.CheckString(ue.Username); err != nil {
			return NewApiResponse[ResponseUserEditCurrentProfile](err, Unsatisfied, nil)
		}
	}
	if ue.Email != "" {
		if err := emailValidator.CheckString(ue.Email); err != nil {
			return NewApiResponse[ResponseUserEditCurrentProfile](err, Unsatisfied, nil)
		}
	}
	if ue.QQ > 0 {
		if err := checkQQ(ue.QQ); err != nil {
			return NewApiResponse[ResponseUserEditCurrentProfile](err, Unsatisfied, nil)
		}
	}

	user, err := database.GetUserById(ue.ID)
	if errors.Is(err, database.ErrUserNotFound) {
		return NewApiResponse[ResponseUserEditCurrentProfile](&ErrUserNotFound, Unsatisfied, nil)
	} else if err != nil {
		return NewApiResponse[ResponseUserEditCurrentProfile](&ErrDatabaseFail, Unsatisfied, nil)
	}

	updateInfo := make(map[string]interface{})

	if ue.Username != "" || ue.Email != "" {
		exist, _ := database.IsUserIdentifierTaken(0, ue.Username, ue.Email)
		if exist {
			return NewApiResponse[ResponseUserEditCurrentProfile](&ErrIdentifierTaken, Unsatisfied, nil)
		}
		if ue.Username != "" && ue.Username != user.Username {
			user.Username = ue.Username
			updateInfo["username"] = ue.Username
		}
		if ue.Email != "" && ue.Email != user.Email {
			user.Email = ue.Email
			updateInfo["email"] = ue.Email
		}
	}
	if ue.QQ > 0 && ue.QQ != user.QQ {
		user.QQ = ue.QQ
		updateInfo["qq"] = ue.QQ
	}

	if ue.OriginPassword != "" {
		password, err := user.UpdatePassword(ue.OriginPassword, ue.NewPassword)
		if errors.Is(err, database.ErrUserNotFound) {
			return NewApiResponse[ResponseUserEditCurrentProfile](&ErrUserNotFound, Unsatisfied, nil)
		} else if errors.Is(err, database.ErrOldPassword) {
			return NewApiResponse[ResponseUserEditCurrentProfile](&ErrOriginPassword, Unsatisfied, nil)
		} else if err != nil {
			return NewApiResponse[ResponseUserEditCurrentProfile](&ErrDatabaseFail, Unsatisfied, nil)
		}
		updateInfo["password"] = password
	}

	if err := user.UpdateInfo(updateInfo); err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			return NewApiResponse[ResponseUserEditCurrentProfile](&ErrUserNotFound, Unsatisfied, nil)
		} else {
			return NewApiResponse[ResponseUserEditCurrentProfile](&ErrDatabaseFail, Unsatisfied, nil)
		}
	}

	return NewApiResponse(&SuccessEditCurrentProfile, Unsatisfied, &ResponseUserEditCurrentProfile{
		Username:       user.Username,
		Email:          user.Email,
		Cid:            user.Cid,
		QQ:             user.QQ,
		Rating:         user.Rating,
		TotalPilotTime: user.TotalPilotTime,
		TotalAtcTime:   user.TotalAtcTime,
		Permission:     user.Permission,
	})
}
