// Package service
package service

import (
	"github.com/half-nothing/fsd-server/internal/interfaces/fsd"
	"github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"html/template"
)

type EmailServiceInterface interface {
	RenderTemplate(template *template.Template, data interface{}) (string, error)
	VerifyCode(email string, code int, cid int) error
	SendEmailCode(email string, cid int) error
	SendEmailVerifyCode(req *RequestEmailVerifyCode) *ApiResponse[ResponseEmailVerifyCode]
	SendPermissionChangeEmail(user *operation.User, operator *operation.User) error
	SendRatingChangeEmail(user *operation.User, operator *operation.User, oldRating, newRating fsd.Rating) error
	SendKickedFromServerEmail(user *operation.User, operator *operation.User, reason string) error
}

type RequestEmailVerifyCode struct {
	Email string `json:"email"`
	Cid   int    `json:"cid"`
}

type ResponseEmailVerifyCode struct {
	Email string `json:"email"`
}
