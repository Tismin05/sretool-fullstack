package diagnostic

import (
	"context"
)

// Diagnostic 诊断接口
type Diagnostic interface {
	// Diagnose 执行系统诊断
	Diagnose(ctx context.Context) (*DiagnosticResult, error)
}

// DiagnosticResult 诊断结果
type DiagnosticResult struct {
	CPU      CPUDiagnostic      `json:"cpu"`
	Memory   MemoryDiagnostic   `json:"memory"`
	Disk     []DiskDiagnostic   `json:"disk"`
	Network  []NetworkDiagnostic `json:"network"`
	TopProcs []ProcDiagnostic   `json:"top_procs"`
}

// CPUDiagnostic CPU 诊断结果
type CPUDiagnostic struct {
	LoadStatus      string  `json:"load_status"`
	UsagePercent   float64 `json:"usage_percent"`
	Load1          float64 `json:"load1"`
	Load5          float64 `json:"load5"`
	Load15         float64 `json:"load15"`
	LoadPerCore    float64 `json:"load_per_core"`
	Recommendation string  `json:"recommendation"`
}

// MemoryDiagnostic 内存诊断结果
type MemoryDiagnostic struct {
	Status          string  `json:"status"`
	UsedPercent     float64 `json:"used_percent"`
	Available       uint64  `json:"available"`
	SwapUsedPercent float64 `json:"swap_used_percent"`
	Recommendation  string  `json:"recommendation"`
}

// DiskDiagnostic 磁盘诊断结果
type DiskDiagnostic struct {
	MountPoint        string  `json:"mount_point"`
	Status            string  `json:"status"`
	UsedPercent      float64 `json:"used_percent"`
	InodesUsedPercent float64 `json:"inodes_used_percent"`
	Await            float64 `json:"await"`
	Util             float64 `json:"util"`
	Recommendation   string  `json:"recommendation"`
}

// NetworkDiagnostic 网络诊断结果
type NetworkDiagnostic struct {
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	ErrorRate      float64 `json:"error_rate"`
	DropRate       float64 `json:"drop_rate"`
	RxSpeed        float64 `json:"rx_speed"`
	TxSpeed        float64 `json:"tx_speed"`
	Recommendation string  `json:"recommendation"`
}

// ProcDiagnostic 进程诊断结果
type ProcDiagnostic struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	CPU     float64 `json:"cpu"`
	Mem     float64 `json:"mem"`
	State   string  `json:"state"`
	Command string  `json:"command"`
}
