import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
  }
}

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: AppLayout,
    children: [
      { path: '', redirect: '/status' },
      {
        path: 'status',
        name: 'system-status',
        component: () => import('@/pages/SystemStatusPage.vue'),
        meta: { title: 'Status do sistema' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/pages/NotFoundPage.vue'),
    meta: { title: 'Página não encontrada' },
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · WorkEvent` : 'WorkEvent'
})
