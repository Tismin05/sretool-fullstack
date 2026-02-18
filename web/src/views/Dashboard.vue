<template>
  <div class="dashboard">
    <!-- Header -->
    <header class="dashboard-header">
      <div class="header-left">
        <div class="logo">
          <el-icon :size="32" color="#00d9ff"><Monitor /></el-icon>
          <span class="logo-text gradient-text">tisminSRETool</span>
        </div>
        <div class="host-info">
          <el-tag :type="statusType" effect="dark" size="small">
            {{ metricsStore.hostName }}
          </el-tag>
        </div>
      </div>
      <div class="header-right">
        <div class="status-indicator" :class="{ connected: metricsStore.wsConnected }">
          <span class="status-dot"></span>
          {{ metricsStore.wsConnected ? 'Live' : 'Disconnected' }}
        </div>
        <div class="last-update">
          <el-icon><Clock /></el-icon>
          {{ lastUpdateText }}
        </div>
        <el-button type="primary" text @click="refresh">
          <el-icon><Refresh /></el-icon>
          Refresh
        </el-button>
      </div>
    </header>

    <!-- 状态概览卡片 -->
    <div class="status-cards">
      <div class="status-card" :class="getStatusClass('cpu')">
        <div class="card-icon">
          <el-icon :size="24"><Cpu /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-label">CPU Usage</div>
          <div class="card-value">{{ metricsStore.cpuUsage.toFixed(1) }}%</div>
          <el-progress
            :percentage="metricsStore.cpuUsage"
            :stroke-width="6"
            :color="getProgressColor(metricsStore.cpuUsage)"
          />
        </div>
      </div>

      <div class="status-card" :class="getStatusClass('memory')">
        <div class="card-icon">
          <el-icon :size="24"><MemoryStick /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-label">Memory Usage</div>
          <div class="card-value">{{ metricsStore.memoryUsage.toFixed(1) }}%</div>
          <el-progress
            :percentage="metricsStore.memoryUsage"
            :stroke-width="6"
            :color="getProgressColor(metricsStore.memoryUsage)"
          />
        </div>
      </div>

      <div class="status-card" :class="getStatusClass('disk')">
        <div class="card-icon">
          <el-icon :size="24"><HardDrive /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-label">Disk Usage</div>
          <div class="card-value">{{ maxDiskUsage.toFixed(1) }}%</div>
          <el-progress
            :percentage="maxDiskUsage"
            :stroke-width="6"
            :color="getProgressColor(maxDiskUsage)"
          />
        </div>
      </div>

      <div class="status-card" :class="getStatusClass('network')">
        <div class="card-icon">
          <el-icon :size="24"><Wifi /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-label">Network I/O</div>
          <div class="card-value">{{ totalNetSpeed }}</div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-grid">
      <!-- CPU Chart -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="chart-header">
            <span><el-icon><Cpu /></el-icon> CPU Performance</span>
            <div class="cpu-load">
              <span>Load: {{ metricsStore.metrics?.cpu.load1?.toFixed(2) }}</span>
              <span class="muted">/ {{ metricsStore.metrics?.cpu.load5?.toFixed(2) }}</span>
              <span class="muted">/ {{ metricsStore.metrics?.cpu.load15?.toFixed(2) }}</span>
            </div>
          </div>
        </template>
        <CpuChart :data="metricsStore.metrics" />
      </el-card>

      <!-- Memory Chart -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="chart-header">
            <span><el-icon><MemoryStick /></el-icon> Memory</span>
            <span class="memory-info">
              {{ formatBytes(metricsStore.metrics?.memory.used || 0) }} /
              {{ formatBytes(metricsStore.metrics?.memory.total || 0) }}
            </span>
          </div>
        </template>
        <MemoryChart :data="metricsStore.metrics" />
      </el-card>

      <!-- Disk Usage -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="chart-header">
            <span><el-icon><HardDrive /></el-icon> Disk Usage</span>
          </div>
        </template>
        <DiskChart :data="metricsStore.diskList" />
      </el-card>

      <!-- Network -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <div class="chart-header">
            <span><el-icon><Wifi /></el-icon> Network Traffic</span>
          </div>
        </template>
        <NetChart :data="metricsStore.netList" />
      </el-card>
    </div>

    <!-- 进程信息 -->
    <el-card class="process-card" shadow="never">
      <template #header>
        <div class="chart-header">
          <span><el-icon><Operation /></el-icon> Top Processes</span>
        </div>
      </template>
      <el-table
        :data="metricsStore.metrics?.procs || []"
        style="width: 100%"
        :row-class-name="tableRowClassName"
      >
        <el-table-column prop="pid" label="PID" width="80" />
        <el-table-column prop="name" label="Process Name" />
        <el-table-column label="CPU" width="120">
          <template #default="{ row }">
            {{ row.cpu?.toFixed(0) || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="Memory" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.mem || 0) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useMetricsStore } from '@/stores/metrics'
