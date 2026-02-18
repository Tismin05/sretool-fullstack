import axios from 'axios'
import type { Metrics, HealthStatus, AppConfig } from '@/types'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000
})

export const metricsApi = {
  // 获取当前指标
  getMetrics: () => api.get<Metrics>('/metrics'),

  // 健康检查
  getHealth: () => api.get<HealthStatus>('/health'),

  // 诊断状态
  getDiagnostic: () => api.get('/diagnostic'),

  // 获取配置
  getConfig: () => api.get<AppConfig>('/config'),

  // 重新加载配置
  reloadConfig: () => api.put('/config/reload'),

  // 获取告警
  getAlerts: () => api.get('/alerts')
}

// WebSocket 连接
export function createWebSocketConnection(
  onMessage: (data: Metrics) => void,
  onError?: (error: Event) => void,
  onClose?: () => void
): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`

  const ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('WebSocket connected')
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as Metrics
      onMessage(data)
    } catch (e) {
      console.error('Failed to parse WebSocket message:', e)
    }
  }

  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    onError?.(error)
  }

  ws.onclose = () => {
    console.log('WebSocket disconnected')
    onClose?.()
  }

  return ws
}

// 格式化字节数
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化百分比
export function formatPercent(value: number): string {
  return value.toFixed(1) + '%'
}

// 格式化时间戳
export function formatTimestamp(timestamp: string): string {
  return new Date(timestamp).toLocaleString()
}

export default api
