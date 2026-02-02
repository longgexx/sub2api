<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex justify-end gap-3">
          <button
            @click="loadAnnouncements"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button @click="showCreateDialog = true" class="btn btn-primary">
            <Icon name="plus" size="md" class="mr-1" />
            {{ t('admin.announcement.create') }}
          </button>
        </div>
      </template>

      <template #filters>
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="max-w-md flex-1">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.announcement.searchPlaceholder')"
              class="input"
              @input="handleSearch"
            />
          </div>
          <div class="flex gap-2">
            <Select
              v-model="filters.status"
              :options="filterStatusOptions"
              class="w-36"
              @change="loadAnnouncements"
            />
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="announcements" :loading="loading">
          <template #cell-title="{ value, row }">
            <div class="flex items-center gap-2">
              <span :class="['type-indicator', `type-${row.type}`]"></span>
              <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
            </div>
          </template>

          <template #cell-type="{ value }">
            <span :class="['badge', `badge-${value}`]">
              <Icon :name="getTypeIcon(value)" size="xs" class="mr-1" />
              {{ t(`announcement.type.${value}`) }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', getStatusBadgeClass(value)]">
              {{ t(`admin.announcement.status.${value}`) }}
            </span>
          </template>

          <template #cell-priority="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-300">{{ value }}</span>
          </template>

          <template #cell-publish_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : '-' }}
            </span>
          </template>

          <template #cell-expires_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : t('admin.announcement.neverExpires') }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDateTime(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-1">
              <button
                v-if="row.status === 'draft'"
                @click="handlePublish(row)"
                class="action-btn action-btn-success"
                :title="t('admin.announcement.publish')"
              >
                <Icon name="play" size="sm" />
              </button>
              <button
                v-if="row.status === 'published'"
                @click="handleArchive(row)"
                class="action-btn action-btn-warning"
                :title="t('admin.announcement.archive')"
              >
                <Icon name="inbox" size="sm" />
              </button>
              <button
                @click="handleEdit(row)"
                class="action-btn action-btn-default"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="handleDelete(row)"
                class="action-btn action-btn-danger"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
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

    <!-- Create/Edit Dialog -->
    <BaseDialog
      :show="showCreateDialog || showEditDialog"
      :title="showEditDialog ? t('admin.announcement.edit') : t('admin.announcement.create')"
      width="wide"
      @close="closeDialog"
    >
      <form id="announcement-form" @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.announcement.fields.title') }}</label>
          <input
            v-model="form.title"
            type="text"
            required
            class="input"
            :placeholder="t('admin.announcement.placeholders.title')"
          />
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.announcement.fields.type') }}</label>
            <Select v-model="form.type" :options="typeOptions" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.announcement.fields.priority') }}</label>
            <input
              v-model.number="form.priority"
              type="number"
              min="0"
              class="input"
            />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.announcement.fields.content') }}</label>
          <textarea
            v-model="form.content"
            rows="6"
            required
            class="input"
            :placeholder="t('admin.announcement.placeholders.content')"
          ></textarea>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('admin.announcement.contentHint') }}
          </p>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">
              {{ t('admin.announcement.fields.publishAt') }}
              <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
            </label>
            <input
              v-model="form.publish_at_str"
              type="datetime-local"
              class="input"
            />
          </div>
          <div>
            <label class="input-label">
              {{ t('admin.announcement.fields.expiresAt') }}
              <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
            </label>
            <input
              v-model="form.expires_at_str"
              type="datetime-local"
              class="input"
            />
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="closeDialog" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="announcement-form" :disabled="submitting" class="btn btn-primary">
            {{ submitting ? t('common.saving') : (showEditDialog ? t('common.save') : t('common.create')) }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.announcement.deleteTitle')"
      :message="t('admin.announcement.deleteConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { announcementAPI, type Announcement, type AnnouncementType } from '@/api'
import { formatDateTime } from '@/utils/format'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

// State
const announcements = ref<Announcement[]>([])
const loading = ref(false)
const submitting = ref(false)
const searchQuery = ref('')

const filters = reactive({
  status: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

// Dialogs
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)

const editingAnnouncement = ref<Announcement | null>(null)
const deletingAnnouncement = ref<Announcement | null>(null)

// Form
const form = reactive({
  title: '',
  content: '',
  type: 'info' as AnnouncementType,
  priority: 0,
  publish_at_str: '',
  expires_at_str: ''
})

// Options
const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.announcement.allStatus') },
  { value: 'draft', label: t('admin.announcement.status.draft') },
  { value: 'published', label: t('admin.announcement.status.published') },
  { value: 'archived', label: t('admin.announcement.status.archived') }
])

const typeOptions = computed(() => [
  { value: 'info', label: t('announcement.type.info') },
  { value: 'warning', label: t('announcement.type.warning') },
  { value: 'important', label: t('announcement.type.important') }
])

const columns = computed<Column[]>(() => [
  { key: 'title', label: t('admin.announcement.columns.title') },
  { key: 'type', label: t('admin.announcement.columns.type') },
  { key: 'status', label: t('admin.announcement.columns.status') },
  { key: 'priority', label: t('admin.announcement.columns.priority'), sortable: true },
  { key: 'publish_at', label: t('admin.announcement.columns.publishAt'), sortable: true },
  { key: 'expires_at', label: t('admin.announcement.columns.expiresAt'), sortable: true },
  { key: 'created_at', label: t('admin.announcement.columns.createdAt'), sortable: true },
  { key: 'actions', label: t('admin.announcement.columns.actions') }
])

// Helpers
type IconName = 'infoCircle' | 'exclamationTriangle' | 'exclamationCircle'
const getTypeIcon = (type: AnnouncementType): IconName => {
  const icons: Record<AnnouncementType, IconName> = {
    info: 'infoCircle',
    warning: 'exclamationTriangle',
    important: 'exclamationCircle'
  }
  return icons[type] || 'infoCircle'
}

