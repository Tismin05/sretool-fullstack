package alert

import (
	"context"
	"fmt"
	"sretool-fullstack/internal/model"
)

// RuleChecker 告警规则检查器
type RuleChecker struct {
	config model.AlertConfig
}

// NewRuleChecker 创建规则检查器
func NewRuleChecker(config model.AlertConfig) *RuleChecker {
	return &RuleChecker{config: config}
}

// Check 检查指标是否触发告警
func (r *RuleChecker) Check(ctx context.Context, m model.Metrics) ([]Alert, []error) {
	if !r.config.Enabled {
		return nil, nil
	}

	var alerts []Alert

	// CPU 检查
	if alerts = append(alerts, r.checkCPU(m)...); len(alerts) > 0 && alerts[len(alerts)-1].Level == "" {
		alerts = alerts[:len(alerts)-1]
	}

	// 内存检查
	alerts = append(alerts, r.checkMem(m)...)

	// 磁盘检查
	alerts = append(alerts, r.checkDisk(m)...)

	// Inodes 检查
	alerts = append(alerts, r.checkInodes(m)...)

	// 网络检查
	alerts = append(alerts, r.checkNet(m)...)

	return alerts, nil
}

func (r *RuleChecker) checkCPU(m model.Metrics) []Alert {
	var alerts []Alert
	if m.CPU.UsagePercent > r.config.CPUThreshold {
		alerts = append(alerts, Alert{
			Level:     LevelError,
			Category:  CategoryCPU,
			Metric:    "usage_percent",
			Message:   fmt.Sprintf("CPU usage %.1f%% exceeds threshold %.1f%%", m.CPU.UsagePercent, r.config.CPUThreshold),
			Value:     m.CPU.UsagePercent,
			Threshold: r.config.CPUThreshold,
			Unit:      "%",
			Host:      m.Host,
		})
	}
	return alerts
}

func (r *RuleChecker) checkMem(m model.Metrics) []Alert {
	var alerts []Alert
	if m.Mem.UsedPercent > r.config.MemoryThreshold {
		alerts = append(alerts, Alert{
			Level:     LevelError,
			Category:  CategoryMemory,
			Metric:    "used_percent",
			Message:   fmt.Sprintf("Memory usage %.1f%% exceeds threshold %.1f%%", m.Mem.UsedPercent, r.config.MemoryThreshold),
			Value:     m.Mem.UsedPercent,
			Threshold: r.config.MemoryThreshold,
			Unit:      "%",
			Host:      m.Host,
		})
	}
	return alerts
}

func (r *RuleChecker) checkDisk(m model.Metrics) []Alert {
	var alerts []Alert
	for _, disk := range m.Disk {
		if disk.UsedPercent > r.config.DiskThreshold {
			alerts = append(alerts, Alert{
				Level:     LevelWarn,
				Category:  CategoryDisk,
				Metric:    "used_percent",
				Message:   fmt.Sprintf("Disk %s usage %.1f%% exceeds threshold %.1f%%", disk.MountPoint, disk.UsedPercent, r.config.DiskThreshold),
				Value:     disk.UsedPercent,
				Threshold: r.config.DiskThreshold,
				Unit:      "%",
				Host:      m.Host,
			})
		}
		if disk.Await > r.config.DiskAwaitThreshold && r.config.DiskAwaitThreshold > 0 {
			alerts = append(alerts, Alert{
				Level:     LevelWarn,
				Category:  CategoryDisk,
				Metric:    "await",
				Message:   fmt.Sprintf("Disk %s await %.1fms exceeds threshold %.1fms", disk.MountPoint, disk.Await, r.config.DiskAwaitThreshold),
				Value:     disk.Await,
				Threshold: r.config.DiskAwaitThreshold,
				Unit:      "ms",
				Host:      m.Host,
			})
		}
		if disk.Util > r.config.DiskUtilThreshold {
			alerts = append(alerts, Alert{
				Level:     LevelWarn,
				Category:  CategoryDisk,
				Metric:    "util",
				Message:   fmt.Sprintf("Disk %s util %.1f%% exceeds threshold %.1f%%", disk.MountPoint, disk.Util, r.config.DiskUtilThreshold),
				Value:     disk.Util,
				Threshold: r.config.DiskUtilThreshold,
				Unit:      "%",
				Host:      m.Host,
			})
		}
	}
	return alerts
}

func (r *RuleChecker) checkInodes(m model.Metrics) []Alert {
	var alerts []Alert
	for _, disk := range m.Disk {
		if disk.InodesUsedPercent > r.config.InodesThreshold {
			alerts = append(alerts, Alert{
				Level:     LevelError,
				Category:  CategoryInodes,
				Metric:    "inodes_used_percent",
				Message:   fmt.Sprintf("Disk %s inodes %.1f%% exceeds threshold %.1f%%", disk.MountPoint, disk.InodesUsedPercent, r.config.InodesThreshold),
				Value:     disk.InodesUsedPercent,
				Threshold: r.config.InodesThreshold,
				Unit:      "%",
				Host:      m.Host,
			})
		}
	}
	return alerts
}

func (r *RuleChecker) checkNet(m model.Metrics) []Alert {
	var alerts []Alert
	for _, net := range m.Net {
		totalPackets := net.RxPackets + net.TxPackets
		if totalPackets > 0 {
			errorRate := float64(net.RxErrors+net.TxErrors) / float64(totalPackets) * 100
			dropRate := float64(net.RxDropped+net.TxDropped) / float64(totalPackets) * 100

			if errorRate > 0.01 || dropRate > 0.01 {
				alerts = append(alerts, Alert{
					Level:     LevelWarn,
					Category:  CategoryNetwork,
					Metric:    "error_rate",
					Message:   fmt.Sprintf("Network interface %s error rate %.4f%%, drop rate %.4f%%", net.Name, errorRate, dropRate),
					Value:     errorRate,
					Threshold: 0.01,
					Unit:      "%",
					Host:      m.Host,
				})
			}
		}
	}
	return alerts
}
