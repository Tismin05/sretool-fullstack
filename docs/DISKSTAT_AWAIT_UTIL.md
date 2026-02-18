# DiskStat 完整方案 - 添加 Await 和 Util 字段

## 一、问题说明

### AlertConfig 定义了 4 个磁盘阈值，但 DiskStat 只有 2 个对应字段：

| AlertConfig 字段 | 含义 | DiskStat 字段 | 状态 |
|-----------------|------|--------------|------|
| `DiskThreshold` | 磁盘空间使用率 | `UsedPercent` | ✅ 已有 |
| `DiskAwaitThreshold` | 平均 IO 等待时间(毫秒) | ❌ 无 | 需要添加 |
| `DiskUtilThreshold` | 磁盘利用率(%) | ❌ 无 | 需要添加 |
| `InodesThreshold` | Inodes 使用率 | `InodesUsedPercent` | ✅ 已有 |

---

## 二、修改步骤

### Step 1: 修改 `internal/model/metrics.go`

在 `DiskStat` 结构体中添加 `Await` 和 `Util` 字段：

```go
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
    Read              uint64  `json:"read"`               // 读 I/O 次数
    ReadSectors       uint64  `json:"read_sectors"`       // 读扇区数（新增）
    ReadSpeed         float64 `json:"read_speed"`
    Write             uint64  `json:"write"`              // 写 I/O 次数
    WriteSectors      uint64  `json:"write_sectors"`      // 写扇区数（新增）
    WriteSpeed        float64 `json:"write_speed"`

    // 新增：IO 性能指标
    Await float64 `json:"await"`  // 平均等待时间（毫秒）
    Util  float64 `json:"util"`   // 磁盘利用率（百分比）
}
```

**注意**：`Read` 和 `Write` 字段原本存储的是 `ReadIOs` 和 `WriteIOs`（I/O 次数），需要额外添加 `ReadSectors` 和 `WriteSectors` 字段来存储扇区数，用于后续计算。

---

### Step 2: 修改 `internal/collector/linux_proc.go`

#### 2.1 修改 `DiskIOStat` 结构体

添加 `IOQueueTime` 字段（第 12 字段，毫秒）：

```go
// 3) 读取 /proc/diskstats (IO 计数)，只保留物理磁盘
type DiskIOStat struct {
    Name          string
    ReadIOs       uint64  // 读 I/O 次数
    ReadSectors   uint64  // 读扇区数
    WriteIOs      uint64  // 写 I/O 次数
    WriteSectors  uint64  // 写扇区数
    IOQueueTime   uint64  // I/O 花费时间（毫秒，新增）
}
```

#### 2.2 修改 `readDiskStats` 函数

解析第 12 字段（IO 花费时间）：

```go
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

        // 过滤虚拟设备和分区
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
        // 新增：解析 IO 花费时间（第 13 字段，从 0 开始计数）
        ioQueueTime, _ := strconv.ParseUint(fields[12], 10, 64)

        stats[name] = DiskIOStat{
            Name:         name,
            ReadIOs:      readIO,
            ReadSectors:  readSectors,
            WriteIOs:     writeIO,
            WriteSectors: writeSectors,
            IOQueueTime:  ioQueueTime,  // 新增
        }
    }
    return stats, nil
}
```

#### 2.3 修改 `CollectDisk` 函数

在构建 `DiskStat` 时添加新字段：

```go
func CollectDisk(ctx context.Context) ([]model.DiskStat, error) {
    // ... 现有代码 ...

    for deviceName, ioStat := range ioStats {
        // ... 现有代码 ...

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
            ReadSectors:       ioStat.ReadSectors,   // 新增
            Write:             ioStat.WriteIOs,
            WriteSectors:      ioStat.WriteSectors,  // 新增
            IOQueueTime:       ioStat.IOQueueTime,  // 新增
            // 注意：Await 和 Util 在 calculater.go 中计算
        })
    }
    return out, nil
}
```

---

### Step 3: 修改 `internal/engine/calculater.go`

