<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsViewTopUser } from '@/api/admin/opsView'
import Icon from '@/components/icons/Icon.vue'
import Skeleton from '@/components/common/Skeleton.vue'
import EmptyState from '@/components/common/EmptyState.vue'

interface Props {
  users: OpsViewTopUser[]
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

// Format date - use toLocaleString to ensure time is displayed in all browsers
const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const hasData = computed(() => props.users.length > 0)

// Get rank badge color
const getRankBadgeClass = (rank: number): string => {
  if (rank === 1) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  if (rank === 2) return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
  if (rank === 3) return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
  return 'bg-gray-50 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
}
</script>

<template>
  <div class="rounded-2xl bg-white p-5 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700">
    <div class="mb-4 flex items-center gap-2">
      <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-100 dark:bg-amber-900/30">
        <Icon name="users" size="sm" class="text-amber-600 dark:text-amber-400" />
      </div>
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {{ t('admin.opsView.topUsers.title') }}
      </h3>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 5" :key="i" class="flex items-center gap-4">
        <Skeleton variant="rect" :width="32" :height="32" class="rounded-lg" />
        <Skeleton variant="text" width="30%" />
        <Skeleton variant="text" width="15%" />
        <Skeleton variant="text" width="15%" />
        <Skeleton variant="text" width="15%" />
        <Skeleton variant="text" width="20%" />
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!hasData" class="flex h-48 items-center justify-center">
      <EmptyState :title="t('common.noData')" :description="t('admin.opsView.topUsers.empty')" />
    </div>

    <!-- Table -->
    <div v-else class="max-h-96 overflow-y-auto overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-100 text-xs text-gray-500 dark:border-gray-700 dark:text-gray-400">
            <th class="pb-3 text-left font-medium">{{ t('admin.opsView.topUsers.rank') }}</th>
            <th class="pb-3 text-left font-medium">{{ t('admin.opsView.topUsers.user') }}</th>
            <th class="pb-3 text-right font-medium">{{ t('admin.opsView.topUsers.consumption') }}</th>
            <th class="pb-3 text-right font-medium">{{ t('admin.opsView.topUsers.requests') }}</th>
            <th class="pb-3 text-right font-medium">{{ t('admin.opsView.topUsers.avgCost') }}</th>
            <th class="pb-3 text-right font-medium">{{ t('admin.opsView.topUsers.lastActive') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(user, index) in users"
            :key="user.user_id"
            class="border-b border-gray-50 last:border-0 dark:border-gray-800"
          >
            <td class="py-3">
              <span
                :class="[
                  'inline-flex h-7 w-7 items-center justify-center rounded-lg text-xs font-bold',
                  getRankBadgeClass(index + 1)
                ]"
              >
                {{ index + 1 }}
              </span>
            </td>
            <td class="py-3">
              <div class="flex flex-col">
                <span class="font-medium text-gray-900 dark:text-white">
                  {{ user.username || user.email.split('@')[0] }}
                </span>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ user.email }}</span>
              </div>
            </td>
            <td class="py-3 text-right font-semibold text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(user.consumption) }}
            </td>
            <td class="py-3 text-right text-gray-600 dark:text-gray-400">
              {{ formatNumber(user.requests) }}
            </td>
            <td class="py-3 text-right text-gray-600 dark:text-gray-400">
              {{ formatCurrency(user.avg_cost) }}
            </td>
            <td class="py-3 text-right text-xs text-gray-500 dark:text-gray-400">
              {{ formatDate(user.last_active_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
