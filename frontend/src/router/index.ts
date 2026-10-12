import { createRouter, createWebHashHistory } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'
import { useAuth } from '@/composables/useAuth'
import ConfigsView from '@/views/ConfigsView.vue'
import DashboardView from '@/views/DashboardView.vue'
import LoginView from '@/views/LoginView.vue'
import MergesView from '@/views/MergesView.vue'
import PasswordView from '@/views/PasswordView.vue'
import UsersView from '@/views/UsersView.vue'
import { getCurrentUser } from '@/api/user'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: { name: 'dashboard' } },
        { path: 'dashboard', name: 'dashboard', component: DashboardView },
        { path: 'configs', name: 'configs', component: ConfigsView },
        { path: 'merges', name: 'merges', component: MergesView },
        {
          path: 'merges/:id/template',
          name: 'merge-template',
          component: () => import('@/views/MergeTemplateView.vue'),
        },
        { path: 'password', name: 'password', component: PasswordView },
        { path: 'users', name: 'users', component: UsersView, meta: { requiresAdmin: true } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuth()
  if (to.meta.requiresAuth) {
    const ok = await auth.ensure()
    if (!ok) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    if (to.meta.requiresAdmin) {
      try {
        if (!(await getCurrentUser()).isAdmin) return { name: 'dashboard' }
      } catch {
        return { name: 'dashboard' }
      }
    }
  } else if (to.name === 'login' && auth.authenticated.value === true) {
    return { name: 'dashboard' }
  }
})

export default router
