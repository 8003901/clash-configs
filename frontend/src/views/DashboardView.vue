<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { listConfigs } from '@/api/config'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatBytes, formatExpire, parseUserInfo } from '@/lib/format'
import type { ClashConfig } from '@/types'

const configs = ref<ClashConfig[]>([])
const loading = ref(false)

const usages = computed(() => configs.value.map((c) => parseUserInfo(c.subscriptionUserinfo)))
const totalUpload = computed(() => usages.value.reduce((sum, u) => sum + u.upload, 0))
const totalDownload = computed(() => usages.value.reduce((sum, u) => sum + u.download, 0))
const usedBytes = computed(() => totalUpload.value + totalDownload.value)
const totalBytes = computed(() => usages.value.reduce((sum, u) => sum + u.total, 0))
const earliestExpire = computed(() => {
  const expires = usages.value.filter((u) => u.expire > 0).map((u) => u.expire)
  return expires.length ? Math.min(...expires) : 0
})

async function load() {
  loading.value = true
  try {
    configs.value = await listConfigs()
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">仪表盘</h1>
      <p class="text-sm text-muted-foreground">订阅流量用量汇总</p>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>订阅数量</CardDescription>
          <CardTitle class="text-2xl">{{ configs.length }}</CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>已用流量</CardDescription>
          <CardTitle class="text-2xl">{{ formatBytes(usedBytes) }}</CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>总流量</CardDescription>
          <CardTitle class="text-2xl">{{ formatBytes(totalBytes) }}</CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>最早到期</CardDescription>
          <CardTitle class="text-2xl">{{ formatExpire(earliestExpire) }}</CardTitle>
        </CardHeader>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>订阅明细</CardTitle>
        <CardDescription>各订阅的上传 / 下载 / 总量与到期时间</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>更新周期</TableHead>
              <TableHead class="text-right">已用</TableHead>
              <TableHead class="text-right">总量</TableHead>
              <TableHead class="text-right">到期</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(config, i) in configs" :key="config.id ?? i">
              <TableCell class="font-medium">{{ config.name ?? '-' }}</TableCell>
              <TableCell>
                <Badge variant="outline">{{ config.updateSchedule }}</Badge>
              </TableCell>
              <TableCell class="text-right">
                {{ formatBytes(usages[i].upload + usages[i].download) }}
              </TableCell>
              <TableCell class="text-right">{{ formatBytes(usages[i].total) }}</TableCell>
              <TableCell class="text-right">{{ formatExpire(usages[i].expire) }}</TableCell>
            </TableRow>
            <TableRow v-if="!loading && configs.length === 0">
              <TableCell colspan="5" class="text-center text-muted-foreground">
                暂无订阅配置
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
