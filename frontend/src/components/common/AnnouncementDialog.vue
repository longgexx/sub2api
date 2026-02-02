<template>
  <BaseDialog
    :show="show"
    :title="$t('announcement.dialog.title')"
    width="wide"
    :close-on-escape="true"
    :close-on-click-outside="false"
    @close="handleClose"
  >
    <div class="announcement-dialog">
      <!-- Empty state -->
      <div v-if="announcements.length === 0" class="empty-state">
        <Icon name="inbox" size="xl" class="text-gray-400 dark:text-dark-500" />
        <p class="mt-2 text-gray-500 dark:text-dark-400">{{ $t('announcement.dialog.empty') }}</p>
      </div>

      <!-- Announcement list -->
      <div v-else class="announcement-list">
        <div
          v-for="announcement in announcements"
          :key="announcement.id"
          :class="[
            'announcement-item',
            `announcement-${announcement.type}`,
            { 'announcement-read': announcement.is_read }
          ]"
        >
          <!-- Header -->
          <div class="announcement-header">
            <div class="announcement-title-row">
              <span :class="['announcement-type-badge', `badge-${announcement.type}`]">
                <Icon :name="getTypeIcon(announcement.type)" size="sm" />
                {{ $t(`announcement.type.${announcement.type}`) }}
              </span>
              <h4 class="announcement-title">{{ announcement.title }}</h4>
            </div>
            <div class="announcement-meta">
              <span class="announcement-date">
                {{ formatDate(announcement.created_at) }}
              </span>
              <button
                v-if="!announcement.is_read"
                class="mark-read-btn"
                @click="handleMarkAsRead(announcement.id)"
                :disabled="markingRead"
              >
                <Icon name="check" size="sm" />
                {{ $t('announcement.dialog.markRead') }}
              </button>
              <span v-else class="read-badge">
                <Icon name="checkCircle" size="sm" />
                {{ $t('announcement.dialog.read') }}
              </span>
            </div>
          </div>

          <!-- Content -->
          <div class="announcement-content" v-html="formatContent(announcement.content)"></div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <button
          v-if="hasUnread"
          class="btn btn-secondary"
          @click="handleMarkAllAsRead"
          :disabled="markingRead"
        >
          <Icon name="check" size="sm" />
          {{ $t('announcement.dialog.markAllRead') }}
        </button>
        <button class="btn btn-primary" @click="handleClose">
          {{ $t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { announcementAPI, type AnnouncementWithReadStatus, type AnnouncementType } from '@/api'
import { useAppStore } from '@/stores'

interface Props {
  show: boolean
  announcements: AnnouncementWithReadStatus[]
}

interface Emits {
  (e: 'close'): void
  (e: 'update:announcements', announcements: AnnouncementWithReadStatus[]): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const markingRead = ref(false)

const hasUnread = computed(() => props.announcements.some((a) => !a.is_read))

const getTypeIcon = (type: AnnouncementType): 'infoCircle' | 'exclamationTriangle' | 'exclamationCircle' => {
  const icons: Record<AnnouncementType, 'infoCircle' | 'exclamationTriangle' | 'exclamationCircle'> = {
    info: 'infoCircle',
    warning: 'exclamationTriangle',
    important: 'exclamationCircle'
  }
  return icons[type] || 'infoCircle'
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return ''
  // Use i18n locale for date formatting
  const locale = t('locale') === 'zh' ? 'zh-CN' : 'en-US'
  return date.toLocaleDateString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatContent = (content: string): string => {
  if (!content) return ''
  // First escape HTML to prevent XSS
  let html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')

  // Then apply simple formatting
  html = html
    .replace(/\n/g, '<br>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')

  // Sanitize with DOMPurify as an extra layer of protection
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['br', 'strong', 'em', 'code'],
    ALLOWED_ATTR: []
  })
}

const handleClose = () => {
  emit('close')
}

const handleMarkAsRead = async (id: number) => {
  markingRead.value = true
  try {
    await announcementAPI.markAsRead(id)
    const updated = props.announcements.map((a) => (a.id === id ? { ...a, is_read: true } : a))
    emit('update:announcements', updated)
  } catch {
    appStore.showError(t('announcement.dialog.markReadError'))
  } finally {
    markingRead.value = false
  }
}

const handleMarkAllAsRead = async () => {
  markingRead.value = true
  try {
    await announcementAPI.markAllAsRead()
    const updated = props.announcements.map((a) => ({ ...a, is_read: true }))
    emit('update:announcements', updated)
    appStore.showSuccess(t('announcement.dialog.markAllReadSuccess'))
  } catch {
    appStore.showError(t('announcement.dialog.markReadError'))
  } finally {
    markingRead.value = false
  }
}
</script>

<style scoped>
.announcement-dialog {
  max-height: 60vh;
  overflow-y: auto;
}

.empty-state {
  @apply flex flex-col items-center justify-center py-12;
}

.announcement-list {
  @apply space-y-4;
}

.announcement-item {
  @apply rounded-lg border p-4 transition-colors;
  @apply border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800;
}

.announcement-item.announcement-read {
  @apply opacity-70;
}

.announcement-info {
  @apply border-l-4 border-l-blue-500;
}

.announcement-warning {
  @apply border-l-4 border-l-yellow-500;
}

.announcement-important {
  @apply border-l-4 border-l-red-500;
}

.announcement-header {
  @apply mb-3 flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between;
}

.announcement-title-row {
  @apply flex flex-wrap items-center gap-2;
}

.announcement-type-badge {
  @apply inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium;
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

.announcement-title {
  @apply text-base font-semibold text-gray-900 dark:text-dark-100;
}

.announcement-meta {
  @apply flex flex-shrink-0 items-center gap-3;
}

.announcement-date {
  @apply text-xs text-gray-500 dark:text-dark-400;
}

.mark-read-btn {
  @apply inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium transition-colors;
  @apply text-primary-600 hover:bg-primary-50 dark:text-primary-400 dark:hover:bg-primary-900/20;
}

.mark-read-btn:disabled {
  @apply cursor-not-allowed opacity-50;
}

.read-badge {
  @apply inline-flex items-center gap-1 text-xs text-green-600 dark:text-green-400;
}

.announcement-content {
  @apply text-sm leading-relaxed text-gray-700 dark:text-dark-300;
}

.announcement-content :deep(code) {
  @apply rounded bg-gray-100 px-1 py-0.5 font-mono text-xs dark:bg-dark-700;
}

.dialog-footer {
  @apply flex justify-end gap-3;
}

.btn {
  @apply inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors;
}

.btn-primary {
  @apply bg-primary-600 text-white hover:bg-primary-700;
}

.btn-secondary {
  @apply bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600;
}

.btn:disabled {
  @apply cursor-not-allowed opacity-50;
}
</style>
