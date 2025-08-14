// Package service
package interfaces

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

type ApiStatus struct {
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
	Uid        uint   `json:"uid"`
	Cid        int    `json:"cid"`
	Username   string `json:"username"`
	Permission int64  `json:"permission"`
	FlushToken bool   `json:"flushToken"`
	jwt.RegisteredClaims
}

type JwtHeader struct {
	Uid        uint
	Permission int64
}

func NewClaims(user *database.User, flushToken bool) *Claims {
	config, _ := c.GetConfig()
	expiredDuration := config.Server.HttpServer.JWT.ExpiresDuration
	if flushToken {
		expiredDuration += config.Server.HttpServer.JWT.RefreshDuration
	}
	return &Claims{
		Uid:        user.ID,
		Cid:        user.Cid,
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

func (claim *Claims) GenerateKey() string {
	config, _ := c.GetConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, _ := token.SignedString([]byte(config.Server.HttpServer.JWT.Secret))
	return tokenString
}

func (res *ApiResponse[T]) Response(ctx echo.Context) error {
	return ctx.JSON(res.HttpCode, res)
}

var (
	ErrIllegalParam = ApiStatus{"PARAM_ERROR", "参数不正确", BadRequest}
	ErrLackParam    = ApiStatus{"PARAM_LACK_ERROR", "缺少参数", BadRequest}
	ErrNoPermission = ApiStatus{"NO_PERMISSION", "无权这么做", PermissionDenied}
	ErrDatabaseFail = ApiStatus{"DATABASE_ERROR", "服务器内部错误", ServerInternalError}
)

func NewErrorResponse(ctx echo.Context, codeStatus *ApiStatus) error {
	return NewApiResponse[any](codeStatus, Unsatisfied, nil).Response(ctx)
}

func NewApiResponse[T any](codeStatus *ApiStatus, httpCode HttpCode, data *T) *ApiResponse[T] {
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
