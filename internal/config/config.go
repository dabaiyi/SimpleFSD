package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/half-nothing/simple-fsd/internal/utils"
	"github.com/thanhpk/randstr"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"
)

const (
	AirportDataFileUrl              = "https://raw.githubusercontent.com/Flyleague-Collection/SimpleFSD/refs/heads/main/data/airport.json"
	EmailVerifyTemplateFileUrl      = "https://raw.githubusercontent.com/Flyleague-Collection/SimpleFSD/refs/heads/main/template/email_verify.template"
	ATCRatingChangeTemplateFileUrl  = "https://raw.githubusercontent.com/Flyleague-Collection/SimpleFSD/refs/heads/main/template/atc_rating_change.template"
	PermissionChangeTemplateFileUrl = "https://raw.githubusercontent.com/Flyleague-Collection/SimpleFSD/refs/heads/main/template/permission_change.template"
	KickedFromServerTemplateFileUrl = "https://raw.githubusercontent.com/Flyleague-Collection/SimpleFSD/refs/heads/main/template/kicked_from_server.template"
)

type Item interface {
	CheckValid() (bool, error)
}

func checkPort(port uint) (bool, error) {
	if port <= 0 {
		return false, errors.New("port must be greater than zero")
	}
	if port > 65535 {
		return false, errors.New("port must be less than 65535")
	}
	if port < 1024 {
		WarnF("The %d port may have a special usage, use it with caution")
	}
	return true, nil
}

type AirportData struct {
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Alt          float64 `json:"alt"`
	AirportRange int     `json:"airport_range"`
}

func (config *AirportData) CheckValid() (bool, error) {
	if config.AirportRange <= 0 {
		return false, errors.New("airport_range must be greater than zero")
	}
	return true, nil
}

type OtherConfig struct {
	SimulatorServer bool `json:"simulator_server"`
	BcryptCost      int  `json:"bcrypt_cost"`
}

func defaultOtherConfig() *OtherConfig {
	return &OtherConfig{
		SimulatorServer: false,
		BcryptCost:      12,
	}
}

func (config *OtherConfig) CheckValid() (bool, error) {
	if config.BcryptCost < bcrypt.MinCost || config.BcryptCost > bcrypt.MaxCost {
		return false, errors.New("bcrypt_cost out of range, must between 4 and 31")
	}
	return true, nil
}

type FSDServerConfig struct {
	FSDName              string                  `json:"fsd_name"` // 应用名称
	Host                 string                  `json:"host"`
	Port                 uint                    `json:"port"`
	Address              string                  `json:"-"`
	AirportDataFile      string                  `json:"airport_data_file"`
	AirportData          map[string]*AirportData `json:"-"`
	PosUpdatePoints      int                     `json:"pos_update_points"`
	HeartbeatInterval    string                  `json:"heartbeat_interval"`
	HeartbeatDuration    time.Duration           `json:"-"`
	SessionCleanTime     string                  `json:"session_clean_time"`    // 会话保留时间
	SessionCleanDuration time.Duration           `json:"-"`                     // 内部使用字段
	MaxWorkers           int                     `json:"max_workers"`           // 并发线程数
	MaxBroadcastWorkers  int                     `json:"max_broadcast_workers"` // 广播并发线程数
	Motd                 []string                `json:"motd"`
}

func defaultFSDServerConfig() *FSDServerConfig {
	return &FSDServerConfig{
		FSDName:             "Simple-Fsd",
		Host:                "0.0.0.0",
		Port:                6809,
		AirportDataFile:     "data/airport.json",
		PosUpdatePoints:     1,
		HeartbeatInterval:   "60s",
		SessionCleanTime:    "40s",
		MaxWorkers:          128,
		MaxBroadcastWorkers: 128,
		Motd:                make([]string, 0),
	}
}

