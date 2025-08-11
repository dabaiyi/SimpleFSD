// Package controller
package controller

import (
	"github.com/half-nothing/fsd-server/internal/server/service"
	"github.com/labstack/echo/v4"
)

func SendVerifyEmail(ctx echo.Context) error {
	data := service.EmailVerifyCodeData{}
	if err := ctx.Bind(&data); err != nil {
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.SendEmailVerifyCode().Response(ctx)
}
