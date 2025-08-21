package main

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/server/defination/fsd"
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
		c.FatalF("Error occurred while initializing database, details: %v", err)
		return
	}
	if config.Server.HttpServer.Enabled {
		go server.StartHttpServer(config)
	}
	if config.Server.GRPCServer.Enabled {
		go server.StartGRPCServer(config.Server.GRPCServer)
	}
	server.StartFSDServer(config)
}
