package main

import (
	c "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/database"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/event"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/logger"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server"
)

func main() {
	_, err := c.ReadConfig()
	if err != nil {
		logger.FatalF("Error occurred while reading config %v", err)
		return
	}
	loggerCallback := logger.Init()
	logger.Info("Application initializing...")
	cleaner := event.NewCleaner()
	cleaner.Init(loggerCallback)
	defer cleaner.Clean()
	err = database.ConnectDatabase()
	if err != nil {
		logger.FatalF("Error occurred while initializing database, details: %v", err)
		return
	}
	logger.Info("Database initialized and connection established")
	server.StartServer()
}
