package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"tisminSRETool/internal/model"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}
)

// WebSocketHub WebSocket hub 管理所有连接
type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast   chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

// NewWebSocketHub 创建新的WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Start 启动hub
func (h *WebSocketHub) Start() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Println("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息
func (h *WebSocketHub) Broadcast(metrics *model.Metrics) {
	data, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Failed to marshal metrics: %v", err)
		return
	}
	h.broadcast <- data
}

// GetHub 获取WebSocket hub（需要在server中初始化）
func (s *Server) GetHub() *WebSocketHub {
	return s.wsHub
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.wsHub.register <- conn

	// 发送当前最新数据
	metrics := s.getLatestMetrics()
	if metrics != nil {
		data, _ := json.Marshal(metrics)
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// 保持连接
	go func() {
		defer func() {
			s.wsHub.unregister <- conn
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// StartWebSocketHub 启动WebSocket hub
func (s *Server) StartWebSocketHub() {
	if s.wsHub == nil {
		s.wsHub = NewWebSocketHub()
	}
	go s.wsHub.Start()
}

// BroadcastMetrics 广播指标数据
func (s *Server) BroadcastMetrics(metrics *model.Metrics) {
	if s.wsHub != nil {
		s.wsHub.Broadcast(metrics)
	}
}

// WaitGroup 用于优雅关闭
func (h *WebSocketHub) WaitGroup() {
	// WebSocket hub不需要waitgroup
	time.Sleep(100 * time.Millisecond)
}
