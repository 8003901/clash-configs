<script setup lang="ts">
import { Copy, ExternalLink, Pencil, Plus, RefreshCw } from '@lucide/vue'
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { listConfigs } from '@/api/config'
import {
  createMerge,
  getMerge,
  listMerges,
  refreshMergeToken,
  updateMerge,
} from '@/api/merge'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
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
import { Textarea } from '@/components/ui/textarea'
import { BACKEND_URL } from '@/lib/config'
import { formatDateTime } from '@/lib/format'
import type { ClashConfig, ClashConfigsMerge } from '@/types'

interface FormState {
  id: string | null
  name: string
  selectedIds: string[]
  config: string
}

const merges = ref<ClashConfigsMerge[]>([])
const configs = ref<ClashConfig[]>([])
const loading = ref(false)
const saving = ref(false)

const dialogOpen = ref(false)
const form = reactive<FormState>({
  id: null,
  name: '',
  selectedIds: [],
  config: '',
})

async function load() {
  loading.value = true
  try {
    ;[merges.value, configs.value] = await Promise.all([listMerges(), listConfigs()])
  } catch {
    toast.error('加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.id = null
  form.name = ''
  form.selectedIds = []
  form.config = ''
  dialogOpen.value = true
}

async function openEdit(merge: ClashConfigsMerge) {
  if (!merge.id) return
  try {
    const detail = await getMerge(merge.id)
    form.id = detail.id
    form.name = detail.name ?? ''
    form.selectedIds = (detail.configs ?? []).map((c) => c.id ?? '').filter(Boolean)
    form.config = detail.config ?? ''
    dialogOpen.value = true
  } catch {
    toast.error('加载合并配置详情失败')
  }
}

function toggleConfig(id: string, checked: boolean) {
  if (checked) {
    if (!form.selectedIds.includes(id)) form.selectedIds.push(id)
  } else {
    form.selectedIds = form.selectedIds.filter((x) => x !== id)
  }
}

async function submit() {
  if (!form.name.trim()) {
    toast.error('名称不能为空')
    return
  }
  saving.value = true
  try {
    if (form.id) {
      await updateMerge(form.id, {
        name: form.name.trim(),
        config: form.config,
        configs: form.selectedIds.map((id) => ({ id })),
      })
    } else {
      await createMerge({
        name: form.name.trim(),
        configIds: form.selectedIds.length ? form.selectedIds : null,
      })
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

async function refreshToken(merge: ClashConfigsMerge) {
  if (!merge.id) return
  try {
    await refreshMergeToken(merge.id)
    toast.success('Token 已刷新')
    await load()
  } catch {
    toast.error('刷新 Token 失败')
  }
}

async function copyText(text: string, message = '已复制') {
  try {
    await navigator.clipboard.writeText(text)
    toast.success(message)
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

function subscriptionUrl(token: string) {
  return `${BACKEND_URL}/configs?token=${token}`
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">合并配置</h1>
        <p class="text-sm text-muted-foreground">将多个订阅合并成一个，生成订阅链接</p>
      </div>
      <Button @click="openCreate">
        <Plus class="size-4" />
        新增合并
      </Button>
    </div>

    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>名称</TableHead>
          <TableHead>订阅数</TableHead>
          <TableHead>Token</TableHead>
          <TableHead>更新时间</TableHead>
          <TableHead class="text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="(merge, i) in merges" :key="merge.id ?? i">
          <TableCell class="font-medium">{{ merge.name ?? '-' }}</TableCell>
          <TableCell>
            <Badge variant="secondary">{{ merge.configs?.length ?? 0 }} 个</Badge>
          </TableCell>
          <TableCell class="max-w-40 truncate font-mono text-xs text-muted-foreground">
            {{ merge.token ?? '-' }}
          </TableCell>
          <TableCell class="text-muted-foreground">
            {{ formatDateTime(merge.updatedAt) }}
          </TableCell>
          <TableCell class="text-right">
            <div class="flex justify-end gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                title="复制订阅链接"
                @click="copyText(subscriptionUrl(merge.token ?? ''), '订阅链接已复制')"
              >
                <ExternalLink class="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                title="复制 Token"
                @click="copyText(merge.token ?? '', 'Token 已复制')"
              >
                <Copy class="size-4" />
              </Button>
              <Button variant="ghost" size="icon-sm" title="刷新 Token" @click="refreshToken(merge)">
                <RefreshCw class="size-4" />
              </Button>
              <Button variant="ghost" size="icon-sm" title="编辑" @click="openEdit(merge)">
                <Pencil class="size-4" />
              </Button>
            </div>
          </TableCell>
        </TableRow>
        <TableRow v-if="!loading && merges.length === 0">
          <TableCell colspan="5" class="text-center text-muted-foreground">
            暂无合并配置
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑合并配置' : '新增合并配置' }}</DialogTitle>
          <DialogDescription>填写名称并勾选要合并的订阅</DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-2">
          <div class="space-y-2">
            <Label for="merge-name">名称</Label>
            <Input id="merge-name" v-model="form.name" placeholder="例如：我的合并订阅" />
          </div>
          <div class="space-y-2">
            <Label>选择订阅</Label>
            <div class="max-h-48 space-y-0.5 overflow-y-auto rounded-md border p-1.5">
              <label
                v-for="(config, i) in configs"
                :key="config.id ?? i"
                class="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-muted"
              >
                <Checkbox
                  :checked="form.selectedIds.includes(config.id ?? '')"
                  @update:checked="toggleConfig(config.id ?? '', $event)"
                />
                <span class="truncate">{{ config.name }}</span>
              </label>
              <p v-if="configs.length === 0" class="p-2 text-sm text-muted-foreground">
                暂无可选订阅，请先到「订阅配置」添加
              </p>
            </div>
          </div>
          <div v-if="form.id" class="space-y-2">
            <Label for="merge-config">配置模板（JSON，可选编辑）</Label>
            <Textarea
              id="merge-config"
              v-model="form.config"
              class="h-40 font-mono text-xs"
              placeholder="{}"
            />
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
  </div>
</template>
