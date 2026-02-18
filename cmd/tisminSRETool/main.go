package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sretool-fullstack/internal/alert"
	"sretool-fullstack/internal/collector"
	"sretool-fullstack/internal/config"
	"sretool-fullstack/internal/scheduler"
	"sretool-fullstack/internal/server"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("c", "configs/config.yaml", "Path to config file")
	showVersion := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("sretool-fullstack version %s, built at %s\n", version, buildTime)
		os.Exit(0)
	}

	fmt.Printf("sretool-fullstack v%s starting...\n", version)

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 获取 Viper 实例用于运行时配置访问
	vip, err := config.GetViper(*configPath)
	if err != nil {
		log.Fatalf("Failed to get Viper: %v", err)
	}

	// 创建采集器
	linuxCollector, err := collector.NewLinuxCollector()
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	// 创建告警检查器
	var ruleChecker *alert.RuleChecker
	if cfg.Alert.Enabled {
		ruleChecker = alert.NewRuleChecker(cfg.Alert)
	}

	// 创建告警发送器
	var emailSender alert.AlertSender
	if cfg.Alert.Enabled && cfg.Email.Host != "" {
		emailSender = alert.NewEmailSender()
	}

	// 创建调度器
	sched := scheduler.New(linuxCollector, ruleChecker, cfg.Email, cfg.App.RefreshInterval)
	if emailSender != nil {
		sched.SetAlertSender(emailSender)
	}

	// 创建 HTTP 服务器
	srv := server.New(vip, nil)

	// 创建根上下文用于优雅退出
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 启动调度器
	sched.Start(rootCtx)

	// 启动 HTTP 服务器
	if err := srv.Start(rootCtx); err != nil {
		log.Printf("Server start error: %v", err)
	}

	// 定期更新服务器指标
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-rootCtx.Done():
				return
			case <-ticker.C:
				metrics := sched.GetLatestMetrics()
				srv.SetMetrics(metrics)
			}
		}
	}()

	// 等待退出信号
	<-rootCtx.Done()

	// 优雅关闭
	log.Println("Shutting down...")

	// 停止调度器
	sched.Stop()

	// 停止 HTTP 服务器
	if err := srv.Stop(context.Background()); err != nil {
		log.Printf("Server stop error: %v", err)
	}

	log.Println("sretool-fullstack stopped")
}
