<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip
} from 'chart.js'
import { Bar } from 'vue-chartjs'
import type { OpsViewUserGrowthPoint } from '@/api/admin/opsView'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'

ChartJS.register(Title, Tooltip, Legend, BarElement, LineElement, LinearScale, PointElement, CategoryScale)

interface Props {
  growth: OpsViewUserGrowthPoint[]
  loading: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const colors = computed(() => ({
  newUsers: '#8b5cf6',
  activeUsers: '#3b82f6',
  retention: '#10b981',
  grid: isDarkMode.value ? '#374151' : '#f3f4f6',
  text: isDarkMode.value ? '#9ca3af' : '#6b7280'
}))

// Format date label - parse date string directly to avoid timezone issues
const formatDateLabel = (dateStr: string): string => {
  const [, month, day] = dateStr.split('-')
  return `${parseInt(month)}/${parseInt(day)}`
}

const hasData = computed(() => props.growth.length > 0)

const chartData = computed(() => {
  if (!hasData.value) return null
  const c = colors.value
  return {
    labels: props.growth.map((p) => formatDateLabel(p.date)),
    datasets: [
      {
        type: 'bar' as const,
        label: t('admin.opsView.userGrowth.newUsers'),
        data: props.growth.map((p) => p.new_users),
        backgroundColor: c.newUsers,
        borderRadius: 4,
        barPercentage: 0.6,
        yAxisID: 'y'
      },
      {
        type: 'line' as const,
        label: t('admin.opsView.userGrowth.activeUsers'),
        data: props.growth.map((p) => p.active_users),
        borderColor: c.activeUsers,
        backgroundColor: 'transparent',
        tension: 0.35,
        pointRadius: 2,
        pointHitRadius: 10,
        yAxisID: 'y'
      },
      {
        type: 'line' as const,
        label: t('admin.opsView.userGrowth.retentionD1'),
        data: props.growth.map((p) => p.retention_d1),
        borderColor: c.retention,
        backgroundColor: 'transparent',
        tension: 0.35,
        pointRadius: 2,
        pointHitRadius: 10,
        yAxisID: 'y1'
      }
    ]
  }
})

const options = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: 'index' as const },
    plugins: {
      legend: {
        position: 'top' as const,
        align: 'end' as const,
        labels: { color: c.text, usePointStyle: true, boxWidth: 6, font: { size: 10 } }
      },
      tooltip: {
        backgroundColor: isDarkMode.value ? '#1f2937' : '#ffffff',
        titleColor: isDarkMode.value ? '#f3f4f6' : '#111827',
        bodyColor: isDarkMode.value ? '#d1d5db' : '#4b5563',
        borderColor: c.grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true,
        callbacks: {
          label: (context: any) => {
            const value = context.raw as number
            if (context.dataset.yAxisID === 'y1') {
              return `${context.dataset.label}: ${value.toFixed(1)}%`
            }
            return `${context.dataset.label}: ${value}`
          }
        }
      }
    },
    scales: {
      x: {
        type: 'category' as const,
        grid: { display: false },
        ticks: {
          color: c.text,
          font: { size: 10 },
          maxTicksLimit: 10,
          autoSkip: true,
          autoSkipPadding: 10
        }
      },
      y: {
        type: 'linear' as const,
        display: true,
        position: 'left' as const,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 }, precision: 0 },
        title: {
          display: true,
          text: t('admin.opsView.userGrowth.users'),
          color: c.text,
          font: { size: 10 }
        }
      },
      y1: {
        type: 'linear' as const,
        display: true,
        position: 'right' as const,
        grid: { display: false },
        min: 0,
        max: 100,
        ticks: {
          color: c.text,
          font: { size: 10 },
          callback: (value: string | number) => `${value}%`
        },
        title: {
          display: true,
          text: t('admin.opsView.userGrowth.retention'),
          color: c.text,
          font: { size: 10 }
        }
      }
    }
  }
})

type ChartState = 'loading' | 'ready' | 'empty'

const state = computed<ChartState>(() => {
  if (chartData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})
</script>

<template>
  <div class="flex h-80 flex-col rounded-2xl bg-white p-5 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700">
    <div class="mb-4 flex shrink-0 items-center justify-between">
      <h3 class="flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-900/30">
          <Icon name="users" size="sm" class="text-purple-600 dark:text-purple-400" />
        </div>
        {{ t('admin.opsView.userGrowth.title') }}
      </h3>
    </div>

    <div class="min-h-0 flex-1">
      <Bar v-if="state === 'ready' && chartData" :data="chartData as any" :options="options" />
      <div v-else class="flex h-full items-center justify-center">
        <div v-if="state === 'loading'" class="animate-pulse text-sm text-gray-400">
          {{ t('common.loading') }}
        </div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.opsView.userGrowth.empty')" />
      </div>
    </div>
  </div>
</template>
