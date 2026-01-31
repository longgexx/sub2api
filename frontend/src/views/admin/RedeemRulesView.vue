<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex justify-end gap-3">
          <button
            @click="loadRules"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button @click="openCreateDialog" class="btn btn-primary">
            {{ t('admin.redeemRules.createRule') }}
          </button>
        </div>
      </template>

      <template #filters>
        <div class="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-800/50 dark:bg-amber-900/20">
          <div class="flex items-start gap-3">
            <Icon name="infoCircle" size="md" class="mt-0.5 text-amber-600 dark:text-amber-400" />
            <div class="text-sm text-amber-700 dark:text-amber-400">
              <p class="font-medium">{{ t('admin.redeemRules.description') }}</p>
              <p class="mt-1 text-xs opacity-80">{{ t('admin.redeemRules.descriptionDetail') }}</p>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="rules" :loading="loading">
          <template #cell-type="{ value }">
            <span
              :class="[
                'badge',
                value === 'balance' ? 'badge-success' : 'badge-primary'
              ]"
            >
              {{ t('admin.redeem.types.' + value) }}
            </span>
          </template>

          <template #cell-trigger_value="{ value, row }">
            <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">
              {{ row.type === 'balance' ? `$${value.toFixed(2)}` : value }}
            </span>
          </template>

          <template #cell-fallback_value="{ value, row }">
            <span class="font-mono text-sm font-medium text-amber-600 dark:text-amber-400">
              {{ row.type === 'balance' ? `$${value.toFixed(2)}` : value }}
            </span>
          </template>

          <template #cell-max_times_per_user="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value }} {{ t('admin.redeemRules.times') }}
            </span>
          </template>

          <template #cell-is_active="{ value }">
            <span
              :class="[
                'badge',
                value ? 'badge-success' : 'badge-gray'
              ]"
            >
              {{ value ? t('common.enabled') : t('common.disabled') }}
            </span>
          </template>

          <template #cell-description="{ value }">
            <span class="text-sm text-gray-500 dark:text-gray-400">
              {{ value || '-' }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-2">
              <button
                @click="openEditDialog(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-300"
              >
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit') }}</span>
              </button>
              <button
                @click="handleToggleActive(row)"
                :class="[
                  'flex flex-col items-center gap-0.5 rounded-lg p-1.5 transition-colors',
                  row.is_active
                    ? 'text-amber-500 hover:bg-amber-50 hover:text-amber-600 dark:hover:bg-amber-900/20'
                    : 'text-green-500 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20'
                ]"
              >
                <Icon :name="row.is_active ? 'eyeOff' : 'eye'" size="sm" />
                <span class="text-xs">{{ row.is_active ? t('common.disable') : t('common.enable') }}</span>
              </button>
              <button
                @click="handleDelete(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              >
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('common.delete') }}</span>
              </button>
            </div>
          </template>
        </DataTable>
      </template>
    </TablePageLayout>

    <!-- Create/Edit Dialog -->
    <Teleport to="body">
      <div v-if="showDialog" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="closeDialog"></div>
        <div class="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-dark-800">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ editingRule ? t('admin.redeemRules.editRule') : t('admin.redeemRules.createRule') }}
          </h2>
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div v-if="!editingRule">
              <label class="input-label">{{ t('admin.redeemRules.codeType') }}</label>
              <Select v-model="form.type" :options="typeOptions" />
            </div>
            <div v-else>
              <label class="input-label">{{ t('admin.redeemRules.codeType') }}</label>
              <p class="text-sm text-gray-600 dark:text-gray-400">
                {{ t('admin.redeem.types.' + form.type) }}
              </p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.redeemRules.triggerValue') }}</label>
              <div class="relative">
                <span v-if="form.type === 'balance'" class="absolute inset-y-0 left-3 flex items-center text-gray-500">$</span>
                <input
                  v-model.number="form.trigger_value"
                  type="number"
                  :step="form.type === 'balance' ? '0.01' : '1'"
                  min="0.01"
                  required
                  class="input"
                  :class="{ 'pl-7': form.type === 'balance' }"
                />
              </div>
              <p class="input-hint">{{ t('admin.redeemRules.triggerValueHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.redeemRules.maxTimesPerUser') }}</label>
              <input
                v-model.number="form.max_times_per_user"
                type="number"
                min="1"
                required
                class="input"
              />
              <p class="input-hint">{{ t('admin.redeemRules.maxTimesHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.redeemRules.fallbackValue') }}</label>
              <div class="relative">
                <span v-if="form.type === 'balance'" class="absolute inset-y-0 left-3 flex items-center text-gray-500">$</span>
                <input
                  v-model.number="form.fallback_value"
                  type="number"
                  :step="form.type === 'balance' ? '0.01' : '1'"
                  min="0"
                  required
                  class="input"
                  :class="{ 'pl-7': form.type === 'balance' }"
                />
              </div>
              <p class="input-hint">{{ t('admin.redeemRules.fallbackValueHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.redeemRules.ruleDescription') }}</label>
              <textarea
                v-model="form.description"
                rows="2"
                class="input"
                :placeholder="t('admin.redeemRules.descriptionPlaceholder')"
              ></textarea>
            </div>

            <div class="flex items-center">
              <input
                id="is_active"
                v-model="form.is_active"
                type="checkbox"
                class="checkbox"
              />
              <label for="is_active" class="ml-2 text-sm text-gray-700 dark:text-gray-300">
                {{ t('admin.redeemRules.enableRule') }}
              </label>
            </div>

            <div class="flex justify-end gap-3 pt-2">
              <button type="button" @click="closeDialog" class="btn btn-secondary">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" :disabled="submitting" class="btn btn-primary">
                {{ submitting ? t('common.saving') : t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.redeemRules.deleteRule')"
      :message="t('admin.redeemRules.deleteRuleConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { RedeemRule } from '@/api/admin/redeemRules'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const rules = ref<RedeemRule[]>([])
const loading = ref(false)
const submitting = ref(false)
const showDialog = ref(false)
const showDeleteDialog = ref(false)
const editingRule = ref<RedeemRule | null>(null)
const deletingRule = ref<RedeemRule | null>(null)

const form = reactive({
  type: 'balance',
  trigger_value: 5,
  max_times_per_user: 1,
  fallback_value: 2,
  is_active: true,
  description: ''
})

const columns = computed<Column[]>(() => [
  { key: 'type', label: t('admin.redeemRules.columns.type') },
  { key: 'trigger_value', label: t('admin.redeemRules.columns.triggerValue') },
  { key: 'max_times_per_user', label: t('admin.redeemRules.columns.maxTimes') },
  { key: 'fallback_value', label: t('admin.redeemRules.columns.fallbackValue') },
  { key: 'is_active', label: t('admin.redeemRules.columns.status') },
  { key: 'description', label: t('admin.redeemRules.columns.description') },
  { key: 'actions', label: t('common.actions') }
])

const typeOptions = computed(() => [
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') }
])

const loadRules = async () => {
  loading.value = true
  try {
    rules.value = await adminAPI.redeemRules.list()
  } catch (error) {
    appStore.showError(t('admin.redeemRules.failedToLoad'))
    console.error('Error loading redeem rules:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.type = 'balance'
  form.trigger_value = 5
  form.max_times_per_user = 1
  form.fallback_value = 2
  form.is_active = true
  form.description = ''
}

const openCreateDialog = () => {
  editingRule.value = null
  resetForm()
  showDialog.value = true
}

const openEditDialog = (rule: RedeemRule) => {
  editingRule.value = rule
  form.type = rule.type
  form.trigger_value = rule.trigger_value
  form.max_times_per_user = rule.max_times_per_user
  form.fallback_value = rule.fallback_value
  form.is_active = rule.is_active
  form.description = rule.description || ''
  showDialog.value = true
}

const closeDialog = () => {
  showDialog.value = false
  editingRule.value = null
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (editingRule.value) {
      await adminAPI.redeemRules.update(editingRule.value.id, {
        trigger_value: form.trigger_value,
        max_times_per_user: form.max_times_per_user,
        fallback_value: form.fallback_value,
        is_active: form.is_active,
        description: form.description || undefined
      })
      appStore.showSuccess(t('admin.redeemRules.ruleUpdated'))
    } else {
      await adminAPI.redeemRules.create({
        type: form.type,
        trigger_value: form.trigger_value,
        max_times_per_user: form.max_times_per_user,
        fallback_value: form.fallback_value,
        is_active: form.is_active,
        description: form.description || undefined
      })
      appStore.showSuccess(t('admin.redeemRules.ruleCreated'))
    }
    closeDialog()
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeemRules.failedToSave'))
    console.error('Error saving rule:', error)
  } finally {
    submitting.value = false
  }
}

const handleToggleActive = async (rule: RedeemRule) => {
  try {
    await adminAPI.redeemRules.update(rule.id, {
      is_active: !rule.is_active
    })
    appStore.showSuccess(
      rule.is_active
        ? t('admin.redeemRules.ruleDisabled')
        : t('admin.redeemRules.ruleEnabled')
    )
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeemRules.failedToUpdate'))
    console.error('Error toggling rule:', error)
  }
}

const handleDelete = (rule: RedeemRule) => {
  deletingRule.value = rule
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingRule.value) return

  try {
    await adminAPI.redeemRules.delete(deletingRule.value.id)
    appStore.showSuccess(t('admin.redeemRules.ruleDeleted'))
    showDeleteDialog.value = false
    deletingRule.value = null
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeemRules.failedToDelete'))
    console.error('Error deleting rule:', error)
  }
}

onMounted(() => {
  loadRules()
})
</script>
