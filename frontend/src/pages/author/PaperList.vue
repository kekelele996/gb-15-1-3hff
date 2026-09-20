<template>
  <el-card shadow="never">
    <template #header>
      <div class="row-between">
        <span>我的投稿（{{ pagination.total.value }}）</span>
        <el-button type="primary" @click="router.push('/papers/create')">新建投稿</el-button>
      </div>
    </template>
    <el-table :data="pagination.items.value" v-loading="pagination.loading.value" stripe>
      <el-table-column prop="title" label="标题" min-width="240" show-overflow-tooltip />
      <el-table-column label="学科" width="120">
        <template #default="{ row }">{{ subjectText(row.subject) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <StatusBadge :status="row.status" kind="paper" />
        </template>
      </el-table-column>
      <el-table-column label="相似度" width="100">
        <template #default="{ row }">
          <span :class="{ 'high-similarity': row.similarity > 30 }">{{ formatPercent(row.similarity) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="版本" width="70">
        <template #default="{ row }">V{{ row.version }}</template>
      </el-table-column>
      <el-table-column label="投稿时间" width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/papers/${row.id}`)">详情</el-button>
          <el-button
            v-if="row.status === 'revision'"
            link
            type="warning"
            @click="router.push(`/papers/${row.id}/revise`)"
          >
            修改
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <EmptyState
      v-if="pagination.total.value === 0 && !pagination.loading.value"
      description="还没有投稿，点击右上角新建投稿"
    />
    <el-pagination
      v-model:current-page="pagination.page.value"
      v-model:page-size="pagination.size.value"
      :total="pagination.total.value"
      layout="total, prev, pager, next"
      @current-change="() => pagination.load()"
      class="pager"
    />
  </el-card>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listMyPapers } from '../../api/paper'
import type { Paper } from '../../api/types'
import EmptyState from '../../components/EmptyState.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import { usePagination } from '../../hooks/usePagination'
import { formatPercent, formatTime, subjectText } from '../../utils/format'

const router = useRouter()
const pagination = usePagination<Paper>((params) => listMyPapers(params))
onMounted(() => pagination.load())
</script>