func (config *FSDServerConfig) CheckValid() (bool, error) {
	if config.MaxBroadcastWorkers > runtime.NumCPU()*50 {
		config.MaxBroadcastWorkers = runtime.NumCPU() * 50
	}

	if pass, err := checkPort(config.Port); !pass {
		return pass, err
	}

	config.Address = fmt.Sprintf("%s:%d", config.Host, config.Port)

	if config.PosUpdatePoints < 0 {
		return false, fmt.Errorf("invalid json field pos_update_points, pos_update_points must larger than 0")
	}

	if bytes, err := cachedContent(config.AirportDataFile, AirportDataFileUrl); err != nil {
		WarnF("fail to load airport data, airport check disable, %v", err)
		config.AirportData = nil
	} else if err := json.Unmarshal(bytes, &config.AirportData); err != nil {
		return false, fmt.Errorf("invalid json file %s, %v", config.AirportDataFile, err)
	} else {
		InfoF("Airport data loaded, found %d airports", len(config.AirportData))
	}

	if duration, err := time.ParseDuration(config.SessionCleanTime); err != nil {
		return false, fmt.Errorf("invalid json field session_clean_time, duration parse error, %v", err)
	} else {
		config.SessionCleanDuration = duration
	}

	if duration, err := time.ParseDuration(config.HeartbeatInterval); err != nil {
		return false, fmt.Errorf("invalid json field heartbead_interval, duration parse error, %v", err)
	} else {
		config.HeartbeatDuration = duration
	}

	return true, nil
}

type JWTConfig struct {
	Secret          string        `json:"secret"`
	ExpiresTime     string        `json:"expires_time"`
	ExpiresDuration time.Duration `json:"-"`
	RefreshTime     string        `json:"refresh_time"`
	RefreshDuration time.Duration `json:"-"`
}

func defaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret:      randstr.String(64),
		ExpiresTime: "15m",
		RefreshTime: "24h",
	}
}

func (config *JWTConfig) CheckValid() (bool, error) {
	if duration, err := time.ParseDuration(config.ExpiresTime); err != nil {
		return false, fmt.Errorf("invalid json field http_server.email.jwt_expires_time, %v", err)
	} else {
		config.ExpiresDuration = duration
	}

	if duration, err := time.ParseDuration(config.RefreshTime); err != nil {
		return false, fmt.Errorf("invalid json field http_server.email.jwt_refresh_time, %v", err)
	} else {
		config.RefreshDuration = duration
	}

	if config.Secret == "" {
		config.Secret = randstr.String(64)
		DebugF("Generate random JWT Secret: %s", config.Secret)
	}
	return true, nil
}

type SSLConfig struct {
	Enable          bool   `json:"enable"`
	EnableHSTS      bool   `json:"enable_hsts"`
	HstsExpiredTime int    `json:"hsts_expired_time"`
	IncludeDomain   bool   `json:"include_domain"`
	CertFile        string `json:"cert_file"`
	KeyFile         string `json:"key_file"`
}

func defaultSSLConfig() *SSLConfig {
	return &SSLConfig{
		Enable:          false,
		EnableHSTS:      false,
		HstsExpiredTime: 5184000,
		IncludeDomain:   false,
		CertFile:        "",
		KeyFile:         "",
	}
}

func (config *SSLConfig) CheckValid() (bool, error) {
	if config.Enable {
		if config.CertFile == "" || config.KeyFile == "" {
			WarnF("HTTPS server requires both cert and key files. Cert: %s, Key: %s. Falling back to HTTP", config.CertFile, config.KeyFile)
			config.Enable = false
		}
	}
	if !config.Enable && config.EnableHSTS {
		Warn("You can not enable HSTS when ssl is not enable!")
		config.EnableHSTS = false
		config.HstsExpiredTime = 0
		config.IncludeDomain = false
	}
	return true, nil
}

type EmailTemplateConfig struct {
	EmailVerifyTemplateFile      string             `json:"email_verify_template_file"`
	EmailVerifyTemplate          *template.Template `json:"-"`
	ATCRatingChangeTemplateFile  string             `json:"atc_rating_change_template_file"`
	ATCRatingChangeTemplate      *template.Template `json:"-"`
	EnableRatingChangeEmail      bool               `json:"enable_rating_change_email"`
	PermissionChangeTemplateFile string             `json:"permission_change_template_file"`
	PermissionChangeTemplate     *template.Template `json:"-"`
	EnablePermissionChangeEmail  bool               `json:"enable_permission_change_email"`
	KickedFromServerTemplateFile string             `json:"kicked_from_server_template_file"`
	KickedFromServerTemplate     *template.Template `json:"-"`
	EnableKickedFromServerEmail  bool               `json:"enable_kicked_from_server_email"`
}

