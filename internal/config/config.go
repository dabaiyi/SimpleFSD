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
	"sync"
	"time"
)

type OtherConfig struct {
	SimulatorServer bool `json:"simulator_server"`
	BcryptCost      int  `json:"bcrypt_cost"`
}

type FSDServerConfig struct {
	FSDName              string        `json:"fsd_name"` // 应用名称
	Host                 string        `json:"host"`
	Port                 uint64        `json:"port"`
	Address              string        `json:"-"`
	HeartbeatInterval    string        `json:"heartbeat_interval"`
	HeartbeatDuration    time.Duration `json:"-"`
	SessionCleanTime     string        `json:"session_clean_time"`    // 会话保留时间
	SessionCleanDuration time.Duration `json:"-"`                     // 内部使用字段
	MaxWorkers           int           `json:"max_workers"`           // 并发线程数
	MaxBroadcastWorkers  int           `json:"max_broadcast_workers"` // 广播并发线程数
	Motd                 []string      `json:"motd"`
}

type HttpServerConfig struct {
	Enabled       bool          `json:"enabled"`
	Host          string        `json:"host"`
	Port          uint64        `json:"port"`
	Address       string        `json:"-"`
	MaxWorkers    int           `json:"max_workers"` // 并发线程数
	CacheTime     string        `json:"cache_time"`
	CacheDuration time.Duration `json:"-"`
	EnableSSL     bool          `json:"enable_ssl"`
	CertFile      string        `json:"cert_file"`
	KeyFile       string        `json:"key_file"`
}

type GRPCServerConfig struct {
	Enabled       bool          `json:"enabled"`
	Host          string        `json:"host"`
	Port          uint64        `json:"port"`
	Address       string        `json:"-"`
	CacheTime     string        `json:"cache_time"`
	CacheDuration time.Duration `json:"-"`
}

type ServerConfig struct {
	General    OtherConfig      `json:"general"`
	FSDServer  FSDServerConfig  `json:"fsd_server"`
	HttpServer HttpServerConfig `json:"http_server"`
	GRPCServer GRPCServerConfig `json:"grpc_server"`
}

type DatabaseConfig struct {
	Type                 string        `json:"type"`
	DBType               DatabaseType  `json:"-"`
	Database             string        `json:"database"`
	Host                 string        `json:"host"`
	Port                 int           `json:"port"`
	Username             string        `json:"username"`
	Password             string        `json:"password"`
	EnableSSL            bool          `json:"enable_ssl"`
	ConnectIdleTimeout   string        `json:"connect_idle_timeout"` // 连接空闲超时时间
	ConnectIdleDuration  time.Duration `json:"-"`
	QueryTimeout         string        `json:"connect_timeout"` // 每次查询超时时间
	QueryDuration        time.Duration `json:"-"`
	ServerMaxConnections int           `json:"server_max_connections"` // 最大连接池大小
}

type Config struct {
	DebugMode     bool           `json:"debug_mode"` // 是否启用调试模式
	ConfigVersion string         `json:"config_version"`
	Server        ServerConfig   `json:"server"`
	Database      DatabaseConfig `json:"database"`
	Rating        map[string]int `json:"rating"`
}

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite3"
)

var (
	config              *Config
	configOnce          sync.Once // 新增同步控制
	configError         error
	allowedDatabaseType = []DatabaseType{MySQL, PostgreSQL, SQLite}
)

func newConfig() *Config {
	return &Config{
		DebugMode:     false,
		ConfigVersion: confVersion.String(),
		Server: ServerConfig{
			General: OtherConfig{
				SimulatorServer: false,
				BcryptCost:      12,
			},
			FSDServer: FSDServerConfig{
				FSDName:             "Simple-Fsd",
				Host:                "0.0.0.0",
				Port:                6809,
				HeartbeatInterval:   "60s",
				SessionCleanTime:    "40s",
				MaxWorkers:          128,
				MaxBroadcastWorkers: 128,
				Motd:                make([]string, 0),
			},
			HttpServer: HttpServerConfig{
				Enabled:    false,
				Host:       "0.0.0.0",
				Port:       6810,
				MaxWorkers: 128,
				CacheTime:  "15s",
				EnableSSL:  false,
				CertFile:   "",
				KeyFile:    "",
			},
			GRPCServer: GRPCServerConfig{
				Enabled:   false,
				Host:      "0.0.0.0",
				Port:      6811,
				CacheTime: "15s",
			},
		},
		Database: DatabaseConfig{
			Type:                 "sqlite3",
			Database:             "database.db",
			Host:                 "",
			Port:                 0,
			Username:             "",
			Password:             "",
			EnableSSL:            false,
			ConnectIdleTimeout:   "1h",
			QueryTimeout:         "5s",
			ServerMaxConnections: 32,
		},
		Rating: make(map[string]int),
	}
}

