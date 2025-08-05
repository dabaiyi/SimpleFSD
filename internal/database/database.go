package database

import (
	. "fmt"
	c "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"time"
)

var (
	database     *gorm.DB
	queryTimeout time.Duration
)

func ConnectDatabase() error {
	config, _ := c.GetConfig()
	queryTimeout = utils.ParseStringTime(config.DatabaseConfig.QueryTimeout)

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
	dbPool.SetConnMaxLifetime(utils.ParseStringTime(config.DatabaseConfig.ConnectIdleTimeout))
	err = dbPool.Ping()
	if err != nil {
		return Errorf("error occured while pinging database: %v", err)
	}
	return nil
}
