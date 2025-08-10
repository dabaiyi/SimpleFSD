// Package controller
package controller

import (
	"github.com/half-nothing/fsd-server/internal/server/service"
	"github.com/labstack/echo/v4"
)

func UserRegister(ctx echo.Context) error {
	data := service.RegisterUserData{}
	if err := ctx.Bind(&data); err != nil {
		return service.NewErrorResponse(ctx, &service.ParamLackError)
	}
	return data.RegisterUser().Response(ctx)
}