在 `CalculateRate` 函数中添加 Await 和 Util 的计算逻辑：

```go
func CalculateRate(prev, cur model.Metrics, interval time.Duration) model.Metrics {
    if interval.Seconds() <= 0 {
        return cur
    }
    seconds := interval.Seconds()
    res := cur
    var used uint64

    // CPU相关信息计算
    diffTotal := cur.CPU.TotalTicks - prev.CPU.TotalTicks
    diffIdle := cur.CPU.IdleTicks - prev.CPU.IdleTicks
    if diffTotal > 0 && diffIdle > 0 && diffTotal > diffIdle {
        used = diffTotal - diffIdle
    }
    res.CPU.UsagePercent = float64(used) / float64(diffTotal) * 100

    // 磁盘相关信息计算
    for i := range res.Disk {
        if i < len(prev.Disk) && res.Disk[i].Device == prev.Disk[i].Device {
            // 读写速度（IO 次数/秒）
            res.Disk[i].ReadSpeed = float64(res.Disk[i].Read-prev.Disk[i].Read) / seconds
            res.Disk[i].WriteSpeed = float64(res.Disk[i].Write-prev.Disk[i].Write) / seconds

            // ========== 新增：计算 Await ==========
            // Await = I/O 等待时间 / I/O 次数
            // 需要计算两次采集之间的增量
            diffReadIOs := res.Disk[i].Read - prev.Disk[i].Read
            diffWriteIOs := res.Disk[i].Write - prev.Disk[i].Write
            diffIOTime := float64(res.Disk[i].IOQueueTime - prev.Disk[i].IOQueueTime) // 毫秒
            totalIOs := diffReadIOs + diffWriteIOs

            if totalIOs > 0 && diffIOTime > 0 {
                // Await: 平均每次 I/O 的等待时间（毫秒）
                res.Disk[i].Await = diffIOTime / float64(totalIOs)

                // Util: 磁盘利用率 = I/O 时间占比
                // interval.Seconds() * 1000 = 总毫秒数
                // diffIOTime / (interval * 1000) * 100 = 利用率百分比
                totalMs := seconds * 1000
                res.Disk[i].Util = diffIOTime / totalMs * 100
            }
        }
    }

    // 网络相关信息计算
    for i := range res.Net {
        if i < len(prev.Net) && res.Net[i].Name == prev.Net[i].Name {
            res.Net[i].RxSpeed = float64(res.Net[i].RxBytes-prev.Net[i].RxBytes) / seconds
            res.Net[i].TxSpeed = float64(res.Net[i].TxBytes-prev.Net[i].TxBytes) / seconds
        }
    }
    return res
}
```

**注意**：上面的代码假设 `DiskStat` 中有 `IOQueueTime` 字段。如果不添加该字段到 `DiskStat`，可以改用扇区数近似计算：

```go
// 替代方案：不依赖 IOQueueTime，用扇区数近似计算
// 仅供无法获取 IOQueueTime 时使用
diffReadSectors := res.Disk[i].ReadSectors - prev.Disk[i].ReadSectors
diffWriteSectors := res.Disk[i].WriteSectors - prev.Disk[i].WriteSectors
diffTotalSectors := diffReadSectors + diffWriteSectors

// 假设平均每个扇区 512 字节，磁盘转速 7200 RPM，seek 时间约 8ms
// 这是一个粗略估算，不准确但可以作为 fallback
if diffTotalSectors > 0 && seconds > 0 {
    // 用扇区数除以时间估算吞吐量（简化计算）
    // 实际 Await 和 Util 需要内核的 iostat 数据
}
```

---

### Step 4: 修改告警规则检查 `internal/alert/rules.go`

添加 Await 和 Util 的阈值检查：

