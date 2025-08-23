// Package service
package service

import (
	"github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	"github.com/labstack/echo/v4"
)

type UserServiceInterface interface {
	UserRegister(req *RequestUserRegister) *ApiResponse[ResponseUserRegister]
	UserLogin(req *RequestUserLogin) *ApiResponse[ResponseUserLogin]
	CheckAvailability(req *RequestUserAvailability) *ApiResponse[ResponseUserAvailability]
	GetCurrentProfile(req *RequestUserCurrentProfile) *ApiResponse[ResponseUserCurrentProfile]
	EditCurrentProfile(req *RequestUserEditCurrentProfile) *ApiResponse[ResponseUserEditCurrentProfile]
	GetUserProfile(req *RequestUserProfile) *ApiResponse[ResponseUserProfile]
	EditUserProfile(req *RequestUserEditProfile) *ApiResponse[ResponseUserEditProfile]
	GetUserList(req *RequestUserList) *ApiResponse[ResponseUserList]
	EditUserPermission(req *RequestUserEditPermission) *ApiResponse[ResponseUserEditPermission]
	EditUserRating(req *RequestUserEditRating) *ApiResponse[ResponseUserEditRating]
}

type RequestUserRegister struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Cid       int    `json:"cid"`
	EmailCode int    `json:"email_code"`
}

type ResponseUserRegister struct {
	User       *operation.User `json:"user"`
	Token      string          `json:"token"`
	FlushToken string          `json:"flush_token"`
}

type RequestUserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseUserLogin struct {
	User       *operation.User `json:"user"`
	Token      string          `json:"token"`
	FlushToken string          `json:"flush_token"`
}

type RequestUserAvailability struct {
	Username string `query:"username"`
	Email    string `query:"email"`
	Cid      string `query:"cid"`
}

type ResponseUserAvailability bool

type RequestUserCurrentProfile struct {
	Uid uint
}

type ResponseUserCurrentProfile operation.User

type RequestUserEditCurrentProfile struct {
	ID             uint   `json:"id"`
	Cid            int    `json:"cid"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	EmailCode      int    `json:"email_code"`
	QQ             int    `json:"qq"`
	OriginPassword string `json:"origin_password"`
	NewPassword    string `json:"new_password"`
}

type ResponseUserEditCurrentProfile operation.User
type RequestUserProfile struct {
	JwtHeader
	TargetUid uint `param:"uid"`
}

type ResponseUserProfile operation.User

type RequestUserList struct {
	JwtHeader
	Page     int `query:"page_number"`
	PageSize int `query:"page_size"`
}

type ResponseUserList struct {
	Items    []*operation.User `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

type RequestUserEditProfile struct {
	JwtHeader
	TargetUid uint `param:"uid"`
	RequestUserEditCurrentProfile
}

type ResponseUserEditProfile operation.User

type RequestUserEditPermission struct {
	JwtHeader
	TargetUid   uint     `param:"uid"`
	Permissions echo.Map `json:"permissions"`
}

type ResponseUserEditPermission operation.User

type RequestUserEditRating struct {
	JwtHeader
	TargetUid uint `param:"uid"`
	Rating    int  `json:"rating"`
}

type ResponseUserEditRating operation.User
