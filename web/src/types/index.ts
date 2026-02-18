// System Metrics Types
export interface CPUStat {
  cores: number
  usage_percent: number
  per_cpu_usage: number[]
  load1: number
  load5: number
  load15: number
  total_ticks: number
  idle_ticks: number
}

export interface MemoryStat {
  total: number
  free: number
  available: number
  used: number
  used_percent: number
  swap_total: number
  swap_free: number
  swap_used: number
  swap_used_percent: number
}

export interface DiskStat {
  mount_point: string
  device: string
  total: number
  used: number
  free: number
  used_percent: number
  inodes_total: number
  inodes_used: number
  inodes_free: number
  inodes_used_percent: number
  read: number
  read_sectors: number
  read_speed: number
  write: number
  write_sectors: number
  write_speed: number
  await: number
  util: number
  io_queue_time: number
}

export interface NetStat {
  name: string
  rx_bytes: number
  rx_packets: number
  rx_errors: number
  rx_dropped: number
  tx_bytes: number
  tx_packets: number
  tx_errors: number
  tx_dropped: number
  rx_speed: number
  tx_speed: number
}

export interface ProcStat {
  pid: number
  name: string
  cpu: number
  mem: number
}

export interface Metrics {
  cpu: CPUStat
  memory: MemoryStat
  disk: DiskStat[]
  net: NetStat[]
  procs: ProcStat[]
  host: string
  update_timestamp: string
}

export interface HealthStatus {
  status: 'healthy' | 'warning' | 'critical'
  timestamp: string
  checks: {
    cpu: string
    memory: string
    disk: string
    network: string
  }
}

export interface AppConfig {
  app: {
    name: string
    version: string
    refresh_interval: string
    log_level: string
  }
  diagnostic: {
    enabled: boolean
    show_top_n_list: number
  }
  alert: {
    enabled: boolean
    cpu_threshold: number
    memory_threshold: number
    disk_threshold: number
    inodes_threshold: number
  }
}
