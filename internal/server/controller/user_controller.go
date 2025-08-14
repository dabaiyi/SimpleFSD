// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	. "github.com/half-nothing/fsd-server/internal/server/service/interfaces"
	"github.com/labstack/echo/v4"
)

type UserHandlerInterface interface {
	UserRegisterHandler(ctx echo.Context) error
	UserLoginHandler(ctx echo.Context) error
	CheckUserAvailabilityHandler(ctx echo.Context) error
	GetCurrentUserProfileHandler(ctx echo.Context) error
	EditCurrentProfileHandler(ctx echo.Context) error
	GetUserProfileHandler(ctx echo.Context) error
	EditProfileHandler(ctx echo.Context) error
	GetAllUsers(ctx echo.Context) error
}

type UserController struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserController {
	return &UserController{service}
}

func (controller *UserController) UserRegisterHandler(ctx echo.Context) error {
	data := &RequestRegisterUser{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	return controller.service.RegisterUser(data).Response(ctx)
}

func (controller *UserController) UserLoginHandler(ctx echo.Context) error {
	data := &RequestUserLogin{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	return controller.service.UserLogin(data).Response(ctx)
}

func (controller *UserController) CheckUserAvailabilityHandler(ctx echo.Context) error {
	data := &RequestUserAvailability{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	return controller.service.CheckAvailability(data).Response(ctx)
}

func (controller *UserController) GetCurrentUserProfileHandler(ctx echo.Context) error {
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data := &RequestUserCurrentProfile{Uid: claim.Uid}
	return controller.service.GetCurrentProfile(data).Response(ctx)
}

func (controller *UserController) EditCurrentProfileHandler(ctx echo.Context) error {
	data := &RequestUserEditCurrentProfile{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.ID = claim.Uid
	data.Cid = claim.Cid
	return controller.service.EditCurrentProfile(data).Response(ctx)
}

func (controller *UserController) GetUserProfileHandler(ctx echo.Context) error {
	data := &RequestUserProfile{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.service.GetUserProfile(data).Response(ctx)
}

func (controller *UserController) EditProfileHandler(ctx echo.Context) error {
	data := &RequestUserEditProfile{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Cid = claim.Cid
	data.Permission = claim.Permission
	return controller.service.EditUserProfile(data).Response(ctx)
}

func (controller *UserController) GetAllUsers(ctx echo.Context) error {
	data := &RequestUserList{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.service.GetUserList(data).Response(ctx)
}
