package main

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/server/packet"
)

func main() {
	config, err := c.ReadConfig()
	if err != nil {
		c.FatalF("Error occurred while reading config %v", err)
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
	c.Info("Database initialized and connection established")
	packet.SyncRatingConfig()
	if config.ServerConfig.EnableGRPC {
		go server.StartGRPCServer()
	}
	server.StartFSDServer()
}