func defaultEmailTemplateConfig() *EmailTemplateConfig {
	return &EmailTemplateConfig{
		EmailVerifyTemplateFile:      "template/email_verify.template",
		ATCRatingChangeTemplateFile:  "template/atc_rating_change.template",
		EnableRatingChangeEmail:      true,
		PermissionChangeTemplateFile: "template/permission_change.template",
		EnablePermissionChangeEmail:  true,
		KickedFromServerTemplateFile: "template/kicked_from_server.template",
		EnableKickedFromServerEmail:  true,
	}
}

func (config *EmailTemplateConfig) CheckValid() (bool, error) {
	if bytes, err := cachedContent(config.EmailVerifyTemplateFile, EmailVerifyTemplateFileUrl); err != nil {
		return false, fmt.Errorf("fail to load email_verify_template_file, %v", err)
	} else if parse, err := template.New("email_verify").Parse(string(bytes)); err != nil {
		return false, fmt.Errorf("fail to parse email_verify_template, %v", err)
	} else {
		config.EmailVerifyTemplate = parse
	}

	if config.EnableRatingChangeEmail {
		if bytes, err := cachedContent(config.ATCRatingChangeTemplateFile, ATCRatingChangeTemplateFileUrl); err != nil {
			return false, fmt.Errorf("fail to load atc_rating_change_template_file, %v", err)
		} else if parse, err := template.New("atc_rating_change").Parse(string(bytes)); err != nil {
			return false, fmt.Errorf("fail to parse atc_rating_change_template, %v", err)
		} else {
			config.ATCRatingChangeTemplate = parse
		}
	}

	if config.EnablePermissionChangeEmail {
		if bytes, err := cachedContent(config.PermissionChangeTemplateFile, PermissionChangeTemplateFileUrl); err != nil {
			return false, fmt.Errorf("fail to load permission_change_template_file, %v", err)
		} else if parse, err := template.New("permission_change").Parse(string(bytes)); err != nil {
			return false, fmt.Errorf("fail to parse permission_change_template, %v", err)
		} else {
			config.PermissionChangeTemplate = parse
		}
	}

	if config.EnableKickedFromServerEmail {
		if bytes, err := cachedContent(config.KickedFromServerTemplateFile, KickedFromServerTemplateFileUrl); err != nil {
			return false, fmt.Errorf("fail to load permission_change_template_file, %v", err)
		} else if parse, err := template.New("kicked_from_server").Parse(string(bytes)); err != nil {
			return false, fmt.Errorf("fail to parse permission_change_template, %v", err)
		} else {
			config.KickedFromServerTemplate = parse
		}
	}

	return true, nil
}

type EmailConfig struct {
	Host                  string               `json:"host"`
	Port                  int                  `json:"port"`
	EmailServer           *gomail.Dialer       `json:"-"`
	Username              string               `json:"username"`
	Password              string               `json:"password"`
	VerifyExpiredTime     string               `json:"verify_expired_time"`
	VerifyExpiredDuration time.Duration        `json:"-"`
	SendInterval          string               `json:"send_interval"`
	SendDuration          time.Duration        `json:"-"`
	Template              *EmailTemplateConfig `json:"template"`
}

func defaultEmailConfig() *EmailConfig {
	return &EmailConfig{
		Host:              "smtp.qq.com",
		Port:              465,
		Username:          "example@qq.com",
		Password:          "123456",
		VerifyExpiredTime: "5m",
		SendInterval:      "1m",
		Template:          defaultEmailTemplateConfig(),
	}
}

func (config *EmailConfig) CheckValid() (bool, error) {
	if duration, err := time.ParseDuration(config.VerifyExpiredTime); err != nil {
		return false, fmt.Errorf("invalid json field http_server.email.verify_expired_time, %v", err)
	} else {
		config.VerifyExpiredDuration = duration
	}

	if duration, err := time.ParseDuration(config.SendInterval); err != nil {
		return false, fmt.Errorf("invalid json field http_server.email.send_interval, %v", err)
	} else {
		config.SendDuration = duration
	}

	if pass, err := config.Template.CheckValid(); !pass {
		return false, err
	}

	config.EmailServer = gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	return true, nil
}

type HttpServerLimit struct {
	RateLimit         int           `json:"rate_limit"`
	RateLimitWindow   string        `json:"rate_limit_window"`
	RateLimitDuration time.Duration `json:"-"`
	UsernameLengthMin int           `json:"username_length_min"`
	UsernameLengthMax int           `json:"username_length_max"`
	EmailLengthMin    int           `json:"email_length_min"`
	EmailLengthMax    int           `json:"email_length_max"`
	PasswordLengthMin int           `json:"password_length_min"`
	PasswordLengthMax int           `json:"password_length_max"`
	CidMin            int           `json:"cid_min"`
	CidMax            int           `json:"cid_max"`
}

