// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/simple-fsd/internal/config"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/labstack/echo/v4"
)

type ActivityControllerInterface interface {
	GetActivities(ctx echo.Context) error
	GetActivitiesPage(ctx echo.Context) error
	GetActivityInfo(ctx echo.Context) error
	AddActivity(ctx echo.Context) error
	DeleteActivity(ctx echo.Context) error
	ControllerJoin(ctx echo.Context) error
	ControllerLeave(ctx echo.Context) error
	PilotJoin(ctx echo.Context) error
	PilotLeave(ctx echo.Context) error
	EditActivity(ctx echo.Context) error
	EditActivityStatus(ctx echo.Context) error
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

func (controller ActivityController) GetActivitiesPage(ctx echo.Context) error {
	data := &RequestGetActivitiesPage{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.activityService.GetActivitiesPage(data).Response(ctx)
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
	data := &RequestDeleteActivity{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.activityService.DeleteActivity(data).Response(ctx)
}

func (controller ActivityController) ControllerJoin(ctx echo.Context) error {
	data := &RequestControllerJoin{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	data.Rating = claim.Rating
	return controller.activityService.ControllerJoin(data).Response(ctx)
}

func (controller ActivityController) ControllerLeave(ctx echo.Context) error {
	data := &RequestControllerLeave{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	data.Cid = claim.Cid
	return controller.activityService.ControllerLeave(data).Response(ctx)
}

func (controller ActivityController) PilotJoin(ctx echo.Context) error {
	data := &RequestPilotJoin{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	data.Cid = claim.Cid
	return controller.activityService.PilotJoin(data).Response(ctx)
}

func (controller ActivityController) PilotLeave(ctx echo.Context) error {
	data := &RequestPilotLeave{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	data.Cid = claim.Cid
	return controller.activityService.PilotLeave(data).Response(ctx)
}

func (controller ActivityController) EditActivity(ctx echo.Context) error {
	data := &RequestEditActivity{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.activityService.EditActivity(data).Response(ctx)
}

func (controller ActivityController) EditActivityStatus(ctx echo.Context) error {
	data := &RequestEditActivityStatus{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.activityService.EditActivityStatus(data).Response(ctx)
}

func (controller ActivityController) EditPilotStatus(ctx echo.Context) error {
	data := &RequestEditPilotStatus{}
	if err := ctx.Bind(data); err != nil {
		c.ErrorF("error binding data: %v", err)
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	data.Cid = claim.Cid
	return controller.activityService.EditPilotStatus(data).Response(ctx)
}
