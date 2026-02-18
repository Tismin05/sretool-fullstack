package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"sretool-fullstack/internal/model"
)

// Server HTTP 服务器
type Server struct {
	httpServer *http.Server
	engine     *gin.Engine
	vip        *viper.Viper
	metrics    *model.Metrics
	wsHub      *WebSocketHub
}

// New 创建新的服务器
func New(vip *viper.Viper, metrics *model.Metrics) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &Server{
		engine:  engine,
		vip:     vip,
		metrics: metrics,
	}

	s.setupRoutes()

	addr := fmt.Sprintf(":%d", vip.GetInt("app.port"))
	if addr == ":0" {
		addr = ":8080"
	}

	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 全局中间件
	s.engine.Use(gin.Logger())
	s.engine.Use(gin.Recovery())

	// 健康检查
	s.engine.GET("/api/v1/health", s.handleHealth)

	// API 路由组
	api := s.engine.Group("/api/v1")
	{
		// 指标 API
		api.GET("/metrics", s.handleMetrics)

		// 配置 API
		api.GET("/config", s.handleGetConfig)
		api.PUT("/config/reload", s.handleReloadConfig)

		// 告警 API
		api.GET("/alerts", s.handleGetAlerts)

		// 诊断 API
		api.GET("/diagnostic", s.handleDiagnostic)
	}

	// WebSocket 路由
	s.engine.GET("/ws", s.handleWebSocket)

	// 静态文件服务（前端）
	s.engine.Static("/static", "./web/dist/static")
	s.engine.LoadHTMLGlob("./web/dist/index.html")
	s.engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	go func() {
		addr := s.httpServer.Addr
		fmt.Printf("Starting HTTP server on %s\n", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server...")

	// 给服务器一些时间来完成正在处理的请求
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	fmt.Println("HTTP server stopped")
	return nil
}

// Run 运行服务器直到收到退出信号
func (s *Server) Run(ctx context.Context) error {
	// 启动服务器
	if err := s.Start(ctx); err != nil {
		return err
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
		fmt.Println("Received shutdown signal")
	case <-ctx.Done():
		fmt.Println("Context cancelled")
	}

	return s.Stop(context.Background())
}

// GetRouter 获取 gin 引擎（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.engine
}
