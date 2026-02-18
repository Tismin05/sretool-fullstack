package model

// Metrics 系统核心指标
type Metrics struct {
	CPU             CPUStat    `json:"cpu"`
	Mem             MemoryStat `json:"memory"`
	Disk            []DiskStat `json:"disk"`
	Net             []NetStat  `json:"net"`
	Procs           []ProcStat `json:"procs"`
	Host            string     `json:"host"`
	UpdateTimestamp string     `json:"update_timestamp"`
}

type CPUStat struct {
	Cores        int       `json:"cores"`
	UsagePercent float64   `json:"usage_percent"`
	PerCPUUsage  []float64 `json:"per_cpu_usage"`
	Load1        float64   `json:"load1"`
	Load5        float64   `json:"load5"`
	Load15       float64   `json:"load15"`
	TotalTicks   uint64    `json:"total_ticks"`
	IdleTicks    uint64    `json:"idle_ticks"`
}

type MemoryStat struct {
	Total           uint64  `json:"total"`
	Free            uint64  `json:"free"`
	Available       uint64  `json:"available"`
	Used            uint64  `json:"used"`
	UsedPercent     float64 `json:"used_percent"`
	SwapTotal       uint64  `json:"swap_total"`
	SwapFree        uint64  `json:"swap_free"`
	SwapUsed        uint64  `json:"swap_used"`
	SwapUsedPercent float64 `json:"swap_used_percent"`
}

type DiskStat struct {
	MountPoint        string  `json:"mount_point"`
	Device            string  `json:"device"`
	Total             uint64  `json:"total"`
	Used              uint64  `json:"used"`
	Free              uint64  `json:"free"`
	UsedPercent       float64 `json:"used_percent"`
	InodesTotal       uint64  `json:"inodes_total"`
	InodesUsed        uint64  `json:"inodes_used"`
	InodesFree        uint64  `json:"inodes_free"`
	InodesUsedPercent float64 `json:"inodes_used_percent"`
	Read              uint64  `json:"read"`
	ReadSectors       uint64  `json:"read_sectors"`
	ReadSpeed         float64 `json:"read_speed"`
	Write             uint64  `json:"write"`
	WriteSectors      uint64  `json:"write_sectors"`
	WriteSpeed        float64 `json:"write_speed"`
	Await             float64 `json:"await"`
	Util              float64 `json:"util"`
	IOQueueTime       uint64  `json:"io_queue_time"`
}

type NetStat struct {
	Name      string  `json:"name"`
	RxBytes   uint64  `json:"rx_bytes"`
	RxPackets uint64  `json:"rx_packets"`
	RxErrors  uint64  `json:"rx_errors"`
	RxDropped uint64  `json:"rx_dropped"`
	TxBytes   uint64  `json:"tx_bytes"`
	TxPackets uint64  `json:"tx_packets"`
	TxErrors  uint64  `json:"tx_errors"`
	TxDropped uint64  `json:"tx_dropped"`
	RxSpeed   float64 `json:"rx_speed"`
	TxSpeed   float64 `json:"tx_speed"`
}

type ProcStat struct {
	PID  int     `json:"pid"`
	Name string  `json:"name"`
	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
}
