<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <template #header>
        <div class="auth-title">注册账号</div>
        <div class="auth-subtitle">加入 PaperFlow 学术论文管理平台</div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="72px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="3-64 位" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="姓名" prop="real_name">
          <el-input v-model="form.real_name" placeholder="真实姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="选填" />
        </el-form-item>
        <el-form-item label="单位" prop="institution">
          <el-input v-model="form.institution" placeholder="选填" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="form.role">
            <el-radio value="author">作者</el-radio>
            <el-radio value="reviewer">审稿人</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-button type="primary" size="large" class="full-width" :loading="loading" @click="submit">
          注 册
        </el-button>
        <div class="auth-links">
          <span>已有账号？</span>
          <router-link to="/login">去登录</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { register } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({
  username: '',
  password: '',
  real_name: '',
  email: '',
  institution: '',
  role: 'author'
})
const rules = {
  username: [
    { required: true, min: 3, max: 64, message: '用户名 3-64 位', trigger: 'blur' }
  ],
  password: [
    { required: true, min: 6, max: 64, message: '密码至少 6 位', trigger: 'blur' }
  ],
  real_name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }]
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const user = await register({ ...form })
    ElMessage.success(`注册成功，欢迎 ${user.real_name}`)
    await auth.login(form.username, form.password)
    router.push('/papers')
  } catch {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}
</script>
