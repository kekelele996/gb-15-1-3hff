<template>
  <div v-loading="loading">
    <el-page-header class="mb" @back="router.back()">
      <template #content>
        <span style="font-weight: 600">论文管理</span>
      </template>
    </el-page-header>
    <template v-if="paper">
      <div class="mt-16">
        <PaperInfoCard :paper="paper" />
      </div>

      <el-card shadow="never" class="mt-16">
        <template #header>
          <div class="row-between">
            <span>审稿人分配</span>
            <el-button
              v-if="['initial_review', 'external_review', 'revision'].includes(paper.status)"
              type="primary"
              size="small"
              @click="assignVisible = true"
            >
              追加审稿人
            </el-button>
          </div>
        </template>
        <EmptyState v-if="!paper.reviews?.length" description="尚未分配审稿人" />
        <el-table v-else :data="paper.reviews" size="small" border>
          <el-table-column label="轮次" width="80">
            <template #default="{ row }">第{{ row.round || 1 }}轮</template>
          </el-table-column>
          <el-table-column label="审稿人" width="140">
            <template #default="{ row }">{{ row.reviewer?.real_name || row.reviewer?.username || '-' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{ row }"><StatusBadge :status="row.status" kind="review" /></template>
          </el-table-column>
          <el-table-column label="评审等级" width="140">
            <template #default="{ row }">
              <StatusBadge v-if="row.decision" :status="row.decision" kind="decision" />
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="comments" label="评审意见" show-overflow-tooltip />
          <el-table-column prop="confidential_comments" label="给编辑的保密意见" show-overflow-tooltip />
          <el-table-column label="截止日期" width="150">
            <template #default="{ row }">{{ formatTime(row.due_date) }}</template>
          </el-table-column>
          <el-table-column label="完成时间" width="150">
            <template #default="{ row }">{{ formatTime(row.completed_at) }}</template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-card shadow="never" class="mt-16">
        <template #header>查重检测</template>
        <el-descriptions v-if="plagiarism" :column="3" border>
          <el-descriptions-item label="状态">
            <StatusBadge :status="plagiarism.status" kind="plagiarism" />
          </el-descriptions-item>
          <el-descriptions-item label="重复率">
            <span :class="{ 'high-similarity': plagiarism.similarity > 30 }">
              {{ formatPercent(plagiarism.similarity) }}
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="检测时间">{{ formatTime(plagiarism.checked_at) }}</el-descriptions-item>
        </el-descriptions>
        <el-button class="mt-16" type="primary" plain size="small" @click="rerunPlagiarism">
          重跑查重
        </el-button>
      </el-card>

      <el-card shadow="never" class="mt-16">
        <template #header>终审决定</template>
        <template v-if="summary">
          <el-descriptions :column="5" border size="small" class="summary-bar">
            <el-descriptions-item label="已完成">{{ summary.completed_reviewers }} 人</el-descriptions-item>
            <el-descriptions-item label="待接受">{{ summary.invited }} 人</el-descriptions-item>
            <el-descriptions-item label="审稿中">{{ summary.accepted }} 人</el-descriptions-item>
            <el-descriptions-item label="已拒绝">{{ summary.declined }} 人</el-descriptions-item>
            <el-descriptions-item label="已超期">{{ summary.expired }} 人</el-descriptions-item>
          </el-descriptions>
          <el-alert
            v-if="!summary.can_finalize"
            class="mt-16"
            type="warning"
            :closable="false"
            title="暂不满足终审条件"
            :description="`终审需至少 2 位审稿人完成评审且无待处理邀请：${summary.block_reasons.join('；')}`"
            show-icon
          />
          <el-alert
            v-else
            class="mt-16"
            type="success"
            :closable="false"
            title="外审进度已满足终审条件"
            show-icon
          />
        </template>
        <el-form label-width="90px" style="max-width: 640px" class="mt-16">
          <el-form-item label="决定">
            <el-radio-group v-model="decision">
              <el-radio value="accepted">录用</el-radio>
              <el-radio value="rejected">拒稿</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="终审意见">
            <el-input v-model="comment" type="textarea" :rows="3" placeholder="选填" />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :disabled="!canSubmitDecision"
              :loading="decisionLoading"
              @click="submitDecision"
            >
              提交终审决定
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </template>
  </div>

  <el-dialog v-model="assignVisible" title="追加审稿人" width="480px">
    <el-alert
      type="info"
      :closable="false"
      title="同一审稿人存在有效任务（待接受/审稿中）时不可重复邀请；已拒绝或已超期的审稿人可补邀"
      style="margin-bottom: 12px"
    />
    <el-select v-model="assignReviewerId" placeholder="选择审稿人" style="width: 100%">
      <el-option v-for="r in reviewers" :key="r.id" :label="`${r.real_name}（${r.username}）`" :value="r.id" />
    </el-select>
    <template #footer>
      <el-button @click="assignVisible = false">取消</el-button>
      <el-button type="primary" :loading="assignLoading" @click="submitAssign">确认分配</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { assignReviewer, getReviewSummary } from '../../api/review'
import { finalDecision, getPaper, getPlagiarism, listReviewers, rerunPlagiarism as rerunApi } from '../../api/paper'
import type { Paper, PlagiarismResult, ReviewSummary } from '../../api/types'
import EmptyState from '../../components/EmptyState.vue'
import PaperInfoCard from '../../components/PaperInfoCard.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import { formatPercent, formatTime } from '../../utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const decisionLoading = ref(false)
const assignLoading = ref(false)
const assignVisible = ref(false)
const paper = ref<Paper | null>(null)
const plagiarism = ref<PlagiarismResult | null>(null)
const summary = ref<ReviewSummary | null>(null)
const reviewers = ref<Array<{ id: number; real_name: string; username: string }>>([])
const decision = ref('accepted')
const comment = ref('')
const assignReviewerId = ref(0)

const canSubmitDecision = computed(() => {
  if (!paper.value) return false
  if (!['initial_review', 'external_review', 'revision'].includes(paper.value.status)) return false
  return summary.value?.can_finalize === true
})

async function load() {
  const id = route.params.id as string
  loading.value = true
  try {
    paper.value = await getPaper(id)
    plagiarism.value = await getPlagiarism(id)
    summary.value = await getReviewSummary(id)
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

async function submitDecision() {
  if (!paper.value) return
  decisionLoading.value = true
  try {
    await finalDecision(paper.value.id, { decision: decision.value, comment: comment.value })
    ElMessage.success('终审决定已提交')
    await load()
  } catch {
    // 拦截器已提示
  } finally {
    decisionLoading.value = false
  }
}

async function submitAssign() {
  if (!paper.value || !assignReviewerId.value) {
    ElMessage.warning('请选择审稿人')
    return
  }
  assignLoading.value = true
  try {
    await assignReviewer(paper.value.id, assignReviewerId.value)
    ElMessage.success('审稿人已分配')
    assignVisible.value = false
    assignReviewerId.value = 0
    await load()
  } catch {
    // 拦截器已提示
  } finally {
    assignLoading.value = false
  }
}

async function rerunPlagiarism() {
  if (!paper.value) return
  try {
    plagiarism.value = await rerunApi(paper.value.id)
    paper.value = await getPaper(paper.value.id)
    ElMessage.success('查重已重跑')
  } catch {
    // 拦截器已提示
  }
}

onMounted(async () => {
  await load()
  reviewers.value = await listReviewers()
})
</script>

<style scoped>
.mb {
  margin-bottom: 4px;
}
.summary-bar {
  max-width: 720px;
}
</style>
