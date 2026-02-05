<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Title,
  Tooltip
} from 'chart.js'
import { Line } from 'vue-chartjs'
import type { OpsViewTrendPoint } from '@/api/admin/opsView'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'

ChartJS.register(Title, Tooltip, Legend, LineElement, LinearScale, PointElement, CategoryScale, Filler)

interface Props {
  trend: OpsViewTrendPoint[]
  loading: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const colors = computed(() => ({
  consumption: '#10b981',
  consumptionAlpha: '#10b98120',
  payment: '#3b82f6',
  paymentAlpha: '#3b82f620',
  redeem: '#f59e0b',
  redeemAlpha: '#f59e0b20',
  totalRecharge: '#8b5cf6',
  totalBalance: '#ec4899',
  grid: isDarkMode.value ? '#374151' : '#f3f4f6',
  text: isDarkMode.value ? '#9ca3af' : '#6b7280'
}))

// Format date label - parse date string directly to avoid timezone issues
const formatDateLabel = (dateStr: string): string => {
  const [, month, day] = dateStr.split('-')
  return `${parseInt(month)}/${parseInt(day)}`
}

const hasData = computed(() => props.trend.length > 0)

const chartData = computed(() => {
  if (!hasData.value) return null
  const c = colors.value
  return {
    labels: props.trend.map((p) => formatDateLabel(p.date)),
    datasets: [
      {
        label: t('admin.opsView.trend.consumption'),
        data: props.trend.map((p) => p.consumption),
        borderColor: c.consumption,
        backgroundColor: c.consumptionAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.opsView.trend.paymentRecharge'),
        data: props.trend.map((p) => p.payment_recharge),
        borderColor: c.payment,
        backgroundColor: c.paymentAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.opsView.trend.redeemRecharge'),
        data: props.trend.map((p) => p.redeem_recharge),
        borderColor: c.redeem,
        backgroundColor: c.redeemAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.opsView.trend.totalRecharge'),
        data: props.trend.map((p) => p.total_recharge),
        borderColor: c.totalRecharge,
        backgroundColor: 'transparent',
        fill: false,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10,
        borderDash: [5, 5]
      },
      {
        label: t('admin.opsView.trend.totalBalance'),
        data: props.trend.map((p) => p.total_balance),
        borderColor: c.totalBalance,
        backgroundColor: 'transparent',
        fill: false,
        tension: 0.35,
        pointRadius: 0,
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
            return `${context.dataset.label}: $${value.toFixed(2)}`
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
        ticks: {
          color: c.text,
          font: { size: 10 },
          callback: (value: string | number) => `$${value}`
        }
      },
      y1: {
        type: 'linear' as const,
        display: true,
        position: 'right' as const,
        grid: { display: false },
        ticks: {
          color: c.totalBalance,
          font: { size: 10 },
          callback: (value: string | number) => `$${value}`
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
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-100 dark:bg-emerald-900/30">
          <Icon name="trendingUp" size="sm" class="text-emerald-600 dark:text-emerald-400" />
        </div>
        {{ t('admin.opsView.trend.title') }}
      </h3>
    </div>

    <div class="min-h-0 flex-1">
      <Line v-if="state === 'ready' && chartData" :data="chartData" :options="options" />
      <div v-else class="flex h-full items-center justify-center">
        <div v-if="state === 'loading'" class="animate-pulse text-sm text-gray-400">
          {{ t('common.loading') }}
        </div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.opsView.trend.empty')" />
      </div>
    </div>
  </div>
</template>
