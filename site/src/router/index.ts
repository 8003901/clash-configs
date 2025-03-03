/**
 * router/index.ts
 *
 * Automatic routes for `./src/pages/*.vue`
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router/auto'
import Login from "@/components/Login.vue";
import HelloWorld from "@/components/HelloWorld.vue";
// import { routes } from 'vue-router/auto-routes'

// 定义路由
const routes = [
  // {
  //   path: '/',
  //   name: 'Main',
  //   component: Main,
  //   children: [
  //     {
  //       path: '',
  //       name: 'home',
  //       component: HelloWorld
  //     },
  //     {
  //       path: 'roles',
  //       name: 'roles',
  //       component: Role
  //     },
  //     {
  //       path: 'users',
  //       name: 'users',
  //       component: User
  //     },
  //     {
  //       path: 'menus',
  //       name: 'menus',
  //       component: Menus
  //     },
  //   ],
  // },
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (!localStorage.getItem('vuetify:dynamic-reload')) {
      console.log('Reloading page to fix dynamic import error')
      localStorage.setItem('vuetify:dynamic-reload', 'true')
      location.assign(to.fullPath)
    } else {
      console.error('Dynamic import error, reloading page did not fix it', err)
    }
  } else {
    console.error(err)
  }
})

router.isReady().then(() => {
  localStorage.removeItem('vuetify:dynamic-reload')
})

export default router
