import request from '../utils/request'
import type { PageResult } from './types'

export interface AuditLogItem {
  id: number
  user_id: number
  username: string
  action: string
  entity: string
  entity_id: string
  detail: string
  ip: string
  request_id: string
  created_at: string
}

export function listAuditLogs(params: Record<string, unknown>) {
  return request.get('/audit-logs', { params }) as Promise<PageResult<AuditLogItem>>
}
