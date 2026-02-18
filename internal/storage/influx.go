package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"tisminSRETool/internal/model"
)

// InfluxStorage InfluxDB 存储
type InfluxStorage struct {
	client       influxdb2.Client
	writeAPI     api.WriteAPI
	queryAPI     api.QueryAPI
	bucket       string
	org          string
}

// InfluxConfig InfluxDB 配置
type InfluxConfig struct {
	URL    string // http://localhost:8086
	Token  string
	Org    string
	Bucket string
}

// NewInfluxStorage 创建 InfluxDB 存储
func NewInfluxStorage(cfg InfluxConfig) (*InfluxStorage, error) {
	if cfg.URL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("InfluxDB config is incomplete")
	}

	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)
	queryAPI := client.QueryAPI(cfg.Org)

	// 测试连接
	ctx := context.Background()
	_, err := client.Ready(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	return &InfluxStorage{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		bucket:   cfg.Bucket,
		org:      cfg.Org,
	}, nil
}

// Write 写入指标数据
func (s *InfluxStorage) Write(ctx context.Context, m *model.Metrics) error {
	timestamp := time.Now()

	// 写入 CPU 数据
	point := influxdb2.NewPoint(
		"cpu",
		map[string]string{"host": m.Host},
		map[string]interface{}{
			"usage_percent": m.CPU.UsagePercent,
			"load1":         m.CPU.Load1,
			"load5":         m.CPU.Load5,
			"load15":        m.CPU.Load15,
		},
		timestamp,
	)
	s.writeAPI.WritePoint(point)

	// 写入内存数据
	point = influxdb2.NewPoint(
		"memory",
		map[string]string{"host": m.Host},
		map[string]interface{}{
			"total":           m.Mem.Total,
			"used":            m.Mem.Used,
			"free":            m.Mem.Free,
			"available":       m.Mem.Available,
			"used_percent":    m.Mem.UsedPercent,
			"swap_used":       m.Mem.SwapUsed,
			"swap_used_percent": m.Mem.SwapUsedPercent,
		},
		timestamp,
	)
	s.writeAPI.WritePoint(point)

	// 写入磁盘数据
	for _, disk := range m.Disk {
		point = influxdb2.NewPoint(
			"disk",
			map[string]string{
				"host":       m.Host,
				"mountpoint": disk.MountPoint,
			},
			map[string]interface{}{
				"total":             disk.Total,
				"used":              disk.Used,
				"free":              disk.Free,
				"used_percent":      disk.UsedPercent,
				"inodes_used":       disk.InodesUsed,
				"inodes_used_percent": disk.InodesUsedPercent,
			},
			timestamp,
		)
		s.writeAPI.WritePoint(point)
	}

	// 写入网络数据
	for _, net := range m.Net {
		point = influxdb2.NewPoint(
			"net",
			map[string]string{
				"host": m.Host,
				"name": net.Name,
			},
			map[string]interface{}{
				"rx_bytes":   net.RxBytes,
				"rx_packets": net.RxPackets,
				"rx_errors":  net.RxErrors,
				"tx_bytes":   net.TxBytes,
				"tx_packets": net.TxPackets,
				"tx_errors":  net.TxErrors,
			},
			timestamp,
		)
		s.writeAPI.WritePoint(point)
	}

	// 刷新缓冲区
	s.writeAPI.Flush()

	return nil
}

// QueryHistory 查询历史数据
func (s *InfluxStorage) QueryHistory(ctx context.Context, metric, host string, start, stop time.Duration) ([]map[string]interface{}, error) {
	now := time.Now()
	startTime := now.Add(-start)
	stopTime := now.Add(-stop)

	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "%s")
		|> filter(fn: (r) => r.host == "%s")
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")`,
		s.bucket, startTime.Format(time.RFC3339), stopTime.Format(time.RFC3339), metric, host)

	result, err := s.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var results []map[string]interface{}
	for result.Next() {
		record := result.Record()
		row := map[string]interface{}{
			"time":   record.Time().Unix(),
			"values": record.Values(),
		}
		results = append(results, row)
	}

	return results, nil
}

// Close 关闭连接
func (s *InfluxStorage) Close() {
	s.writeAPI.Flush()
	s.client.Close()
}

// Enabled 检查是否启用
func (s *InfluxStorage) Enabled() bool {
	return s != nil && s.client != nil
}
