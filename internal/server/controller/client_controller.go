// Package controller
package controller

import (
	"github.com/golang-jwt/jwt/v5"
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	"github.com/labstack/echo/v4"
	"net/http"
)

type ClientControllerInterface interface {
	GetServerWhazzup(ctx echo.Context) error
	GetOnlineClients(ctx echo.Context) error
	SendMessageToClient(ctx echo.Context) error
	KillClient(ctx echo.Context) error
}

type ClientController struct {
	clientService ClientServiceInterface
}

func NewClientController(clientService ClientServiceInterface) *ClientController {
	return &ClientController{clientService: clientService}
}

func (controller *ClientController) GetOnlineClients(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, controller.clientService.GetOnlineClient())
}

func (controller *ClientController) SendMessageToClient(ctx echo.Context) error {
	data := &RequestSendMessageToClient{}
	if err := ctx.Bind(data); err != nil {
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Cid = claim.Cid
	data.Permission = claim.Permission
	return controller.clientService.SendMessageToClient(data).Response(ctx)
}

func (controller *ClientController) KillClient(ctx echo.Context) error {
	data := &RequestKillClient{}
	if err := ctx.Bind(data); err != nil {
		return NewErrorResponse(ctx, &ErrLackParam)
	}
	token := ctx.Get("user").(*jwt.Token)
	claim := token.Claims.(*Claims)
	data.Uid = claim.Uid
	data.Permission = claim.Permission
	return controller.clientService.KillClient(data).Response(ctx)
}
