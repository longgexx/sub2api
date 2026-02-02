<template>
  <AppLayout>
    <div class="announcements-page">
      <!-- Page Header -->
      <div class="page-header">
        <div class="header-title">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-dark-100">
            {{ $t('announcement.page.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ $t('announcement.page.description') }}
          </p>
        </div>
        <div class="header-actions">
          <!-- Type Filter -->
          <select
            v-model="filterType"
            class="filter-select"
          >
            <option value="all">{{ $t('announcement.page.filterAll') }}</option>
            <option value="info">{{ $t('announcement.type.info') }}</option>
            <option value="warning">{{ $t('announcement.type.warning') }}</option>
            <option value="important">{{ $t('announcement.type.important') }}</option>
          </select>
          <!-- Mark All Read Button -->
          <button
            v-if="hasUnread"
            class="btn btn-secondary"
            :disabled="markingRead"
            @click="handleMarkAllAsRead"
          >
            <Icon name="check" size="sm" />
            {{ $t('announcement.dialog.markAllRead') }}
          </button>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredAnnouncements.length === 0" class="empty-state">
        <Icon name="inbox" size="xl" class="text-gray-400 dark:text-dark-500" />
        <p class="mt-2 text-gray-500 dark:text-dark-400">
          {{ $t('announcement.page.empty') }}
        </p>
      </div>

      <!-- Announcement List -->
      <div v-else class="announcement-list">
        <div
          v-for="announcement in processedAnnouncements"
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
                :disabled="markingReadIds.has(announcement.id)"
                @click="handleMarkAsRead(announcement.id)"
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
          <div class="announcement-content" v-html="announcement.sanitizedContent"></div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { announcementAPI, type AnnouncementWithReadStatus, type AnnouncementType } from '@/api'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const announcements = ref<AnnouncementWithReadStatus[]>([])
const loading = ref(false)
const markingRead = ref(false)
const markingReadIds = ref<Set<number>>(new Set())
const filterType = ref<'all' | AnnouncementType>('all')

const filteredAnnouncements = computed(() => {
  let result = announcements.value
  if (filterType.value !== 'all') {
    result = result.filter(a => a.type === filterType.value)
  }
  // 按优先级降序，然后按创建时间降序（带防御性处理）
  return [...result].sort((a, b) => {
    const pa = a.priority ?? 0
    const pb = b.priority ?? 0
    if (pb !== pa) {
      return pb - pa
    }
    const ta = Date.parse(a.created_at) || 0
    const tb = Date.parse(b.created_at) || 0
    return tb - ta
  })
})

// 预处理公告内容，避免每次渲染重复执行 sanitize
const processedAnnouncements = computed(() => {
  return filteredAnnouncements.value.map(a => ({
    ...a,
    sanitizedContent: formatContent(a.content)
  }))
})

const hasUnread = computed(() => announcements.value.some(a => !a.is_read))

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
  let html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')

  html = html
    .replace(/\n/g, '<br>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')

  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['br', 'strong', 'em', 'code'],
    ALLOWED_ATTR: []
  })
}

const loadAnnouncements = async () => {
  loading.value = true
  try {
    announcements.value = await announcementAPI.getActiveAnnouncements()
  } catch {
    appStore.showError(t('announcement.page.loadError'))
  } finally {
    loading.value = false
  }
}

const handleMarkAsRead = async (id: number) => {
  markingReadIds.value.add(id)
  try {
    await announcementAPI.markAsRead(id)
    const idx = announcements.value.findIndex(a => a.id === id)
    if (idx !== -1) {
      announcements.value[idx] = { ...announcements.value[idx], is_read: true }
    }
  } catch {
    appStore.showError(t('announcement.dialog.markReadError'))
  } finally {
    markingReadIds.value.delete(id)
  }
}

const handleMarkAllAsRead = async () => {
  markingRead.value = true
  try {
    await announcementAPI.markAllAsRead()
    announcements.value = announcements.value.map(a => ({ ...a, is_read: true }))
    appStore.showSuccess(t('announcement.dialog.markAllReadSuccess'))
  } catch {
    appStore.showError(t('announcement.dialog.markAllReadError'))
  } finally {
    markingRead.value = false
  }
}

onMounted(() => {
  loadAnnouncements()
})
</script>

<style scoped>
.announcements-page {
  @apply space-y-6;
}

.page-header {
  @apply flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between;
}

.header-actions {
  @apply flex flex-wrap items-center gap-3;
}

.filter-select {
  @apply rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm;
  @apply focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-dark-200;
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

.btn {
  @apply inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors;
}

.btn-secondary {
  @apply bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600;
}

.btn:disabled {
  @apply cursor-not-allowed opacity-50;
}
</style>
