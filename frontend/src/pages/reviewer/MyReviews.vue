<template>
  <el-card shadow="never">
    <template #header>
      <div class="row-between">
        <span>我的审稿任务（{{ pagination.total.value }}）</span>
        <el-select v-model="status" placeholder="全部状态" style="width: 160px" @change="onStatusChange">
          <el-option label="全部状态" value="" />
          <el-option v-for="(text, key) in REVIEW_STATUS_MAP" :key="key" :label="text" :value="key" />
        </el-select>
      </div>
    </template>
    <el-table :data="pagination.items.value" v-loading="pagination.loading.value" stripe>
      <el-table-column label="论文" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">{{ row.paper?.title || `论文 #${row.paper_id}` }}</template>
      </el-table-column>
      <el-table-column label="学科" width="120">
        <template #default="{ row }">{{ subjectText(row.paper?.subject || '') }}</template>
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
      <el-table-column label="截止日期" width="150">
        <template #default="{ row }">{{ formatTime(row.due_date) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'invited'">
            <el-button link type="success" @click="respond(row, true)">接受</el-button>
            <el-button link type="danger" @click="respond(row, false)">拒绝</el-button>
          </template>
          <el-button
            v-if="row.status === 'accepted'"
            link
            type="primary"
            @click="router.push(`/reviews/${row.id}`)"
          >
            提交审稿
          </el-button>
          <el-button v-if="row.status === 'completed'" link type="primary" @click="router.push(`/reviews/${row.id}`)">
            查看
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <EmptyState
      v-if="pagination.total.value === 0 && !pagination.loading.value"
      description="暂无审稿任务"
    />
    <el-pagination
      v-model:current-page="pagination.page.value"
      v-model:page-size="pagination.size.value"
      :total="pagination.total.value"
      layout="total, prev, pager, next"
      @current-change="() => pagination.load({ status: status.value })"
      class="pager"
    />
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { REVIEW_STATUS_MAP } from '../../constants'
import { listMyReviews, respondReview } from '../../api/review'
import type { ReviewItem } from '../../api/types'
import EmptyState from '../../components/EmptyState.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import { usePagination } from '../../hooks/usePagination'
import { formatTime, subjectText } from '../../utils/format'

const router = useRouter()
const status = ref('')
const pagination = usePagination<ReviewItem>((params) => listMyReviews({ ...params, status: status.value }))

function onStatusChange() {
  pagination.load({ status: status.value })
}

async function respond(row: ReviewItem, accept: boolean) {
  try {
    await ElMessageBox.confirm(
      accept ? `确认接受论文「${row.paper?.title || row.paper_id}」的审稿邀请？` : '确认拒绝该审稿邀请？',
      accept ? '接受邀请' : '拒绝邀请',
      { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    await respondReview(row.id, { accept })
    ElMessage.success(accept ? '已接受，请在截止日前完成审稿' : '已拒绝')
    pagination.load({ status: status.value })
  } catch {
    // 拦截器已提示
  }
}

onMounted(() => pagination.load({ status: '' }))
</script>
