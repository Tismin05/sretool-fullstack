package collector

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"sretool-fullstack/internal/model"
	"sretool-fullstack/pkg/utils"

	"golang.org/x/sys/unix"
)

// LinuxCollector 基于 /proc 文件系统的采集器
type LinuxCollector struct {
	host string
}

// 确保 LinuxCollector 实现了 Collector 接口
var _ Collector = &LinuxCollector{}

// NewLinuxCollector 创建新的 Linux 采集器
func NewLinuxCollector() (*LinuxCollector, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &LinuxCollector{host: hostname}, nil
}

// Collect 采集所有系统指标
func (c *LinuxCollector) Collect(ctx context.Context) (*model.Metrics, *model.CollectErrors) {
	metrics := &model.Metrics{
		Host:            c.host,
		UpdateTimestamp: time.Now().Format(time.RFC3339),
	}

	errs := c.collectAll(ctx, metrics)
	return metrics, errs
}

// collectAll 并行采集所有子系统指标
func (c *LinuxCollector) collectAll(ctx context.Context, m *model.Metrics) *model.CollectErrors {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := &model.CollectErrors{}

	wg.Add(5)

	// 1. CPU 采集
	go func() {
		defer wg.Done()
		cpuStat, err := CollectCPUStat(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs.CPU = append(errs.CPU, err)
		}
		m.CPU = cpuStat
	}()

	// 2. 内存采集
	go func() {
		defer wg.Done()
		memStat, err := CollectMeminfo(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs.Mem = append(errs.Mem, err)
		}
		m.Mem = *memStat
	}()

	// 3. 磁盘采集
	go func() {
		defer wg.Done()
		diskStats, err := CollectDisk(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs.Disk = append(errs.Disk, err)
		}
		m.Disk = diskStats
	}()

	// 4. 网络采集
	go func() {
		defer wg.Done()
		netStats, err := CollectNetinfo(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs.Net = append(errs.Net, err)
		}
		m.Net = netStats
	}()

	// 5. 进程采集
	go func() {
		defer wg.Done()
		procStats, err := CollectProcs(ctx, 10)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			log.Printf("Process collection warning: %v", err)
		}
		m.Procs = procStats
	}()

	wg.Wait()

	if !errs.HasError() {
		return nil
	}
	return errs
}

// collectCPUCores 采集 CPU 核心数
func collectCPUCores(ctx context.Context) (cores int, err error) {
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/cpuinfo", 0, -1)
	if err != nil {
		return 0, err
	}

	cpuCores := 0
	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		field := strings.SplitN(line, ":", 2)
		if len(field) != 2 {
			continue
		}
		key := strings.TrimSpace(field[0])
		if key == "processor" {
			cpuCores++
		}
	}
	if cpuCores == 0 {
		return 0, fmt.Errorf("no processor entries found in /proc/cpuinfo")
	}
	return cpuCores, nil
}

// CollectCPUStat 采集 CPU 统计信息
func CollectCPUStat(ctx context.Context) (model.CPUStat, error) {
	cores, err := collectCPUCores(ctx)
	if err != nil {
		return model.CPUStat{}, err
	}

	perCPU, err := collectCPUInfo(ctx, cores)
	if err != nil {
		return model.CPUStat{}, err
	}

	avg := 0.0
	for _, v := range perCPU {
		avg += v
	}
	if len(perCPU) > 0 {
		avg /= float64(len(perCPU))
	}

	loads, err := collectLoadAvg(ctx)
	if err != nil {
		return model.CPUStat{}, err
	}
	if len(loads) < 3 {
		return model.CPUStat{}, fmt.Errorf("invalid loadavg length: %d", len(loads))
	}

	return model.CPUStat{
		Cores:        cores,
		UsagePercent: avg,
		PerCPUUsage:  perCPU,
		Load1:        loads[0],
		Load5:        loads[1],
		Load15:       loads[2],
	}, nil
}

