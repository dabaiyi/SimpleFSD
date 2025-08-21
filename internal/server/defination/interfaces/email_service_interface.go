// Package interfaces
package interfaces

import (
	"github.com/half-nothing/fsd-server/internal/server/defination/database"
	"github.com/half-nothing/fsd-server/internal/server/defination/fsd"
	"html/template"
)

type EmailServiceInterface interface {
	RenderTemplate(template *template.Template, data interface{}) (string, error)
	VerifyCode(email string, code int, cid int) error
	SendEmailCode(email string, cid int) error
	SendEmailVerifyCode(req *RequestEmailVerifyCode) *ApiResponse[ResponseEmailVerifyCode]
	SendPermissionChangeEmail(user *database.User, operator *database.User) error
	SendRatingChangeEmail(user *database.User, operator *database.User, oldRating, newRating fsd.Rating) error
	SendKickedFromServerEmail(user *database.User, operator *database.User, reason string) error
}

type RequestEmailVerifyCode struct {
	Email string `json:"email"`
	Cid   int    `json:"cid"`
}

type ResponseEmailVerifyCode struct {
	Email string `json:"email"`
}
