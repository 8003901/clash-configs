import { createRouter, createWebHashHistory } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'
import { useAuth } from '@/composables/useAuth'
import ConfigsView from '@/views/ConfigsView.vue'
import DashboardView from '@/views/DashboardView.vue'
import LoginView from '@/views/LoginView.vue'
import MergesView from '@/views/MergesView.vue'
import PasswordView from '@/views/PasswordView.vue'

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
        { path: 'password', name: 'password', component: PasswordView },
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
  } else if (to.name === 'login' && auth.authenticated.value === true) {
    return { name: 'dashboard' }
  }
})

export default router
