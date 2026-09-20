<template>
  <el-card shadow="never">
    <template #header>
      <div class="row-between">
        <span>编辑部初审队列（待初审 {{ pagination.total.value }}）</span>
        <el-select v-model="status" placeholder="全部状态" style="width: 160px" @change="onStatusChange">
          <el-option label="全部状态" value="" />
          <el-option v-for="(text, key) in PAPER_STATUS_MAP" :key="key" :label="text" :value="key" />
        </el-select>
      </div>
    </template>
    <el-table :data="pagination.items.value" v-loading="pagination.loading.value" stripe>
      <el-table-column prop="title" label="标题" min-width="240" show-overflow-tooltip />
      <el-table-column label="学科" width="120">
        <template #default="{ row }">{{ subjectText(row.subject) }}</template>
      </el-table-column>
      <el-table-column label="作者" width="120">
        <template #default="{ row }">{{ row.submitter?.real_name || row.submitter?.username || '-' }}</template>
      </el-table-column>
      <el-table-column label="相似度" width="100">
        <template #default="{ row }">{{ formatPercent(row.similarity) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }"><StatusBadge :status="row.status" kind="paper" /></template>
      </el-table-column>
      <el-table-column label="投稿时间" width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 'submitted'" link type="primary" @click="openReview(row)">
            初审
          </el-button>
          <el-button link type="primary" @click="router.push(`/editor/papers/${row.id}`)">管理</el-button>
        </template>
      </el-table-column>
    </el-table>
    <EmptyState
      v-if="pagination.total.value === 0 && !pagination.loading.value"
      description="暂无待初审稿件"
    />
    <el-pagination
      v-model:current-page="pagination.page.value"
      v-model:page-size="pagination.size.value"
      :total="pagination.total.value"
      layout="total, prev, pager, next"
      @current-change="() => pagination.load({ status: status })"
      class="pager"
    />
  </el-card>

  <el-dialog v-model="dialog.visible" title="编辑部初审" width="560px">
    <el-form ref="formRef" :model="dialog" label-width="90px">
      <el-form-item label="论文">
        <span>{{ dialog.paper?.title }}</span>
      </el-form-item>
      <el-form-item label="初审结论">
        <el-radio-group v-model="dialog.pass">
          <el-radio :value="true">通过</el-radio>
          <el-radio :value="false">退回</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="dialog.pass" label="分配审稿人" prop="reviewerId">
        <el-select v-model="dialog.reviewerId" placeholder="选择审稿人" style="width: 100%">
          <el-option v-for="r in reviewers" :key="r.id" :label="`${r.real_name}（${r.username}）`" :value="r.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="意见" prop="reason">
        <el-input v-model="dialog.reason" type="textarea" :rows="3" placeholder="初审意见 / 退回理由" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog.visible = false">取消</el-button>
      <el-button type="primary" :loading="dialog.loading" @click="submitReview">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { PAPER_STATUS_MAP } from '../../constants'
import { initialReview, listPapers, listReviewers } from '../../api/paper'
import type { Paper } from '../../api/types'
import EmptyState from '../../components/EmptyState.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import { usePagination } from '../../hooks/usePagination'
import { formatPercent, formatTime, subjectText } from '../../utils/format'

const router = useRouter()
const status = ref('')
const reviewers = ref<Array<{ id: number; real_name: string; username: string }>>([])
const formRef = ref<FormInstance>()
const pagination = usePagination<Paper>((params) => listPapers({ ...params, status: status.value }))
const dialog = reactive({
  visible: false,
  paper: null as Paper | null,
  pass: true,
  reviewerId: 0,
  reason: '',
  loading: false
})

function onStatusChange() {
  pagination.load({ status: status.value })
}

function openReview(row: Paper) {
  dialog.paper = row
  dialog.pass = true
  dialog.reviewerId = 0
  dialog.reason = ''
  dialog.visible = true
}

async function submitReview() {
  if (!dialog.paper) return
  if (dialog.pass && !dialog.reviewerId) {
    ElMessage.warning('通过初审必须分配审稿人')
    return
  }
  dialog.loading = true
  try {
    await initialReview(dialog.paper.id, {
      pass: dialog.pass,
      reviewer_id: dialog.reviewerId,
      reason: dialog.reason
    })
    ElMessage.success('初审已提交')
    dialog.visible = false
    pagination.load({ status: status.value })
  } catch {
    // 拦截器已提示
  } finally {
    dialog.loading = false
  }
}

onMounted(async () => {
  pagination.load({ status: '' })
  reviewers.value = await listReviewers()
})
</script>
