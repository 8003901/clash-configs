/**
 * 后端服务的基础地址，用于拼接对外提供的订阅链接（/configs?token=...）。
 * 开发环境默认指向 http://localhost:8080；生产环境请通过 VITE_BACKEND_URL 指定。
 */
export const BACKEND_URL = import.meta.env.VITE_BACKEND_URL ?? 'http://localhost:8080'
