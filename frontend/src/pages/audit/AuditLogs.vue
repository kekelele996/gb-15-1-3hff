<template>
  <el-card shadow="never">
    <template #header>
      <span>操作审计日志（{{ pagination.total.value }}）</span>
    </template>
    <el-table :data="pagination.items.value" v-loading="pagination.loading.value" stripe size="small">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户" width="110" />
      <el-table-column prop="action" label="动作" min-width="200" />
      <el-table-column label="实体" width="90">
        <template #default="{ row }">{{ entityText(row.entity) }}</template>
      </el-table-column>
      <el-table-column prop="entity_id" label="实体ID" width="90" />
      <el-table-column prop="detail" label="详情" min-width="180" show-overflow-tooltip />
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column prop="request_id" label="请求ID" width="200" show-overflow-tooltip />
      <el-table-column label="时间" width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
    </el-table>
    <EmptyState v-if="pagination.total.value === 0 && !pagination.loading.value" description="暂无审计日志" />
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
import { listAuditLogs } from '../../api/audit'
import type { AuditLogItem } from '../../api/audit'
import EmptyState from '../../components/EmptyState.vue'
import { usePagination } from '../../hooks/usePagination'
import { formatTime } from '../../utils/format'

const pagination = usePagination<AuditLogItem>((params) => listAuditLogs(params))

const entityText = (e: string) => {
  const map: Record<string, string> = {
    paper: '论文',
    review: '审稿',
    revision: '修稿',
    user: '用户',
    file: '文件',
    audit: '审计'
  }
  return map[e] || e
}

onMounted(() => pagination.load())
</script>
