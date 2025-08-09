package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/url"
	"os"
	"runtime"
	"slices"
	"time"
)

type Config struct {
	DebugMode            bool          `json:"debug_mode"`            // 是否启用调试模式
	AppName              string        `json:"app_name"`              // 应用名称
	AppVersion           string        `json:"app_version"`           // 应用版本
	MaxWorkers           int           `json:"max_workers"`           // 并发线程数
	MaxBroadcastWorkers  int           `json:"max_broadcast_workers"` // 广播并发线程数
	SessionCleanTime     string        `json:"session_clean_time"`    // 会话保留时间
	SessionCleanDuration time.Duration `json:"-"`                     // 内部使用字段
	ServerConfig         struct {
		Host              string        `json:"host"`
		Port              uint64        `json:"port"`
		EnableGRPC        bool          `json:"enable_grpc"`
		GRPCPort          uint64        `json:"grpc_port"`
		GRPCCacheTime     string        `json:"grpc_cache_time"`
		GRPCCacheDuration time.Duration `json:"-"`
		HeartbeatInterval string        `json:"heartbeat_interval"`
		HeartbeatDuration time.Duration `json:"-"`
		Motd              []string      `json:"motd"`
	} `json:"server_config"`
	DatabaseConfig struct {
		Type                 string        `json:"type"`
		DBType               DatabaseType  `json:"-"`
		Host                 string        `json:"host"`
		Port                 int           `json:"port"`
		Username             string        `json:"username"`
		Password             string        `json:"password"`
		Database             string        `json:"database"`
		EnableSSL            bool          `json:"enable_ssl"`
		ConnectIdleTimeout   string        `json:"connect_idle_timeout"` // 连接空闲超时时间
		ConnectIdleDuration  time.Duration `json:"-"`
		QueryTimeout         string        `json:"connect_timeout"` // 每次查询超时时间
		QueryDuration        time.Duration `json:"-"`
		ServerMaxConnections int           `json:"server_max_connections"` // 最大连接池大小
	} `json:"database_config"`
	RatingConfig map[string]int `json:"rating_config"`
}

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite3"
)

var (
	config              Config
	initialized         = false
	allowedDatabaseType = []DatabaseType{MySQL, PostgreSQL, SQLite}
)

// ReadConfig 从配置文件读取配置
func ReadConfig() (*Config, error) {
	// 读取配置文件
	bytes, err := os.ReadFile("config.json")

	if err != nil {
		// 如果配置文件不存在，创建默认配置
		config.RatingConfig = make(map[string]int)
		if err := SaveConfig(); err != nil {
			return nil, err
		}
		return nil, errors.New("the configuration file does not exist and has been created. Please try again after editing the configuration file")
	}

	// 解析JSON配置
	err = json.Unmarshal(bytes, &config)

	if err != nil {
		return nil, fmt.Errorf("the configuration file does not contain valid JSON, %v", err)
	}

	config.DatabaseConfig.DBType = DatabaseType(config.DatabaseConfig.Type)

	if !slices.Contains(allowedDatabaseType, config.DatabaseConfig.DBType) {
		return nil, fmt.Errorf("database type %s is not allowed, support database is %v, please check the configuration file", config.DatabaseConfig.DBType, allowedDatabaseType)
	}

	if config.MaxBroadcastWorkers > runtime.NumCPU()*50 {
		config.MaxBroadcastWorkers = runtime.NumCPU() * 50
	}

	config.SessionCleanDuration, err = time.ParseDuration(config.SessionCleanTime)
	if err != nil {
		return nil, fmt.Errorf("time duration could not be parsed correctly, %v", err)
	}

	config.ServerConfig.GRPCCacheDuration, err = time.ParseDuration(config.ServerConfig.GRPCCacheTime)
	if err != nil {
		return nil, fmt.Errorf("time duration could not be parsed correctly, %v", err)
	}

	config.ServerConfig.HeartbeatDuration, err = time.ParseDuration(config.ServerConfig.HeartbeatInterval)
	if err != nil {
		return nil, fmt.Errorf("time duration could not be parsed correctly, %v", err)
	}

	config.DatabaseConfig.ConnectIdleDuration, err = time.ParseDuration(config.DatabaseConfig.ConnectIdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("time duration could not be parsed correctly, %v", err)
	}

	config.DatabaseConfig.QueryDuration, err = time.ParseDuration(config.DatabaseConfig.QueryTimeout)
	if err != nil {
		return nil, fmt.Errorf("time duration could not be parsed correctly, %v", err)
	}

	initialized = true
	return &config, nil
}

func SaveConfig() error {
	writer, err := os.OpenFile("config.json", os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	return nil
}

// GetConfig 获取配置，如果未初始化则先读取配置
func GetConfig() (*Config, error) {
	if initialized {
		return &config, nil
	}
	return ReadConfig()
}

func (dbt DatabaseType) GetConnection() gorm.Dialector {
	switch dbt {
	case MySQL:
		return mySQLConnection()
	case PostgreSQL:
		return postgreSQLConnection()
	case SQLite:
		return sqliteConnection()
	default:
		return nil
	}
}

func mySQLConnection() gorm.Dialector {
	encodedUser := url.QueryEscape(config.DatabaseConfig.Username)
	encodedPass := url.QueryEscape(config.DatabaseConfig.Password)
	var enableSSL string
	if config.DatabaseConfig.EnableSSL {
		enableSSL = "true"
	} else {
		enableSSL = "false"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&tls=%s",
		encodedUser,
		encodedPass,
		config.DatabaseConfig.Host,
		config.DatabaseConfig.Port,
		config.DatabaseConfig.Database,
		enableSSL,
	)
	return mysql.Open(dsn)
}

func postgreSQLConnection() gorm.Dialector {
	encodedUser := url.QueryEscape(config.DatabaseConfig.Username)
	encodedPass := url.QueryEscape(config.DatabaseConfig.Password)
	var enableSSL string
	if config.DatabaseConfig.EnableSSL {
		enableSSL = "enable"
	} else {
		enableSSL = "disable"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		config.DatabaseConfig.Host,
		encodedUser,
		encodedPass,
		config.DatabaseConfig.Database,
		config.DatabaseConfig.Port,
		enableSSL,
	)
	return postgres.Open(dsn)
}

func sqliteConnection() gorm.Dialector {
	return sqlite.Open(config.DatabaseConfig.Database)
}
