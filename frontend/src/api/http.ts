import axios from 'axios'

import { API_BASE } from '@/lib/config'

/**
 * 统一 axios 实例：
 * - baseURL 开发时为 /api（Vite 代理剥掉前缀），生产为空（后端同源托管）
 * - 携带会话 cookie（withCredentials）
 * - X-Requested-With: XMLHttpRequest 让 Spring Security 对未认证请求返回 401 而非 302
 */
export const http = axios.create({
  baseURL: API_BASE,
  withCredentials: true,
  headers: {
    'X-Requested-With': 'XMLHttpRequest',
    Accept: 'application/json',
  },
})
