// Package service
package service

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/interfaces/fsd"
	. "github.com/half-nothing/fsd-server/internal/interfaces/service"
	"github.com/half-nothing/fsd-server/internal/utils"
)

type ServerService struct {
	config       *c.ServerConfig
	serverConfig *utils.CachedValue[ServerConfig]
}

func NewServerService(config *c.ServerConfig) *ServerService {
	service := &ServerService{
		config: config,
	}
	service.serverConfig = utils.NewCachedValue[ServerConfig](0, func() *ServerConfig { return service.getServerConfig() })
	return service
}

func (serverService *ServerService) getServerConfig() *ServerConfig {
	return &ServerConfig{
		Limits: &ServerLimits{
			UsernameLengthMin: serverService.config.HttpServer.Limits.UsernameLengthMin,
			UsernameLengthMax: serverService.config.HttpServer.Limits.UsernameLengthMax,
			PasswordLengthMin: serverService.config.HttpServer.Limits.PasswordLengthMin,
			PasswordLengthMax: serverService.config.HttpServer.Limits.PasswordLengthMax,
			EmailLengthMin:    serverService.config.HttpServer.Limits.EmailLengthMin,
			EmailLengthMax:    serverService.config.HttpServer.Limits.EmailLengthMax,
			CidMin:            serverService.config.HttpServer.Limits.CidMin,
			CidMax:            serverService.config.HttpServer.Limits.CidMax,
			SimulatorServer:   serverService.config.General.SimulatorServer,
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
