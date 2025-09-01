package database

import (
	"context"
	. "fmt"
	c "github.com/half-nothing/simple-fsd/internal/config"
	. "github.com/half-nothing/simple-fsd/internal/interfaces/operation"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type DBCloseCallback struct {
	db *gorm.DB
}

func NewDBCloseCallback(db *gorm.DB) *DBCloseCallback {
	return &DBCloseCallback{db: db}
}

func (dc *DBCloseCallback) Invoke(ctx context.Context) error {
	c.InfoF("Closing operation connection")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := dc.db.DB()
	if err != nil {
		return err
	}
	err = db.Close()
	return err
}

func ConnectDatabase(config *c.Config) (*DatabaseOperations, error) {
	queryTimeout := config.Database.QueryDuration

	connection := config.Database.GetConnection()

	gormConfig := gorm.Config{}
	gormConfig.DefaultTransactionTimeout = 5 * time.Second
	gormConfig.PrepareStmt = true
	gormConfig.TranslateError = true

	if config.DebugMode {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(connection, &gormConfig)
	if err != nil {
		return nil, Errorf("error occured while connecting to operation: %v", err)
	}

	if err = db.Migrator().AutoMigrate(&User{}, &FlightPlan{}, &History{}, &Activity{}, &ActivityATC{}, &ActivityPilot{}, &ActivityFacility{}, &AuditLog{}); err != nil {
		return nil, Errorf("error occured while migrating operation: %v", err)
	}

	dbPool, err := db.DB()
	if err != nil {
		return nil, Errorf("error occured while creating operation pool: %v", err)
	}

	maxOpenConnections := config.Database.ServerMaxConnections * 4 / 5 // 不超过数据库最大连接的80%
	maxIdleConnections := maxOpenConnections / 5                       // 空闲连接约为最大连接的20%

	dbPool.SetMaxIdleConns(maxIdleConnections)
	dbPool.SetMaxOpenConns(maxOpenConnections)
	dbPool.SetConnMaxLifetime(config.Database.ConnectIdleDuration)

	err = dbPool.Ping()
	if err != nil {
		return nil, Errorf("error occured while pinging operation: %v", err)
	}
	c.Info("Database initialized and connection established")

	c.GetCleaner().Add(NewDBCloseCallback(db))

	userOperation := NewUserOperation(db, queryTimeout, config.Server.General)
	flightPlanOperation := NewFlightPlanOperation(db, queryTimeout, config.Server.General)
	historyOperation := NewHistoryOperation(db, queryTimeout)
	activityOperation := NewActivityOperation(db, queryTimeout)
	auditLogOperation := NewAuditLogOperation(db, queryTimeout)

	return NewDatabaseOperations(userOperation, flightPlanOperation, historyOperation, activityOperation, auditLogOperation), nil
}
