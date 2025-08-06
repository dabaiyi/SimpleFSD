package main

import (
	c "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/database"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/packet"
)

func main() {
	_, err := c.ReadConfig()
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
	packet.SyncRatingConfig()
	c.Info("Database initialized and connection established")
	server.StartServer()
}
