/**
 * 后端服务的基础地址，用于拼接对外提供的订阅链接（/configs?token=...）。
 * 开发环境默认指向 http://localhost:8080；生产环境请通过 VITE_BACKEND_URL 指定。
 */
export const BACKEND_URL = import.meta.env.VITE_BACKEND_URL ?? 'http://localhost:8080'

/** API 前缀：开发时走 Vite 代理（/api，代理会剥掉前缀），生产由后端同源托管（无前缀） */
export const API_BASE = import.meta.env.PROD ? '' : '/api'
