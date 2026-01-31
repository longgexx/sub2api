/**
 * Admin Redeem Rules API endpoints
 * Handles trial limitation rules management
 */

import { apiClient } from '../client'

export interface RedeemRule {
  id: number
  type: string
  trigger_value: number
  max_times_per_user: number
  fallback_value: number
  is_active: boolean
  description?: string
  created_at: string
  updated_at: string
}

export interface CreateRedeemRuleRequest {
  type: string
  trigger_value: number
  max_times_per_user: number
  fallback_value: number
  is_active?: boolean
  description?: string
}

export interface UpdateRedeemRuleRequest {
  trigger_value?: number
  max_times_per_user?: number
  fallback_value?: number
  is_active?: boolean
  description?: string
}

export async function list(): Promise<RedeemRule[]> {
  const { data } = await apiClient.get<RedeemRule[]>('/admin/redeem-rules')
  return data
}

export async function getById(id: number): Promise<RedeemRule> {
  const { data } = await apiClient.get<RedeemRule>(`/admin/redeem-rules/${id}`)
  return data
}

export async function create(request: CreateRedeemRuleRequest): Promise<RedeemRule> {
  const { data } = await apiClient.post<RedeemRule>('/admin/redeem-rules', request)
  return data
}

export async function update(id: number, request: UpdateRedeemRuleRequest): Promise<RedeemRule> {
  const { data } = await apiClient.put<RedeemRule>(`/admin/redeem-rules/${id}`, request)
  return data
}

export async function deleteRule(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/redeem-rules/${id}`)
  return data
}

const redeemRulesAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteRule
}

export default redeemRulesAPI