func collectCPUInfo(ctx context.Context, cores int) (perCPU []float64, err error) {
	var perCPUUsage []float64
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/stat", 0, -1)
	if err != nil {
		return nil, err
	}
	for _, line := range lines[1 : cores+1] {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			log.Printf("invalid format of /proc/stat: %s", line)
			continue
		}
		if fields[0] == "cpu" {
			continue
		}
		if len(fields) < 5 {
			log.Printf("invalid format of /proc/stat: %s", line)
			continue
		}

		usr, _ := strconv.ParseFloat(fields[1], 64)
		nice, _ := strconv.ParseFloat(fields[2], 64)
		system, _ := strconv.ParseFloat(fields[3], 64)
		idle, _ := strconv.ParseFloat(fields[4], 64)

		iowait := 0.0
		if len(fields) > 5 {
			iowait, _ = strconv.ParseFloat(fields[5], 64)
		}
		irq := 0.0
		if len(fields) > 6 {
			irq, _ = strconv.ParseFloat(fields[6], 64)
		}
		softirq := 0.0
		if len(fields) > 7 {
			softirq, _ = strconv.ParseFloat(fields[7], 64)
		}
		steal := 0.0
		if len(fields) > 8 {
			steal, _ = strconv.ParseFloat(fields[8], 64)
		}
		guest := 0.0
		if len(fields) > 9 {
			guest, _ = strconv.ParseFloat(fields[9], 64)
		}
		guestnice := 0.0
		if len(fields) > 10 {
			guestnice, _ = strconv.ParseFloat(fields[10], 64)
		}

		total := usr + nice + system + idle + iowait + irq + softirq + steal + guest + guestnice
		if total <= 0 {
			continue
		}
		busy := total - idle - iowait
		usage := busy / total * 100
		perCPUUsage = append(perCPUUsage, usage)
	}
	return perCPUUsage, nil
}

func collectLoadAvg(ctx context.Context) ([]float64, error) {
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/loadavg", 0, -1)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty /proc/loadavg")
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid /proc/loadavg: %s", lines[0])
	}
	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)
	return []float64{load1, load5, load15}, nil
}

// CollectMeminfo 采集内存信息
func CollectMeminfo(ctx context.Context) (*model.MemoryStat, error) {
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/meminfo", 0, -1)
	if err != nil {
		log.Printf("error collecting meminfo: %s", err)
		return nil, err
	}

	ret := &model.MemoryStat{}
	var buffer uint64
	var cache uint64
	var available uint64

	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fields := strings.SplitN(line, ":", 2)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		valueFields := strings.Fields(strings.TrimSpace(fields[1]))
		if len(valueFields) == 0 {
			continue
		}
		value := valueFields[0]

		switch key {
		case "MemTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Total = t * 1024
		case "MemFree":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Free = t * 1024
		case "MemAvailable":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			available = t * 1024
		case "SwapTotal":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.SwapTotal = t * 1024
		case "SwapFree":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.SwapFree = t * 1024
		case "Buffers":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			buffer = t * 1024
		case "Cached":
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return ret, err
			}
			cache = t * 1024
		}
	}

	if available > 0 {
		ret.Available = available
		ret.Used = ret.Total - available
	} else {
		ret.Used = ret.Total - ret.Free - buffer - cache
	}
	if ret.Total > 0 {
		ret.UsedPercent = float64(ret.Used) / float64(ret.Total) * 100
	}
	if ret.SwapTotal >= ret.SwapFree {
		ret.SwapUsed = ret.SwapTotal - ret.SwapFree
		if ret.SwapTotal > 0 {
			ret.SwapUsedPercent = float64(ret.SwapUsed) / float64(ret.SwapTotal) * 100
		}
	}
	return ret, nil
}

