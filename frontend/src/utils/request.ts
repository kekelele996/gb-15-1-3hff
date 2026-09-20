import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('paperflow_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res && typeof res === 'object' && 'code' in res) {
      if (res.code !== 0) {
        ElMessage.error(res.message || '请求失败')
        return Promise.reject(new Error(res.message || '请求失败'))
      }
      return res.data
    }
    return res
  },
  (error) => {
    const res = error.response?.data
    const msg = res?.message || error.message || '网络错误'
    if (error.response?.status === 401) {
      localStorage.removeItem('paperflow_token')
      localStorage.removeItem('paperflow_user')
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login'
      }
    } else if (error.response?.status === 403) {
      ElMessage.error(res?.message || '无权限执行该操作')
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(error)
  }
)

export default request
