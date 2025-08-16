// Package controller
package controller

import (
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	"github.com/labstack/echo/v4"
)

type ServerControllerInterface interface {
	GetServerConfig(ctx echo.Context) error
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
