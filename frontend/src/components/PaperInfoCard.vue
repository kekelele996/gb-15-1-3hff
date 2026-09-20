<template>
  <el-card v-if="paper" shadow="never" class="paper-info-card">
    <template #header>
      <div class="card-header">
        <div>
          <span class="paper-title">{{ paper.title }}</span>
          <StatusBadge :status="paper.status" kind="paper" style="margin-left: 8px" />
        </div>
        <el-tag type="info" effect="plain">V{{ paper.version }}</el-tag>
      </div>
    </template>
    <div class="info-grid">
      <div class="info-item">
        <span class="label">学科分类</span>{{ subjectText(paper.subject) }}
      </div>
      <div class="info-item">
        <span class="label">关键词</span>{{ paper.keywords }}
      </div>
      <div class="info-item">
        <span class="label">投稿作者</span>{{ paper.submitter?.real_name || paper.submitter?.username || '-' }}
      </div>
      <div class="info-item">
        <span class="label">查重相似度</span>
        <span :class="{ 'high-similarity': paper.similarity > 30 }">{{ formatPercent(paper.similarity) }}</span>
      </div>
      <div class="info-item">
        <span class="label">投稿时间</span>{{ formatTime(paper.created_at) }}
      </div>
      <div class="info-item">
        <span class="label">稿件文件</span>{{ paper.file_name }}
      </div>
    </div>
    <el-divider content-position="left">摘要</el-divider>
    <p class="abstract">{{ paper.abstract }}</p>
    <el-divider content-position="left">作者信息</el-divider>
    <pre class="authors">{{ prettyAuthors }}</pre>
    <template v-if="paper.initial_review_comment">
      <el-divider content-position="left">初审意见</el-divider>
      <p class="comment">{{ paper.initial_review_comment }}</p>
    </template>
    <template v-if="paper.final_comment">
      <el-divider content-position="left">终审意见</el-divider>
      <p class="comment">{{ paper.final_comment }}</p>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Paper } from '../api/types'
import StatusBadge from './StatusBadge.vue'
import { formatPercent, formatTime, subjectText } from '../utils/format'

const props = defineProps<{ paper: Paper }>()

const prettyAuthors = computed(() => {
  try {
    const arr = JSON.parse(props.paper.authors_meta || '[]') as Array<{ name: string; institution?: string }>
    return arr.map((a) => `${a.name}（${a.institution || '未填单位'}）`).join('\n')
  } catch {
    return props.paper.authors_meta || '-'
  }
})
</script>
