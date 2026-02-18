<template>
  <div class="chart-container">
    <v-chart :option="chartOption" autoresize />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { GaugeChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { Metrics } from '@/types'

use([CanvasRenderer, GaugeChart, BarChart, GridComponent, TooltipComponent])

const props = defineProps<{
  data: Metrics | null
}>()

const chartOption = computed(() => {
  const cpuUsage = props.data?.cpu.usage_percent || 0

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: '#1a2234',
      borderColor: '#2a3548',
      textStyle: { color: '#e8edf5' }
    },
    series: [
      // 仪表盘
      {
        type: 'gauge',
        startAngle: 200,
        endAngle: -20,
        min: 0,
        max: 100,
        splitNumber: 10,
        radius: '90%',
        center: ['50%', '55%'],
        axisLine: {
          lineStyle: {
            width: 12,
            color: [
              [cpuUsage / 100, getColor(cpuUsage)],
              [1, '#2a3548']
            ]
          }
        },
        pointer: {
          icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
          length: '60%',
          width: 10,
          offsetCenter: [0, '-10%'],
          itemStyle: { color: '#00d9ff' }
        },
        axisTick: {
          length: 8,
          lineStyle: { color: 'auto', width: 1 }
        },
        splitLine: {
          length: 15,
          lineStyle: { color: 'auto', width: 2 }
        },
        axisLabel: {
          color: '#8b95a5',
          fontSize: 10,
          distance: -40
        },
        title: {
          offsetCenter: [0, '30%'],
          fontSize: 14,
          color: '#8b95a5'
        },
        detail: {
          fontSize: 32,
          offsetCenter: [0, '70%'],
          valueAnimation: true,
          formatter: '{value}%',
          color: '#e8edf5',
          fontFamily: 'JetBrains Mono',
          fontWeight: 600
        },
        data: [{ value: cpuUsage.toFixed(1) }]
      }
    ]
  }
})

function getColor(value: number): string {
  if (value >= 90) return '#ff4757'
  if (value >= 70) return '#ff9500'
  return '#00ff9d'
}
</script>

<style scoped>
.chart-container {
  height: 220px;
  width: 100%;
}
</style>
