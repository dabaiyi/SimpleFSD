package main

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/database"
	"github.com/half-nothing/fsd-server/internal/fsd_server"
	"github.com/half-nothing/fsd-server/internal/grpc_server"
	"github.com/half-nothing/fsd-server/internal/http_server"
	"github.com/half-nothing/fsd-server/internal/interfaces/fsd"
)

func main() {
	config := c.GetConfig()
	if err := fsd.SyncRatingConfig(config); err != nil {
		c.FatalF("Error occurred while handle rating config, details: %v", err)
		return
	}
	loggerCallback := c.Init(config)
	c.Info("Application initializing...")
	cleaner := c.GetCleaner()
	cleaner.Init(loggerCallback)
	defer cleaner.Clean()
	if err := database.ConnectDatabase(config); err != nil {
		c.FatalF("Error occurred while initializing operation, details: %v", err)
		return
	}
	if config.Server.HttpServer.Enabled {
		go http_server.StartHttpServer(config)
	}
	if config.Server.GRPCServer.Enabled {
		go grpc_server.StartGRPCServer(config.Server.GRPCServer)
	}
	fsd_server.StartFSDServer(config)
}
