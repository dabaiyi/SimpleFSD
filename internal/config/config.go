package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
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
	SimulatorServer      bool          `json:"simulator_server"`      // 是否为模拟机服务器
	ServerConfig         struct {
		Host              string        `json:"host"`
		Port              uint64        `json:"port"`
		EnableGRPC        bool          `json:"enable_grpc"`
		GRPCPort          uint64        `json:"grpc_port"`
		HeartbeatInterval string        `json:"heartbeat_interval"`
		HeartbeatDuration time.Duration `json:"-"`
		Motd              []string      `json:"motd"`
	} `json:"server_config"`
	DatabaseConfig struct {
		Host                 string        `json:"host"`
		Port                 int           `json:"port"`
		Username             string        `json:"username"`
		Password             string        `json:"password"`
		Database             string        `json:"database"`
		ConnectIdleTimeout   string        `json:"connect_idle_timeout"` // 连接空闲超时时间
		ConnectIdleDuration  time.Duration `json:"-"`
		QueryTimeout         string        `json:"connect_timeout"` // 每次查询超时时间
		QueryDuration        time.Duration `json:"-"`
		ServerMaxConnections int           `json:"server_max_connections"` // 最大连接池大小
	} `json:"database_config"`
	RatingConfig map[string]int `json:"rating_config"`
}

var (
	config      Config
	initialized = false
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

	initialized = true

	if config.MaxBroadcastWorkers > runtime.NumCPU()*50 {
		config.MaxBroadcastWorkers = runtime.NumCPU() * 50
	}

	config.SessionCleanDuration, err = time.ParseDuration(config.SessionCleanTime)
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