// readMounts 读取挂载点信息
func readMounts(ctx context.Context) (map[string]string, error) {
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/mounts", 0, -1)
	if err != nil {
		return nil, err
	}
	mounts := make(map[string]string)
	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		if isVirtualFS(fsType) {
			continue
		}

		deviceName := strings.TrimPrefix(device, "/dev/")
		mounts[deviceName] = mountPoint
	}
	return mounts, nil
}

func isVirtualFS(fsType string) bool {
	virtualFSTypes := map[string]bool{
		"tmpfs": true, "devtmpfs": true, "overlay": true, "aufs": true,
		"devpts": true, "sysfs": true, "proc": true, "cgroup": true,
		"cgroup2": true, "securityfs": true, "pstore": true, "efivarfs": true,
		"bpf": true, "tracefs": true, "hugetlbfs": true, "mqueue": true,
		"fusectl": true, "configfs": true, "debugfs": true, "selinuxfs": true,
	}
	return virtualFSTypes[fsType]
}

func statFS(path string) (total, free, avail, inodes, inodesFree uint64, err error) {
	var st unix.Statfs_t
	if err = unix.Statfs(path, &st); err != nil {
		return
	}
	total = st.Blocks * uint64(st.Bsize)
	free = st.Bfree * uint64(st.Bsize)
	avail = st.Bavail * uint64(st.Bsize)
	inodes = st.Files
	inodesFree = st.Ffree
	return
}

// DiskIOStat 磁盘 IO 统计
type DiskIOStat struct {
	Name         string
	ReadIOs      uint64
	ReadSectors  uint64
	WriteIOs     uint64
	WriteSectors uint64
	IOQueuesTime uint64
}

func readDiskStats(ctx context.Context) (map[string]DiskIOStat, error) {
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/diskstats", 0, -1)
	if err != nil {
		return nil, err
	}
	stats := make(map[string]DiskIOStat)
	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}
		name := strings.TrimSpace(fields[2])

		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}
		if isPartition(name) {
			continue
		}

		readIO, _ := strconv.ParseUint(fields[3], 10, 64)
		readSectors, _ := strconv.ParseUint(fields[5], 10, 64)
		writeIO, _ := strconv.ParseUint(fields[7], 10, 64)
		writeSectors, _ := strconv.ParseUint(fields[9], 10, 64)
		ioQueuesTime, _ := strconv.ParseUint(fields[12], 10, 64)

		stats[name] = DiskIOStat{
			Name:         name,
			ReadIOs:      readIO,
			ReadSectors:  readSectors,
			WriteIOs:     writeIO,
			WriteSectors: writeSectors,
			IOQueuesTime: ioQueuesTime,
		}
	}
	return stats, nil
}

func isPartition(name string) bool {
	if strings.HasPrefix(name, "nvme") && strings.Contains(name, "p") {
		return true
	}
	if len(name) > 3 {
		_, err := strconv.Atoi(name[3:])
		return err == nil
	}
	return false
}

// CollectDisk 采集磁盘信息
func CollectDisk(ctx context.Context) ([]model.DiskStat, error) {
	mounts, err := readMounts(ctx)
	if err != nil {
		return nil, err
	}

	ioStats, err := readDiskStats(ctx)
	if err != nil {
		return nil, err
	}

	var out []model.DiskStat
	for deviceName, ioStat := range ioStats {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		mountPoint, ok := mounts[deviceName]
		if !ok {
			continue
		}

		total, free, _, inodes, inodesFree, err := statFS(mountPoint)
		if err != nil {
			continue
		}
		used := total - free
		usedPct := 0.0
		if total > 0 {
			usedPct = float64(used) / float64(total) * 100
		}

		out = append(out, model.DiskStat{
			MountPoint:        mountPoint,
			Device:            deviceName,
			Total:             total,
			Free:              free,
			Used:              used,
			UsedPercent:       usedPct,
			InodesTotal:       inodes,
			InodesFree:        inodesFree,
			InodesUsed:        inodes - inodesFree,
			InodesUsedPercent: utils.Pct(inodes-inodesFree, inodes),
			Read:              ioStat.ReadIOs,
			Write:             ioStat.WriteIOs,
			IOQueueTime:       ioStat.IOQueuesTime,
		})
	}
	return out, nil
}

