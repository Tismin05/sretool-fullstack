package model

import "time"

// Config 应用配置
type Config struct {
	App        Appconfig        `mapstructure:"app"`
	Diagnostic DiagnosticConfig `mapstructure:"diagnostic"`
	Alert      AlertConfig      `mapstructure:"alert"`
	Email      EmailConfig      `mapstructure:"email"`
	InfluxDB   InfluxDBConfig   `mapstructure:"influxdb"`
	Webhook    WebhookConfig    `mapstructure:"webhook"`
	DingTalk   DingTalkConfig   `mapstructure:"dingtalk"`
}

// Appconfig 应用配置
type Appconfig struct {
	Name            string        `mapstructure:"name"`
	Version         string        `mapstructure:"version"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	LogLevel        string        `mapstructure:"log_level"`
	LogPath         string        `mapstructure:"log_path"`
	Port            int           `mapstructure:"port"`
}

// DiagnosticConfig 诊断配置
type DiagnosticConfig struct {
	Enabled      bool `mapstructure:"enabled"`
	ShowTopNList int  `mapstructure:"show_top_n_list"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled                      bool    `mapstructure:"enabled"`
	CPUThreshold                 float64 `mapstructure:"cpu_threshold"`
	MemoryThreshold              float64 `mapstructure:"memory_threshold"`
	DiskThreshold                float64 `mapstructure:"disk_threshold"`
	DiskAwaitThreshold           float64 `mapstructure:"disk_await_threshold"`
	DiskUtilThreshold            float64 `mapstructure:"disk_util_threshold"`
	InodesThreshold              float64 `mapstructure:"inodes_threshold"`
	NetworkBandwidthThreshold    float64 `mapstructure:"network_bandwidth_threshold"`
	NetworkPacketLossThreshold   float64 `mapstructure:"network_packet_loss_threshold"`
	NetworkRTTThreshold          float64 `mapstructure:"network_rtt_threshold"`
	TCPTimeWaitThreshold         uint64  `mapstructure:"tcp_time_wait_threshold"`
	TCPCLOSEWaitThreshold        uint64  `mapstructure:"tcp_close_wait_threshold"`
	TotalTCPThreshold            uint64  `mapstructure:"total_tcp_threshold"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string   `mapstructure:"host"`
	Port     int      `mapstructure:"port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	From     string   `mapstructure:"from"`
	To       []string `mapstructure:"to"`
}

// InfluxDBConfig InfluxDB 配置
type InfluxDBConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
	Token   string `mapstructure:"token"`
	Org     string `mapstructure:"org"`
	Bucket  string `mapstructure:"bucket"`
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	Enabled    bool              `mapstructure:"enabled"`
	URL       string            `mapstructure:"url"`
	Method    string            `mapstructure:"method"`
	Headers   map[string]string `mapstructure:"headers"`
	Timeout   int               `mapstructure:"timeout"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	WebhookURL string `mapstructure:"webhook_url"`
	Secret    string `mapstructure:"secret"`
}
