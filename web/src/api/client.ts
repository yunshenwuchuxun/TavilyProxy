import axios from 'axios'
import { ref } from 'vue'

const STORAGE_KEY = 'tavily_proxy_admin_token'

const authTokenRef = ref<string>(localStorage.getItem(STORAGE_KEY) ?? '')

export function getAuthToken(): string {
  return authTokenRef.value
}

export function setAuthToken(value: string): void {
  localStorage.setItem(STORAGE_KEY, value)
  authTokenRef.value = value
}

export function clearAuthToken(): void {
  localStorage.removeItem(STORAGE_KEY)
  authTokenRef.value = ''
}

export const api = axios.create()

api.interceptors.request.use((config) => {
  const token = getAuthToken()
  if (token) {
    config.headers = config.headers ?? {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error?.response?.status === 401) {
      clearAuthToken()
      window.dispatchEvent(new Event('auth-required'))
    }
    return Promise.reject(error)
  }
)
