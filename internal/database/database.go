package database

import (
	"context"
	. "fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	database2 "github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var (
	database     *gorm.DB
	config       *c.Config
	queryTimeout time.Duration
)

type DBCloseCallback struct {
}

func NewDBCloseCallback() *DBCloseCallback {
	return &DBCloseCallback{}
}

func (dc *DBCloseCallback) Invoke(ctx context.Context) error {
	c.InfoF("Closing operation connection")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.DB()
	if err != nil {
		return err
	}
	err = db.Close()
	return err
}

func ConnectDatabase(config *c.Config) error {
	queryTimeout = config.Database.QueryDuration

	connection := config.Database.GetConnection()

	connectionConfig := gorm.Config{}
	connectionConfig.DefaultTransactionTimeout = 5 * time.Second
	connectionConfig.PrepareStmt = true

	if config.DebugMode {
		connectionConfig.Logger = logger.Default.LogMode(logger.Error)
	} else {
		connectionConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(connection, &connectionConfig)
	if err != nil {
		return Errorf("error occured while connecting to operation: %v", err)
	}
	database = db

	if err = db.Migrator().AutoMigrate(&database2.User{}, &database2.FlightPlan{}, &database2.History{}, &database2.Activity{}, &database2.ActivityATC{}, &database2.ActivityPilot{}, &database2.ActivityFacility{}); err != nil {
		return Errorf("error occured while migrating operation: %v", err)
	}

	dbPool, err := db.DB()
	if err != nil {
		return Errorf("error occured while creating operation pool: %v", err)
	}

	maxOpenConnections := config.Database.ServerMaxConnections * 4 / 5 // 不超过数据库最大连接的80%
	maxIdleConnections := maxOpenConnections / 5                       // 空闲连接约为最大连接的20%

	dbPool.SetMaxIdleConns(maxIdleConnections)
	dbPool.SetMaxOpenConns(maxOpenConnections)
	dbPool.SetConnMaxLifetime(config.Database.ConnectIdleDuration)

	err = dbPool.Ping()
	if err != nil {
		return Errorf("error occured while pinging operation: %v", err)
	}
	c.Info("Database initialized and connection established")

	c.GetCleaner().Add(NewDBCloseCallback())
	return nil
}
