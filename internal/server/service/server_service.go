// Package service
package service

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/defination/fsd"
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	"github.com/half-nothing/fsd-server/internal/utils"
)

type ServerService struct {
	config       *c.Config
	serverConfig *utils.CachedValue[ServerConfig]
}

func NewServerService(config *c.Config) *ServerService {
	service := &ServerService{
		config: config,
	}
	service.serverConfig = utils.NewCachedValue[ServerConfig](0, func() *ServerConfig { return service.getServerConfig() })
	return service
}

func (serverService *ServerService) getServerConfig() *ServerConfig {
	return &ServerConfig{
		Limits: &ServerLimits{
			UsernameLengthMin: serverService.config.Server.HttpServer.UsernameLengthMin,
			UsernameLengthMax: serverService.config.Server.HttpServer.UsernameLengthMax,
			PasswordLengthMin: serverService.config.Server.HttpServer.PasswordLengthMin,
			PasswordLengthMax: serverService.config.Server.HttpServer.PasswordLengthMax,
			EmailLengthMin:    serverService.config.Server.HttpServer.EmailLengthMin,
			EmailLengthMax:    serverService.config.Server.HttpServer.EmailLengthMax,
			CidMin:            serverService.config.Server.HttpServer.CidMin,
			CidMax:            serverService.config.Server.HttpServer.CidMax,
			SimulatorServer:   serverService.config.Server.General.SimulatorServer,
		},
		Facilities: &fsd.Facilities,
		Ratings:    &fsd.Ratings,
	}
}

var SuccessGetServerConfig = ApiStatus{StatusName: "GET_SERVER_CONFIG", Description: "成功获取服务器配置", HttpCode: Ok}

func (serverService *ServerService) GetServerConfig() *ApiResponse[ResponseGetServerConfig] {
	return NewApiResponse(&SuccessGetServerConfig, Unsatisfied, &ResponseGetServerConfig{
		ServerConfig: serverService.serverConfig.GetValue(),
	})
}
