# tisminSRETool

一个生产级 Go 编写的 Linux 服务器 SRE 监控工具。采集系统指标（CPU、内存、磁盘、网络），执行系统诊断，并通过多渠道发送告警通知。

## 功能特性

- **Linux /proc 采集**：直接从 `/proc` 文件系统采集，无需外部依赖
- **全面指标采集**：CPU、内存、磁盘、网络、进程信息，支持实时速率计算
- **RESTful API**：基于 Gin 框架内置 HTTP 服务器
- **WebSocket 支持**：实时推送指标到前端
- **精美 Web UI**：Vue 3 + Element Plus + ECharts 仪表盘
- **多渠道告警**：支持邮件、Webhook、钉钉告警
- **InfluxDB 集成**：历史数据存储和趋势分析
- **优雅退出**：完整的 Context 支持超时和取消控制

## 快速开始

### 1. 构建

```bash
# 克隆项目
git clone https://github.com/your-repo/tisminSRETool.git
cd tisminSRETool

# 构建应用
go build -o tisminSRETool ./cmd/tisminSRETool
```

### 2. 配置

编辑 `configs/config.yaml`:

```yaml
app:
  name: "tisminSRETool"
  version: "1.0.0"
  refresh_interval: "10s"
  log_level: "info"
  port: 8080

# 启用告警（可选）
alert:
  enabled: true
  cpu_threshold: 80.0
  memory_threshold: 80.0
  disk_threshold: 85.0

# 邮件告警（可选）
email:
  host: "smtp.example.com"
  port: 587
  username: "your-email@example.com"
  password: "your-password"
  from: "alerts@example.com"
  to:
    - "admin@example.com"

# InfluxDB 存储（可选）
influxdb:
  enabled: false
  url: "http://localhost:8086"
  token: "your-influxdb-token"
  org: "tismin"
  bucket: "tisminSRETool"
```

### 3. 运行

```bash
# 使用配置文件运行
./tisminSRETool -c configs/config.yaml

# 或使用默认配置运行
./tisminSRETool
```

### 4. 访问 Web UI

打开浏览器：`http://your-server:8080`

## 部署教程

### Linux 服务器（生产环境）

#### 1. 下载与构建

```bash
# 安装 Go 1.25+
wget https://go.dev/dl/go1.25.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 克隆并构建
git clone https://github.com/your-repo/tisminSRETool.git
cd tisminSRETool
go build -o tisminSRETool ./cmd/tisminSRETool
```

#### 2. Systemd 服务

创建 `/etc/systemd/system/tisminsretool.service`:

```ini
[Unit]
Description=tisminSRETool - SRE 监控工具
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/tisminSRETool
ExecStart=/opt/tisminSRETool/tisminSRETool -c /opt/tisminSRETool/configs/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# 安装并启动
sudo cp tisminSRETool /opt/tisminSRETool/
sudo cp -r configs /opt/tisminSRETool/
sudo systemctl daemon-reload
sudo systemctl enable tisminsretool
sudo systemctl start tisminsretool

# 查看状态
sudo systemctl status tisminsretool
```

#### 3. 防火墙配置

```bash
# 开放 8080 端口
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# 或使用 ufw
sudo ufw allow 8080/tcp
```

### Docker 部署（推荐）

#### 1. 创建 Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o tisminSRETool ./cmd/tisminSRETool

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/tisminSRETool .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./tisminSRETool", "-c", "configs/config.yaml"]
```

#### 2. 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  tisminsretool:
    build: .
    container_name: tisminsretool
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs
    restart: unless-stopped
    network_mode: host

  # 可选：InfluxDB
  # influxdb:
  #   image: influxdb:2.7
  #   container_name: influxdb
  #   ports:
  #     - "8086:8086"
  #   volumes:
  #     - influxdb-data:/var/lib/influxdb2
  #   environment:
  #     - DOCKER_INFLUXDB_INIT_MODE=setup
  #     - DOCKER_INFLUXDB_INIT_USERNAME=admin
  #     - DOCKER_INFLUXDB_INIT_PASSWORD=adminpassword
  #     - DOCKER_INFLUXDB_INIT_ORG=tismin
  #     - DOCKER_INFLUXDB_INIT_BUCKET=tisminSRETool
  #     - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=my-super-secret-admin-token

# volumes:
#   influxdb-data:
```

#### 3. 启动

```bash
docker-compose up -d
```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket 支持
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/metrics` | GET | 获取当前指标 |
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/diagnostic` | GET | 系统诊断 |
| `/api/v1/config` | GET | 获取配置 |
| `/api/v1/config/reload` | PUT | 重新加载配置 |
| `/api/v1/alerts` | GET | 获取告警 |
| `/ws` | WebSocket | 实时指标推送 |

## 项目结构

```
tisminSRETool/
├── cmd/tisminSRETool/     # 应用入口
├── configs/               # 配置文件
├── internal/
│   ├── alert/            # 告警模块
│   ├── collector/        # 指标采集
│   ├── config/           # 配置加载
│   ├── diagnostic/       # 系统诊断
│   ├── engine/           # 速率计算
│   ├── model/            # 数据模型
│   ├── scheduler/        # 任务调度
│   ├── server/           # HTTP 服务器
│   └── storage/          # InfluxDB 存储
├── web/                  # Vue 3 前端
├── pkg/                  # 工具包
└── docs/                 # 文档
```

## 技术栈

- **后端**: Go 1.25+, Gin, Viper, zap
- **前端**: Vue 3, TypeScript, Element Plus, ECharts
- **数据库**: InfluxDB（可选）
- **监控**: Prometheus 兼容

## 配置选项

### 告警阈值

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `cpu_threshold` | CPU 使用率 % | 80% |
| `memory_threshold` | 内存使用率 % | 80% |
| `disk_threshold` | 磁盘使用率 % | 85% |
| `disk_await_threshold` | 磁盘等待时间 ms | 50ms |
| `inodes_threshold` | Inodes 使用率 % | 80% |

### 环境变量

```bash
# 应用配置
APP_NAME=tisminSRETool
APP_LOG_LEVEL=info
APP_PORT=8080

# 告警配置
ALERT_ENABLED=true
ALERT_CPU_THRESHOLD=80

# SMTP 配置
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-email
SMTP_PASSWORD=your-password
```

## 构建前端

```bash
cd web
npm install
npm run build
```

构建后的前端将托管在 `/static` 路径。

## 开发

```bash
# 运行后端
go run ./cmd/tisminSRETool

# 运行前端开发服务器
cd web
npm run dev
```

## 许可证

MIT

---

[English Version](./README.md)
