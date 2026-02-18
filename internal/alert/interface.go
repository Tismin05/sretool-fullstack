package alert

import (
	"context"
	"time"

	"tisminSRETool/internal/model"
)

// Alert 告警信息结构体
type Alert struct {
	Level     AlertLevel    // 告警级别：info/warn/error
	Category  AlertCategory // 告警类别：cpu/memory/disk/network/inodes/tcp
	Metric    string        // 指标名称
	Message   string        // 告警消息
	Value     float64       // 当前值
	Threshold float64       // 阈值
	Unit      string        // 单位
	Timestamp time.Time     // 发生时间
	Host      string        // 主机名
}

// AlertLevel 告警级别
type AlertLevel string

const (
	LevelInfo  AlertLevel = "info"
	LevelWarn  AlertLevel = "warning"
	LevelError AlertLevel = "error"
)

// AlertCategory 告警类别
type AlertCategory string

const (
	CategoryCPU     AlertCategory = "cpu"
	CategoryMemory  AlertCategory = "memory"
	CategoryDisk    AlertCategory = "disk"
	CategoryNetwork AlertCategory = "network"
	CategoryInodes  AlertCategory = "inodes"
	CategoryTCP     AlertCategory = "tcp"
)

// AlertChecker 告警检查器接口
type AlertChecker interface {
	Check(ctx context.Context, m model.Metrics) ([]Alert, []error)
}

// AlertSender 告警发送器接口
type AlertSender interface {
	Send(ctx context.Context, alerts []Alert, cfg model.EmailConfig) error
}
