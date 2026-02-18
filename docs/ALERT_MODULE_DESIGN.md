# Alert 模块完整实现方案

## 一、当前问题

### 1.1 文件结构问题

| 文件 | 问题 |
|------|------|
| `interface.go` | 空文件，未定义任何接口 |
| `rules.go` | 网络检查循环体为空；Inodes 检查只有 if 没有 append；变量拼写错误 |
| `sender.go` | 存在重复的 `package alert` 声明和 import 语句（第 10-17 行）；拼写错误 `Sumject` → `Subject` |

---

## 二、设计目标

1. **模块化**: 接口抽象，支持多渠道告警发送
2. **可扩展**: 方便添加新的告警规则和发送渠道
3. **可靠性**: 支持告警抑制、聚合、恢复通知
4. **完整性**: 覆盖配置文件中定义的所有阈值

---

## 三、接口定义 (interface.go)

### 3.1 告警结构体

```go
// Alert 告警信息结构体
type Alert struct {
    Level     AlertLevel      // 告警级别：info/warn/error
    Category  AlertCategory   // 告警类别：cpu/memory/disk/network/inodes/tcp
    Metric    string          // 指标名称
    Message   string          // 告警消息
    Value     float64        // 当前值
    Threshold float64        // 阈值
    Unit      string          // 单位
    Timestamp time.Time       // 发生时间
    Host      string          // 主机名
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
```

### 3.2 告警检查器接口

```go
// AlertChecker 告警检查器接口
type AlertChecker interface {
    // Check 检查指标是否触发告警
    // 返回触发的告警列表
    Check(ctx context.Context, m model.Metrics) ([]Alert, error)
}
```

### 3.3 告警发送器接口

```go
// AlertSender 告警发送器接口
type AlertSender interface {
    // Send 发送告警
    // 支持 ctx 控制超时和取消
    Send(ctx context.Context, alerts []Alert, config model.EmailConfig) error
}
```

### 3.4 告警管理器接口

```go
// AlertManager 告警管理器接口
type AlertManager interface {
    // Run 启动告警管理器
    // 接收指标通道，定期检查并发送告警
    Run(ctx context.Context, metricsCh <-chan model.Metrics)

    // Stop 停止告警管理器
    Stop()
}
```

---

## 四、规则检查器实现 (rules.go)

### 4.1 RuleChecker 结构体

```go
type RuleChecker struct {
    config model.AlertConfig
}

func NewRuleChecker(config model.AlertConfig) *RuleChecker {
    return &RuleChecker{config: config}
}
```

### 4.2 Check 方法实现

```go
func (r *RuleChecker) Check(ctx context.Context, m model.Metrics) ([]Alert, error) {
    if !r.config.Enabled {
        return nil, nil
    }

    var alerts []Alert

    // 1. CPU 阈值检查
    alerts = append(alerts, r.checkCPU(m)...)

    // 2. 内存阈值检查
    alerts = append(alerts, r.checkMemory(m)...)

    // 3. 磁盘阈值检查
    alerts = append(alerts, r.checkDisk(m)...)

    // 4. Inodes 阈值检查
    alerts = append(alerts, r.checkInodes(m)...)

    // 5. 网络阈值检查
    alerts = append(alerts, r.checkNetwork(m)...)

    // 6. TCP 连接状态检查
    alerts = append(alerts, r.checkTCP(m)...)

    return alerts, nil
}
```

### 4.3 各类检查方法

#### 4.3.1 CPU 检查

```go
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
        })
    }
    return alerts
}
```

#### 4.3.2 内存检查

```go
func (r *RuleChecker) checkMemory(m model.Metrics) []Alert {
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
        })
    }
    return alerts
}
```

#### 4.3.3 磁盘检查

