<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">PaperFlow</div>
      <el-menu :default-active="activeMenu" router class="menu">
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <span>{{ m.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="page-title">{{ route.meta.title }}</div>
        <el-dropdown @command="onCommand">
          <span class="user-info">
            <el-avatar :size="28" class="avatar">{{ avatarText }}</el-avatar>
            {{ auth.user?.real_name }}（{{ auth.roleText }}）
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowDown } from '@element-plus/icons-vue'
import { useAuthStore } from '../../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const allMenus = [
  { path: '/papers', title: '我的投稿', icon: 'Document', roles: ['author', 'admin'] },
  { path: '/papers/create', title: '新建投稿', icon: 'EditPen', roles: ['author', 'admin'] },
  { path: '/editor/initial-review', title: '编辑部初审', icon: 'Checked', roles: ['editor', 'admin'] },
  { path: '/editor/statistics', title: '数据统计', icon: 'DataAnalysis', roles: ['editor', 'admin'] },
  { path: '/reviews', title: '我的审稿', icon: 'Tickets', roles: ['reviewer', 'admin'] },
  { path: '/library', title: '论文库', icon: 'Collection', roles: ['author', 'reviewer', 'editor', 'admin'] },
  { path: '/audit', title: '审计日志', icon: 'Memo', roles: ['editor', 'admin'] }
]

const menus = computed(() => allMenus.filter((m) => m.roles.includes(auth.role)))
const activeMenu = computed(() => route.path)
const avatarText = computed(() => (auth.user?.real_name || auth.user?.username || 'U').slice(0, 1))

function onCommand(cmd: string) {
  if (cmd === 'logout') {
    auth.logout()
    router.push('/login')
  }
}
</script>
