# sretool-fullstack

A production-ready Go-based SRE monitoring tool for Linux servers. Collects system metrics (CPU, Memory, Disk, Network), performs system diagnostics, and sends alert notifications via multiple channels.

## Features

- **Linux /proc Based**: Direct collection from `/proc` filesystem, no external dependencies
- **Comprehensive Metrics**: CPU, Memory, Disk, Network, Process information with real-time rate calculation
- **RESTful API**: Built-in HTTP server with Gin framework
- **WebSocket Support**: Real-time metrics push to frontend
- **Beautiful Web UI**: Vue 3 + Element Plus + ECharts dashboard
- **Multi-channel Alerts**: Email, Webhook, DingTalk support
- **InfluxDB Integration**: Historical data storage and trend analysis
- **Graceful Shutdown**: Full Context support for timeout and cancellation

## Quick Start

### 1. Build

```bash
# Clone the project
git clone https://github.com/your-repo/sretool-fullstack.git
cd sretool-fullstack

# Build the application
go build -o sretool-fullstack ./cmd/sretool-fullstack
```

### 2. Configure

Edit `configs/config.yaml`:

```yaml
app:
  name: "sretool-fullstack"
  version: "1.0.0"
  refresh_interval: "10s"
  log_level: "info"
  port: 8080

# Enable alerts (optional)
alert:
  enabled: true
  cpu_threshold: 80.0
  memory_threshold: 80.0
  disk_threshold: 85.0

# Email alerts (optional)
email:
  host: "smtp.example.com"
  port: 587
  username: "your-email@example.com"
  password: "your-password"
  from: "alerts@example.com"
  to:
    - "admin@example.com"

# InfluxDB storage (optional)
influxdb:
  enabled: false
  url: "http://localhost:8086"
  token: "your-influxdb-token"
  org: "tismin"
  bucket: "sretool-fullstack"
```

### 3. Run

```bash
# Run with config file
./sretool-fullstack -c configs/config.yaml

# Or run with default config
./sretool-fullstack
```

### 4. Access Web UI

Open browser: `http://your-server:8080`

## Deployment

### Linux Server (Production)

#### 1. Download & Build

```bash
# Install Go 1.25+
wget https://go.dev/dl/go1.25.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/your-repo/sretool-fullstack.git
cd sretool-fullstack
go build -o sretool-fullstack ./cmd/sretool-fullstack
```

#### 2. Systemd Service

Create `/etc/systemd/system/tisminsretool.service`:

```ini
[Unit]
Description=sretool-fullstack - SRE Monitoring Tool
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/sretool-fullstack
ExecStart=/opt/sretool-fullstack/sretool-fullstack -c /opt/sretool-fullstack/configs/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Install and start
sudo cp sretool-fullstack /opt/sretool-fullstack/
sudo cp -r configs /opt/sretool-fullstack/
sudo systemctl daemon-reload
sudo systemctl enable tisminsretool
sudo systemctl start tisminsretool

# Check status
sudo systemctl status tisminsretool
```

#### 3. Firewall

```bash
# Allow port 8080
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# Or using ufw
sudo ufw allow 8080/tcp
```

### Docker Deployment (Recommended)

#### 1. Create Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o sretool-fullstack ./cmd/sretool-fullstack

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/sretool-fullstack .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./sretool-fullstack", "-c", "configs/config.yaml"]
```

#### 2. Create docker-compose.yml

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

  # Optional: InfluxDB
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
  #     - DOCKER_INFLUXDB_INIT_BUCKET=sretool-fullstack
  #     - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=my-super-secret-admin-token

# volumes:
#   influxdb-data:
```

#### 3. Start

```bash
docker-compose up -d
```

### Reverse Proxy (Nginx)

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

    # WebSocket support
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/metrics` | GET | Get current metrics |
| `/api/v1/health` | GET | Health check |
| `/api/v1/diagnostic` | GET | System diagnostic |
| `/api/v1/config` | GET | Get configuration |
| `/api/v1/config/reload` | PUT | Reload configuration |
| `/api/v1/alerts` | GET | Get alerts |
| `/ws` | WebSocket | Real-time metrics |

## Project Structure

```
sretool-fullstack/
├── cmd/sretool-fullstack/     # Application entry
├── configs/               # Configuration files
├── internal/
│   ├── alert/            # Alert module
│   ├── collector/        # Metrics collection
│   ├── config/           # Configuration loader
│   ├── diagnostic/       # System diagnostics
│   ├── engine/           # Rate calculation
│   ├── model/            # Data models
│   ├── scheduler/        # Task scheduler
│   ├── server/           # HTTP server
│   └── storage/          # InfluxDB storage
├── web/                  # Vue 3 frontend
├── pkg/                  # Utility packages
└── docs/                # Documentation
```

## Technology Stack

- **Backend**: Go 1.25+, Gin, Viper, zap
- **Frontend**: Vue 3, TypeScript, Element Plus, ECharts
- **Database**: InfluxDB (optional)
- **Monitoring**: Prometheus compatible

## Configuration Options

### Alert Thresholds

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cpu_threshold` | CPU usage % | 80% |
| `memory_threshold` | Memory usage % | 80% |
| `disk_threshold` | Disk usage % | 85% |
| `disk_await_threshold` | Disk await ms | 50ms |
| `inodes_threshold` | Inodes usage % | 80% |

### Environment Variables

```bash
# Application
APP_NAME=sretool-fullstack
APP_LOG_LEVEL=info
APP_PORT=8080

# Alerts
ALERT_ENABLED=true
ALERT_CPU_THRESHOLD=80

# SMTP
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-email
SMTP_PASSWORD=your-password
```

## Building Frontend

```bash
cd web
npm install
npm run build
```

The built frontend will be served from `/static` path.

## Development

```bash
# Run backend
go run ./cmd/sretool-fullstack

# Run frontend dev server
cd web
npm run dev
```

## License

MIT

---

[中文版](./README_ZH.md)
