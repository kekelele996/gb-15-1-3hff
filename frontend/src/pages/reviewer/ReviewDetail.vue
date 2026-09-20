<template>
  <div v-loading="loading">
    <el-page-header class="mb" @back="router.back()">
      <template #content>
        <span style="font-weight: 600">审稿详情</span>
      </template>
    </el-page-header>
    <template v-if="review">
      <div class="mt-16">
        <PaperInfoCard v-if="paper" :paper="paper" />
      </div>
      <el-card shadow="never" class="mt-16" style="max-width: 840px">
        <template #header>提交评审意见</template>
        <template v-if="review.status === 'accepted'">
          <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
            <el-form-item label="评审等级" prop="decision">
              <el-select v-model="form.decision" placeholder="请选择评审等级" style="width: 100%">
                <el-option v-for="(text, key) in REVIEW_DECISION_MAP" :key="key" :label="text" :value="key" />
              </el-select>
            </el-form-item>
            <el-form-item label="评审意见" prop="comments">
              <el-input v-model="form.comments" type="textarea" :rows="6" placeholder="面向作者的详细评审意见（至少 10 字）" />
            </el-form-item>
            <el-form-item label="给编辑的意见" prop="confidentialComments">
              <el-input v-model="form.confidentialComments" type="textarea" :rows="3" placeholder="仅编辑可见，选填" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="submitting" @click="submit">提交评审意见</el-button>
            </el-form-item>
          </el-form>
        </template>
        <el-alert v-else :title="`当前审稿状态：${reviewStatusText(review.status)}`" type="info" :closable="false" />
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { REVIEW_DECISION_MAP } from '../../constants'
import { getPaper } from '../../api/paper'
import { listMyReviews, submitReview } from '../../api/review'
import type { Paper, ReviewItem } from '../../api/types'
import PaperInfoCard from '../../components/PaperInfoCard.vue'
import { reviewStatusText } from '../../utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const review = ref<ReviewItem | null>(null)
const paper = ref<Paper | null>(null)
const form = reactive({
  decision: '',
  comments: '',
  confidentialComments: ''
})
const rules: Record<string, unknown> = {
  decision: [{ required: true, message: '请选择评审等级', trigger: 'change' }],
  comments: [{ required: true, min: 10, message: '评审意见至少 10 字', trigger: 'blur' }]
}

onMounted(async () => {
  const id = route.params.id as string
  loading.value = true
  try {
    const res = await listMyReviews({ page: 1, size: 100 })
    review.value = res.items.find((r) => String(r.id) === id) || null
    if (review.value) {
      paper.value = await getPaper(review.value.paper_id)
    }
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
})

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || !review.value) return
  submitting.value = true
  try {
    await submitReview(review.value.id, {
      decision: form.decision,
      comments: form.comments,
      confidential_comments: form.confidentialComments
    })
    ElMessage.success('评审意见已提交')
    router.push('/reviews')
  } catch {
    // 拦截器已提示
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.mb {
  margin-bottom: 4px;
}
</style>