```go
func (r *RuleChecker) checkDisk(m model.Metrics) []Alert {
    var alerts []Alert
    for _, disk := range m.Disk {
        if disk.UsedPercent > r.config.DiskThreshold {
            alerts = append(alerts, Alert{
                Level:     LevelError,
                Category:  CategoryDisk,
                Metric:    "used_percent",
                Message:   fmt.Sprintf("Disk %s usage %.1f%% exceeds threshold %.1f%%", disk.MountPoint, disk.UsedPercent, r.config.DiskThreshold),
                Value:     disk.UsedPercent,
                Threshold: r.config.DiskThreshold,
                Unit:      "%",
                Host:      m.Host,
            })
        }
    }
    return alerts
}
```

#### 4.3.4 Inodes 检查（当前为空，需要实现）

```go
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
```

#### 4.3.5 网络检查（当前为空，需要实现）

根据配置文件的字段：
- `NetworkBandwidthThreshold` - 网卡带宽使用率阈值
- `NetworkPacketLossThreshold` - 丢包率阈值
- `NetworkRTTThreshold` - 网络延迟阈值

```go
func (r *RuleChecker) checkNetwork(m model.Metrics) []Alert {
    var alerts []Alert
    // 注意：gopsutil 未直接提供带宽使用率和 RTT
    // 这里需要额外采集或依赖其他数据源
    // 当前配置中有这些字段但 metrics.go 中没有对应数据
    // 建议：添加对应指标或在文档中说明需要扩展

    for _, net := range m.Net {
        // 检查错误率和丢包率
        totalPackets := net.RxPackets + net.TxPackets
        if totalPackets > 0 {
            errorRate := float64(net.RxErrors+net.TxErrors) / float64(totalPackets) * 100
            dropRate := float64(net.RxDropped+net.TxDropped) / float64(totalPackets) * 100

            if errorRate > 0.01 || dropRate > 0.01 { // 超过 0.01% 告警
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
```

#### 4.3.6 TCP 连接检查（当前为空，需要实现）

配置文件中有：
- `TCPTimeWaitThreshold` - TIME_WAIT 连接数阈值
- `TCPCLOSEWaitThreshold` - CLOSE_WAIT 连接数阈值
- `TotalTCPThreshold` - 总 TCP 连接数阈值

需要在 metrics.go 中添加 TCP 统计结构体：

```go
type TCPStat struct {
    Established int `json:"established"`
    TimeWait    int `json:"time_wait"`
    CloseWait   int `json:"close_wait"`
    Listen      int `json:"listen"`
    Total       int `json:"total"`
}
```

```go
func (r *RuleChecker) checkTCP(m model.Metrics) []Alert {
    var alerts []Alert

    // 需要先在 Metrics 中添加 TCP 统计字段
    // if m.TCP.TimeWait > int(r.config.TCPTimeWaitThreshold) {
    //     alerts = append(alerts, Alert{...})
    // }

    return alerts
}
```

---

## 五、告警发送器实现 (sender.go)

### 5.1 EmailSender 结构体

```go
type EmailSender struct {
    // 可配置重试次数
    maxRetries int
}

func NewEmailSender(maxRetries int) *EmailSender {
    return &EmailSender{maxRetries: maxRetries}
}
```

### 5.2 Send 方法实现

```go
func (s *EmailSender) Send(ctx context.Context, alerts []Alert, config model.EmailConfig) error {
    if len(alerts) == 0 {
        return nil
    }

    subject := s.buildSubject(alerts)
    body := s.buildBody(alerts)

    var lastErr error
    for i := 0; i <= s.maxRetries; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        lastErr = s.sendEmail(subject, body, config)
        if lastErr == nil {
            return nil
        }

        // 指数退避重试
        if i < s.maxRetries {
            time.Sleep(time.Duration(1<<i) * time.Second)
        }
    }

    return lastErr
}
```

### 5.3 邮件内容构建

