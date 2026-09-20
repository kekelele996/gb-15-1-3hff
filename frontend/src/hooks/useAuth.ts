import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'

export function useAuth() {
  const auth = useAuthStore()
  const isLoggedIn = computed(() => auth.isLoggedIn)
  const user = computed(() => auth.user)
  const role = computed(() => auth.role)
  const can = (roles: string[]) => roles.includes(auth.role)
  const logout = () => auth.logout()
  return { isLoggedIn, user, role, can, logout }
}
