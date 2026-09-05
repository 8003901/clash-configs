<script setup lang="ts">
import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { json, jsonParseLinter } from '@codemirror/lang-json'
import { linter } from '@codemirror/lint'
import { basicSetup, EditorView } from 'codemirror'
import { tags } from '@lezer/highlight'
import {
  ArrowLeft,
  Braces,
  CircleCheck,
  Copy,
  Loader2,
  TriangleAlert,
} from '@lucide/vue'
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { getMerge, updateMerge } from '@/api/merge'
import { Button } from '@/components/ui/button'
import type { ClashConfigsMerge } from '@/types'

const route = useRoute()
const router = useRouter()

const merge = ref<ClashConfigsMerge | null>(null)
const loading = ref(false)
const saving = ref(false)
const jsonError = ref<string | null>(null)
const container = ref<HTMLElement | null>(null)

let view: EditorView | null = null

// 编辑器外观：复用应用 CSS 变量，自动适配明暗主题
const jsonTheme = EditorView.theme({
  '&': {
    backgroundColor: 'var(--card)',
    color: 'var(--foreground)',
    fontSize: '13px',
    height: '100%',
  },
  '.cm-scroller': {
    fontFamily:
      'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
    lineHeight: '1.6',
  },
  '.cm-content': {
    caretColor: 'var(--primary)',
    padding: '12px 0',
  },
  '.cm-cursor, .cm-dropCursor': {
    borderLeftColor: 'var(--primary)',
  },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
    backgroundColor: 'color-mix(in srgb, var(--primary) 22%, transparent)',
  },
  '.cm-gutters': {
    backgroundColor: 'var(--card)',
    color: 'var(--muted-foreground)',
    border: 'none',
    borderRight: '1px solid var(--border)',
  },
  '.cm-activeLine': {
    backgroundColor: 'color-mix(in srgb, var(--muted) 45%, transparent)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'color-mix(in srgb, var(--muted) 45%, transparent)',
    color: 'var(--foreground)',
  },
  '.cm-foldGutter .cm-gutterElement': {
    color: 'var(--muted-foreground)',
  },
  '.cm-tooltip': {
    backgroundColor: 'var(--popover)',
    color: 'var(--popover-foreground)',
    border: '1px solid var(--border)',
  },
  '.cm-panels': {
    backgroundColor: 'var(--card)',
    color: 'var(--foreground)',
  },
  '.cm-panels.cm-panels-bottom': {
    borderTop: '1px solid var(--border)',
  },
  '.cm-searchMatch': {
    backgroundColor: 'color-mix(in srgb, var(--chart-2) 30%, transparent)',
  },
  '.cm-selectionMatch': {
    backgroundColor: 'color-mix(in srgb, var(--chart-1) 22%, transparent)',
  },
})

// JSON 语法高亮：复用 chart 色板，键=蓝、字符串=绿、数字=紫、布尔=橙
const jsonHighlight = HighlightStyle.define([
  { tag: tags.propertyName, color: 'var(--chart-1)' },
  { tag: tags.string, color: 'var(--chart-3)' },
  { tag: tags.number, color: 'var(--chart-4)' },
  { tag: tags.bool, color: 'var(--chart-2)' },
  { tag: tags.null, color: 'var(--muted-foreground)' },
  { tag: tags.separator, color: 'var(--muted-foreground)' },
  { tag: tags.punctuation, color: 'var(--muted-foreground)' },
  { tag: tags.bracket, color: 'var(--foreground)' },
  { tag: tags.brace, color: 'var(--foreground)' },
  { tag: tags.squareBracket, color: 'var(--foreground)' },
])

function currentDoc(): string {
  return view?.state.doc.toString() ?? ''
}

function validate(): boolean {
  const text = currentDoc().trim()
  if (!text) {
    jsonError.value = '配置内容为空'
    return false
  }
  try {
    JSON.parse(text)
    jsonError.value = null
    return true
  } catch (e) {
    jsonError.value = e instanceof Error ? e.message : String(e)
    return false
  }
}

