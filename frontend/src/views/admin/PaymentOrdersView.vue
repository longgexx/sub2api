<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex justify-end gap-3">
          <button
            @click="loadOrders"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </template>

      <template #filters>
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <!-- Stats Cards -->
          <div class="flex gap-4">
            <div class="rounded-lg bg-emerald-50 px-4 py-2 dark:bg-emerald-900/20">
              <p class="text-xs text-emerald-600 dark:text-emerald-400">{{ t('admin.payment.todayPaid') }}</p>
              <p class="text-lg font-semibold text-emerald-700 dark:text-emerald-300">
                ${{ stats?.today_paid_amount?.toFixed(2) || '0.00' }}
              </p>
            </div>
            <div class="rounded-lg bg-primary-50 px-4 py-2 dark:bg-primary-900/20">
              <p class="text-xs text-primary-600 dark:text-primary-400">{{ t('admin.payment.totalPaid') }}</p>
              <p class="text-lg font-semibold text-primary-700 dark:text-primary-300">
                ${{ stats?.total_paid_amount?.toFixed(2) || '0.00' }}
              </p>
            </div>
            <div class="rounded-lg bg-amber-50 px-4 py-2 dark:bg-amber-900/20">
              <p class="text-xs text-amber-600 dark:text-amber-400">{{ t('admin.payment.pendingOrders') }}</p>
              <p class="text-lg font-semibold text-amber-700 dark:text-amber-300">
                {{ stats?.pending_order_count || 0 }}
              </p>
            </div>
          </div>

          <!-- Filter and Search -->
          <div class="flex gap-2">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.payment.searchOrders')"
              class="input w-48"
              @input="handleSearch"
            />
            <Select
              v-model="filters.status"
              :options="filterStatusOptions"
              class="w-36"
              @change="loadOrders"
            />
          </div>
        </div>

        <!-- Monitor Status -->
        <div class="mt-4 flex items-center justify-between rounded-lg bg-gray-50 px-4 py-3 dark:bg-dark-800">
          <div class="flex items-center gap-3">
            <div
              :class="[
                'h-3 w-3 rounded-full',
                monitorStatus?.running ? 'bg-emerald-500 animate-pulse' : 'bg-gray-400'
              ]"
            ></div>
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ t('admin.payment.monitorService') }}:
              {{ monitorStatus?.running ? t('admin.payment.running') : t('admin.payment.stopped') }}
            </span>
            <span v-if="monitorStatus?.last_check_time" class="text-xs text-gray-500 dark:text-dark-400">
              ({{ t('admin.payment.lastCheck') }}: {{ formatDateTime(monitorStatus.last_check_time) }})
            </span>
          </div>
          <div class="flex gap-2">
            <button
              v-if="!monitorStatus?.running"
              @click="startMonitor"
              :disabled="!monitorStatus?.available"
              class="btn btn-sm btn-primary"
            >
              {{ t('admin.payment.startMonitor') }}
            </button>
            <button
              v-else
              @click="stopMonitor"
              class="btn btn-sm btn-secondary"
            >
              {{ t('admin.payment.stopMonitor') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-trade_no="{ value }">
            <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</code>
          </template>

          <template #cell-user="{ value }">
            <span v-if="value" class="text-sm text-gray-700 dark:text-gray-300">
              {{ value.email || value.username || `ID: ${value.id}` }}
            </span>
            <span v-else class="text-gray-400">-</span>
          </template>

          <template #cell-amount="{ value }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              ${{ value.toFixed(2) }}
            </span>
          </template>

          <template #cell-payment_amount="{ value }">
            <span class="font-mono text-sm text-primary-600 dark:text-primary-400">
              ¥{{ value.toFixed(2) }}
            </span>
          </template>

          <template #cell-credit_amount="{ value, row }">
            <span v-if="row.status === 'paid'" class="font-mono text-sm text-emerald-600 dark:text-emerald-400">
              ${{ (value || row.amount).toFixed(2) }}
            </span>
            <span v-else class="text-gray-400">-</span>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'paid'
                  ? 'badge-success'
                  : value === 'pending'
                    ? 'badge-warning'
                    : value === 'expired'
                      ? 'badge-gray'
                      : 'badge-danger'
              ]"
            >
              {{ t(`admin.payment.status.${value}`) }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDateTime(value) }}
            </span>
          </template>

          <template #cell-paid_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : '-' }}
            </span>
          </template>

          <template #cell-payer_account="{ value }">
            <span v-if="value" class="text-sm text-gray-700 dark:text-gray-300">
              {{ value }}
            </span>
            <span v-else class="text-gray-400">-</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-2">
              <button
                v-if="row.status === 'pending'"
                @click="handleManualConfirm(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-emerald-50 hover:text-emerald-600 dark:hover:bg-emerald-900/20 dark:hover:text-emerald-400"
              >
                <Icon name="checkCircle" size="sm" />
                <span class="text-xs">{{ t('admin.payment.confirm') }}</span>
              </button>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Manual Confirm Dialog -->
    <ConfirmDialog
      :show="showConfirmDialog"
      :title="t('admin.payment.manualConfirm')"
      :message="t('admin.payment.manualConfirmHint', { trade_no: selectedOrder?.trade_no, amount: selectedOrder?.amount?.toFixed(2) })"
      :confirm-text="t('admin.payment.confirmPayment')"
      :cancel-text="t('common.cancel')"
      @confirm="confirmManualPayment"
      @cancel="showConfirmDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { paymentAPI, type PaymentOrder, type PaymentStats, type MonitorStatus } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

