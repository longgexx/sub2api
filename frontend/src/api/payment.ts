/**
 * Payment API endpoints
 * Handles payment/topup operations for users
 */

import { apiClient } from './client'

// Types
export interface PaymentOrder {
  id: number
  trade_no: string
  user_id: number
  amount: number
  payment_amount: number
  credit_amount?: number
  status: 'pending' | 'paid' | 'expired' | 'cancelled'
  created_at: string
  paid_at?: string
  expired_at: string
  alipay_trade_no?: string
  payer_account?: string
  user?: {
    id: number
    email: string
    username?: string
  }
}

export interface PaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  qr_code_url: string
  timeout_mins: number
  rate_coefficient: number  // 充值系数
}

export interface PaymentStats {
  today_order_count: number
  today_paid_count: number
  today_paid_amount: number
  total_order_count: number
  total_paid_count: number
  total_paid_amount: number
  pending_order_count: number
}

export interface MonitorStatus {
  running: boolean
  available: boolean
  last_check_time?: string
  seconds_since_check?: number
  message?: string
}

export interface PaginationResult {
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface OrderListResponse {
  items: PaymentOrder[]
  total: number
  page: number
  page_size: number
  pages: number
}

// User Payment API
export async function getConfig(): Promise<PaymentConfig> {
  const { data } = await apiClient.get<PaymentConfig>('/payment/config')
  return data
}

export async function createOrder(amount: number): Promise<PaymentOrder> {
  const { data } = await apiClient.post<PaymentOrder>('/payment/orders', { amount })
  return data
}

export async function listOrders(params?: {
  page?: number
  page_size?: number
  status?: string
}): Promise<OrderListResponse> {
  const { data } = await apiClient.get<OrderListResponse>('/payment/orders', {
    params
  })
  return data
}

export async function getOrder(tradeNo: string): Promise<PaymentOrder> {
  const { data } = await apiClient.get<PaymentOrder>(`/payment/orders/${tradeNo}`)
  return data
}

export async function cancelOrder(tradeNo: string): Promise<void> {
  await apiClient.delete(`/payment/orders/${tradeNo}`)
}

// Admin Payment API
export async function adminListOrders(params?: {
  page?: number
  page_size?: number
  status?: string
  search?: string
}): Promise<OrderListResponse> {
  const { data } = await apiClient.get<OrderListResponse>('/admin/payment/orders', {
    params
  })
  return data
}

export async function adminGetOrder(id: number): Promise<PaymentOrder> {
  const { data } = await apiClient.get<PaymentOrder>(`/admin/payment/orders/${id}`)
  return data
}

export async function adminManualConfirm(id: number): Promise<void> {
  await apiClient.post(`/admin/payment/orders/${id}/confirm`)
}

export async function adminGetStats(): Promise<PaymentStats> {
  const { data } = await apiClient.get<PaymentStats>('/admin/payment/stats')
  return data
}

export async function adminGetMonitorStatus(): Promise<MonitorStatus> {
  const { data } = await apiClient.get<MonitorStatus>('/admin/payment/monitor/status')
  return data
}

export async function adminStartMonitor(): Promise<void> {
  await apiClient.post('/admin/payment/monitor/start')
}

export async function adminStopMonitor(): Promise<void> {
  await apiClient.post('/admin/payment/monitor/stop')
}

export const paymentAPI = {
  // User APIs
  getConfig,
  createOrder,
  listOrders,
  getOrder,
  cancelOrder,

  // Admin APIs
  admin: {
    listOrders: adminListOrders,
    getOrder: adminGetOrder,
    manualConfirm: adminManualConfirm,
    getStats: adminGetStats,
    getMonitorStatus: adminGetMonitorStatus,
    startMonitor: adminStartMonitor,
    stopMonitor: adminStopMonitor
  }
}

export default paymentAPI
