// Package service
package service

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/labstack/echo/v4"
)

type HttpCode int

const (
	Unsatisfied         HttpCode = 0
	Ok                  HttpCode = 200
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
	jwt.RegisteredClaims
}

func NewClaims(user *database.User) *Claims {
	return &Claims{
		Username:   user.Username,
		Permission: user.Permission,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "",
			Subject:   "",
			Audience:  nil,
			ExpiresAt: nil,
			NotBefore: nil,
			IssuedAt:  nil,
			ID:        "",
		},
	}
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
