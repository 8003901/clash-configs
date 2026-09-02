import axios from 'axios'

/**
 * 统一 axios 实例：
 * - baseURL 为 /api，开发时由 Vite 代理到后端（去掉 /api 前缀）
 * - 携带会话 cookie（withCredentials）
 * - X-Requested-With: XMLHttpRequest 让 Spring Security 对未认证请求返回 401 而非 302
 */
export const http = axios.create({
  baseURL: '/api',
  withCredentials: true,
  headers: {
    'X-Requested-With': 'XMLHttpRequest',
    Accept: 'application/json',
  },
})
