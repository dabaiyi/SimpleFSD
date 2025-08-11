package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"time"
)

const (
	AirportDataFileUrl              = "https://raw.githubusercontent.com/Flyleague-Collection/fsd-server/refs/heads/main/data/airport.json"
	EmailVerifyTemplateFileUrl      = "https://raw.githubusercontent.com/Flyleague-Collection/fsd-server/refs/heads/main/template/email_verify.template"
	ATCRatingChangeTemplateFileUrl  = "https://raw.githubusercontent.com/Flyleague-Collection/fsd-server/refs/heads/main/template/atc_rating_change.template"
	PermissionChangeTemplateFileUrl = "https://raw.githubusercontent.com/Flyleague-Collection/fsd-server/refs/heads/main/template/permission_change.template"
)

type AirportData struct {
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Alt          float64 `json:"alt"`
	AirportRange int     `json:"airport_range"`
}

type OtherConfig struct {
	SimulatorServer bool `json:"simulator_server"`
	BcryptCost      int  `json:"bcrypt_cost"`
}

type FSDServerConfig struct {
	FSDName              string                 `json:"fsd_name"` // 应用名称
	Host                 string                 `json:"host"`
	Port                 uint64                 `json:"port"`
	Address              string                 `json:"-"`
	AirportDataFile      string                 `json:"airport_data_file"`
	AirportData          map[string]AirportData `json:"-"`
	SendWallopToADM      bool                   `json:"send_wallop_to_adm"`
	HeartbeatInterval    string                 `json:"heartbeat_interval"`
	HeartbeatDuration    time.Duration          `json:"-"`
	SessionCleanTime     string                 `json:"session_clean_time"`    // 会话保留时间
	SessionCleanDuration time.Duration          `json:"-"`                     // 内部使用字段
	MaxWorkers           int                    `json:"max_workers"`           // 并发线程数
	MaxBroadcastWorkers  int                    `json:"max_broadcast_workers"` // 广播并发线程数
	Motd                 []string               `json:"motd"`
}

type JWTConfig struct {
	Secret          string        `json:"secret"`
	ExpiresTime     string        `json:"expires_time"`
	ExpiresDuration time.Duration `json:"-"`
	RefreshTime     string        `json:"refresh_time"`
	RefreshDuration time.Duration `json:"-"`
}

type SSLConfig struct {
	Enable          bool   `json:"enable"`
	EnableHSTS      bool   `json:"enable_hsts"`
	HstsExpiredTime int    `json:"hsts_expired_time"`
	IncludeDomain   bool   `json:"include_domain"`
	CertFile        string `json:"cert_file"`
	KeyFile         string `json:"key_file"`
}

type EmailTemplateConfig struct {
	EmailVerifyTemplateFile      string             `json:"email_verify_template_file"`
	EmailVerifyTemplate          *template.Template `json:"-"`
	ATCRatingChangeTemplateFile  string             `json:"atc_rating_change_template_file"`
	ATCRatingChangeTemplate      *template.Template `json:"-"`
	PermissionChangeTemplateFile string             `json:"permission_change_template_file"`
	PermissionChangeTemplate     *template.Template `json:"-"`
}

type EmailConfig struct {
	Host                  string              `json:"host"`
	Port                  int                 `json:"port"`
	EmailServer           *gomail.Dialer      `json:"-"`
	Username              string              `json:"username"`
	Password              string              `json:"password"`
	VerifyExpiredTime     string              `json:"verify_expired_time"`
	VerifyExpiredDuration time.Duration       `json:"-"`
	SendInterval          string              `json:"send_interval"`
	SendDuration          time.Duration       `json:"-"`
	Template              EmailTemplateConfig `json:"template"`
}

type HttpServerConfig struct {
	Enabled           bool          `json:"enabled"`
	Host              string        `json:"host"`
	Port              uint64        `json:"port"`
	Address           string        `json:"-"`
	MaxWorkers        int           `json:"max_workers"` // 并发线程数
	CacheTime         string        `json:"cache_time"`
	CacheDuration     time.Duration `json:"-"`
	ProxyType         int           `json:"proxy_type"`
	RateLimit         int           `json:"rate_limit"`
	RateLimitWindow   string        `json:"rate_limit_window"`
	RateLimitDuration time.Duration `json:"-"`
	Email             EmailConfig   `json:"email"`
	JWT               JWTConfig     `json:"jwt"`
	SSL               SSLConfig     `json:"ssl"`
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
	configOnce          sync.Once
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
				AirportDataFile:     "data/airport.json",
				AirportData:         make(map[string]AirportData),
				SendWallopToADM:     true,
				HeartbeatInterval:   "60s",
				SessionCleanTime:    "40s",
				MaxWorkers:          128,
				MaxBroadcastWorkers: 128,
				Motd:                make([]string, 0),
			},
			HttpServer: HttpServerConfig{
				Enabled:         false,
				Host:            "0.0.0.0",
				Port:            6810,
				MaxWorkers:      128,
				CacheTime:       "15s",
				ProxyType:       0,
				RateLimit:       100,
				RateLimitWindow: "1m",
				Email: EmailConfig{
					Host:              "smtp.qq.com",
					Port:              465,
					Username:          "example@qq.com",
					Password:          "123456",
					VerifyExpiredTime: "5m",
					SendInterval:      "1m",
					Template: EmailTemplateConfig{
						EmailVerifyTemplateFile:      "template/email_verify.template",
						ATCRatingChangeTemplateFile:  "template/atc_rating_change.template",
						PermissionChangeTemplateFile: "template/permission_change.template",
					},
				},
				JWT: JWTConfig{
					Secret:      "123456",
					ExpiresTime: "1h",
					RefreshTime: "1h",
				},
				SSL: SSLConfig{
					Enable:   false,
					CertFile: "",
					KeyFile:  "",
				},
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

func createFileWithContent(filePath string, content []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, content, 0644)
}

