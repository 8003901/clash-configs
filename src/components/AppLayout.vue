<script setup lang="ts">
import { ref } from 'vue'
import { GitMerge, KeyRound, LayoutDashboard, LogOut, Menu, Moon, Radio, Sun } from '@lucide/vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuth } from '@/composables/useAuth'
import { useTheme } from '@/composables/useTheme'

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const { isDark, toggle } = useTheme()

const expanded = ref(false)

const navItems = [
  { name: 'dashboard', label: '仪表盘', icon: LayoutDashboard },
  { name: 'configs', label: '订阅配置', icon: Radio },
  { name: 'merges', label: '合并配置', icon: GitMerge },
  { name: 'password', label: '修改密码', icon: KeyRound },
]

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="flex min-h-screen w-full bg-background">
    <!-- 桌面端图标窄栏（hover 展开显示文字） -->
    <aside
      class="hidden shrink-0 flex-col border-r bg-sidebar transition-[width] duration-200 md:flex"
      :class="expanded ? 'w-56' : 'w-[72px]'"
      @mouseenter="expanded = true"
      @mouseleave="expanded = false"
    >
      <div
        class="flex h-14 items-center gap-2 border-b"
        :class="expanded ? 'justify-start px-4' : 'justify-center px-0'"
      >
        <Radio class="size-5 shrink-0 text-primary" />
        <span v-if="expanded" class="truncate font-semibold">Clash 配置中心</span>
      </div>

      <nav class="flex-1 space-y-1 p-2">
        <RouterLink
          v-for="item in navItems"
          :key="item.name"
          :to="{ name: item.name }"
          :title="item.label"
          class="flex items-center gap-3 rounded-xl py-2.5 text-sm font-medium transition-colors"
          :class="[
            expanded ? 'justify-start px-3' : 'justify-center px-0',
            route.name === item.name
              ? 'bg-sidebar-accent text-sidebar-accent-foreground'
              : 'text-muted-foreground hover:bg-muted hover:text-foreground',
          ]"
        >
          <component :is="item.icon" class="size-5 shrink-0" />
          <span v-if="expanded" class="truncate">{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="space-y-1 border-t p-2">
        <Button
          variant="ghost"
          class="w-full shrink-0"
          :class="expanded ? 'justify-start gap-3 px-3' : 'justify-center px-0'"
          :title="isDark ? '切换到浅色' : '切换到深色'"
          @click="toggle"
        >
          <Sun v-if="isDark" class="size-5 shrink-0" />
          <Moon v-else class="size-5 shrink-0" />
          <span v-if="expanded">{{ isDark ? '浅色模式' : '深色模式' }}</span>
        </Button>
        <Button
          variant="ghost"
          class="w-full shrink-0"
          :class="expanded ? 'justify-start gap-3 px-3' : 'justify-center px-0'"
          :title="'退出登录'"
          @click="handleLogout"
        >
          <LogOut class="size-5 shrink-0" />
          <span v-if="expanded">退出登录</span>
        </Button>
      </div>
    </aside>

    <!-- 移动端顶栏 -->
    <div class="flex flex-1 flex-col">
      <header class="flex h-14 items-center justify-between border-b px-4 md:hidden">
        <span class="font-semibold">Clash 配置中心</span>
        <div class="flex items-center gap-1">
          <Button variant="ghost" size="icon" :title="isDark ? '切换到浅色' : '切换到深色'" @click="toggle">
            <Sun v-if="isDark" class="size-5" />
            <Moon v-else class="size-5" />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="icon">
                <Menu class="size-5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-48">
              <DropdownMenuLabel>导航</DropdownMenuLabel>
              <DropdownMenuItem
                v-for="item in navItems"
                :key="item.name"
                @click="router.push({ name: item.name })"
              >
                <component :is="item.icon" class="size-4" />
                {{ item.label }}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="handleLogout">
                <LogOut class="size-4" />
                退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <main class="flex-1 p-4 md:p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
