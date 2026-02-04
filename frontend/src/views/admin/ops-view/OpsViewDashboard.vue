<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <!-- Page Header -->
        <div class="flex items-center justify-between">
          <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.opsView.title') }}
          </h1>
          <button
            @click="refreshData"
            class="btn btn-secondary flex items-center gap-2"
            :disabled="loading"
          >
            <Icon name="refresh" size="sm" />
            {{ t('common.refresh') }}
          </button>
        </div>

        <!-- KPI Overview Cards -->
        <OpsViewKpiCards :overview="overview" :loading="overviewLoading" />

        <!-- Charts Section -->
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <!-- Revenue Trend Chart -->
          <OpsViewTrendChart :trend="trend" :loading="trendLoading" />

          <!-- User Growth Chart -->
          <OpsViewUserGrowthChart :growth="userGrowth" :loading="userGrowthLoading" />
        </div>

        <!-- Top Users Table -->
        <OpsViewTopUsersTable :users="topUsers" :loading="topUsersLoading" />
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsViewAPI } from '@/api/admin/opsView'
import type {
  OpsViewOverview,
  OpsViewTrendPoint,
  OpsViewUserGrowthPoint,
  OpsViewTopUser
} from '@/api/admin/opsView'

import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import OpsViewKpiCards from './components/OpsViewKpiCards.vue'
import OpsViewTrendChart from './components/OpsViewTrendChart.vue'
import OpsViewUserGrowthChart from './components/OpsViewUserGrowthChart.vue'
import OpsViewTopUsersTable from './components/OpsViewTopUsersTable.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Loading states
const loading = ref(false)
const overviewLoading = ref(false)
const trendLoading = ref(false)
const userGrowthLoading = ref(false)
const topUsersLoading = ref(false)

// Error tracking
const errorCount = ref(0)

// Data
const overview = ref<OpsViewOverview | null>(null)
const trend = ref<OpsViewTrendPoint[]>([])
const userGrowth = ref<OpsViewUserGrowthPoint[]>([])
const topUsers = ref<OpsViewTopUser[]>([])

// Get user timezone
const getTimezone = () => Intl.DateTimeFormat().resolvedOptions().timeZone

// Load functions
const loadOverview = async () => {
  overviewLoading.value = true
  try {
    overview.value = await opsViewAPI.getOverview(getTimezone())
  } catch (error) {
    console.error('Failed to load overview:', error)
    errorCount.value++
  } finally {
    overviewLoading.value = false
  }
}

const loadTrend = async () => {
  trendLoading.value = true
  try {
    const response = await opsViewAPI.getTrend({ days: 30, timezone: getTimezone() })
    trend.value = response.trend
  } catch (error) {
    console.error('Failed to load trend:', error)
    errorCount.value++
  } finally {
    trendLoading.value = false
  }
}

const loadUserGrowth = async () => {
  userGrowthLoading.value = true
  try {
    const response = await opsViewAPI.getUserGrowth({ days: 30, timezone: getTimezone() })
    userGrowth.value = response.growth
  } catch (error) {
    console.error('Failed to load user growth:', error)
    errorCount.value++
  } finally {
    userGrowthLoading.value = false
  }
}

const loadTopUsers = async () => {
  topUsersLoading.value = true
  try {
    const response = await opsViewAPI.getTopUsers({ limit: 100, timezone: getTimezone() })
    topUsers.value = response.users
  } catch (error) {
    console.error('Failed to load top users:', error)
    errorCount.value++
  } finally {
    topUsersLoading.value = false
  }
}

const loadAllData = async () => {
  loading.value = true
  errorCount.value = 0

  // Use Promise.allSettled since individual load functions catch their own errors
  // This ensures we wait for all requests to complete
  await Promise.allSettled([
    loadOverview(),
    loadTrend(),
    loadUserGrowth(),
    loadTopUsers()
  ])

  // Show error toast if any data failed to load
  if (errorCount.value > 0) {
    appStore.showError(t('admin.opsView.loadError'))
  }

  loading.value = false
}

const refreshData = () => {
  loadAllData()
}

onMounted(() => {
  loadAllData()
})
</script>
