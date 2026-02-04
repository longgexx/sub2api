<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsViewOverview } from '@/api/admin/opsView'
import Icon from '@/components/icons/Icon.vue'
import Skeleton from '@/components/common/Skeleton.vue'

interface Props {
  overview: OpsViewOverview | null
  loading: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()

// Format currency with K/M suffix
const formatCurrency = (value: number): string => {
  if (value >= 1_000_000) {
    return `$${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `$${(value / 1_000).toFixed(2)}K`
  }
  return `$${value.toFixed(2)}`
}

// Format number with K/M suffix
const formatNumber = (value: number): string => {
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(1)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(1)}K`
  }
  return value.toLocaleString()
}

const kpiCards = computed(() => {
  if (!props.overview) return null
  return [
    {
      key: 'totalBalance',
      label: t('admin.opsView.kpi.totalBalance'),
      value: formatCurrency(props.overview.total_balance),
      icon: 'dollar' as const,
      iconBg: 'bg-amber-100 dark:bg-amber-900/30',
      iconColor: 'text-amber-600 dark:text-amber-400',
      valueColor: 'text-amber-600 dark:text-amber-400'
    },
    {
      key: 'todayConsumption',
      label: t('admin.opsView.kpi.todayConsumption'),
      value: formatCurrency(props.overview.today_consumption),
      icon: 'trendingUp' as const,
      iconBg: 'bg-emerald-100 dark:bg-emerald-900/30',
      iconColor: 'text-emerald-600 dark:text-emerald-400',
      valueColor: 'text-emerald-600 dark:text-emerald-400'
    },
    {
      key: 'todayRecharge',
      label: t('admin.opsView.kpi.todayRecharge'),
      value: formatCurrency(props.overview.today_recharge),
      icon: 'creditCard' as const,
      iconBg: 'bg-blue-100 dark:bg-blue-900/30',
      iconColor: 'text-blue-600 dark:text-blue-400',
      valueColor: 'text-blue-600 dark:text-blue-400'
    },
    {
      key: 'totalUsers',
      label: t('admin.opsView.kpi.totalUsers'),
      value: formatNumber(props.overview.total_users),
      icon: 'users' as const,
      iconBg: 'bg-purple-100 dark:bg-purple-900/30',
      iconColor: 'text-purple-600 dark:text-purple-400',
      valueColor: 'text-purple-600 dark:text-purple-400'
    },
    {
      key: 'todayNewUsers',
      label: t('admin.opsView.kpi.todayNewUsers'),
      value: formatNumber(props.overview.today_new_users),
      icon: 'userPlus' as const,
      iconBg: 'bg-green-100 dark:bg-green-900/30',
      iconColor: 'text-green-600 dark:text-green-400',
      valueColor: 'text-green-600 dark:text-green-400'
    },
    {
      key: 'todayActiveUsers',
      label: t('admin.opsView.kpi.todayActiveUsers'),
      value: formatNumber(props.overview.today_active_users),
      icon: 'bolt' as const,
      iconBg: 'bg-indigo-100 dark:bg-indigo-900/30',
      iconColor: 'text-indigo-600 dark:text-indigo-400',
      valueColor: 'text-indigo-600 dark:text-indigo-400'
    }
  ]
})
</script>

<template>
  <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
    <!-- Loading State -->
    <template v-if="loading">
      <div
        v-for="i in 6"
        :key="i"
        class="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700"
      >
        <div class="mb-3 flex items-center gap-2">
          <Skeleton variant="rect" :width="32" :height="32" class="rounded-lg" />
          <Skeleton variant="text" width="60%" />
        </div>
        <Skeleton variant="text" width="50%" :height="28" />
      </div>
    </template>

    <!-- KPI Cards -->
    <template v-else-if="kpiCards">
      <div
        v-for="card in kpiCards"
        :key="card.key"
        class="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700"
      >
        <div class="mb-3 flex items-center gap-2">
          <div :class="['flex h-8 w-8 items-center justify-center rounded-lg', card.iconBg]">
            <Icon :name="card.icon" size="sm" :class="card.iconColor" />
          </div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ card.label }}
          </span>
        </div>
        <p :class="['text-xl font-bold', card.valueColor]">
          {{ card.value }}
        </p>
      </div>
    </template>
  </div>
</template>