```go
func (r *RuleChecker) checkDisk(m model.Metrics) []Alert {
    var alerts []Alert
    for _, disk := range m.Disk {
        // 磁盘使用率检查
        if disk.UsedPercent > r.config.DiskThreshold {
            alerts = append(alerts, Alert{
                Level:     LevelError,
                Category:  CategoryDisk,
                Metric:    "used_percent",
                Message:   fmt.Sprintf("Disk %s usage %.1f%% exceeds threshold %.1f%%", disk.MountPoint, disk.UsedPercent, r.config.DiskThreshold),
                Value:     disk.UsedPercent,
                Threshold: r.config.DiskThreshold,
                Unit:      "%",
            })
        }

        // 新增：磁盘 await 检查
        if disk.Await > r.config.DiskAwaitThreshold && r.config.DiskAwaitThreshold > 0 {
            alerts = append(alerts, Alert{
                Level:     LevelWarn,
                Category:  CategoryDisk,
                Metric:    "await",
                Message:   fmt.Sprintf("Disk %s await %.1fms exceeds threshold %.1fms", disk.MountPoint, disk.Await, r.config.DiskAwaitThreshold),
                Value:     disk.Await,
                Threshold: r.config.DiskAwaitThreshold,
                Unit:      "ms",
            })
        }

        // 新增：磁盘 util 检查
        if disk.Util > r.config.DiskUtilThreshold && r.config.DiskUtilThreshold > 0 {
            alerts = append(alerts, Alert{
                Level:     LevelWarn,
                Category:  CategoryDisk,
                Metric:    "util",
                Message:   fmt.Sprintf("Disk %s util %.1f%% exceeds threshold %.1f%%", disk.Util, disk.Util, r.config.DiskUtilThreshold),
                Value:     disk.Util,
                Threshold: r.config.DiskUtilThreshold,
                Unit:      "%",
            })
        }
    }
    return alerts
}
```

---

## 三、/proc/diskstats 字段参考

```
/proc/diskstats 格式（14 个字段）：
  1: major number
  2: minor number
  3: device name
  4: reads completed successfully
  5: reads merged
  6: sectors read
  7: time spent reading (ms)
  8: writes completed successfully
  9: writes merged
 10: sectors written
 11: time spent writing (ms)
 12: I/Os currently in progress      <-- 当前进行中的 I/O
13: time spent doing I/Os (ms)       <-- I/O 花费总时间（累计）
14: weighted time spent doing I/Os (ms)

字段索引（从 0 开始）：fields[12] 和 fields[13]
```

**推荐使用字段 [13]**：`time spent doing I/Os`（累计 I/O 时间），用于计算两次采集之间的增量。

---

## 四、配置文件对应

`configs/config.yaml` 中的告警阈值配置：

```yaml
alert:
  enabled: true
  disk_threshold: 85.0          # 磁盘使用率阈值 (%)
  disk_await_threshold: 50.0    # 磁盘平均等待时间阈值 (ms)
  disk_util_threshold: 80.0     # 磁盘利用率阈值 (%)
  inodes_threshold: 80.0
```

---

## 五、修改文件清单

| 文件 | 修改内容 |
|------|---------|
| `internal/model/metrics.go` | 添加 `ReadSectors`, `WriteSectors`, `Await`, `Util` 字段 |
| `internal/collector/linux_proc.go` | 修改 `DiskIOStat` 和 `readDiskStats`，解析 IO 时间 |
| `internal/engine/calculater.go` | 添加 Await 和 Util 的计算逻辑 |
| `internal/alert/rules.go` | 添加 Await 和 Util 的告警检查 |
| `configs/config.yaml` | 添加对应阈值配置（可选） |

---

## 六、注意事项

1. **首次计算时 prev 为空**：第一次调用 `CalculateRate` 时，`prev` 没有数据，`Await` 和 `Util` 会为 0，这是正常现象

2. **interval 不能太短**：如果采集间隔太短（例如 < 1 秒），计算结果可能波动较大，建议间隔 >= 5 秒

3. **IOQueueTime 可能溢出**：该字段是累计值，在长时间运行后可能溢出（回绕），需要处理

4. **向后兼容性**：如果使用 gopsutil 采集器，需要同步修改 `local_MacOS.go` 和 `local_Linux.go`
