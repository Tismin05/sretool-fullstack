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
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { NetStat } from '@/types'
import { formatBytes } from '@/api/metrics'

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent])

const props = defineProps<{
  data: NetStat[]
}>()

const chartOption = computed(() => {
  const nets = props.data || []

  // 过滤掉loopback
  const filteredNets = nets.filter(n => !n.name.startsWith('lo'))

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: '#1a2234',
      borderColor: '#2a3548',
      textStyle: { color: '#e8edf5' },
      formatter: (params: any) => {
        const net = filteredNets[params[0].dataIndex]
        if (!net) return ''
        return `
          <div style="font-weight: 600">${net.name}</div>
          <div style="color: #00d9ff">↓ ${formatBytes(net.rx_speed || 0)}/s</div>
          <div style="color: #00ff9d">↑ ${formatBytes(net.tx_speed || 0)}>
          <div/s</div style="color: #8b95a5; font-size: 11px; margin-top: 4px">
            RX: ${formatBytes(net.rx_bytes)} | TX: ${formatBytes(net.tx_bytes)}
          </div>
        `
      }
    },
    legend: {
      data: ['Download', 'Upload'],
      bottom: 0,
      textStyle: { color: '#8b95a5' },
      itemWidth: 12,
      itemHeight: 12
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: filteredNets.map(n => n.name),
      axisLine: { lineStyle: { color: '#2a3548' } },
      axisLabel: {
        color: '#8b95a5',
        fontSize: 10
      }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#2a3548' } },
      axisLabel: {
        color: '#8b95a5',
        formatter: (value: number) => formatBytes(value) + '/s'
      },
      splitLine: { lineStyle: { color: '#1a2234' } }
    },
    series: [
      {
        name: 'Download',
        type: 'bar',
        stack: 'total',
        barWidth: '40%',
        data: filteredNets.map(n => n.rx_speed || 0),
        itemStyle: {
          color: '#00d9ff',
          borderRadius: [0, 0, 4, 4]
        }
      },
      {
        name: 'Upload',
        type: 'bar',
        stack: 'total',
        barWidth: '40%',
        data: filteredNets.map(n => n.tx_speed || 0),
        itemStyle: {
          color: '#00ff9d',
          borderRadius: [4, 4, 0, 0]
        }
      }
    ]
  }
})
</script>

<style scoped>
.chart-container {
  height: 220px;
  width: 100%;
}
</style>
