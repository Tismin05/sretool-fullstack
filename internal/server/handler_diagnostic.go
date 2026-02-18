package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleDiagnostic 处理诊断请求
func (s *Server) handleDiagnostic(c *gin.Context) {
	// 使用采集器创建诊断
	// 这里简化处理，直接返回当前指标的健康状态
	metrics := s.getLatestMetrics()
	if metrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "metrics not available",
		})
		return
	}

	// 简单的健康状态判断
	health := gin.H{
		"cpu":    "healthy",
		"memory": "healthy",
		"disk":   "healthy",
		"network": "healthy",
	}

	// CPU 检查
	if metrics.CPU.UsagePercent > 90 {
		health["cpu"] = "critical"
	} else if metrics.CPU.UsagePercent > 70 {
		health["cpu"] = "warning"
	}

	// 内存检查
	if metrics.Mem.UsedPercent > 90 {
		health["memory"] = "critical"
	} else if metrics.Mem.UsedPercent > 70 {
		health["memory"] = "warning"
	}

	// 磁盘检查
	for _, disk := range metrics.Disk {
		if disk.UsedPercent > 90 {
			health["disk"] = "critical"
			break
		} else if disk.UsedPercent > 80 {
			health["disk"] = "warning"
		}
	}

	// 网络检查
	for _, net := range metrics.Net {
		totalPackets := net.RxPackets + net.TxPackets
		if totalPackets > 0 {
			errorRate := float64(net.RxErrors+net.TxErrors) / float64(totalPackets) * 100
			if errorRate > 1 {
				health["network"] = "critical"
				break
			} else if errorRate > 0.1 {
				health["network"] = "warning"
			}
		}
	}

	// 总体状态
	status := "healthy"
	for _, v := range health {
		if v == "critical" {
			status = "critical"
			break
		}
		if v == "warning" && status == "healthy" {
			status = "warning"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"timestamp": metrics.UpdateTimestamp,
		"checks":    health,
	})
}
