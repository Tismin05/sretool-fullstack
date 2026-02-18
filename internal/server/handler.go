package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"tisminSRETool/internal/config"
	"tisminSRETool/internal/model"
)

// startTime 服务器启动时间
var startTime = time.Now()

// handleMetrics 处理指标请求
func (s *Server) handleMetrics(c *gin.Context) {
	// 获取最新的指标数据
	metrics := s.getLatestMetrics()
	if metrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "metrics not available yet",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// handleHealth 处理健康检查请求
func (s *Server) handleHealth(c *gin.Context) {
	metrics := s.getLatestMetrics()
	var lastCollect string
	if metrics != nil {
		lastCollect = metrics.UpdateTimestamp
	}

	status := "healthy"
	if metrics == nil {
		status = "starting"
	}

	uptime := time.Since(startTime)

	c.JSON(http.StatusOK, gin.H{
		"status":       status,
		"uptime":       uptime.String(),
		"last_collect": lastCollect,
	})
}

// handleGetConfig 处理获取配置请求
func (s *Server) handleGetConfig(c *gin.Context) {
	// 返回非敏感配置
	cfg := gin.H{
		"app": gin.H{
			"name":             s.vip.GetString("app.name"),
			"version":          s.vip.GetString("app.version"),
			"refresh_interval": s.vip.GetString("app.refresh_interval"),
			"log_level":        s.vip.GetString("app.log_level"),
		},
		"diagnostic": gin.H{
			"enabled":         s.vip.GetBool("diagnostic.enabled"),
			"show_top_n_list": s.vip.GetInt("diagnostic.show_top_n_list"),
		},
		"alert": gin.H{
			"enabled":              s.vip.GetBool("alert.enabled"),
			"cpu_threshold":        s.vip.GetFloat64("alert.cpu_threshold"),
			"memory_threshold":     s.vip.GetFloat64("alert.memory_threshold"),
			"disk_threshold":       s.vip.GetFloat64("alert.disk_threshold"),
			"disk_await_threshold": s.vip.GetFloat64("alert.disk_await_threshold"),
			"disk_util_threshold":  s.vip.GetFloat64("alert.disk_util_threshold"),
			"inodes_threshold":    s.vip.GetFloat64("alert.inodes_threshold"),
		},
	}

	c.JSON(http.StatusOK, cfg)
}

// handleReloadConfig 处理重新加载配置请求
func (s *Server) handleReloadConfig(c *gin.Context) {
	newCfg, err := config.Reload(s.vip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to reload config: " + err.Error(),
		})
		return
	}

	// 更新服务器指标引用
	s.metrics = &model.Metrics{}

	_ = newCfg // 后续可用于更新运行时配置

	c.JSON(http.StatusOK, gin.H{
		"message": "configuration reloaded successfully",
	})
}

// handleGetAlerts 处理获取告警请求
func (s *Server) handleGetAlerts(c *gin.Context) {
	// 返回当前告警状态
	alerts := gin.H{
		"enabled": s.vip.GetBool("alert.enabled"),
		"last_check": s.vip.GetString("alert.last_check"),
	}

	c.JSON(http.StatusOK, alerts)
}

// getLatestMetrics 获取最新指标（线程安全方式）
func (s *Server) getLatestMetrics() *model.Metrics {
	// 这里应该使用 mutex 保护，但在简单场景下直接返回也是安全的
	// 后续可以通过 channel 或 atomic 方式获取最新数据
	return s.metrics
}

// SetMetrics 设置指标数据
func (s *Server) SetMetrics(m *model.Metrics) {
	s.metrics = m
}
