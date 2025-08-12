// Package service
package service

import (
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/labstack/echo/v4"
	"time"
)

type HttpCode int

const (
	Unsatisfied         HttpCode = 0
	Ok                  HttpCode = 200
	NoContent           HttpCode = 204
	BadRequest          HttpCode = 400
	Unauthorized        HttpCode = 401
	PermissionDenied    HttpCode = 403
	NotFound            HttpCode = 404
	ServerInternalError HttpCode = 500
)

func (hc HttpCode) Code() int {
	return int(hc)
}

type CodeStatus struct {
	StatusName  string
	Description string
	HttpCode    HttpCode
}

type ApiResponse[T any] struct {
	HttpCode int    `json:"-"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Data     *T     `json:"data"`
}

type Claims struct {
	Username   string `json:"username"`
	Permission int64  `json:"permission"`
	FlushToken bool   `json:"flushToken"`
	jwt.RegisteredClaims
}

var (
	ErrIllegalParam = CodeStatus{"PARAM_ERROR", "参数不正确", BadRequest}
	ErrLackParam    = CodeStatus{"PARAM_LACK_ERROR", "缺少参数", BadRequest}
	ErrNoPermission = CodeStatus{"NO_PERMISSION", "无权这么做", PermissionDenied}
)

func NewClaims(user *database.User, flushToken bool) *Claims {
	config, _ := c.GetConfig()
	expiredDuration := config.Server.HttpServer.JWT.ExpiresDuration
	if flushToken {
		expiredDuration += config.Server.HttpServer.JWT.RefreshDuration
	}
	return &Claims{
		Username:   user.Username,
		Permission: user.Permission,
		FlushToken: flushToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "FsdHttpServer",
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiredDuration)),
		},
	}
}

func (claim *Claims) generateKey() string {
	config, _ := c.GetConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, _ := token.SignedString([]byte(config.Server.HttpServer.JWT.Secret))
	return tokenString
}

func NewErrorResponse(ctx echo.Context, codeStatus *CodeStatus) error {
	return NewApiResponse[any](codeStatus, Unsatisfied, nil).Response(ctx)
}

func NewApiResponse[T any](codeStatus *CodeStatus, httpCode HttpCode, data *T) *ApiResponse[T] {
	if httpCode == Unsatisfied {
		httpCode = codeStatus.HttpCode
	}
	if httpCode == Unsatisfied {
		httpCode = Ok
	}
	return &ApiResponse[T]{
		HttpCode: httpCode.Code(),
		Code:     codeStatus.StatusName,
		Message:  codeStatus.Description,
		Data:     data,
	}
}

func (res *ApiResponse[T]) Response(ctx echo.Context) error {
	return ctx.JSON(res.HttpCode, res)
}