import { formatBytes } from '@/api/metrics'
import CpuChart from '@/components/CpuChart.vue'
import MemoryChart from '@/components/MemoryChart.vue'
import DiskChart from '@/components/DiskChart.vue'
import NetChart from '@/components/NetChart.vue'

const metricsStore = useMetricsStore()

// 计算最大磁盘使用率
const maxDiskUsage = computed(() => {
  const disks = metricsStore.diskList
  if (!disks || disks.length === 0) return 0
  return Math.max(...disks.map(d => d.used_percent))
})

// 计算总网络速度
const totalNetSpeed = computed(() => {
  const nets = metricsStore.netList
  if (!nets || nets.length === 0) return '0 B/s'
  let total = 0
  for (const net of nets) {
    total += (net.rx_speed || 0) + (net.tx_speed || 0)
  }
  return formatBytes(total) + '/s'
})

// 状态类型
const statusType = computed(() => {
  const health = metricsStore.health?.status
  if (health === 'critical') return 'danger'
  if (health === 'warning') return 'warning'
  return 'success'
})

// 获取状态类名
function getStatusClass(type: string): string {
  const checks = metricsStore.health?.checks
  if (!checks) return ''
  const status = checks[type]
  if (status === 'critical') return 'status-critical'
  if (status === 'warning') return 'status-warning'
  return 'status-healthy'
}

// 获取进度条颜色
function getProgressColor(value: number): string {
  if (value >= 90) return '#ff4757'
  if (value >= 70) return '#ff9500'
  return '#00ff9d'
}

// 最后更新时间
const lastUpdateText = computed(() => {
  if (!metricsStore.lastUpdate) return 'Never'
  const date = new Date(metricsStore.lastUpdate)
  return date.toLocaleTimeString()
})

// 刷新数据
async function refresh() {
  await metricsStore.fetchMetrics()
  await metricsStore.fetchHealth()
}

// 表格行样式
function tableRowClassName({ row }: { row: { cpu: number } }) {
  if (row.cpu > 50) return 'danger-row'
  if (row.cpu > 20) return 'warning-row'
  return ''
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
  max-width: 1600px;
  margin: 0 auto;
}

/* Header */
.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: var(--bg-card);
  border-radius: 16px;
  margin-bottom: 24px;
  border: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-text {
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-muted);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent-red);
}

.status-indicator.connected .status-dot {
  background: var(--accent-green);
  animation: pulse-glow 2s infinite;
}

.last-update {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
}

/* Status Cards */
.status-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.status-card {
  background: var(--bg-card);
  border-radius: 16px;
  padding: 20px;
  border: 1px solid var(--border-color);
  display: flex;
  gap: 16px;
  transition: all 0.3s ease;
}

.status-card:hover {
  border-color: var(--accent-cyan);
  box-shadow: var(--shadow-glow);
}

.card-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-cyan);
}

.card-content {
  flex: 1;
}

.card-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.card-value {
  font-size: 28px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 8px;
}

/* Charts Grid */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.chart-card {
  background: var(--bg-card);
  border-radius: 16px;
  border: 1px solid var(--border-color);
}

.chart-card :deep(.el-card__header) {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-header .muted {
  color: var(--text-muted);
  font-weight: 400;
  font-size: 12px;
}

.cpu-load, .memory-info {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-secondary);
}

/* Process Card */
.process-card {
  background: var(--bg-card);
  border-radius: 16px;
  border: 1px solid var(--border-color);
}

/* 响应式 */
@media (max-width: 1200px) {
  .status-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .status-cards {
    grid-template-columns: 1fr;
  }

  .charts-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-header {
    flex-direction: column;
    gap: 16px;
  }

  .header-left, .header-right {
    width: 100%;
    justify-content: center;
  }
}

:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: var(--bg-tertiary);
  --el-table-row-hover-bg-color: var(--bg-tertiary);
  --el-table-border-color: var(--border-color);
  --el-table-text-color: var(--text-primary);
  --el-table-header-text-color: var(--text-secondary);
}

:deep(.danger-row) {
  --el-table-tr-bg-color: rgba(255, 71, 87, 0.1);
}

:deep(.warning-row) {
  --el-table-tr-bg-color: rgba(255, 149, 0, 0.1);
}
</style>