// CollectNetinfo 采集网络信息
func CollectNetinfo(ctx context.Context) ([]model.NetStat, error) {
	m := make([]model.NetStat, 0)
	lines, err := utils.ReadLinesOffsetNWithContext(ctx, "/proc/net/dev", 2, -1)
	if err != nil {
		log.Printf("error collecting net io: %s", err)
		return nil, err
	}

	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		separation := strings.LastIndex(line, ":")
		if separation == -1 {
			continue
		}
		parts := make([]string, 2)
		parts[0] = line[:separation]
		parts[1] = line[separation+1:]

		interfaceName := strings.TrimSpace(parts[0])
		if interfaceName == "" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 12 {
			log.Printf("invalid format of /proc/net/dev: %s", line)
			continue
		}
		recvBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		recvPackets, _ := strconv.ParseUint(fields[1], 10, 64)
		recvErrors, _ := strconv.ParseUint(fields[2], 10, 64)
		recvDrops, _ := strconv.ParseUint(fields[3], 10, 64)
		sendBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		sendPackets, _ := strconv.ParseUint(fields[9], 10, 64)
		sendErrors, _ := strconv.ParseUint(fields[10], 10, 64)
		sendDrops, _ := strconv.ParseUint(fields[11], 10, 64)

		netStat := model.NetStat{
			Name:      interfaceName,
			RxBytes:   recvBytes,
			RxPackets: recvPackets,
			RxErrors:  recvErrors,
			RxDropped: recvDrops,
			TxBytes:   sendBytes,
			TxPackets: sendPackets,
			TxErrors:  sendErrors,
			TxDropped: sendDrops,
		}
		m = append(m, netStat)
	}
	return m, nil
}

// CollectProcs 采集进程信息
func CollectProcs(ctx context.Context, topN int) ([]model.ProcStat, error) {
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var procStats []model.ProcStat

	for _, entry := range procs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		name := entry.Name()
		pid, err := strconv.Atoi(name)
		if err != nil {
			continue
		}

		statPath := fmt.Sprintf("/proc/%s/stat", name)
		data, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		// 解析 stat 文件
		content := string(data)
		idx := strings.Index(content, "(")
		if idx == -1 {
			continue
		}
		endIdx := strings.LastIndex(content, ")")
		if endIdx == -1 || endIdx < idx {
			continue
		}

		procName := content[idx+1 : endIdx]
		parts := strings.Fields(content[endIdx+2:])
		if len(parts) < 14 {
			continue
		}

		// 获取 CPU 时间
		utime, _ := strconv.ParseUint(parts[12], 10, 64)
		stime, _ := strconv.ParseUint(parts[13], 10, 64)
		totalTime := utime + stime

		// 获取内存使用
		memRss, _ := strconv.ParseUint(parts[23], 10, 64)

		procStats = append(procStats, model.ProcStat{
			PID:  pid,
			Name: procName,
			CPU:  float64(totalTime), // 需要除以 CPU 时钟 tick
			Mem:  float64(memRss),   // 需要转换为百分比
		})

		if len(procStats) >= topN*2 {
			break
		}
	}

	// 简单排序取 Top N（按 CPU）
	if len(procStats) > topN {
		// 简单选择排序
		for i := 0; i < len(procStats) && i < topN; i++ {
			maxIdx := i
			for j := i + 1; j < len(procStats); j++ {
				if procStats[j].CPU > procStats[maxIdx].CPU {
					maxIdx = j
				}
			}
			procStats[i], procStats[maxIdx] = procStats[maxIdx], procStats[i]
		}
		procStats = procStats[:topN]
	}

	return procStats, nil
}
