<script setup lang="ts">
import { Pencil, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { createConfig, deleteConfig, getConfig, listConfigs, updateConfig } from '@/api/config'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatDateTime } from '@/lib/format'
import type { ClashConfig, UpdateSchedule } from '@/types'

interface FormState {
  id: string | null
  name: string
  url: string
  updateSchedule: UpdateSchedule
  enabled: boolean
}

const configs = ref<ClashConfig[]>([])
const loading = ref(false)
const saving = ref(false)

const dialogOpen = ref(false)
const form = reactive<FormState>({
  id: null,
  name: '',
  url: '',
  updateSchedule: 'DAY',
  enabled: true,
})

const deleteOpen = ref(false)
const deleteTarget = ref<ClashConfig | null>(null)

async function load() {
  loading.value = true
  try {
    configs.value = await listConfigs()
  } catch {
    toast.error('加载订阅配置失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.id = null
  form.name = ''
  form.url = ''
  form.updateSchedule = 'DAY'
  form.enabled = true
  dialogOpen.value = true
}

function openEdit(config: ClashConfig) {
  form.id = config.id
  form.name = config.name ?? ''
  form.url = config.url ?? ''
  form.updateSchedule = config.updateSchedule ?? 'DAY'
  form.enabled = config.enabled
  dialogOpen.value = true
}

async function submit() {
  if (!form.name.trim() || !form.url.trim()) {
    toast.error('名称和订阅地址不能为空')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      url: form.url.trim(),
      updateSchedule: form.updateSchedule,
      enabled: form.enabled,
    }
    if (form.id) {
      await updateConfig(form.id, payload)
    } else {
      await createConfig(payload)
    }
    dialogOpen.value = false
    toast.success('已保存')
    await load()
  } catch {
    toast.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function renew(config: ClashConfig) {
  if (!config.id) return
  try {
    await getConfig(config.id, true)
    toast.success(`已更新「${config.name}」`)
    await load()
  } catch {
    toast.error('更新失败')
  }
}

function askDelete(config: ClashConfig) {
  deleteTarget.value = config
  deleteOpen.value = true
}

async function confirmDelete() {
  const id = deleteTarget.value?.id
  if (!id) return
  try {
    await deleteConfig(id)
    toast.success('已删除')
    await load()
  } finally {
    deleteOpen.value = false
    deleteTarget.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">订阅配置</h1>
        <p class="text-sm text-muted-foreground">管理 Clash 订阅源</p>
      </div>
      <Button @click="openCreate">
        <Plus class="size-4" />
        新增配置
      </Button>
    </div>

    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>名称</TableHead>
          <TableHead>订阅地址</TableHead>
          <TableHead>更新周期</TableHead>
          <TableHead>启用</TableHead>
          <TableHead>更新时间</TableHead>
          <TableHead class="text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="(config, i) in configs" :key="config.id ?? i">
          <TableCell class="font-medium">{{ config.name ?? '-' }}</TableCell>
          <TableCell class="max-w-64 truncate text-muted-foreground">
            {{ config.url ?? '-' }}
          </TableCell>
          <TableCell>
            <Badge variant="outline">{{ config.updateSchedule === 'WEEK' ? '每周' : '每天' }}</Badge>
          </TableCell>
          <TableCell>
            <Badge :variant="config.enabled ? 'secondary' : 'destructive'">
              {{ config.enabled ? '启用' : '停用' }}
            </Badge>
          </TableCell>
          <TableCell class="text-muted-foreground">
            {{ formatDateTime(config.updatedAt) }}
          </TableCell>
          <TableCell class="text-right">
            <div class="flex justify-end gap-1">
              <Button variant="ghost" size="icon-sm" title="手动更新" @click="renew(config)">
                <RefreshCw class="size-4" />
              </Button>
              <Button variant="ghost" size="icon-sm" title="编辑" @click="openEdit(config)">
                <Pencil class="size-4" />
              </Button>
              <Button variant="ghost" size="icon-sm" title="删除" @click="askDelete(config)">
                <Trash2 class="size-4" />
              </Button>
            </div>
          </TableCell>
        </TableRow>
        <TableRow v-if="!loading && configs.length === 0">
          <TableCell colspan="6" class="text-center text-muted-foreground">
            暂无订阅配置，点击右上角「新增配置」添加
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑配置' : '新增配置' }}</DialogTitle>
          <DialogDescription>填写订阅信息，保存后会自动从订阅地址拉取最新内容</DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-2">
          <div class="space-y-2">
            <Label for="cfg-name">名称</Label>
            <Input id="cfg-name" v-model="form.name" placeholder="例如：机场 A" />
          </div>
          <div class="space-y-2">
            <Label for="cfg-url">订阅地址</Label>
            <Input id="cfg-url" v-model="form.url" placeholder="https://example.com/sub" />
          </div>
          <div class="space-y-2">
            <Label>更新周期</Label>
            <Select v-model="form.updateSchedule">
              <SelectTrigger class="w-full">
                <SelectValue placeholder="选择更新周期" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="DAY">每天</SelectItem>
                <SelectItem value="WEEK">每周</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center justify-between">
            <Label for="cfg-enabled">启用</Label>
            <Switch id="cfg-enabled" v-model:checked="form.enabled" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="dialogOpen = false">取消</Button>
          <Button :disabled="saving" @click="submit">
            {{ saving ? '保存中…' : '保存' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <AlertDialog v-model:open="deleteOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>
            确定要删除「{{ deleteTarget?.name }}」吗？此操作不可恢复。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction variant="destructive" @click="confirmDelete">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