// readConfig 从配置文件读取配置
func readConfig() (*Config, error) {
	// 读取配置文件
	bytes, err := os.ReadFile("config.json")

	if err != nil {
		// 如果配置文件不存在，创建默认配置
		if err := newConfig().SaveConfig(); err != nil {
			return nil, err
		}
		return nil, errors.New("the configuration file does not exist and has been created. Please try again after editing the configuration file")
	}

	// 解析JSON配置
	err = json.Unmarshal(bytes, &config)

	if err != nil {
		return nil, fmt.Errorf("the configuration file does not contain valid JSON, %v", err)
	}

	err = config.handleConfigVersion()
	if err != nil {
		return nil, err
	}

	err = config.handleConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) handleConfigVersion() error {
	version, err := newVersion(c.ConfigVersion)
	if err != nil {
		return err
	}
	result := confVersion.checkVersion(version)
	if result != AllMatch {
		return fmt.Errorf("config version mismatch, expected %s, got %s", confVersion.String(), version.String())
	}
	return nil
}

func (c *Config) handleConfig() error {
	config.Database.DBType = DatabaseType(config.Database.Type)
	if !slices.Contains(allowedDatabaseType, config.Database.DBType) {
		return fmt.Errorf("database type %s is not allowed, support database is %v, please check the configuration file", config.Database.DBType, allowedDatabaseType)
	}

	if config.Server.FSDServer.MaxBroadcastWorkers > runtime.NumCPU()*50 {
		config.Server.FSDServer.MaxBroadcastWorkers = runtime.NumCPU() * 50
	}

	var err error
	config.Server.FSDServer.SessionCleanDuration, err = time.ParseDuration(config.Server.FSDServer.SessionCleanTime)
	if err != nil {
		return fmt.Errorf("invalid json field session_clean_time, %v", err)
	}

	config.Server.FSDServer.HeartbeatDuration, err = time.ParseDuration(config.Server.FSDServer.HeartbeatInterval)
	if err != nil {
		return fmt.Errorf("invalid json field heartbead_interval, %v", err)
	}

	config.Server.FSDServer.Address = fmt.Sprintf("%s:%d", config.Server.FSDServer.Host, config.Server.FSDServer.Port)

	config.Server.HttpServer.CacheDuration, err = time.ParseDuration(config.Server.HttpServer.CacheTime)
	if err != nil {
		return fmt.Errorf("invalid json field http_server.cache_time, %v", err)
	}

	config.Server.HttpServer.Address = fmt.Sprintf("%s:%d", config.Server.HttpServer.Host, config.Server.HttpServer.Port)

	config.Server.GRPCServer.CacheDuration, err = time.ParseDuration(config.Server.GRPCServer.CacheTime)
	if err != nil {
		return fmt.Errorf("invalid json field grpc_server.cache_time, %v", err)
	}

	config.Server.GRPCServer.Address = fmt.Sprintf("%s:%d", config.Server.GRPCServer.Host, config.Server.GRPCServer.Port)

	config.Database.ConnectIdleDuration, err = time.ParseDuration(config.Database.ConnectIdleTimeout)
	if err != nil {
		return fmt.Errorf("invalid json field connect_idel_timeout, %v", err)
	}

	config.Database.QueryDuration, err = time.ParseDuration(config.Database.QueryTimeout)
	if err != nil {
		return fmt.Errorf("invalid json field query_timeout, %v", err)
	}

	return nil
}

func (c *Config) SaveConfig() error {
	writer, err := os.OpenFile("config.json", os.O_WRONLY|os.O_CREATE, 0655)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "\t")
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
	configOnce.Do(func() {
		config, configError = readConfig()
	})
	return config, configError
}

func (db *DatabaseConfig) GetConnection() gorm.Dialector {
	switch db.DBType {
	case MySQL:
		return mySQLConnection(db)
	case PostgreSQL:
		return postgreSQLConnection(db)
	case SQLite:
		return sqliteConnection(db)
	default:
		return nil
	}
}

func mySQLConnection(db *DatabaseConfig) gorm.Dialector {
	var enableSSL string
	if db.EnableSSL {
		enableSSL = "true"
	} else {
		enableSSL = "false"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&tls=%s",
		url.QueryEscape(db.Username),
		url.QueryEscape(db.Password),
		db.Host,
		db.Port,
		url.QueryEscape(db.Database),
		enableSSL,
	)
	return mysql.Open(dsn)
}

func postgreSQLConnection(db *DatabaseConfig) gorm.Dialector {
	var enableSSL string
	if db.EnableSSL {
		enableSSL = "enable"
	} else {
		enableSSL = "disable"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		db.Host,
		url.QueryEscape(db.Username),
		url.QueryEscape(db.Password),
		url.QueryEscape(db.Database),
		db.Port,
		enableSSL,
	)
	return postgres.Open(dsn)
}

func sqliteConnection(db *DatabaseConfig) gorm.Dialector {
	return sqlite.Open(db.Database)
}
