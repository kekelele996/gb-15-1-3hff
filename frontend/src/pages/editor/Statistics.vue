<template>
  <div v-loading="loading">
    <el-row :gutter="16" class="mt-16">
      <el-col v-for="card in cards" :key="card.label" :span="6">
        <el-card shadow="hover">
          <div style="font-size: 13px; color: #909399">{{ card.label }}</div>
          <div style="font-size: 26px; font-weight: 600; color: #303133">{{ card.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mt-16">
      <el-col :span="14">
        <el-card shadow="never">
          <template #header>投稿量趋势（近 30 天）</template>
          <EChart :option="trendOption" :height="320" />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>学科投稿分布</template>
          <EChart :option="subjectOption" :height="320" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mt-16">
      <el-col :span="24">
        <el-card shadow="never">
          <template #header>审稿人工作量排名</template>
          <EChart :option="reviewerOption" :height="280" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  getOverview,
  getReviewerWorkload,
  getSubjects,
  getTrend
} from '../../api/statistics'
import type { Overview } from '../../api/statistics'
import EChart from '../../components/EChart.vue'
import { subjectText } from '../../utils/format'

const loading = ref(false)
const overview = ref<Overview | null>(null)
const trend = ref<Array<{ day: string; count: number }>>([])
const subjects = ref<Array<{ subject: string; count: number }>>([])
const reviewers = ref<Array<{ reviewer_name: string; total: number; completed: number }>>([])

const cards = computed(() => {
  const o = overview.value
  return [
    { label: '投稿总量', value: o?.total ?? 0 },
    { label: '待初审', value: o?.submitted ?? 0 },
    { label: '外审中', value: o?.external_review ?? 0 },
    { label: '已录用', value: o?.accepted ?? 0 },
    { label: '已拒稿', value: o?.rejected ?? 0 },
    { label: '录用率', value: o ? `${o.acceptance_rate.toFixed(1)}%` : '0%' },
    { label: '平均审稿周期(天)', value: o ? o.avg_review_days.toFixed(1) : '0' },
    { label: '修改中', value: o?.revision ?? 0 }
  ]
})

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 40, right: 20, top: 30, bottom: 40 },
  xAxis: { type: 'category', data: trend.value.map((t) => t.day) },
  yAxis: { type: 'value', minInterval: 1 },
  series: [
    {
      name: '投稿量',
      type: 'line',
      smooth: true,
      areaStyle: { opacity: 0.15 },
      data: trend.value.map((t) => t.count)
    }
  ]
}))

const subjectOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
  series: [
    {
      type: 'pie',
      radius: ['40%', '65%'],
      data: subjects.value.map((s) => ({ name: subjectText(s.subject), value: s.count }))
    }
  ]
}))

const reviewerOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 80, right: 40, top: 30, bottom: 40 },
  xAxis: { type: 'value', minInterval: 1 },
  yAxis: { type: 'category', data: reviewers.value.map((r) => r.reviewer_name) },
  series: [
    {
      name: '审稿任务',
      type: 'bar',
      data: reviewers.value.map((r) => r.total)
    }
  ]
}))

onMounted(async () => {
  loading.value = true
  try {
    const [o, t, s, r] = await Promise.all([
      getOverview(),
      getTrend(30),
      getSubjects(),
      getReviewerWorkload()
    ])
    overview.value = o
    trend.value = t
    subjects.value = s
    reviewers.value = r
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
})
</script>
