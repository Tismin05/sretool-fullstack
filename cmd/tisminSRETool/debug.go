//go:build debug
// +build debug

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sretool-fullstack/internal/collector"
)

func main() {
	fmt.Println("Starting sretool-fullstack debug mode...")

	// 创建采集器
	c, err := collector.NewLinuxCollector()
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	// 创建上下文
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	collectCtx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancel()

	// 执行采集
	startTime := time.Now()
	fmt.Printf("Collecting system metrics... %s\n", startTime)

	metrics, collectErrs := c.Collect(collectCtx)
	if collectErrs != nil && collectErrs.HasError() {
		log.Printf("Collection errors: %v", collectErrs)
	}

	// 输出结果
	output, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics: %v", err)
	}

	fmt.Println("\nCollection successful! Current system metrics:")
	fmt.Println(string(output))

	fmt.Printf("\nTimestamp: %s\n", metrics.UpdateTimestamp)
}
