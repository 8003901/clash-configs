import axios from 'axios'
import { http } from './http'

/** 通过访问受保护接口判断当前会话是否已认证 */
export async function checkSession(): Promise<boolean> {
  try {
    await http.get('/clash_configs')
    return true
  } catch {
    return false
  }
}

/** 使用 Spring Security 表单登录（form-encoded + session cookie） */
export async function login(username: string, password: string): Promise<boolean> {
  const body = new URLSearchParams()
  body.append('username', username)
  body.append('password', password)
  try {
    await axios.post('/api/login', body, {
      withCredentials: true,
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      transformResponse: [(data) => data],
    })
  } catch {
    // 登录成功后 Spring 会 302 跳转到 /，浏览器跟随返回纯文本；
    // 这里不依赖响应内容，交由 checkSession 校验。
  }
  return checkSession()
}

export async function logout(): Promise<void> {
  try {
    await axios.post('/api/logout', null, {
      withCredentials: true,
      transformResponse: [(data) => data],
    })
  } catch {
    // 忽略登出接口的返回
  }
}
