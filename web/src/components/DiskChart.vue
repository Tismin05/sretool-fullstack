<template>
  <div class="chart-container">
    <v-chart :option="chartOption" autoresize />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { DiskStat } from '@/types'
import { formatBytes } from '@/api/metrics'

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent])

const props = defineProps<{
  data: DiskStat[]
}>()

const chartOption = computed(() => {
  const disks = props.data || []

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: '#1a2234',
      borderColor: '#2a3548',
      textStyle: { color: '#e8edf5' },
      formatter: (params: any) => {
        const data = params[0]
        const disk = disks[data.dataIndex]
        if (!disk) return ''
        return `
          <div style="font-weight: 600">${disk.mount_point}</div>
          <div>Total: ${formatBytes(disk.total)}</div>
          <div>Used: ${formatBytes(disk.used)}</div>
          <div>Free: ${formatBytes(disk.free)}</div>
          <div>Inodes: ${disk.inodes_used_percent.toFixed(1)}%</div>
        `
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: disks.map(d => d.mount_point),
      axisLine: { lineStyle: { color: '#2a3548' } },
      axisLabel: {
        color: '#8b95a5',
        fontSize: 10,
        rotate: 15
      }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: '#2a3548' } },
      axisLabel: {
        color: '#8b95a5',
        formatter: '{value}%'
      },
      splitLine: { lineStyle: { color: '#1a2234' } }
    },
    series: [
      {
        name: 'Usage',
        type: 'bar',
        barWidth: '50%',
        data: disks.map(d => ({
          value: d.used_percent,
          itemStyle: {
            color: getColor(d.used_percent),
            borderRadius: [4, 4, 0, 0]
          }
        })),
        showBackground: true,
        backgroundStyle: {
          color: '#1a2234',
          borderRadius: [4, 4, 0, 0]
        }
      }
    ]
  }
})

function getColor(value: number): string {
  if (value >= 90) return '#ff4757'
  if (value >= 80) return '#ff9500'
  if (value >= 60) return '#00d9ff'
  return '#00ff9d'
}
</script>

<style scoped>
.chart-container {
  height: 220px;
  width: 100%;
}
</style>
