/**
 * Admin Operations View API endpoints
 * Provides business metrics and analytics for operators
 */

import { apiClient } from '../client'

// ============ Types ============

export interface OpsViewOverview {
  // 余额指标
  total_balance: number
  // 消耗指标
  today_consumption: number
  week_consumption: number
  month_consumption: number
  // 充值指标
  today_recharge: number
  week_recharge: number
  month_recharge: number
  // 充值明细
  today_payment_recharge: number
  today_redeem_recharge: number
  week_payment_recharge: number
  week_redeem_recharge: number
  month_payment_recharge: number
  month_redeem_recharge: number
  // 用户指标
  total_users: number
  today_active_users: number
  today_new_users: number
  week_active_users: number
  week_new_users: number
  week_paying_users: number
  week_retention: number
  month_active_users: number
  month_new_users: number
}

export interface OpsViewTrendPoint {
  date: string
  consumption: number
  payment_recharge: number
  redeem_recharge: number
  total_recharge: number
}

export interface OpsViewUserGrowthPoint {
  date: string
  new_users: number
  active_users: number
  paying_users: number
  retention_d1: number
  retention_d7: number
  retention_d30: number
}

export interface OpsViewTopUser {
  user_id: number
  email: string
  username: string
  consumption: number
  requests: number
  avg_cost: number
  last_active_at: string
  registered_at: string
}

// ============ API Parameters ============

export interface OpsViewTrendParams {
  days?: number
  timezone?: string
}

export interface OpsViewTopUsersParams {
  start_date?: string
  end_date?: string
  timezone?: string
  limit?: number
}

// ============ API Responses ============

export interface OpsViewTrendResponse {
  trend: OpsViewTrendPoint[]
  days: number
}

export interface OpsViewUserGrowthResponse {
  growth: OpsViewUserGrowthPoint[]
  days: number
}

export interface OpsViewTopUsersResponse {
  users: OpsViewTopUser[]
  start_date: string
  end_date: string
  limit: number
}

// ============ API Functions ============

/**
 * Get operations overview data
 */
export async function getOverview(timezone?: string): Promise<OpsViewOverview> {
  const { data } = await apiClient.get<OpsViewOverview>('/admin/ops-view/overview', {
    params: { timezone }
  })
  return data
}

/**
 * Get consumption and recharge trend data
 */
export async function getTrend(params?: OpsViewTrendParams): Promise<OpsViewTrendResponse> {
  const { data } = await apiClient.get<OpsViewTrendResponse>('/admin/ops-view/trend', { params })
  return data
}

/**
 * Get user growth data
 */
export async function getUserGrowth(
  params?: OpsViewTrendParams
): Promise<OpsViewUserGrowthResponse> {
  const { data } = await apiClient.get<OpsViewUserGrowthResponse>('/admin/ops-view/users/growth', {
    params
  })
  return data
}

/**
 * Get top users by consumption
 */
export async function getTopUsers(
  params?: OpsViewTopUsersParams
): Promise<OpsViewTopUsersResponse> {
  const { data } = await apiClient.get<OpsViewTopUsersResponse>('/admin/ops-view/users/top', {
    params
  })
  return data
}

export const opsViewAPI = {
  getOverview,
  getTrend,
  getUserGrowth,
  getTopUsers
}

export default opsViewAPI