func defaultHttpServerLimit() *HttpServerLimit {
	return &HttpServerLimit{
		RateLimit:         15,
		RateLimitWindow:   "1m",
		UsernameLengthMin: 4,
		UsernameLengthMax: 16,
		EmailLengthMin:    4,
		EmailLengthMax:    64,
		PasswordLengthMin: 6,
		PasswordLengthMax: 64,
		CidMin:            1,
		CidMax:            9999,
	}
}

func (config *HttpServerLimit) CheckValid() (bool, error) {
	if duration, err := time.ParseDuration(config.RateLimitWindow); err != nil {
		return false, fmt.Errorf("invalid json field http_server.rate_limit_window, %v", err)
	} else {
		config.RateLimitDuration = duration
	}

	if config.UsernameLengthMin <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.username_length_min, value must larger than 0")
	}
	if config.UsernameLengthMin > 64 {
		return false, fmt.Errorf("invalid json field http_server.limits.username_length_min, value must less than 64")
	}
	if config.UsernameLengthMax <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.username_length_max, value must larger than 0")
	}
	if config.UsernameLengthMax > 64 {
		return false, fmt.Errorf("invalid json field http_server.limits.username_length_max, value must less than 64")
	}
	if config.UsernameLengthMin >= config.UsernameLengthMax {
		return false, fmt.Errorf("invalid json field http_server.limits.username_length_min, value must less than http_server.limits.username_length_max")
	}

	if config.EmailLengthMin <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.email_length_min, value must larger than 0")
	}
	if config.EmailLengthMin > 128 {
		return false, fmt.Errorf("invalid json field http_server.limits.email_length_min, value must less than 128")
	}
	if config.EmailLengthMax <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.email_length_max, value must larger than 0")
	}
	if config.EmailLengthMax > 128 {
		return false, fmt.Errorf("invalid json field http_server.limits.email_length_max, value must less than 128")
	}
	if config.EmailLengthMin >= config.EmailLengthMax {
		return false, fmt.Errorf("invalid json field http_server.limits.email_length_min, value must less than http_server.limits.email_length_max")
	}

	if config.PasswordLengthMin <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.password_length_min, value must larger than 0")
	}
	if config.PasswordLengthMin > 128 {
		return false, fmt.Errorf("invalid json field http_server.limits.password_length_min, value must less than 128")
	}
	if config.PasswordLengthMax <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.password_length_max, value must larger than 0")
	}
	if config.PasswordLengthMax > 128 {
		return false, fmt.Errorf("invalid json field http_server.limits.password_length_max, value must less than 128")
	}
	if config.PasswordLengthMin >= config.PasswordLengthMax {
		return false, fmt.Errorf("invalid json field http_server.limits.password_length_min, value must less than http_server.limits.password_length_max")
	}

	if config.CidMin <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.cid_min, value must larger than 0")
	}
	if config.CidMax <= 0 {
		return false, fmt.Errorf("invalid json field http_server.limits.cid_max, value must larger than 0")
	}
	if config.CidMin >= config.CidMax {
		return false, fmt.Errorf("invalid json field http_server.limits.cid_min, value must less than http_server.limits.cid_max")
	}

	return true, nil
}

type HttpServerStoreFileLimit struct {
	MaxFileSize    int64    `json:"max_file_size"`
	AllowedFileExt []string `json:"allowed_file_ext"`
	StorePrefix    string   `json:"store_prefix"`
	StoreInServer  bool     `json:"store_in_server"`
	RootPath       string   `json:"-"`
}

func (config *HttpServerStoreFileLimit) CheckValid() (bool, error) {
	if config.MaxFileSize < 0 {
		return false, fmt.Errorf("invalid json field http_server.store.max_file_size, cannot be negative")
	}
	return true, nil
}

type HttpServerStoreFileLimits struct {
	ImageLimit *HttpServerStoreFileLimit `json:"image_limit"`
}

func defaultHttpServerStoreFileLimits() *HttpServerStoreFileLimits {
	return &HttpServerStoreFileLimits{
		ImageLimit: &HttpServerStoreFileLimit{
			MaxFileSize:    5 * 1024 * 1024,
			AllowedFileExt: []string{".jpg", ".png", ".bmp", ".jpeg"},
			StorePrefix:    "images",
			StoreInServer:  true,
		},
	}
}

