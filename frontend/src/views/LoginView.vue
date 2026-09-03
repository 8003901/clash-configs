<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { login } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/composables/useAuth'

const route = useRoute()
const router = useRouter()
const auth = useAuth()

const username = ref('admin')
const password = ref('')
const loading = ref(false)

async function handleSubmit() {
  if (!username.value.trim() || !password.value) {
    toast.error('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const ok = await login(username.value.trim(), password.value)
    if (ok) {
      auth.set(true)
      const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
      router.push(redirect)
    } else {
      toast.error('登录失败，用户名或密码错误')
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-muted/40 p-4">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">Clash 配置中心</CardTitle>
        <CardDescription>请登录以管理订阅与合并配置</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="space-y-4" @submit.prevent="handleSubmit">
          <div class="space-y-2">
            <Label for="username">用户名</Label>
            <Input id="username" v-model="username" autocomplete="username" placeholder="admin" />
          </div>
          <div class="space-y-2">
            <Label for="password">密码</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              placeholder="请输入密码"
            />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? '登录中…' : '登录' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
