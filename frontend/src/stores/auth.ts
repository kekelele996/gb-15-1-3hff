import { defineStore } from 'pinia'
import { login as apiLogin, getMe } from '../api/auth'
import type { UserSummary } from '../api/types'
import { ROLE_MAP } from '../constants'

function readUser(): UserSummary | null {
  try {
    return JSON.parse(localStorage.getItem('paperflow_user') || 'null')
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('paperflow_token') || '',
    user: readUser()
  }),
  getters: {
    isLoggedIn: (s) => !!s.token,
    role: (s) => s.user?.role || '',
    roleText: (s) => (s.user ? ROLE_MAP[s.user.role] || s.user.role : '')
  },
  actions: {
    async login(username: string, password: string) {
      const res = await apiLogin({ username, password })
      this.token = res.token
      this.user = res.user
      localStorage.setItem('paperflow_token', res.token)
      localStorage.setItem('paperflow_user', JSON.stringify(res.user))
      return res.user
    },
    async fetchMe() {
      const user = await getMe()
      this.user = user
      localStorage.setItem('paperflow_user', JSON.stringify(user))
      return user
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('paperflow_token')
      localStorage.removeItem('paperflow_user')
    }
  }
})
