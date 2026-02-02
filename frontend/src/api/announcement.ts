/**
 * Announcement API endpoints
 * Handles announcement operations for users and admins
 */

import { apiClient } from './client'

// Types
export type AnnouncementType = 'info' | 'warning' | 'important'
export type AnnouncementStatus = 'draft' | 'published' | 'archived'

export interface Announcement {
  id: number
  title: string
  content: string
  type: AnnouncementType
  priority: number
  status: AnnouncementStatus
  publish_at: string | null
  expires_at: string | null
  created_by: number | null
  created_at: string
  updated_at: string
  is_read?: boolean
}

export interface AnnouncementWithReadStatus extends Announcement {
  is_read: boolean
}

export interface UnreadCountResponse {
  count: number
}

export interface AnnouncementListResponse {
  items: Announcement[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface CreateAnnouncementRequest {
  title: string
  content: string
  type?: AnnouncementType
  priority?: number
  publish_at?: number | null
  expires_at?: number | null
}

export interface UpdateAnnouncementRequest {
  title?: string
  content?: string
  type?: AnnouncementType
  priority?: number
  status?: AnnouncementStatus
  publish_at?: number | null
  expires_at?: number | null
}

// User Announcement API
export async function getActiveAnnouncements(): Promise<AnnouncementWithReadStatus[]> {
  const { data } = await apiClient.get<AnnouncementWithReadStatus[]>('/announcements')
  return data
}

export async function getUnreadAnnouncements(): Promise<AnnouncementWithReadStatus[]> {
  const { data } = await apiClient.get<AnnouncementWithReadStatus[]>('/announcements/unread')
  return data
}

export async function getUnreadCount(): Promise<number> {
  const { data } = await apiClient.get<UnreadCountResponse>('/announcements/unread/count')
  return data.count
}

export async function markAsRead(id: number): Promise<void> {
  await apiClient.post(`/announcements/${id}/read`)
}

export async function markAllAsRead(): Promise<void> {
  await apiClient.post('/announcements/read-all')
}

// Admin Announcement API
export async function adminListAnnouncements(params?: {
  page?: number
  page_size?: number
  status?: string
  search?: string
}): Promise<AnnouncementListResponse> {
  const { data } = await apiClient.get<AnnouncementListResponse>('/admin/announcements', {
    params
  })
  return data
}

export async function adminGetAnnouncement(id: number): Promise<Announcement> {
  const { data } = await apiClient.get<Announcement>(`/admin/announcements/${id}`)
  return data
}

export async function adminCreateAnnouncement(req: CreateAnnouncementRequest): Promise<Announcement> {
  const { data } = await apiClient.post<Announcement>('/admin/announcements', req)
  return data
}

export async function adminUpdateAnnouncement(id: number, req: UpdateAnnouncementRequest): Promise<Announcement> {
  const { data } = await apiClient.put<Announcement>(`/admin/announcements/${id}`, req)
  return data
}

export async function adminDeleteAnnouncement(id: number): Promise<void> {
  await apiClient.delete(`/admin/announcements/${id}`)
}

export async function adminPublishAnnouncement(id: number): Promise<Announcement> {
  const { data } = await apiClient.post<Announcement>(`/admin/announcements/${id}/publish`)
  return data
}

export async function adminArchiveAnnouncement(id: number): Promise<Announcement> {
  const { data } = await apiClient.post<Announcement>(`/admin/announcements/${id}/archive`)
  return data
}

export const announcementAPI = {
  // User APIs
  getActiveAnnouncements,
  getUnreadAnnouncements,
  getUnreadCount,
  markAsRead,
  markAllAsRead,

  // Admin APIs
  admin: {
    list: adminListAnnouncements,
    get: adminGetAnnouncement,
    create: adminCreateAnnouncement,
    update: adminUpdateAnnouncement,
    delete: adminDeleteAnnouncement,
    publish: adminPublishAnnouncement,
    archive: adminArchiveAnnouncement
  }
}

export default announcementAPI
