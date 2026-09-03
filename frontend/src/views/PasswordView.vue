<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { changePassword } from '@/api/user'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const saving = ref(false)

async function handleSubmit() {
  if (!oldPassword.value || !newPassword.value) {
    toast.error('请输入旧密码和新密码')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    toast.error('两次输入的新密码不一致')
    return
  }
  saving.value = true
  try {
    await changePassword(oldPassword.value, newPassword.value)
    toast.success('密码已修改')
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch {
    toast.error('修改失败，请检查旧密码是否正确')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">修改密码</h1>
      <p class="text-sm text-muted-foreground">修改当前登录账号的密码</p>
    </div>

    <Card class="max-w-md">
      <CardHeader>
        <CardTitle>修改密码</CardTitle>
        <CardDescription>修改后下次登录请使用新密码</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="space-y-4" @submit.prevent="handleSubmit">
          <div class="space-y-2">
            <Label for="old-password">旧密码</Label>
            <Input
              id="old-password"
              v-model="oldPassword"
              type="password"
              autocomplete="current-password"
            />
          </div>
          <div class="space-y-2">
            <Label for="new-password">新密码</Label>
            <Input
              id="new-password"
              v-model="newPassword"
              type="password"
              autocomplete="new-password"
            />
          </div>
          <div class="space-y-2">
            <Label for="confirm-password">确认新密码</Label>
            <Input
              id="confirm-password"
              v-model="confirmPassword"
              type="password"
              autocomplete="new-password"
            />
          </div>
          <Button type="submit" :disabled="saving">
            {{ saving ? '保存中…' : '保存' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
