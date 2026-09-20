import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import MainLayout from '../pages/layout/MainLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../pages/auth/Login.vue'),
    meta: { public: true, title: '登录' }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../pages/auth/Register.vue'),
    meta: { public: true, title: '注册' }
  },
  {
    path: '/',
    component: MainLayout,
    redirect: '/papers',
    children: [
      {
        path: 'papers',
        name: 'papers',
        component: () => import('../pages/author/PaperList.vue'),
        meta: { title: '我的投稿', roles: ['author', 'admin'] }
      },
      {
        path: 'papers/create',
        name: 'paper-create',
        component: () => import('../pages/author/PaperCreate.vue'),
        meta: { title: '新建投稿', roles: ['author', 'admin'] }
      },
      {
        path: 'papers/:id',
        name: 'paper-detail',
        component: () => import('../pages/author/PaperDetail.vue'),
        meta: { title: '投稿详情' }
      },
      {
        path: 'papers/:id/revise',
        name: 'paper-revise',
        component: () => import('../pages/author/PaperRevise.vue'),
        meta: { title: '提交修改', roles: ['author', 'admin'] }
      },
      {
        path: 'editor/initial-review',
        name: 'editor-initial',
        component: () => import('../pages/editor/InitialReview.vue'),
        meta: { title: '编辑部初审', roles: ['editor', 'admin'] }
      },
      {
        path: 'editor/papers/:id',
        name: 'editor-paper-detail',
        component: () => import('../pages/editor/PaperManage.vue'),
        meta: { title: '论文管理', roles: ['editor', 'admin'] }
      },
      {
        path: 'editor/statistics',
        name: 'editor-stats',
        component: () => import('../pages/editor/Statistics.vue'),
        meta: { title: '数据统计', roles: ['editor', 'admin'] }
      },
      {
        path: 'reviews',
        name: 'reviews',
        component: () => import('../pages/reviewer/MyReviews.vue'),
        meta: { title: '我的审稿', roles: ['reviewer', 'admin'] }
      },
      {
        path: 'reviews/:id',
        name: 'review-detail',
        component: () => import('../pages/reviewer/ReviewDetail.vue'),
        meta: { title: '审稿详情', roles: ['reviewer', 'admin'] }
      },
      {
        path: 'library',
        name: 'library',
        component: () => import('../pages/library/PaperLibrary.vue'),
        meta: { title: '论文库' }
      },
      {
        path: 'audit',
        name: 'audit',
        component: () => import('../pages/audit/AuditLogs.vue'),
        meta: { title: '审计日志', roles: ['editor', 'admin'] }
      }
    ]
  }
]

const ROLE_HOME: Record<string, string> = {
  admin: '/papers',
  author: '/papers',
  editor: '/editor/initial-review',
  reviewer: '/reviews'
}

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!auth.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.path === '/') {
    return ROLE_HOME[auth.role] || '/papers'
  }
  const roles = to.meta.roles as string[] | undefined
  if (roles && roles.length && !roles.includes(auth.role)) {
    return ROLE_HOME[auth.role] || '/papers'
  }
  return true
})

router.afterEach((to) => {
  document.title = `${String(to.meta.title || '')} - PaperFlow 学术论文管理平台`
})

export default router