func cachedContent(filePath, url string) ([]byte, error) {
	if content, err := os.ReadFile(filePath); err == nil {
		return content, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("file read error: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	if err := createFileWithContent(filePath, content); err != nil {
		return nil, fmt.Errorf("file write error: %w", err)
	}

	return content, nil
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

	config.Server.FSDServer.Address = fmt.Sprintf("%s:%d", config.Server.FSDServer.Host, config.Server.FSDServer.Port)

	if bytes, err := cachedContent(config.Server.FSDServer.AirportDataFile, AirportDataFileUrl); err != nil {
		log.Warnf("fail to load airport data, arrival airport check disable, %v", err)
		config.Server.FSDServer.AirportData = nil
	} else if err := json.Unmarshal(bytes, &config.Server.FSDServer.AirportData); err != nil {
		return fmt.Errorf("invalid json file %s, %v", config.Server.FSDServer.AirportDataFile, err)
	}

	config.Server.FSDServer.SessionCleanDuration, err = time.ParseDuration(config.Server.FSDServer.SessionCleanTime)
	if err != nil {
		return fmt.Errorf("invalid json field fsd_server.session_clean_time, %v", err)
	}

	config.Server.FSDServer.HeartbeatDuration, err = time.ParseDuration(config.Server.FSDServer.HeartbeatInterval)
	if err != nil {
		return fmt.Errorf("invalid json field fsd_server.heartbead_interval, %v", err)
	}

	if config.Server.HttpServer.Enabled {
		config.Server.HttpServer.Address = fmt.Sprintf("%s:%d", config.Server.HttpServer.Host, config.Server.HttpServer.Port)

		config.Server.HttpServer.CacheDuration, err = time.ParseDuration(config.Server.HttpServer.CacheTime)
		if err != nil {
			return fmt.Errorf("invalid json field http_server.email.cache_time, %v", err)
		}

		config.Server.HttpServer.RateLimitDuration, err = time.ParseDuration(config.Server.HttpServer.RateLimitWindow)
		if err != nil {
			return fmt.Errorf("invalid json field http_server.rate_limit_window, %v", err)
		}

		config.Server.HttpServer.Email.VerifyExpiredDuration, err = time.ParseDuration(config.Server.HttpServer.Email.VerifyExpiredTime)
		if err != nil {
			return fmt.Errorf("invalid json field http_server.email.verify_expired_time, %v", err)
		}

		config.Server.HttpServer.Email.SendDuration, err = time.ParseDuration(config.Server.HttpServer.Email.SendInterval)
		if err != nil {
			return fmt.Errorf("invalid json field http_server.email.send_interval, %v", err)
		}

		config.Server.HttpServer.JWT.ExpiresDuration, err = time.ParseDuration(config.Server.HttpServer.JWT.ExpiresTime)
		if err != nil {
			return fmt.Errorf("invalid json field http_server.email.jwt_expires_time, %v", err)
		}

		config.Server.HttpServer.JWT.RefreshDuration, err = time.ParseDuration(config.Server.HttpServer.JWT.RefreshTime)
		if err != nil {
			return fmt.Errorf("invalid json field http_server.email.jwt_refresh_time, %v", err)
		}

		if bytes, err := cachedContent(config.Server.HttpServer.Email.Template.EmailVerifyTemplateFile, EmailVerifyTemplateFileUrl); err != nil {
			return fmt.Errorf("fail to load http_server.email.email_verify_template_file, %v", err)
		} else if parse, err := template.New("email_verify").Parse(string(bytes)); err != nil {
			return fmt.Errorf("fail to parse email_verify_template, %v", err)
		} else {
			config.Server.HttpServer.Email.Template.EmailVerifyTemplate = parse
		}

		if bytes, err := cachedContent(config.Server.HttpServer.Email.Template.ATCRatingChangeTemplateFile, ATCRatingChangeTemplateFileUrl); err != nil {
			return fmt.Errorf("fail to load http_server.email.atc_rating_change_template_file, %v", err)
		} else if parse, err := template.New("atc_rating_change").Parse(string(bytes)); err != nil {
			return fmt.Errorf("fail to parse atc_rating_change_template, %v", err)
		} else {
			config.Server.HttpServer.Email.Template.ATCRatingChangeTemplate = parse
		}

		if bytes, err := cachedContent(config.Server.HttpServer.Email.Template.PermissionChangeTemplateFile, PermissionChangeTemplateFileUrl); err != nil {
			return fmt.Errorf("fail to load http_server.email.permission_change_template_file, %v", err)
		} else if parse, err := template.New("permission_change").Parse(string(bytes)); err != nil {
			return fmt.Errorf("fail to parse permission_change_template, %v", err)
		} else {
			config.Server.HttpServer.Email.Template.PermissionChangeTemplate = parse
		}

		config.Server.HttpServer.Email.EmailServer = gomail.NewDialer(config.Server.HttpServer.Email.Host, config.Server.HttpServer.Email.Port, config.Server.HttpServer.Email.Username, config.Server.HttpServer.Email.Password)
	}

	if config.Server.GRPCServer.Enabled {
		config.Server.GRPCServer.Address = fmt.Sprintf("%s:%d", config.Server.GRPCServer.Host, config.Server.GRPCServer.Port)

		config.Server.GRPCServer.CacheDuration, err = time.ParseDuration(config.Server.GRPCServer.CacheTime)
		if err != nil {
			return fmt.Errorf("invalid json field grpc_server.cache_time, %v", err)
		}
	}

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
