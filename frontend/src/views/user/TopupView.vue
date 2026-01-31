<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Current Balance Card -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div
            class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm"
          >
            <Icon name="creditCard" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">{{ t('payment.currentBalance') }}</p>
          <p class="mt-2 text-4xl font-bold text-white">
            ${{ user?.balance?.toFixed(2) || '0.00' }}
          </p>
        </div>
      </div>

      <!-- Payment Not Enabled -->
      <div
        v-if="!loading && !config?.enabled"
        class="card border-amber-200 bg-amber-50 dark:border-amber-800/50 dark:bg-amber-900/20"
      >
        <div class="p-6">
          <div class="flex items-start gap-4">
            <div
              class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-amber-100 dark:bg-amber-900/30"
            >
              <Icon name="exclamationCircle" size="md" class="text-amber-600 dark:text-amber-400" />
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-semibold text-amber-800 dark:text-amber-300">
                {{ t('payment.notEnabled') }}
              </h3>
              <p class="mt-2 text-sm text-amber-700 dark:text-amber-400">
                {{ t('payment.notEnabledHint') }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Topup Form -->
      <div v-if="config?.enabled && !currentOrder" class="card">
        <div class="p-6">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('payment.topupAmount') }}
          </h2>

          <!-- Preset Amounts -->
          <div class="mb-4 grid grid-cols-4 gap-3">
            <button
              v-for="amount in presetAmounts"
              :key="amount"
              @click="selectedAmount = amount"
              :class="[
                'rounded-xl border-2 py-3 text-center font-semibold transition-all',
                selectedAmount === amount
                  ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                  : 'border-gray-200 bg-white text-gray-700 hover:border-primary-300 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:border-primary-600'
              ]"
            >
              ${{ amount }}
            </button>
          </div>

          <!-- Custom Amount -->
          <div class="mb-6">
            <label for="customAmount" class="input-label">
              {{ t('payment.customAmount') }}
            </label>
            <div class="relative mt-1">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                <span class="text-gray-500 dark:text-dark-400">$</span>
              </div>
              <input
                id="customAmount"
                v-model.number="customAmount"
                type="number"
                :min="config?.min_amount || 1"
                :max="config?.max_amount || 10000"
                step="0.01"
                :placeholder="t('payment.customAmountPlaceholder')"
                class="input py-3 pl-8"
                @focus="selectedAmount = null"
              />
            </div>
            <p class="input-hint">
              {{ t('payment.amountRange', { min: config?.min_amount || 1, max: config?.max_amount || 10000 }) }}
            </p>
            <!-- Show estimated payment amount when rate coefficient is not 1 -->
            <p
              v-if="finalAmount && config?.rate_coefficient && config.rate_coefficient !== 1"
              class="mt-2 text-sm text-primary-600 dark:text-primary-400"
            >
              {{ t('payment.estimatedPayment', { amount: (finalAmount * config.rate_coefficient).toFixed(2) }) }}
            </p>
          </div>

          <button
            @click="createOrder"
            :disabled="!finalAmount || submitting"
            class="btn btn-primary w-full py-3"
          >
            <svg
              v-if="submitting"
              class="-ml-1 mr-2 h-5 w-5 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ submitting ? t('payment.creating') : t('payment.createOrder') }}
          </button>
        </div>
      </div>

      <!-- Payment Order (QR Code) -->
      <div v-if="currentOrder" class="card">
        <div class="p-6">
          <div class="text-center">
            <h2 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('payment.scanToPay') }}
            </h2>
            <p class="mb-4 text-sm text-gray-500 dark:text-dark-400">
              {{ t('payment.scanToPayHint') }}
            </p>

            <!-- QR Code -->
            <div class="mb-4 inline-block rounded-2xl bg-white p-4 shadow-lg">
              <img
                v-if="config?.qr_code_url"
                :src="config.qr_code_url"
                alt="Payment QR Code"
                class="h-48 w-48 object-contain"
              />
              <div v-else class="flex h-48 w-48 items-center justify-center bg-gray-100">
                <Icon name="creditCard" size="xl" class="text-gray-400" />
              </div>
            </div>

            <!-- Payment Amount (Highlighted) -->
            <div class="mb-4 rounded-xl bg-amber-50 p-4 dark:bg-amber-900/30 border-2 border-amber-400 dark:border-amber-600">
              <p class="text-sm font-medium text-amber-700 dark:text-amber-300">
                {{ t('payment.paymentAmount') }}
              </p>
              <p class="text-3xl font-bold text-amber-700 dark:text-amber-300">
                ¥{{ currentOrder.payment_amount.toFixed(2) }}
              </p>
              <p class="mt-2 text-sm font-semibold text-amber-600 dark:text-amber-400">
                ⚠️ {{ t('payment.payExactAmount') }}
              </p>
              <p class="mt-1 text-xs text-amber-600 dark:text-amber-400">
                {{ t('payment.actualCreditHint') }}
              </p>
            </div>

            <!-- Countdown Timer -->
            <div class="mb-4">
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('payment.orderExpires') }}
              </p>
              <p
                :class="[
                  'text-2xl font-bold',
                  remainingTime <= 60 ? 'text-red-600 dark:text-red-400' : 'text-gray-900 dark:text-white'
                ]"
              >
                {{ formatCountdown(remainingTime) }}
              </p>
            </div>

            <!-- Order Status -->
            <div class="mb-4">
              <span
                :class="[
                  'inline-flex items-center rounded-full px-3 py-1 text-sm font-medium',
                  currentOrder.status === 'pending'
                    ? 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
                    : currentOrder.status === 'paid'
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400'
                      : 'bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-400'
                ]"
              >
                {{ t(`payment.status.${currentOrder.status}`) }}
              </span>
            </div>

            <!-- Cancel Button -->
            <button
              v-if="currentOrder.status === 'pending'"
              @click="cancelCurrentOrder"
              :disabled="cancelling"
              class="btn btn-secondary"
            >
              {{ cancelling ? t('payment.cancelling') : t('payment.cancelOrder') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Payment Success -->
      <transition name="fade">
        <div
          v-if="paymentSuccess"
          class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30"
              >
                <Icon name="checkCircle" size="md" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">
                  {{ t('payment.paymentSuccess') }}
                </h3>
                <p class="mt-2 text-sm text-emerald-700 dark:text-emerald-400">
                  {{ t('payment.balanceUpdated', { amount: lastPaidAmount?.toFixed(2) }) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Information Card -->
      <div
        class="card border-primary-200 bg-primary-50 dark:border-primary-800/50 dark:bg-primary-900/20"
      >
        <div class="p-6">
          <div class="flex items-start gap-4">
            <div
              class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30"
            >
              <Icon name="infoCircle" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-semibold text-primary-800 dark:text-primary-300">
                {{ t('payment.instructions') }}
              </h3>
              <ul
                class="mt-2 list-inside list-disc space-y-1 text-sm text-primary-700 dark:text-primary-400"
              >
                <li>{{ t('payment.instruction1') }}</li>
                <li>{{ t('payment.instruction2') }}</li>
                <li>{{ t('payment.instruction3') }}</li>
                <li>{{ t('payment.instruction4') }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Order History -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('payment.orderHistory') }}
          </h2>
        </div>
        <div class="p-6">
          <!-- Loading State -->
          <div v-if="loadingHistory" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          </div>

          <!-- History List -->
          <div v-else-if="orderHistory.length > 0" class="space-y-3">
            <div
              v-for="order in orderHistory"
              :key="order.id"
              class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
            >
              <div class="flex items-center gap-4">
                <div
                  :class="[
                    'flex h-10 w-10 items-center justify-center rounded-xl',
                    order.status === 'paid'
                      ? 'bg-emerald-100 dark:bg-emerald-900/30'
                      : order.status === 'pending'
                        ? 'bg-amber-100 dark:bg-amber-900/30'
                        : 'bg-gray-100 dark:bg-dark-700'
                  ]"
                >
                  <Icon
                    :name="order.status === 'paid' ? 'checkCircle' : order.status === 'pending' ? 'clock' : 'xCircle'"
                    size="md"
                    :class="
                      order.status === 'paid'
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : order.status === 'pending'
                          ? 'text-amber-600 dark:text-amber-400'
                          : 'text-gray-400 dark:text-dark-500'
                    "
                  />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    ${{ (order.status === 'paid' ? (order.credit_amount || order.amount) : order.amount).toFixed(2) }}
                    <span class="ml-1 text-xs text-gray-400">({{ t('payment.actualPaid') }}: ¥{{ order.payment_amount.toFixed(2) }})</span>
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{ formatDateTime(order.created_at) }}
                  </p>
                  <p v-if="order.payer_account" class="text-xs text-primary-500 dark:text-primary-400">
                    {{ order.payer_account }}
                  </p>
                </div>
              </div>
              <div class="text-right">
                <span
                  :class="[
                    'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
                    order.status === 'paid'
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400'
                      : order.status === 'pending'
                        ? 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
                        : 'bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-400'
                  ]"
                >
                  {{ t(`payment.status.${order.status}`) }}
                </span>
                <p class="mt-1 font-mono text-xs text-gray-400 dark:text-dark-500">
                  {{ order.trade_no.slice(0, 12) }}...
                </p>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="empty-state py-8">
            <div
              class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800"
            >
              <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
            </div>
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('payment.noOrders') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { paymentAPI, type PaymentOrder, type PaymentConfig } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