const getStatusBadgeClass = (status: string): string => {
  const classes: Record<string, string> = {
    draft: 'badge-gray',
    published: 'badge-success',
    archived: 'badge-warning'
  }
  return classes[status] || 'badge-gray'
}

const toDateTimeLocalValue = (isoString: string | null): string => {
  if (!isoString) return ''
  const date = new Date(isoString)
  if (isNaN(date.getTime())) return ''
  const pad2 = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}T${pad2(date.getHours())}:${pad2(date.getMinutes())}`
}

// API calls
let abortController: AbortController | null = null

const loadAnnouncements = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true

  try {
    const response = await announcementAPI.admin.list({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      search: searchQuery.value || undefined
    })
    if (currentController.signal.aborted) return

    announcements.value = response.items
    pagination.total = response.total
  } catch (error: any) {
    if (currentController.signal.aborted || error?.name === 'AbortError') return
    appStore.showError(t('admin.announcement.failedToLoad'))
    console.error('Error loading announcements:', error)
  } finally {
    if (abortController === currentController && !currentController.signal.aborted) {
      loading.value = false
      abortController = null
    }
  }
}

let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadAnnouncements()
  }, 300)
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadAnnouncements()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadAnnouncements()
}

// Form handling
const resetForm = () => {
  form.title = ''
  form.content = ''
  form.type = 'info'
  form.priority = 0
  form.publish_at_str = ''
  form.expires_at_str = ''
}

const closeDialog = () => {
  showCreateDialog.value = false
  showEditDialog.value = false
  editingAnnouncement.value = null
  resetForm()
}

const handleEdit = (announcement: Announcement) => {
  editingAnnouncement.value = announcement
  form.title = announcement.title
  form.content = announcement.content
  form.type = announcement.type
  form.priority = announcement.priority
  // datetime-local expects a local-time value (YYYY-MM-DDTHH:mm), not UTC (toISOString()).
  form.publish_at_str = toDateTimeLocalValue(announcement.publish_at)
  form.expires_at_str = toDateTimeLocalValue(announcement.expires_at)
  showEditDialog.value = true
}

// Form validation
const validateForm = (): string | null => {
  // Validate title
  if (!form.title.trim()) {
    return t('admin.announcement.validation.titleRequired')
  }
  if (form.title.length > 255) {
    return t('admin.announcement.validation.titleTooLong')
  }

  // Validate content
  if (!form.content.trim()) {
    return t('admin.announcement.validation.contentRequired')
  }

  // Validate dates
  if (form.publish_at_str && form.expires_at_str) {
    const publishAt = new Date(form.publish_at_str)
    const expiresAt = new Date(form.expires_at_str)
    if (expiresAt <= publishAt) {
      return t('admin.announcement.validation.expiresBeforePublish')
    }
  }

  return null
}

const handleSubmit = async () => {
  // Validate form before submitting
  const validationError = validateForm()
  if (validationError) {
    appStore.showError(validationError)
    return
  }

  submitting.value = true
  try {
    const payload = {
      title: form.title,
      content: form.content,
      type: form.type,
      priority: form.priority,
      publish_at: form.publish_at_str ? Math.floor(new Date(form.publish_at_str).getTime() / 1000) : null,
      expires_at: form.expires_at_str ? Math.floor(new Date(form.expires_at_str).getTime() / 1000) : null
    }

    if (showEditDialog.value && editingAnnouncement.value) {
      await announcementAPI.admin.update(editingAnnouncement.value.id, payload)
      appStore.showSuccess(t('admin.announcement.updated'))
    } else {
      await announcementAPI.admin.create(payload)
      appStore.showSuccess(t('admin.announcement.created'))
    }
    closeDialog()
    loadAnnouncements()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.announcement.failedToSave'))
  } finally {
    submitting.value = false
  }
}

// Publish/Archive
const handlePublish = async (announcement: Announcement) => {
  try {
    await announcementAPI.admin.publish(announcement.id)
    appStore.showSuccess(t('admin.announcement.published'))
    loadAnnouncements()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.announcement.failedToPublish'))
  }
}

const handleArchive = async (announcement: Announcement) => {
  try {
    await announcementAPI.admin.archive(announcement.id)
    appStore.showSuccess(t('admin.announcement.archived'))
    loadAnnouncements()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.announcement.failedToArchive'))
  }
}

// Delete
const handleDelete = (announcement: Announcement) => {
  deletingAnnouncement.value = announcement
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingAnnouncement.value) return

  try {
    await announcementAPI.admin.delete(deletingAnnouncement.value.id)
    appStore.showSuccess(t('admin.announcement.deleted'))
    showDeleteDialog.value = false
    deletingAnnouncement.value = null
    loadAnnouncements()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.announcement.failedToDelete'))
  }
}

onMounted(() => {
  loadAnnouncements()
})

onUnmounted(() => {
  clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>

<style scoped>
.type-indicator {
  @apply h-2 w-2 rounded-full;
}

.type-info {
  @apply bg-blue-500;
}

.type-warning {
  @apply bg-yellow-500;
}

.type-important {
  @apply bg-red-500;
}

.badge-info {
  @apply bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400;
}

.badge-warning {
  @apply bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400;
}

.badge-important {
  @apply bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400;
}

.action-btn {
  @apply flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors;
}

.action-btn-default {
  @apply hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300;
}

.action-btn-success {
  @apply hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400;
}

.action-btn-warning {
  @apply hover:bg-yellow-50 hover:text-yellow-600 dark:hover:bg-yellow-900/20 dark:hover:text-yellow-400;
}

.action-btn-danger {
  @apply hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400;
}
</style>
