<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <template #header>
        <div class="auth-title">PaperFlow 学术论文管理平台</div>
        <div class="auth-subtitle">作者 · 审稿人 · 编辑 一体化发表流程</div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="0">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" size="large" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            :prefix-icon="Lock"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button type="primary" size="large" class="full-width" :loading="loading" @click="submit">
          登 录
        </el-button>
        <div class="auth-links">
          <span>还没有账号？</span>
          <router-link to="/register">立即注册</router-link>
        </div>
        <el-alert type="info" :closable="false" class="seed-tip" title="演示账号">
          editor/editor123（编辑）、reviewer/reviewer123（审稿人）、author/author123（作者）、admin/admin123（管理员）
        </el-alert>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const user = await auth.login(form.username, form.password)
    ElMessage.success(`欢迎，${user.real_name}`)
    const redirect = (route.query.redirect as string) || '/papers'
    router.push(redirect)
  } catch {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}
</script>
