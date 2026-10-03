/** API 前缀：开发时走 Vite 代理（/api，代理会剥掉前缀），生产由后端同源托管（无前缀） */
export const API_BASE = import.meta.env.PROD ? '' : '/api'
