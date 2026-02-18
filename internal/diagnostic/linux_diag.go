package diagnostic

import (
	"context"
	"fmt"

	"tisminSRETool/internal/collector"
	"tisminSRETool/internal/model"
)

// LinuxDiagnostic Linux 系统诊断实现
type LinuxDiagnostic struct {
	collector collector.Collector
	topN      int
}

// NewLinuxDiagnostic 创建 Linux 诊断器
func NewLinuxDiagnostic(c collector.Collector, topN int) *LinuxDiagnostic {
	if topN <= 0 {
		topN = 10
	}
	return &LinuxDiagnostic{
		collector: c,
		topN:      topN,
	}
}

// Diagnose 执行系统诊断
func (d *LinuxDiagnostic) Diagnose(ctx context.Context) (*DiagnosticResult, error) {
	metrics, errs := d.collector.Collect(ctx)
	if errs != nil && errs.HasError() {
		return nil, fmt.Errorf("collection errors: %v", errs)
	}

	if metrics == nil {
		return nil, fmt.Errorf("no metrics collected")
	}

	result := &DiagnosticResult{
		CPU:      d.diagnoseCPU(metrics),
		Memory:   d.diagnoseMemory(metrics),
		Disk:     d.diagnoseDisk(metrics),
		Network:  d.diagnoseNetwork(metrics),
		TopProcs: d.diagnoseProcs(metrics),
	}

	return result, nil
}

func (d *LinuxDiagnostic) diagnoseCPU(m *model.Metrics) CPUDiagnostic {
	diag := CPUDiagnostic{
		UsagePercent: m.CPU.UsagePercent,
		Load1:        m.CPU.Load1,
		Load5:        m.CPU.Load5,
		Load15:       m.CPU.Load15,
	}

	if m.CPU.Cores > 0 {
		diag.LoadPerCore = m.CPU.Load1 / float64(m.CPU.Cores)
	}

	// 判断负载状态
	if m.CPU.Load1 > float64(m.CPU.Cores)*2 {
		diag.LoadStatus = "critical"
		diag.Recommendation = "CPU 负载极高，建议检查高耗 CPU 进程"
	} else if m.CPU.Load1 > float64(m.CPU.Cores) {
		diag.LoadStatus = "high"
		diag.Recommendation = "CPU 负载较高，建议关注"
	} else {
		diag.LoadStatus = "normal"
		diag.Recommendation = "CPU 负载正常"
	}

	// CPU 使用率建议
	if m.CPU.UsagePercent > 90 {
		diag.Recommendation += "，CPU 使用率超过 90%"
	} else if m.CPU.UsagePercent > 70 {
		diag.Recommendation += "，CPU 使用率偏高"
	}

	return diag
}

func (d *LinuxDiagnostic) diagnoseMemory(m *model.Metrics) MemoryDiagnostic {
	diag := MemoryDiagnostic{
		UsedPercent:     m.Mem.UsedPercent,
		Available:       m.Mem.Available,
		SwapUsedPercent: m.Mem.SwapUsedPercent,
	}

	// 判断内存状态
	if m.Mem.UsedPercent > 95 {
		diag.Status = "critical"
		diag.Recommendation = "内存使用率超过 95%，系统可能 OOM，建议立即处理"
	} else if m.Mem.UsedPercent > 85 {
		diag.Status = "warning"
		diag.Recommendation = "内存使用率超过 85%，建议关注"
	} else if m.Mem.UsedPercent > 70 {
		diag.Status = "warning"
		diag.Recommendation = "内存使用率偏高"
	} else {
		diag.Status = "normal"
		diag.Recommendation = "内存使用正常"
	}

	// Swap 建议
	if m.Mem.SwapUsedPercent > 50 {
		diag.Recommendation += fmt.Sprintf(", Swap 使用率 %.1f%% 偏高", m.Mem.SwapUsedPercent)
	}

	return diag
}

func (d *LinuxDiagnostic) diagnoseDisk(m *model.Metrics) []DiskDiagnostic {
	var diags []DiskDiagnostic

	for _, disk := range m.Disk {
		diag := DiskDiagnostic{
			MountPoint:        disk.MountPoint,
			UsedPercent:       disk.UsedPercent,
			InodesUsedPercent: disk.InodesUsedPercent,
			Await:             disk.Await,
			Util:              disk.Util,
		}

		// 判断磁盘状态
		if disk.UsedPercent > 95 || disk.InodesUsedPercent > 95 {
			diag.Status = "critical"
			diag.Recommendation = "磁盘空间或 Inodes 即将用尽"
		} else if disk.UsedPercent > 85 || disk.InodesUsedPercent > 85 {
			diag.Status = "warning"
			diag.Recommendation = "磁盘空间或 Inodes 使用率偏高"
		} else if disk.Await > 100 || disk.Util > 80 {
			diag.Status = "warning"
			diag.Recommendation = fmt.Sprintf("磁盘 IO 压力较大 (await=%.1fms, util=%.1f%%)", disk.Await, disk.Util)
		} else {
			diag.Status = "normal"
			diag.Recommendation = "磁盘状态正常"
		}

		diags = append(diags, diag)
	}

	return diags
}

func (d *LinuxDiagnostic) diagnoseNetwork(m *model.Metrics) []NetworkDiagnostic {
	var diags []NetworkDiagnostic

	for _, net := range m.Net {
		diag := NetworkDiagnostic{
			Name:    net.Name,
			RxSpeed: net.RxSpeed,
			TxSpeed: net.TxSpeed,
		}

		// 计算错误率和丢包率
		totalPackets := net.RxPackets + net.TxPackets
		if totalPackets > 0 {
			diag.ErrorRate = float64(net.RxErrors+net.TxErrors) / float64(totalPackets) * 100
			diag.DropRate = float64(net.RxDropped+net.TxDropped) / float64(totalPackets) * 100
		}

		// 判断网络状态
		if diag.ErrorRate > 1 || diag.DropRate > 1 {
			diag.Status = "critical"
			diag.Recommendation = fmt.Sprintf("网络错误率 %.2f%% 或丢包率 %.2f%% 极高", diag.ErrorRate, diag.DropRate)
		} else if diag.ErrorRate > 0.1 || diag.DropRate > 0.1 {
			diag.Status = "warning"
			diag.Recommendation = fmt.Sprintf("网络存在少量错误或丢包")
		} else {
			diag.Status = "normal"
			diag.Recommendation = "网络状态正常"
		}

		diags = append(diags, diag)
	}

	return diags
}

func (d *LinuxDiagnostic) diagnoseProcs(m *model.Metrics) []ProcDiagnostic {
	var diags []ProcDiagnostic

	// 取前 N 个 CPU 占用最高的进程
	procs := m.Procs
	if len(procs) > d.topN {
		procs = procs[:d.topN]
	}

	for _, p := range procs {
		diag := ProcDiagnostic{
			PID:  p.PID,
			Name: p.Name,
			CPU:  p.CPU,
			Mem:  p.Mem,
		}
		diags = append(diags, diag)
	}

	return diags
}