// Config and state
const loading = ref(true)
const config = ref<PaymentConfig | null>(null)
const presetAmounts = [10, 50, 100, 500]
const selectedAmount = ref<number | null>(null)
const customAmount = ref<number | null>(null)
const submitting = ref(false)
const cancelling = ref(false)

// Current order
const currentOrder = ref<PaymentOrder | null>(null)
const remainingTime = ref(0)
const paymentSuccess = ref(false)
const lastPaidAmount = ref<number | null>(null)

// History
const orderHistory = ref<PaymentOrder[]>([])
const loadingHistory = ref(false)

// Polling interval
let pollInterval: ReturnType<typeof setInterval> | null = null
let countdownInterval: ReturnType<typeof setInterval> | null = null

const finalAmount = computed(() => {
  return selectedAmount.value || customAmount.value
})

const fetchConfig = async () => {
  try {
    config.value = await paymentAPI.getConfig()
  } catch (error) {
    console.error('Failed to fetch payment config:', error)
  } finally {
    loading.value = false
  }
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    const response = await paymentAPI.listOrders({ page_size: 10 })
    orderHistory.value = response.items || []
  } catch (error) {
    console.error('Failed to fetch order history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const createOrder = async () => {
  if (!finalAmount.value) return

  submitting.value = true
  try {
    currentOrder.value = await paymentAPI.createOrder(finalAmount.value)
    startCountdown()
    startPolling()
    appStore.showSuccess(t('payment.orderCreated'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('payment.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

const cancelCurrentOrder = async () => {
  if (!currentOrder.value) return

  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(currentOrder.value.trade_no)
    stopPolling()
    stopCountdown()
    currentOrder.value = null
    appStore.showSuccess(t('payment.orderCancelled'))
    fetchHistory()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('payment.cancelFailed'))
  } finally {
    cancelling.value = false
  }
}

const startCountdown = () => {
  if (!currentOrder.value) return

  const expiredAt = new Date(currentOrder.value.expired_at).getTime()
  const updateCountdown = () => {
    const now = Date.now()
    remainingTime.value = Math.max(0, Math.floor((expiredAt - now) / 1000))

    if (remainingTime.value <= 0) {
      stopCountdown()
      stopPolling()
      currentOrder.value = null
      fetchHistory()
    }
  }

  updateCountdown()
  countdownInterval = setInterval(updateCountdown, 1000)
}

const stopCountdown = () => {
  if (countdownInterval) {
    clearInterval(countdownInterval)
    countdownInterval = null
  }
}

const startPolling = () => {
  pollInterval = setInterval(async () => {
    if (!currentOrder.value) {
      stopPolling()
      return
    }

    try {
      const order = await paymentAPI.getOrder(currentOrder.value.trade_no)
      currentOrder.value = order

      if (order.status === 'paid') {
        stopPolling()
        stopCountdown()
        paymentSuccess.value = true
        lastPaidAmount.value = order.credit_amount || order.amount
        currentOrder.value = null

        // Refresh user data
        await authStore.refreshUser()
        fetchHistory()

        // Auto-hide success message after 5 seconds
        setTimeout(() => {
          paymentSuccess.value = false
        }, 5000)
      } else if (order.status !== 'pending') {
        stopPolling()
        stopCountdown()
        currentOrder.value = null
        fetchHistory()
      }
    } catch (error) {
      console.error('Failed to poll order status:', error)
    }
  }, 3000) // Poll every 3 seconds
}

const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

const formatCountdown = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

onMounted(() => {
  fetchConfig()
  fetchHistory()
})

onUnmounted(() => {
  stopPolling()
  stopCountdown()
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
