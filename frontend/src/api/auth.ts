import request from '../utils/request'
import type { UserSummary } from './types'

export interface LoginResult {
  token: string
  user: UserSummary
}

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data) as Promise<LoginResult>
}

export function register(data: Record<string, unknown>) {
  return request.post('/auth/register', data) as Promise<UserSummary>
}

export function getMe() {
  return request.get('/users/me') as Promise<UserSummary>
}

export function updateMe(data: Record<string, unknown>) {
  return request.put('/users/me', data) as Promise<UserSummary>
}
