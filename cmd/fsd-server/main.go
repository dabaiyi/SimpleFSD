package main

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/server/packet"
)

func main() {
	config, err := c.GetConfig()
	if err != nil {
		c.FatalF("Error occurred while reading config %v", err)
		return
	}
	err = packet.SyncRatingConfig()
	if err != nil {
		c.FatalF("Error occurred while handle rating config, details: %v", err)
		return
	}
	loggerCallback := c.Init()
	c.Info("Application initializing...")
	cleaner := c.NewCleaner()
	cleaner.Init(loggerCallback)
	defer cleaner.Clean()
	err = database.ConnectDatabase()
	if err != nil {
		c.FatalF("Error occurred while initializing database, details: %v", err)
		return
	}
	if config.Server.HttpServer.Enabled {
		go server.StartHttpServer()
	}
	if config.Server.GRPCServer.Enabled {
		go server.StartGRPCServer()
	}
	server.StartFSDServer()
}
