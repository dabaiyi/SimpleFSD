// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	"github.com/labstack/echo/v4"
)

type ActivityControllerInterface interface {
	GetActivities(ctx echo.Context) error
	GetActivityInfo(ctx echo.Context) error
	AddActivity(ctx echo.Context) error
	DeleteActivity(ctx echo.Context) error
	EditActivity(ctx echo.Context) error
	ControllerJoin(ctx echo.Context) error
	ControllerLeave(ctx echo.Context) error
	PilotJoin(ctx echo.Context) error
	PilotLeave(ctx echo.Context) error
	EditPilotStatus(ctx echo.Context) error
}

type ActivityController struct {
	activityService ActivityServiceInterface
}

func NewActivityController(activityService ActivityServiceInterface) *ActivityController {
	return &ActivityController{activityService}
}

func (controller ActivityController) GetActivities(ctx echo.Context) error {
	data := &RequestGetActivities{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	return controller.activityService.GetActivities(data).Response(ctx)
}

func (controller ActivityController) GetActivityInfo(ctx echo.Context) error {
	data := &RequestActivityInfo{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	return controller.activityService.GetActivityInfo(data).Response(ctx)
}

func (controller ActivityController) AddActivity(ctx echo.Context) error {
	data := &RequestAddActivity{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Cid = claim.Cid
	data.Permission = claim.Permission
	return controller.activityService.AddActivity(data).Response(ctx)
}

func (controller ActivityController) DeleteActivity(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (controller ActivityController) EditActivity(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (controller ActivityController) ControllerJoin(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (controller ActivityController) ControllerLeave(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (controller ActivityController) PilotJoin(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (controller ActivityController) PilotLeave(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (controller ActivityController) EditPilotStatus(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}
