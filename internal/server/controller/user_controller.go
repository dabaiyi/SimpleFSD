// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/half-nothing/fsd-server/internal/server/service"
	"github.com/labstack/echo/v4"
)

func UserRegister(ctx echo.Context) error {
	data := service.RegisterUserData{}
	if err := ctx.Bind(&data); err != nil {
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.RegisterUser().Response(ctx)
}

func UserLogin(ctx echo.Context) error {
	data := service.UserLoginData{}
	if err := ctx.Bind(&data); err != nil {
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.UserLogin().Response(ctx)
}

func CheckUserAvailability(ctx echo.Context) error {
	data := service.UserAvailabilityData{}
	if err := ctx.Bind(&data); err != nil {
		return service.NewErrorResponse(ctx, &service.ErrLackParam)
	}
	return data.CheckAvailability().Response(ctx)
}

func GetCurrentUserProfile(ctx echo.Context) error {
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*service.Claims)
	data := service.UserCurrentProfileData{Username: claim.Username}
	return data.GetCurrentProfile().Response(ctx)
}