```go
func (s *EmailSender) buildSubject(alerts []Alert) string {
    errorCount := 0
    warnCount := 0
    for _, a := range alerts {
        if a.Level == LevelError {
            errorCount++
        } else if a.Level == LevelWarn {
            warnCount++
        }
    }

    if errorCount > 0 {
        return fmt.Sprintf("[%s] %d Error Alerts from tisminSRETool", m.Host, errorCount)
    }
    return fmt.Sprintf("[%s] %d Warning Alerts from tisminSRETool", m.Host, warnCount)
}

func (s *EmailSender) buildBody(alerts []Alert) string {
    var buf strings.Builder

    buf.WriteString("Alert Report\n")
    buf.WriteString(strings.Repeat("=", 50))
    buf.WriteString("\n\n")

    for _, a := range alerts {
        buf.WriteString(fmt.Sprintf("[%s] %s\n", a.Level, a.Category))
        buf.WriteString(fmt.Sprintf("  Message: %s\n", a.Message))
        buf.WriteString(fmt.Sprintf("  Host: %s\n", a.Host))
        buf.WriteString(fmt.Sprintf("  Time: %s\n", a.Timestamp.Format(time.RFC3339)))
        buf.WriteString("\n")
    }

    buf.WriteString(strings.Repeat("=", 50))
    buf.WriteString("\n")
    buf.WriteString("Generated by tisminSRETool\n")

    return buf.String()
}
```

### 5.4 修复原代码问题

修复 sender.go 中的问题：

```go
package alert

import (
    "fmt"
    "net/smtp"
    "strings"
    "tisminSRETool/internal/model"
)

// 注意：删除重复的 package 声明和 import 语句

func SendEmail(subject, body string, config model.EmailConfig) error {
    auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
    // 修复拼写错误：Sumject -> Subject
    msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", strings.Join(config.To, ","), subject, body))
    addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
    return smtp.SendMail(addr, auth, config.From, config.To, msg)
}
```

---

## 六、高级功能

### 6.1 告警抑制（避免重复告警）

```go
type AlertSuppression struct {
    mu           sync.Mutex
    lastAlert    map[string]time.Time  // key: "category:metric:host"
    suppressDuration time.Duration      // 抑制时长
}

func NewAlertSuppression(duration time.Duration) *AlertSuppression {
    return &AlertSuppression{
        lastAlert: make(map[string]time.Time),
        suppressDuration: duration,
    }
}

func (a *AlertSuppression) ShouldSend(alert Alert) bool {
    key := fmt.Sprintf("%s:%s:%s", alert.Category, alert.Metric, alert.Host)

    a.mu.Lock()
    defer a.mu.Unlock()

    if lastTime, ok := a.lastAlert[key]; ok {
        if time.Since(lastTime) < a.suppressDuration {
            return false
        }
    }

    a.lastAlert[key] = time.Now()
    return true
}
```

### 6.2 告警恢复通知

```go
type RecoveryTracker struct {
    mu           sync.Mutex
    alertStatus  map[string]bool  // key: 是否处于告警状态
}

func (r *RecoveryTracker) CheckRecovery(alert Alert) (bool, bool) {
    // 返回 (是否恢复, 是否需要发送恢复通知)
    key := fmt.Sprintf("%s:%s:%s", alert.Category, alert.Metric, alert.Host)

    r.mu.Lock()
    defer r.mu.Unlock()

    wasAlerting := r.alertStatus[key]
    r.alertStatus[key] = true  // 设置为告警状态

    if wasAlerting {
        // 之前是告警状态，现在没有新的告警，说明恢复了
        r.alertStatus[key] = false
        return true, true
    }

    return false, false
}
```

### 6.3 AlertManager 完整实现