func (config *HttpServerStoreFileLimits) CheckValid() (bool, error) {
	if pass, err := config.ImageLimit.CheckValid(); !pass {
		return false, err
	}
	return true, nil
}

func (config *HttpServerStoreFileLimits) CheckLocalStore(localStore bool) (bool, error) {
	if !localStore {
		return true, nil
	}
	if !config.ImageLimit.StoreInServer {
		return false, fmt.Errorf("when you use local store, store_in_server must be true")
	}
	return true, nil
}

func (config *HttpServerStoreFileLimits) CreateDir(root string) (bool, error) {
	config.ImageLimit.RootPath = root
	if config.ImageLimit.StoreInServer {
		imagePath := filepath.Join(root, config.ImageLimit.StorePrefix)
		if err := os.MkdirAll(imagePath, 0644); err != nil {
			return false, err
		}
	}
	return true, nil
}

type HttpServerStore struct {
	StoreType       int                        `json:"store_type"`        // 文件存储类型, 0: 本地存储, 1: 阿里云OSS存储, 2: 腾讯云对象存储
	Region          string                     `json:"region"`            // 云存储地域
	Bucket          string                     `json:"bucket"`            // 云存储桶名
	AccessId        string                     `json:"access_id"`         // 访问id
	AccessKey       string                     `json:"access_key"`        // 访问秘钥
	CdnDomain       string                     `json:"cdn_domain"`        // 自定义加速域名
	UseInternalUrl  bool                       `json:"use_internal_url"`  // 上传使用内部域名
	LocalStorePath  string                     `json:"local_store_path"`  // 本地存储路径
	RemoteStorePath string                     `json:"remote_store_path"` // 远程存储路径
	FileLimit       *HttpServerStoreFileLimits `json:"file_limit"`
}

func defaultHttpServerStore() *HttpServerStore {
	return &HttpServerStore{
		StoreType:       0,
		Region:          "",
		Bucket:          "",
		AccessId:        "",
		AccessKey:       "",
		CdnDomain:       "",
		UseInternalUrl:  false,
		LocalStorePath:  "uploads",
		RemoteStorePath: "",
		FileLimit:       defaultHttpServerStoreFileLimits(),
	}
}

func (config *HttpServerStore) CheckValid() (bool, error) {
	if pass, err := config.FileLimit.CheckValid(); !pass {
		return false, err
	}
	if config.LocalStorePath == "" {
		return false, fmt.Errorf("invalid json field http_server.store.local_store_path, path cannot be empty")
	}
	if err := os.MkdirAll(filepath.Clean(config.LocalStorePath), 0644); err != nil {
		return false, fmt.Errorf("error while creating local store path(%s), %v", config.LocalStorePath, err)
	}
	if pass, err := config.FileLimit.CreateDir(config.LocalStorePath); !pass {
		return false, err
	}
	switch config.StoreType {
	case 0:
		if pass, err := config.FileLimit.CheckLocalStore(true); !pass {
			return false, err
		}
		// 本地存储
		// 不用任何额外操作, 仅占位使用
	case 1, 2:
		if pass, err := config.FileLimit.CheckLocalStore(false); !pass {
			return false, err
		}
		// 阿里云OSS存储或者腾讯云对象存储
		if config.Region == "" {
			return false, fmt.Errorf("invalid json field http_server.store.region, region cannot be empty")
		}
		if config.Bucket == "" {
			return false, fmt.Errorf("invalid json field http_server.store.bucket, bucket cannot be empty")
		}
		if config.AccessId == "" {
			return false, fmt.Errorf("invalid json field http_server.store.access_id, access_id cannot be empty")
		}
		if config.AccessKey == "" {
			return false, fmt.Errorf("invalid json field http_server.store.access_key, access_key cannot be empty")
		}
	default:
		return false, fmt.Errorf("invalid json field http_server.store_type %d, only support 0, 1, 2", config.StoreType)
	}
	return true, nil
}

