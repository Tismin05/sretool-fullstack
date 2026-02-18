<template>
  <div class="chart-container">
    <v-chart :option="chartOption" autoresize />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { Metrics } from '@/types'
import { formatBytes } from '@/api/metrics'

use([CanvasRenderer, PieChart, TooltipComponent, LegendComponent])

const props = defineProps<{
  data: Metrics | null
}>()

const chartOption = computed(() => {
  const memory = props.data?.memory
  if (!memory) return {}

  const used = memory.used
  const free = memory.free
  const available = memory.available
  const buffers = used - available - free
  const swapUsed = memory.swap_used

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: '#1a2234',
      borderColor: '#2a3548',
      textStyle: { color: '#e8edf5' },
      formatter: (params: any) => {
        return `${params.name}: ${formatBytes(params.value)}`
      }
    },
    legend: {
      orient: 'horizontal',
      bottom: 0,
      textStyle: { color: '#8b95a5' },
      itemWidth: 12,
      itemHeight: 12
    },
    series: [
      {
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 6,
          borderColor: '#151d2e',
          borderWidth: 2
        },
        label: {
          show: false
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold',
            color: '#e8edf5'
          },
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        },
        data: [
          {
            value: used - buffers,
            name: 'Used',
            itemStyle: { color: '#00d9ff' }
          },
          {
            value: buffers,
            name: 'Buffers',
            itemStyle: { color: '#a855f7' }
          },
          {
            value: free,
            name: 'Free',
            itemStyle: { color: '#2a3548' }
          },
          {
            value: swapUsed,
            name: 'Swap',
            itemStyle: { color: '#ff4757' }
          }
        ]
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
