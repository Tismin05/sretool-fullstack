package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"tisminSRETool/internal/alert"
	"tisminSRETool/internal/collector"
	"tisminSRETool/internal/engine"
	"tisminSRETool/internal/model"
)

// Scheduler 调度器
type Scheduler struct {
	collector    collector.Collector
	ruleChecker  *alert.RuleChecker
	emailConfig  model.EmailConfig
	interval     time.Duration
	metricsCh    chan *model.Metrics
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	latestMetric *model.Metrics
	alertSender  alert.AlertSender
}

// New 创建调度器
func New(c collector.Collector, ruleChecker *alert.RuleChecker, emailConfig model.EmailConfig, interval time.Duration) *Scheduler {
	return &Scheduler{
		collector:    c,
		ruleChecker:  ruleChecker,
		emailConfig:   emailConfig,
		interval:      interval,
		metricsCh:     make(chan *model.Metrics, 10),
		stopCh:        make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("Starting scheduler with interval: %s", s.interval)
	s.wg.Add(1)
	go s.run(ctx)
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	close(s.stopCh)
	s.wg.Wait()
	close(s.metricsCh)
	log.Println("Scheduler stopped")
}

// MetricsCh 返回指标通道
func (s *Scheduler) MetricsCh() <-chan *model.Metrics {
	return s.metricsCh
}

// GetLatestMetrics 获取最新指标
func (s *Scheduler) GetLatestMetrics() *model.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestMetric
}

// run 运行调度循环
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	// 立即执行一次采集
	s.collectAndProcess(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler context cancelled")
			return
		case <-s.stopCh:
			log.Println("Scheduler stopped signal received")
			return
		case <-ticker.C:
			s.collectAndProcess(ctx)
		}
	}
}

// collectAndProcess 执行采集和处理
func (s *Scheduler) collectAndProcess(ctx context.Context) {
	// 创建带超时的上下文
	collectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 采集指标
	metrics, errs := s.collector.Collect(collectCtx)
	if errs != nil && errs.HasError() {
		log.Printf("Collection errors: %v", errs)
	}

	if metrics == nil {
		log.Println("Failed to collect metrics")
		return
	}

	// 获取前一次指标用于速率计算
	prevMetrics := s.GetLatestMetrics()
	if prevMetrics != nil {
		// 计算速率
		ratedMetrics := engine.CalculateRate(*prevMetrics, *metrics, s.interval)
		metrics = &ratedMetrics
	}

	// 更新最新指标
	s.mu.Lock()
	s.latestMetric = metrics
	s.mu.Unlock()

	// 发送到通道
	select {
	case s.metricsCh <- metrics:
	default:
		log.Println("Metrics channel full, dropping metrics")
	}

	// 告警检查
	if s.ruleChecker != nil {
		s.checkAlerts(ctx, metrics)
	}

	if errs != nil && !errs.HasError() {
		return
	}
}

// checkAlerts 检查告警规则
func (s *Scheduler) checkAlerts(ctx context.Context, metrics *model.Metrics) {
	alertList, err := s.ruleChecker.Check(ctx, *metrics)
	if err != nil && len(err) > 0 {
		log.Printf("Alert check errors: %v", err)
	}

	if len(alertList) > 0 {
		log.Printf("Generated %d alerts", len(alertList))

		// 发送告警邮件
		if s.alertSender != nil {
			if err := s.alertSender.Send(context.Background(), alertList, s.emailConfig); err != nil {
				log.Printf("Failed to send alert email: %v", err)
			}
		}
	}
}

// formatAlertEmail 格式化告警邮件内容
func formatAlertEmail(alerts []alert.Alert, host string) string {
	body := fmt.Sprintf("Host: %s\n\nAlerts:\n\n", host)
	for _, a := range alerts {
		body += fmt.Sprintf("[%s] %s: %s\n", a.Level, a.Category, a.Message)
	}
	return body
}

// SetAlertSender 设置告警发送器
func (s *Scheduler) SetAlertSender(sender alert.AlertSender) {
	s.alertSender = sender
}
