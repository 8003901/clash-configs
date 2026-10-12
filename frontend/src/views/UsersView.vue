<script setup lang="ts">
import { onMounted, ref } from 'vue'
import axios from 'axios'
import { Plus } from '@lucide/vue'
import { toast } from 'vue-sonner'
import { createUser, listUsers, type UserSummary } from '@/api/user'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const users = ref<UserSummary[]>([])
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const createOpen = ref(false)
const loading = ref(false)
const saving = ref(false)

async function loadUsers() {
  loading.value = true
  try {
    users.value = await listUsers()
  } catch {
    toast.error('加载账户列表失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  const normalizedUsername = username.value.trim()
  if ([...normalizedUsername].length < 3 || [...normalizedUsername].length > 32) {
    toast.error('用户名须为 3–32 个字符')
    return
  }
  const passwordBytes = new TextEncoder().encode(password.value).length
  if (passwordBytes < 8 || passwordBytes > 72) {
    toast.error('密码须为 8–72 字节')
    return
  }
  if (password.value !== confirmPassword.value) {
    toast.error('两次输入的密码不一致')
    return
  }

  saving.value = true
  try {
    await createUser(normalizedUsername, password.value)
    toast.success('账户已创建')
    createOpen.value = false
    username.value = ''
    password.value = ''
    confirmPassword.value = ''
    await loadUsers()
  } catch (error) {
    const message = axios.isAxiosError<{ error?: string }>(error)
      ? error.response?.data?.error
      : undefined
    toast.error(message || '创建账户失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadUsers)
</script>

<template>
  <div class="space-y-4">
    <div>
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">账户管理</h1>
          <p class="text-sm text-muted-foreground">仅管理员可创建账户；所有账户拥有相同的配置管理权限</p>
        </div>
        <Button @click="createOpen = true">
          <Plus class="size-4" />
          创建账户
        </Button>
      </div>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>现有账户</CardTitle>
        <CardDescription>账户密码不会在此显示</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户名</TableHead>
              <TableHead>类型</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in users" :key="user.id">
              <TableCell class="font-medium">{{ user.username }}</TableCell>
              <TableCell class="text-muted-foreground">
                {{ user.username === 'admin' ? '管理员' : '账户' }}
              </TableCell>
            </TableRow>
            <TableRow v-if="!loading && users.length === 0">
              <TableCell colspan="2" class="text-center text-muted-foreground">
                暂无账户
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="createOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>创建账户</DialogTitle>
          <DialogDescription>请将初始密码安全地告知新用户，并提醒其登录后修改密码</DialogDescription>
        </DialogHeader>
        <form class="space-y-4 py-2" @submit.prevent="handleSubmit">
          <div class="space-y-2">
            <Label for="new-username">用户名</Label>
            <Input id="new-username" v-model="username" autocomplete="off" required />
          </div>
          <div class="space-y-2">
            <Label for="new-user-password">初始密码</Label>
            <Input
              id="new-user-password"
              v-model="password"
              type="password"
              autocomplete="new-password"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="confirm-user-password">确认初始密码</Label>
            <Input
              id="confirm-user-password"
              v-model="confirmPassword"
              type="password"
              autocomplete="new-password"
              required
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="createOpen = false">
              取消
            </Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? '创建中…' : '创建账户' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>
