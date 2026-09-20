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
            <div>
              <span v-if="summary" class="round-progress">
                第 {{ summary.round }} 轮评审：已完成 {{ summary.completed }} / {{ roundTotal }}
              </span>
              <el-button
                v-if="['initial_review', 'external_review', 'revision'].includes(paper.status)"
                type="primary"
                size="small"
                @click="assignVisible = true"
              >
                补邀审稿人
              </el-button>
            </div>
          </div>
        </template>
        <el-progress
          v-if="summary && roundTotal > 0"
          :percentage="roundPercent"
          :format="() => progressText"
          class="mb-16"
        />
        <EmptyState v-if="!paper.reviews?.length" description="尚未分配审稿人" />
        <el-table v-else :data="paper.reviews" size="small" border>
          <el-table-column label="轮次" width="80">
            <template #default="{ row }">第{{ row.round }}轮</template>
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
        <template #header>
          <div class="row-between">
            <span>终审决定</span>
            <span v-if="summary" class="round-progress">当前第 {{ summary.round }} 轮</span>
          </div>
        </template>
        <el-descriptions v-if="summary" :column="5" border class="mb-16">
          <el-descriptions-item label="已完成">{{ summary.completed }} 人</el-descriptions-item>
          <el-descriptions-item label="待接受">{{ summary.pending }} 人</el-descriptions-item>
          <el-descriptions-item label="审稿中">{{ summary.in_progress }} 人</el-descriptions-item>
          <el-descriptions-item label="已拒绝">{{ summary.declined }} 人</el-descriptions-item>
          <el-descriptions-item label="已超期">{{ summary.expired }} 人</el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-if="summary && !summary.can_finalize"
          :title="`暂不能终审：${summary.block_reason}`"
          type="warning"
          show-icon
          :closable="false"
          class="mb-16"
        />
        <el-form label-width="90px" style="max-width: 640px">
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
            <span v-if="summary && !summary.can_finalize" class="gate-hint">
              需当前轮次有效评审至少 2 人完成且无待处理邀请
            </span>
          </el-form-item>
        </el-form>
      </el-card>
    </template>
  </div>

  <el-dialog v-model="assignVisible" title="补邀审稿人" width="480px">
    <el-alert
      title="同一人同一轮只能存在一条有效任务；已拒绝或已超期的审稿人可再次邀请"
      type="info"
      :closable="false"
      class="mb-16"
    />
    <el-select v-model="assignReviewerId" placeholder="选择审稿人" style="width: 100%">
      <el-option
        v-for="r in reviewers"
        :key="r.id"
        :label="`${r.real_name}（${r.username}）${busyReviewerIds.has(r.id) ? '（本轮已有有效任务）' : ''}`"
        :value="r.id"
        :disabled="busyReviewerIds.has(r.id)"
      />
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

const roundTotal = computed(() => {
  if (!summary.value) return 0
  return summary.value.completed + summary.value.pending + summary.value.in_progress
})

const roundPercent = computed(() => {
  if (!summary.value || roundTotal.value === 0) return 0
  return Math.round((summary.value.completed / roundTotal.value) * 100)
})

const progressText = computed(() => `${summary.value?.completed ?? 0}/${roundTotal.value}`)

const canSubmitDecision = computed(() => {
  if (!paper.value || !summary.value) return false
  if (!['initial_review', 'external_review', 'revision'].includes(paper.value.status)) return false
  return summary.value.can_finalize
})

// 当前轮已有有效任务（含已完成）的审稿人，补邀时禁用
const busyReviewerIds = computed(() => {
  const ids = new Set<number>()
  if (!paper.value?.reviews || !summary.value) return ids
  const round = summary.value.round
  const now = Date.now()
  for (const r of paper.value.reviews) {
    if (r.round !== round) continue
    if (r.status === 'completed') {
      ids.add(r.reviewer_id)
    } else if (['invited', 'accepted'].includes(r.status)) {
      if (!r.due_date || new Date(r.due_date).getTime() > now) ids.add(r.reviewer_id)
    }
  }
  return ids
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
.mb-16 {
  margin-bottom: 16px;
}
.round-progress {
  color: #909399;
  font-size: 13px;
  margin-right: 12px;
}
.gate-hint {
  margin-left: 12px;
  color: #e6a23c;
  font-size: 12px;
}
</style>
