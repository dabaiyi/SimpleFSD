// Package controller
package controller

import (
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/labstack/echo/v4"
)

type EmailControllerInterface interface {
	SendVerifyEmail(ctx echo.Context) error
}

type EmailController struct {
	emailService EmailServiceInterface
}

func NewEmailController(emailService EmailServiceInterface) *EmailController {
	return &EmailController{emailService}
}

func (controller *EmailController) SendVerifyEmail(ctx echo.Context) error {
	data := &RequestEmailVerifyCode{}
	if err := ctx.Bind(data); err != nil {
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	return controller.emailService.SendEmailVerifyCode(data).Response(ctx)
}
