// Package service
package service

import "github.com/half-nothing/fsd-server/internal/server/database"

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
	ParamError      = CodeStatus{"PARAM_ERROR", "参数不正确", BadRequest}
	ParamLackError  = CodeStatus{"PARAM_LACK_ERROR", "缺少参数", BadRequest}
	DatabaseError   = CodeStatus{"DATABASE_ERROR", "服务器内部错误", ServerInternalError}
	NoPermission    = CodeStatus{"NO_PERMISSION", "无权这么做", PermissionDenied}
	RegisterSuccess = CodeStatus{"REGISTER_SUCCESS", "注册成功", Ok}
)

func (rud *RegisterUserData) RegisterUser() *ApiResponse[RegisterUserResponse] {
	if rud.Username == "" || rud.Email == "" || rud.Password == "" || rud.Cid == 0 || rud.EmailCode == 0 {
		return NewApiResponse[RegisterUserResponse](&ParamError, Unsatisfied, nil)
	}
	// TODO: 验证邮箱验证码
	//if rud.EmailCode == {
	//
	//}
	user := database.NewUser(rud.Username, rud.Email, rud.Cid, rud.Password)
	err := user.Save()
	if err != nil {
		return NewApiResponse[RegisterUserResponse](&DatabaseError, Unsatisfied, nil)
	}
	return NewApiResponse(&RegisterSuccess, Unsatisfied, &RegisterUserResponse{
		Username:   rud.Username,
		Token:      "",
		FlushToken: "",
	})
}
