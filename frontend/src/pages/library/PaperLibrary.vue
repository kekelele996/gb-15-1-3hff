<template>
  <el-card shadow="never">
    <template #header>
      <div class="row-between">
        <span>论文库（已录用 {{ pagination.total.value }} 篇）</span>
        <div style="display: flex; gap: 8px">
          <el-input v-model="keyword" placeholder="标题 / 摘要 / 关键词" clearable style="width: 240px" @keyup.enter="search" />
          <el-select v-model="subject" placeholder="全部学科" clearable style="width: 160px">
            <el-option v-for="(text, key) in SUBJECT_MAP" :key="key" :label="text" :value="key" />
          </el-select>
          <el-button type="primary" @click="search">检索</el-button>
        </div>
      </div>
    </template>
    <el-table :data="pagination.items.value" v-loading="pagination.loading.value" stripe>
      <el-table-column prop="title" label="标题" min-width="260" show-overflow-tooltip />
      <el-table-column label="学科" width="120">
        <template #default="{ row }">{{ subjectText(row.subject) }}</template>
      </el-table-column>
      <el-table-column prop="keywords" label="关键词" min-width="180" show-overflow-tooltip />
      <el-table-column label="作者" width="120">
        <template #default="{ row }">{{ row.submitter?.real_name || row.submitter?.username || '-' }}</template>
      </el-table-column>
      <el-table-column label="录用时间" width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>
    <EmptyState
      v-if="pagination.total.value === 0 && !pagination.loading.value"
      description="没有找到符合条件的论文"
    />
    <el-pagination
      v-model:current-page="pagination.page.value"
      v-model:page-size="pagination.size.value"
      :total="pagination.total.value"
      layout="total, prev, pager, next"
      @current-change="() => pagination.load(query())"
      class="pager"
    />
  </el-card>

  <el-drawer v-model="drawer" size="520px" title="论文详情">
    <PaperInfoCard v-if="current" :paper="current" />
    <EmptyState v-else description="无数据" />
  </el-drawer>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { SUBJECT_MAP } from '../../constants'
import { getPaper, searchLibrary } from '../../api/paper'
import type { Paper } from '../../api/types'
import EmptyState from '../../components/EmptyState.vue'
import PaperInfoCard from '../../components/PaperInfoCard.vue'
import { usePagination } from '../../hooks/usePagination'
import { formatTime, subjectText } from '../../utils/format'

const keyword = ref('')
const subject = ref('')
const drawer = ref(false)
const current = ref<Paper | null>(null)
const pagination = usePagination<Paper>((params) => searchLibrary({ ...params, ...query() }))

function query() {
  return { keyword: keyword.value, subject: subject.value }
}

function search() {
  pagination.load(query())
}

async function openDetail(row: Paper) {
  current.value = await getPaper(row.id)
  drawer.value = true
}

onMounted(() => pagination.load())
</script>