function format() {
  const text = currentDoc()
  if (!text.trim()) return
  try {
    const formatted = JSON.stringify(JSON.parse(text), null, 2)
    view?.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: formatted } })
    jsonError.value = null
    toast.success('已格式化')
  } catch {
    validate()
    toast.error('格式化失败：JSON 无效')
  }
}

async function copy() {
  try {
    await navigator.clipboard.writeText(currentDoc())
    toast.success('已复制')
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

async function save() {
  if (!merge.value?.id) return
  if (!validate()) {
    toast.error('请先修正 JSON 错误再保存')
    return
  }
  saving.value = true
  try {
    await updateMerge(merge.value.id, {
      name: merge.value.name ?? '',
      config: currentDoc(),
      configs: (merge.value.configs ?? [])
        .map((c) => ({ id: c.id ?? '' }))
        .filter((c) => c.id),
    })
    toast.success('已保存')
  } catch {
    toast.error('保存失败')
  } finally {
    saving.value = false
  }
}

function createEditor(initialDoc: string) {
  if (!container.value) return
  view = new EditorView({
    doc: initialDoc,
    parent: container.value,
    extensions: [
      basicSetup,
      json(),
      linter(jsonParseLinter()),
      syntaxHighlighting(jsonHighlight),
      jsonTheme,
      EditorView.lineWrapping,
      EditorView.updateListener.of((update) => {
        if (update.docChanged) validate()
      }),
    ],
  })
  validate()
}

onMounted(async () => {
  const id = String(route.params.id ?? '')
  loading.value = true
  try {
    merge.value = await getMerge(id)
  } catch {
    toast.error('加载配置模板失败')
    router.push({ name: 'merges' })
    return
  } finally {
    loading.value = false
  }
  await nextTick()
  createEditor(merge.value?.config ?? '{}')
})

onBeforeUnmount(() => {
  view?.destroy()
  view = null
})
</script>

<template>
  <div class="flex h-[calc(100vh-5.5rem)] flex-col gap-3 md:h-[calc(100vh-3rem)]">
    <div class="flex items-center justify-between gap-3">
      <div class="flex min-w-0 items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          title="返回合并配置"
          @click="router.push({ name: 'merges' })"
        >
          <ArrowLeft class="size-4" />
        </Button>
        <div class="min-w-0">
          <h1 class="truncate text-xl font-semibold tracking-tight md:text-2xl">
            {{ merge?.name ?? '配置模板' }}
          </h1>
          <p class="text-sm text-muted-foreground">编辑合并配置的 JSON 模板</p>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-2">
        <div
          class="flex items-center gap-1.5 text-sm"
          :class="jsonError ? 'text-destructive' : 'text-[var(--chart-3)]'"
          :title="jsonError ?? undefined"
        >
          <CircleCheck v-if="!jsonError" class="size-4" />
          <TriangleAlert v-else class="size-4" />
          <span class="hidden sm:inline">{{ jsonError ? 'JSON 无效' : 'JSON 有效' }}</span>
        </div>
        <Button variant="outline" size="sm" @click="format">
          <Braces class="size-4" />
          格式化
        </Button>
        <Button variant="outline" size="sm" @click="copy">
          <Copy class="size-4" />
          复制
        </Button>
        <Button size="sm" :disabled="saving" @click="save">
          <Loader2 v-if="saving" class="size-4 animate-spin" />
          {{ saving ? '保存中…' : '保存' }}
        </Button>
      </div>
    </div>

    <div class="min-h-0 flex-1 overflow-hidden rounded-lg border bg-card">
      <div v-if="loading" class="flex h-full items-center justify-center text-muted-foreground">
        加载中…
      </div>
      <div v-else ref="container" class="h-full w-full" />
    </div>
  </div>
</template>
