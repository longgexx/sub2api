<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { onMounted, watch, ref } from 'vue'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import AnnouncementDialog from '@/components/common/AnnouncementDialog.vue'
import { useAppStore, useAuthStore, useSubscriptionStore } from '@/stores'
import { getSetupStatus } from '@/api/setup'
import { announcementAPI, type AnnouncementWithReadStatus } from '@/api'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()

// Announcement state
const showAnnouncementDialog = ref(false)
const unreadAnnouncements = ref<AnnouncementWithReadStatus[]>([])

// Check for unread announcements
const checkUnreadAnnouncements = async () => {
  try {
    const announcements = await announcementAPI.getUnreadAnnouncements()
    if (announcements.length > 0) {
      unreadAnnouncements.value = announcements
      showAnnouncementDialog.value = true
    }
  } catch (error: any) {
    // Only log error, don't show to user as this is a background check
    // 401 errors are expected when not authenticated
    if (error?.response?.status !== 401) {
      console.error('Failed to check announcements:', error)
    }
  }
}

const handleAnnouncementDialogClose = () => {
  showAnnouncementDialog.value = false
}

const handleAnnouncementsUpdate = (updated: AnnouncementWithReadStatus[]) => {
  unreadAnnouncements.value = updated
}

/**
 * Update favicon dynamically
 * @param logoUrl - URL of the logo to use as favicon
 */
function updateFavicon(logoUrl: string) {
  // Find existing favicon link or create new one
  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.type = logoUrl.endsWith('.svg') ? 'image/svg+xml' : 'image/x-icon'
  link.href = logoUrl
}

// Watch for site settings changes and update favicon/title
watch(
  () => appStore.siteLogo,
  (newLogo) => {
    if (newLogo) {
      updateFavicon(newLogo)
    }
  },
  { immediate: true }
)

watch(
  () => appStore.siteName,
  (newName) => {
    if (newName) {
      document.title = `${newName} - AI API Gateway`
    }
  },
  { immediate: true }
)

// Watch for authentication state and manage subscription data
watch(
  () => authStore.isAuthenticated,
  (isAuthenticated) => {
    if (isAuthenticated) {
      // User logged in: preload subscriptions and start polling
      subscriptionStore.fetchActiveSubscriptions().catch((error) => {
        console.error('Failed to preload subscriptions:', error)
      })
      subscriptionStore.startPolling()

      // Check for unread announcements on login or page refresh with existing session
      // Use a small delay to ensure auth is fully initialized
      setTimeout(() => {
        checkUnreadAnnouncements()
      }, 500)
    } else {
      // User logged out: clear data and stop polling
      subscriptionStore.clear()
    }
  },
  { immediate: true }
)

onMounted(async () => {
  // Check if setup is needed
  try {
    const status = await getSetupStatus()
    if (status.needs_setup && route.path !== '/setup') {
      router.replace('/setup')
      return
    }
  } catch {
    // If setup endpoint fails, assume normal mode and continue
  }

  // Load public settings into appStore (will be cached for other components)
  await appStore.fetchPublicSettings()
})
</script>

<template>
  <NavigationProgress />
  <RouterView />
  <Toast />
  <AnnouncementDialog
    :show="showAnnouncementDialog"
    :announcements="unreadAnnouncements"
    @close="handleAnnouncementDialogClose"
    @update:announcements="handleAnnouncementsUpdate"
  />
</template>