// Simple debounce utility
function debounce<T extends (...args: any[]) => any>(fn: T, delay: number): T {
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  return ((...args: Parameters<T>) => {
    if (timeoutId) clearTimeout(timeoutId)
    timeoutId = setTimeout(() => fn(...args), delay)
  }) as T
}

const { t } = useI18n()
const appStore = useAppStore()

// Data
const orders = ref<PaymentOrder[]>([])
const stats = ref<PaymentStats | null>(null)
const monitorStatus = ref<MonitorStatus | null>(null)
const loading = ref(false)
const searchQuery = ref('')

// Filters
const filters = ref({
  status: ''
})

// Pagination
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0
})

// Dialog state
const showConfirmDialog = ref(false)
const selectedOrder = ref<PaymentOrder | null>(null)

// Filter options
const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.payment.allStatus') },
  { value: 'pending', label: t('admin.payment.status.pending') },
  { value: 'paid', label: t('admin.payment.status.paid') },
  { value: 'expired', label: t('admin.payment.status.expired') },
  { value: 'cancelled', label: t('admin.payment.status.cancelled') }
])

// Table columns
const columns = computed(() => [
  { key: 'trade_no', label: t('admin.payment.tradeNo'), width: '180px' },
  { key: 'user', label: t('admin.payment.user'), width: '160px' },
  { key: 'amount', label: t('admin.payment.amount'), width: '90px' },
  { key: 'payment_amount', label: t('admin.payment.paymentAmount'), width: '100px' },
  { key: 'credit_amount', label: t('admin.payment.creditAmount'), width: '100px' },
  { key: 'status', label: t('admin.payment.statusLabel'), width: '90px' },
  { key: 'payer_account', label: t('admin.payment.payerAccount'), width: '160px' },
  { key: 'created_at', label: t('admin.payment.createdAt'), width: '150px' },
  { key: 'paid_at', label: t('admin.payment.paidAt'), width: '150px' },
  { key: 'actions', label: t('common.actions'), width: '80px' }
])

// Load orders
const loadOrders = async () => {
  loading.value = true
  try {
    const response = await paymentAPI.admin.listOrders({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      status: filters.value.status || undefined,
      search: searchQuery.value || undefined
    })
    orders.value = response.items || []
    pagination.value.total = response.total || 0
  } catch (error) {
    console.error('Failed to load orders:', error)
    appStore.showError(t('admin.payment.loadFailed'))
  } finally {
    loading.value = false
  }
}

// Load stats
const loadStats = async () => {
  try {
    stats.value = await paymentAPI.admin.getStats()
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

// Load monitor status
const loadMonitorStatus = async () => {
  try {
    monitorStatus.value = await paymentAPI.admin.getMonitorStatus()
  } catch (error) {
    console.error('Failed to load monitor status:', error)
  }
}

// Search handler with debounce
const handleSearch = debounce(() => {
  pagination.value.page = 1
  loadOrders()
}, 300)

// Pagination handlers
const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadOrders()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  loadOrders()
}

// Manual confirm
const handleManualConfirm = (order: PaymentOrder) => {
  selectedOrder.value = order
  showConfirmDialog.value = true
}

const confirmManualPayment = async () => {
  if (!selectedOrder.value) return

  try {
    await paymentAPI.admin.manualConfirm(selectedOrder.value.id)
    appStore.showSuccess(t('admin.payment.confirmSuccess'))
    showConfirmDialog.value = false
    loadOrders()
    loadStats()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.confirmFailed'))
  }
}

// Monitor controls
const startMonitor = async () => {
  try {
    await paymentAPI.admin.startMonitor()
    appStore.showSuccess(t('admin.payment.monitorStarted'))
    loadMonitorStatus()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.monitorStartFailed'))
  }
}

const stopMonitor = async () => {
  try {
    await paymentAPI.admin.stopMonitor()
    appStore.showSuccess(t('admin.payment.monitorStopped'))
    loadMonitorStatus()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.payment.monitorStopFailed'))
  }
}

onMounted(() => {
  loadOrders()
  loadStats()
  loadMonitorStatus()
})
</script>