type HttpServerConfig struct {
	Enabled          bool             `json:"enabled"`
	WhazzupUrlHeader string           `json:"whazzup_url_header"`
	Host             string           `json:"host"`
	Port             uint             `json:"port"`
	Address          string           `json:"-"`
	MaxWorkers       int              `json:"max_workers"` // 并发线程数
	CacheTime        string           `json:"whazzup_cache_time"`
	CacheDuration    time.Duration    `json:"-"`
	ProxyType        int              `json:"proxy_type"`
	BodyLimit        string           `json:"body_limit"`
	Store            *HttpServerStore `json:"store"`
	Limits           *HttpServerLimit `json:"limits"`
	Email            *EmailConfig     `json:"email"`
	JWT              *JWTConfig       `json:"jwt"`
	SSL              *SSLConfig       `json:"ssl"`
}

func defaultHttpServerConfig() *HttpServerConfig {
	return &HttpServerConfig{
		Enabled:          false,
		Host:             "0.0.0.0",
		Port:             6810,
		MaxWorkers:       128,
		CacheTime:        "15s",
		WhazzupUrlHeader: "http://127.0.0.1:6810",
		ProxyType:        0,
		BodyLimit:        "10MB",
		Store:            defaultHttpServerStore(),
		Limits:           defaultHttpServerLimit(),
		Email:            defaultEmailConfig(),
		JWT:              defaultJWTConfig(),
		SSL:              defaultSSLConfig(),
	}
}

func (config *HttpServerConfig) CheckValid() (bool, error) {
	if config.Enabled {
		if pass, err := checkPort(config.Port); !pass {
			return pass, err
		}

		config.Address = fmt.Sprintf("%s:%d", config.Host, config.Port)

		if config.BodyLimit == "" {
			WarnF("body_limit is empty, where the length of the request body is not restricted. This is a very dangerous behavior")
		}

		if duration, err := time.ParseDuration(config.CacheTime); err != nil {
			return false, fmt.Errorf("invalid json field http_server.email.cache_time, %v", err)
		} else {
			config.CacheDuration = duration
		}

		config.WhazzupUrlHeader = strings.TrimRight(config.WhazzupUrlHeader, "/")

		if pass, err := config.SSL.CheckValid(); !pass {
			return pass, err
		}
		if pass, err := config.Limits.CheckValid(); !pass {
			return pass, err
		}
		if pass, err := config.Email.CheckValid(); !pass {
			return pass, err
		}
		if pass, err := config.JWT.CheckValid(); !pass {
			return pass, err
		}
		if pass, err := config.SSL.CheckValid(); !pass {
			return pass, err
		}
		if pass, err := config.Store.CheckValid(); !pass {
			return pass, err
		}
	}
	return true, nil
}

type GRPCServerConfig struct {
	Enabled       bool          `json:"enabled"`
	Host          string        `json:"host"`
	Port          uint          `json:"port"`
	Address       string        `json:"-"`
	CacheTime     string        `json:"whazzup_cache_time"`
	CacheDuration time.Duration `json:"-"`
}

func defaultGRPCServerConfig() *GRPCServerConfig {
	return &GRPCServerConfig{
		Enabled:   false,
		Host:      "0.0.0.0",
		Port:      6811,
		CacheTime: "15s",
	}
}

func (config *GRPCServerConfig) CheckValid() (bool, error) {
	if config.Enabled {
		if pass, err := checkPort(config.Port); !pass {
			return pass, err
		}
		config.Address = fmt.Sprintf("%s:%d", config.Host, config.Port)

		if duration, err := time.ParseDuration(config.CacheTime); err != nil {
			return false, fmt.Errorf("invalid json field grpc_server.cache_time, %v", err)
		} else {
			config.CacheDuration = duration
		}
	}
	return true, nil
}

type ServerConfig struct {
	General    *OtherConfig      `json:"general"`
	FSDServer  *FSDServerConfig  `json:"fsd_server"`
	HttpServer *HttpServerConfig `json:"http_server"`
	GRPCServer *GRPCServerConfig `json:"grpc_server"`
}

func defaultServerConfig() *ServerConfig {
	return &ServerConfig{
		General:    defaultOtherConfig(),
		FSDServer:  defaultFSDServerConfig(),
		HttpServer: defaultHttpServerConfig(),
		GRPCServer: defaultGRPCServerConfig(),
	}
}

func (config *ServerConfig) CheckValid() (bool, error) {
	if pass, err := config.General.CheckValid(); !pass {
		return pass, err
	}
	if pass, err := config.FSDServer.CheckValid(); !pass {
		return pass, err
	}
	if pass, err := config.HttpServer.CheckValid(); !pass {
		return pass, err
	}
	if pass, err := config.GRPCServer.CheckValid(); !pass {
		return pass, err
	}
	return true, nil
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

func defaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
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
	}
}

