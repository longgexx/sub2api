/**
 * Redeem code API endpoints
 * Handles redeem code redemption for users
 */

import { apiClient } from './client'
import type { RedeemCodeRequest } from '@/types'

export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  actual_value?: number // 实际到账金额（降级兑换时与 value 不同）
  status: string
  used_at: string
  created_at: string
  // 订阅类型专用字段
  group_id?: number
  validity_days?: number
  group?: {
    id: number
    name: string
  }
}

export interface RedeemResult {
  redeem_code: {
    id: number
    code: string
    type: string
    value: number
    status: string
    used_at: string
    created_at: string
    group_id?: number
    validity_days?: number
    group?: {
      id: number
      name: string
    }
  }
  original_value: number
  actual_value: number
  is_degraded: boolean
  message?: string
}

// 预检结果（用于确认降级兑换）
export interface RedeemPreview {
  redeem_code: {
    id: number
    code: string
    type: string
    value: number
    status: string
    created_at: string
    group_id?: number
    validity_days?: number
    group?: {
      id: number
      name: string
    }
  }
  original_value: number
  actual_value: number
  will_degrade: boolean
  needs_confirm: boolean
  confirm_message?: string
}

// 兑换响应可能是 RedeemResult 或 RedeemPreview
export type RedeemResponse = RedeemResult | RedeemPreview

// 类型守卫：判断是否是需要确认的预检结果
export function isRedeemPreview(response: RedeemResponse): response is RedeemPreview {
  return 'needs_confirm' in response && response.needs_confirm === true
}

// 类型守卫：判断是否是兑换成功的结果
export function isRedeemResult(response: RedeemResponse): response is RedeemResult {
  return 'is_degraded' in response && !('needs_confirm' in response && response.needs_confirm === true)
}

/**
 * Redeem a code
 * @param code - Redeem code string
 * @param confirm - Whether to confirm degraded redemption
 * @returns Redemption result or preview (if needs confirmation)
 */
export async function redeem(code: string, confirm: boolean = false): Promise<RedeemResponse> {
  const payload: RedeemCodeRequest & { confirm?: boolean } = { code }
  if (confirm) {
    payload.confirm = true
  }

  const { data } = await apiClient.post<RedeemResponse>('/redeem', payload)

  return data
}

/**
 * Get user's redemption history
 * @returns List of redeemed codes
 */
export async function getHistory(): Promise<RedeemHistoryItem[]> {
  const { data } = await apiClient.get<RedeemHistoryItem[]>('/redeem/history')
  return data
}

export const redeemAPI = {
  redeem,
  getHistory
}

export default redeemAPI
