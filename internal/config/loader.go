package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"sretool-fullstack/internal/model"
)

// Load 加载配置文件
func Load(configPath string) (*model.Config, error) {
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 设置默认值
	setDefaults(v)

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 绑定环境变量
	bindEnvVars(v)

	var cfg model.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置配置默认值
func setDefaults(v *viper.Viper) {
	// 应用配置
	v.SetDefault("app.name", "sretool-fullstack")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.refresh_interval", 10*time.Second)
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.log_path", "./app.log")

	// 诊断配置
	v.SetDefault("diagnostic.enabled", false)
	v.SetDefault("diagnostic.show_top_n_list", 10)

	// 告警配置
	v.SetDefault("alert.enabled", false)
	v.SetDefault("alert.cpu_threshold", 80.0)
	v.SetDefault("alert.memory_threshold", 80.0)
	v.SetDefault("alert.disk_threshold", 85.0)
	v.SetDefault("alert.disk_await_threshold", 50.0)
	v.SetDefault("alert.disk_util_threshold", 80.0)
	v.SetDefault("alert.inodes_threshold", 80.0)
	v.SetDefault("alert.network_bandwidth_threshold", 80.0)
	v.SetDefault("alert.network_packet_loss_threshold", 1.0)
	v.SetDefault("alert.network_rtt_threshold", 100.0)
	v.SetDefault("alert.tcp_time_wait_threshold", 1000)
	v.SetDefault("alert.tcp_close_wait_threshold", 100)
	v.SetDefault("alert.total_tcp_threshold", 10000)

	// InfluxDB 配置
	v.SetDefault("influxdb.enabled", false)
	v.SetDefault("influxdb.url", "http://localhost:8086")
	v.SetDefault("influxdb.token", "")
	v.SetDefault("influxdb.org", "tismin")
	v.SetDefault("influxdb.bucket", "sretool-fullstack")
}

// bindEnvVars 绑定环境变量
func bindEnvVars(v *viper.Viper) {
	// 应用配置环境变量
	_ = v.BindEnv("app.name", "APP_NAME")
	_ = v.BindEnv("app.version", "APP_VERSION")
	_ = v.BindEnv("app.refresh_interval", "APP_REFRESH_INTERVAL")
	_ = v.BindEnv("app.log_level", "APP_LOG_LEVEL")
	_ = v.BindEnv("app.log_path", "APP_LOG_PATH")

	// 告警配置环境变量
	_ = v.BindEnv("alert.enabled", "ALERT_ENABLED")
	_ = v.BindEnv("alert.cpu_threshold", "ALERT_CPU_THRESHOLD")
	_ = v.BindEnv("alert.memory_threshold", "ALERT_MEMORY_THRESHOLD")
	_ = v.BindEnv("alert.disk_threshold", "ALERT_DISK_THRESHOLD")
	_ = v.BindEnv("alert.inodes_threshold", "ALERT_INODES_THRESHOLD")

	// SMTP 环境变量
	_ = v.BindEnv("email.host", "SMTP_HOST")
	_ = v.BindEnv("email.port", "SMTP_PORT")
	_ = v.BindEnv("email.username", "SMTP_USERNAME")
	_ = v.BindEnv("email.password", "SMTP_PASSWORD")
	_ = v.BindEnv("email.from", "SMTP_FROM")
	_ = v.BindEnv("email.to", "SMTP_TO")
}

// Reload 重新加载配置
func Reload(v *viper.Viper) (*model.Config, error) {
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to reload config: %w", err)
	}

	var cfg model.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Println("Configuration reloaded successfully")
	return &cfg, nil
}

// GetViper 获取 Viper 实例
func GetViper(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	bindEnvVars(v)

	return v, nil
}
