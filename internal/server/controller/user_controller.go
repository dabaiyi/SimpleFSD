// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/service"
	"github.com/labstack/echo/v4"
)

func UserRegister(ctx echo.Context) error {
	data := service.RequestRegisterUser{}
	if err := ctx.Bind(&data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.RegisterUser().Response(ctx)
}

func UserLogin(ctx echo.Context) error {
	data := service.RequestUserLogin{}
	if err := ctx.Bind(&data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.UserLogin().Response(ctx)
}

func CheckUserAvailability(ctx echo.Context) error {
	data := service.RequestUserAvailability{}
	if err := ctx.Bind(&data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.CheckAvailability().Response(ctx)
}

func GetCurrentUserProfile(ctx echo.Context) error {
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*service.Claims)
	data := service.RequestUserCurrentProfile{Uid: claim.Uid}
	return data.GetCurrentProfile().Response(ctx)
}

func EditCurrentProfile(ctx echo.Context) error {
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*service.Claims)
	data := service.RequestUserEditCurrentProfile{}
	if err := ctx.Bind(&data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	data.ID = claim.Uid
	return data.EditCurrentProfile().Response(ctx)
}
