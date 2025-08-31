// Package controller
package controller

import (
	. "github.com/half-nothing/simple-fsd/internal/interfaces/service"
	"github.com/labstack/echo/v4"
)

type ServerControllerInterface interface {
	GetServerConfig(ctx echo.Context) error
	GetServerInfo(ctx echo.Context) error
	GetServerOnlineTime(ctx echo.Context) error
}

type ServerController struct {
	serverService ServerServiceInterface
}

func NewServerController(serverService ServerServiceInterface) *ServerController {
	return &ServerController{serverService}
}

func (controller *ServerController) GetServerConfig(ctx echo.Context) error {
	return controller.serverService.GetServerConfig().Response(ctx)
}

func (controller *ServerController) GetServerInfo(ctx echo.Context) error {
	return controller.serverService.GetServerInfo().Response(ctx)
}

func (controller *ServerController) GetServerOnlineTime(ctx echo.Context) error {
	return controller.serverService.GetTimeRating().Response(ctx)
}
