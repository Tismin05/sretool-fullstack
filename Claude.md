# Claude Code 项目指南 - sretool-fullstack

## 项目概述

sretool-fullstack 是一个用 Go 编写的 SRE (Site Reliability Engineering) 监控工具，用于采集系统指标（CPU、内存、磁盘、网络）、执行系统诊断和发送告警通知。

## 项目结构

```
sretool-fullstack/
├── cmd/sretool-fullstack/          # 程序入口
│   ├── main.go                 # 主程序入口（当前为空）
│   └── debug.go                # 调试/测试入口
├── configs/
│   └── config.yaml             # 配置文件
├── pkg/utils/                  # 公共工具包
│   ├── convert.go              # 单位转换工具
│   └── linereader.go           # 文件行读取工具
├── internal/                    # 内部业务逻辑
│   ├── model/                  # 数据模型层
│   │   ├── config.go           # 配置结构体
│   │   ├── metrics.go          # 系统指标结构体
│   │   ├── diagnostic.go       # 诊断结果结构体
│   │   └── collector_error.go  # 采集错误结构体
│   ├── collector/              # 指标采集器模块
│   │   ├── interface.go        # Collector 接口定义
│   │   ├── local_MacOS.go      # macOS 采集实现
│   │   ├── local_Linux.go      # Linux 采集实现（gopsutil）
│   │   └── linux_proc.go       # Linux 采集实现（/proc）
│   ├── diagnostic/             # 诊断模块
│   │   ├── interface.go        # Diagnostic 接口
│   │   └── diagnostic_Linux.go # Linux 诊断实现
│   ├── alert/                  # 告警模块
│   │   ├── interface.go        # Alert 接口
│   │   ├── rules.go            # 告警规则检查器
│   │   └── sender.go           # 邮件发送实现
│   └── engine/                 # 调度引擎模块
│       └── calculater.go       # 指标计算（速率计算）
├── go.mod
├── go.sum
└── README.md
```

## 核心接口定义

### 1. Collector 接口 (`internal/collector/interface.go`)

```go
type Collector interface {
    Collect(ctx context.Context) (*model.Metrics, *model.CollectErrors)
}
```

- 实现文件：
  - `local_MacOS.go`: MacOSCollector - 使用 gopsutil 库采集 macOS 指标
  - `local_Linux.go`: LinuxCollector - 使用 gopsutil 库采集 Linux 指标
  - `linux_proc.go`: 使用 `// +build proc_refactor` 标签，直接读取 /proc 文件系统

### 2. 指标数据结构 (`internal/model/metrics.go`)

```go
type Metrics struct {
    CPU             CPUStat    // CPU 统计
    Mem             MemoryStat // 内存统计
    Disk            []DiskStat // 磁盘统计（多个挂载点）
    Net             []NetStat  // 网络统计（多个网卡）
    Procs           []ProcStat // 进程信息
    Host            string     // 主机名
    UpdateTimestamp string     // 更新时间戳
}
```

采集的指标包括：
- **CPU**: 核心数、使用率、每核使用率、1/5/15分钟负载
- **内存**: 总/空闲/可用/已用内存、Swap信息
- **磁盘**: 挂载点、总容量、已用容量、Inodes信息、IO读写统计
- **网络**: 网卡名、接收/发送字节数、数据包数、错误数、丢包数

### 3. 配置结构 (`internal/model/config.go`)

```go
type Config struct {
    App         Appconfig       // 应用配置
    Diagnostic  DiagnosticConfig // 诊断配置
    Alert       AlertConfig     // 告警配置
}
```

- AlertConfig 包含 CPU、内存、磁盘、网络、Inodes 的阈值配置
- EmailConfig 用于 SMTP 邮件发送

### 4. 告警规则检查 (`internal/alert/rules.go`)

```go
type RuleChecker struct {
    config model.AlertConfig
}

func (r *RuleChecker) Check(ctx context.Context, m model.Metrics) (error, []error)
```

### 5. 速率计算 (`internal/engine/calculater.go`)

```go
func CalculateRate(prev, cur model.Metrics, interval time.Duration) model.Metrics
```

计算两次采集之间的速率变化（CPU使用率增量、磁盘读写速度、网络传输速度）。

## 开发约束与注意事项

### 1. Context 传递规范

- 所有采集操作必须支持 `context.Context`
- Context 用于控制优雅退出和超时取消
- I/O 操作（如文件读取）使用 `ReadLinesOffsetNWithContext` 检查 `ctx.Err()`

### 2. 错误处理规范

- 使用 `CollectErrors` 结构体按子系统（CPU/Mem/Disk/Net）聚合错误
- 每个子系统有独立的错误列表：`CPU`, `Mem`, `Disk`, `Net`
- 使用 `HasError()` 方法检查是否有错误

### 3. 并发处理

- 采集模块使用 goroutine 并行采集各子系统指标
- 使用 `sync.WaitGroup` 等待所有采集完成
- 使用 `sync.Mutex` 保护共享数据

### 4. 构建标签

- `// +build darwin` - macOS 平台
- `// +build linux` - Linux 平台
- `// +build proc_refactor` - 直接读取 /proc 的实验性实现

### 5. 依赖管理

- 使用 `github.com/shirou/gopsutil/v3` 进行跨平台系统指标采集
- Go 版本要求：1.25.6

## 重要文件清单

| 文件路径 | 作用 |
|---------|------|
| `cmd/sretool-fullstack/debug.go` | 调试入口，包含完整的采集测试代码 |
| `internal/collector/interface.go` | Collector 接口定义 |
| `internal/collector/local_MacOS.go` | macOS 采集实现（主要参考） |
| `internal/model/metrics.go` | 指标数据结构定义 |
| `internal/model/config.go` | 配置数据结构定义 |
| `internal/alert/rules.go` | 告警规则检查实现 |
| `internal/alert/sender.go` | SMTP 邮件发送实现 |
| `internal/engine/calculater.go` | 指标速率计算实现 |
| `pkg/utils/convert.go` | 单位转换工具函数 |
| `configs/config.yaml` | 配置文件模板 |

## 扩展开发指南

### 添加新的采集指标

1. 在 `internal/model/metrics.go` 中添加新的结构体字段
2. 在 `internal/collector/local_MacOS.go` 的 `collectViaLib` 方法中添加对应的采集逻辑
3. 使用 goroutine 并行采集，并使用 mutex 保护共享数据

### 添加新的告警规则

1. 在 `internal/model/config.go` 的 `AlertConfig` 中添加阈值字段
2. 在 `internal/alert/rules.go` 的 `Check` 方法中添加对应的检查逻辑
3. 在 `configs/config.yaml` 中添加对应的配置项

### 添加新的平台支持

1. 创建 `internal/collector/local_<Platform>.go` 文件
2. 实现 `Collector` 接口
3. 添加对应的构建标签 `// +build <platform>`

## 当前实现状态

- ✅ Collector 接口定义完整
- ✅ macOS 采集器实现完整
- ✅ Linux 采集器（gopsutil）实现完整
- ✅ Linux 采集器（/proc）存在但未完成
- ⚠️ 告警模块部分完成（rules.go 有编译错误，sender.go 有重复代码）
- ⚠️ 诊断模块接口存在但未实现
- ⚠️ 主程序入口 main.go 为空

## 代码规范

- 使用 Go 标准包结构和命名约定
- 字段使用 JSON 标签以便序列化
- 配置使用 mapstructure 标签支持 Viper 加载
- 错误处理使用自定义错误类型和聚合结构