func (config *DatabaseConfig) CheckValid() (bool, error) {
	config.DBType = DatabaseType(config.Type)
	if !slices.Contains(allowedDatabaseType, config.DBType) {
		return false, fmt.Errorf("database type %s is not allowed, support database is %v, please check the configuration file", config.DBType, allowedDatabaseType)
	}

	if duration, err := time.ParseDuration(config.ConnectIdleTimeout); err != nil {
		return false, fmt.Errorf("invalid json field connect_idel_timeout, %v", err)
	} else {
		config.ConnectIdleDuration = duration
	}

	if duration, err := time.ParseDuration(config.QueryTimeout); err != nil {
		return false, fmt.Errorf("invalid json field query_timeout, %v", err)
	} else {
		config.QueryDuration = duration
	}
	return true, nil
}

type Config struct {
	DebugMode     bool            `json:"debug_mode"` // 是否启用调试模式
	ConfigVersion string          `json:"config_version"`
	Server        *ServerConfig   `json:"server"`
	Database      *DatabaseConfig `json:"database"`
	Rating        map[string]int  `json:"rating"`
}

func defaultConfig() *Config {
	return &Config{
		DebugMode:     false,
		ConfigVersion: confVersion.String(),
		Server:        defaultServerConfig(),
		Database:      defaultDatabaseConfig(),
		Rating:        make(map[string]int),
	}
}

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite3"
)

var (
	allowedDatabaseType = []DatabaseType{MySQL, PostgreSQL, SQLite}
	config              = utils.NewCachedValue[Config](0, func() *Config {
		if config, err := readConfig(); err != nil {
			FatalF("Error occurred while reading config %v", err)
			panic(err)
		} else {
			return config
		}
	})
)

// readConfig 从配置文件读取配置
func readConfig() (*Config, error) {
	config := defaultConfig()

	// 读取配置文件
	if bytes, err := os.ReadFile("config.json"); err != nil {
		// 如果配置文件不存在，创建默认配置
		if err := config.SaveConfig(); err != nil {
			return nil, err
		}
		return nil, errors.New("the configuration file does not exist and has been created. Please try again after editing the configuration file")
	} else if err := json.Unmarshal(bytes, config); err != nil {
		// 解析JSON配置
		return nil, fmt.Errorf("the configuration file does not contain valid JSON, %v", err)
	} else if pass, err := config.CheckValid(); !pass {
		return nil, err
	}
	return config, nil
}

func (c *Config) CheckValid() (bool, error) {
	if version, err := newVersion(c.ConfigVersion); err != nil {
		return false, err
	} else if result := confVersion.checkVersion(version); result != AllMatch {
		return false, fmt.Errorf("config version mismatch, expected %s, got %s", confVersion.String(), version.String())
	}
	if pass, err := c.Database.CheckValid(); !pass {
		return pass, err
	}
	if pass, err := c.Server.CheckValid(); !pass {
		return pass, err
	}
	return true, nil
}

func (c *Config) SaveConfig() error {
	if writer, err := os.OpenFile("config.json", os.O_WRONLY|os.O_CREATE, 0655); err != nil {
		return err
	} else if data, err := json.MarshalIndent(c, "", "\t"); err != nil {
		return err
	} else if _, err = writer.Write(data); err != nil {
		return err
	} else if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	return config.GetValue()
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

	InfoF("%s not found, downloading from %s", filePath, url)

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

	InfoF("Connection established with %s", url)

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	InfoF("%s successfully downloaded, %d bytes", filePath, len(content))

	if err := createFileWithContent(filePath, content); err != nil {
		return nil, fmt.Errorf("file write error: %w", err)
	}

	return content, nil
}

func (config *DatabaseConfig) GetConnection() gorm.Dialector {
	switch config.DBType {
	case MySQL:
		return mySQLConnection(config)
	case PostgreSQL:
		return postgreSQLConnection(config)
	case SQLite:
		return sqliteConnection(config)
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
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
		enableSSL,
	)
	DebugF("Mysql Connection DSN %s", dsn)
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
		db.Username,
		db.Password,
		db.Database,
		db.Port,
		enableSSL,
	)
	DebugF("PostgreSQL Connection DSN %s", dsn)
	return postgres.Open(dsn)
}

func sqliteConnection(db *DatabaseConfig) gorm.Dialector {
	return sqlite.Open(db.Database)
}
