// Package interfaces
package interfaces

import "html/template"

type EmailServiceInterface interface {
	RenderTemplate(template *template.Template, data interface{}) (string, error)
	VerifyCode(email string, code int, cid int) error
	SendEmailCode(email string, cid int) error
	SendEmailVerifyCode(req *RequestEmailVerifyCode) *ApiResponse[ResponseEmailVerifyCode]
}

type RequestEmailVerifyCode struct {
	Email string `json:"email"`
	Cid   int    `json:"cid"`
}

type ResponseEmailVerifyCode struct {
	Email string `json:"email"`
}
