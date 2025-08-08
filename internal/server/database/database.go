package database

import (
	"context"
	. "fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/url"
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
	c.InfoF("Closing database connection")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.DB()
	if err != nil {
		return err
	}
	err = db.Close()
	return err
}

func ConnectDatabase() error {
	config, _ = c.GetConfig()
	queryTimeout = config.DatabaseConfig.QueryDuration

	encodedUser := url.QueryEscape(config.DatabaseConfig.Username)
	encodedPass := url.QueryEscape(config.DatabaseConfig.Password)
	dsn := Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		encodedUser,
		encodedPass,
		config.DatabaseConfig.Host,
		config.DatabaseConfig.Port,
		config.DatabaseConfig.Database,
	)

	connectionConfig := gorm.Config{}
	connectionConfig.DefaultTransactionTimeout = 5 * time.Second
	connectionConfig.PrepareStmt = true

	if config.DebugMode {
		connectionConfig.Logger = logger.Default.LogMode(logger.Error)
	} else {
		connectionConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(mysql.Open(dsn), &connectionConfig)
	if err != nil {
		return Errorf("error occured while connecting to database: %v", err)
	}
	database = db

	err = db.Migrator().AutoMigrate(&User{}, &FlightPlan{})
	if err != nil {
		return Errorf("error occured while migrating database: %v", err)
	}

	dbPool, err := db.DB()
	if err != nil {
		return Errorf("error occured while creating database pool: %v", err)
	}

	maxOpenConnections := float32(config.DatabaseConfig.ServerMaxConnections) * 0.8 // 不超过数据库最大连接的80%
	maxIdleConnections := maxOpenConnections / 5                                    // 空闲连接约为最大连接的20%

	dbPool.SetMaxIdleConns(int(maxIdleConnections))
	dbPool.SetMaxOpenConns(int(maxOpenConnections))
	dbPool.SetConnMaxLifetime(config.DatabaseConfig.ConnectIdleDuration)
	err = dbPool.Ping()
	if err != nil {
		return Errorf("error occured while pinging database: %v", err)
	}
	c.NewCleaner().Add(NewDBCloseCallback())
	return nil
}