```go
type AlertManagerImpl struct {
    checker     AlertChecker
    sender      AlertSender
    suppression *AlertSuppression
    recovery    *RecoveryTracker
    config      model.AlertConfig
    emailConfig model.EmailConfig
    stopCh      chan struct{}
}

func NewAlertManager(checker AlertChecker, sender AlertSender, alertConfig model.AlertConfig, emailConfig model.EmailConfig) *AlertManagerImpl {
    return &AlertManagerImpl{
        checker:     checker,
        sender:      sender,
        suppression: NewAlertSuppression(5 * time.Minute),
        recovery:    &RecoveryTracker{alertStatus: make(map[string]bool)},
        config:      alertConfig,
        emailConfig: emailConfig,
        stopCh:      make(chan struct{}),
    }
}

func (m *AlertManagerImpl) Run(ctx context.Context, metricsCh <-chan model.Metrics) {
    ticker := time.NewTicker(m.config.CheckInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-m.stopCh:
            return
        case <-ticker.C:
            select {
            case metrics := <-metricsCh:
                m.processMetrics(ctx, metrics)
            default:
                // 没有数据，继续
            }
        }
    }
}

func (m *AlertManagerImpl) processMetrics(ctx context.Context, metrics model.Metrics) {
    alerts, err := m.checker.Check(ctx, metrics)
    if err != nil {
        log.Printf("Alert check error: %v", err)
        return
    }

    // 过滤需要发送的告警
    var toSend []Alert
    for _, alert := range alerts {
        if m.suppression.ShouldSend(alert) {
            toSend = append(toSend, alert)
        }
    }

    if len(toSend) > 0 {
        if err := m.sender.Send(ctx, toSend, m.emailConfig); err != nil {
            log.Printf("Alert send error: %v", err)
        }
    }
}

func (m *AlertManagerImpl) Stop() {
    close(m.stopCh)
}
```

---

## 七、配置文件对应关系

### 7.1 AlertConfig 字段与告警对应

| 配置文件字段 | 对应检查方法 | 状态 |
|-------------|-------------|------|
| `CPUThreshold` | `checkCPU()` | ✅ 已实现 |
| `MemoryThreshold` | `checkMemory()` | ✅ 已实现 |
| `DiskThreshold` | `checkDisk()` | ✅ 已实现 |
| `InodesThreshold` | `checkInodes()` | ❌ 当前为空 |
| `NetworkBandwidthThreshold` | - | ⚠️ 需要扩展 metrics |
| `NetworkPacketLossThreshold` | - | ⚠️ 需要扩展 metrics |
| `NetworkRTTThreshold` | - | ⚠️ 需要扩展 metrics |
| `TCPTimeWaitThreshold` | `checkTCP()` | ❌ 当前为空 |
| `TCPCLOSEWaitThreshold` | `checkTCP()` | ❌ 当前为空 |
| `TotalTCPThreshold` | `checkTCP()` | ❌ 当前为空 |

### 7.2 需要的配置扩展

在 `model/config.go` 中添加：

```go
type AlertConfig struct {
    Enabled         bool          `mapstructure:"enabled"`
    CheckInterval   time.Duration `mapstructure:"check_interval"`  // 告警检查间隔
    // ... 现有字段
}
```

---

## 八、实现优先级建议

| 优先级 | 功能 | 说明 |
|--------|------|------|
| P0 | 修复现有 Bug | 修复 sender.go 重复代码、rules.go 空循环 |
| P1 | 完成 Inodes 检查 | 实现 `checkInodes()` |
| P2 | 完成 TCP 连接检查 | 扩展 metrics 添加 TCP 统计 |
| P3 | 告警抑制 | 避免重复告警 |
| P4 | 告警恢复通知 | 指标恢复正常时通知 |
| P5 | 扩展网络告警 | 支持带宽、RTT 等指标 |

---

## 九、文件修改清单

| 文件 | 修改内容 |
|------|---------|
| `internal/alert/interface.go` | 新增接口定义 |
| `internal/alert/rules.go` | 完成所有检查方法，修复变量命名 |
| `internal/alert/sender.go` | 修复重复代码和拼写错误，增加重试机制 |
| `internal/model/metrics.go` | 添加 TCPStat 结构体（可选） |
| `internal/model/config.go` | 添加 CheckInterval 字段（可选） |
