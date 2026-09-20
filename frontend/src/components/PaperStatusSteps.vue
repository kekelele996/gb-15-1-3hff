<template>
  <el-steps :active="active" align-center finish-status="success" class="paper-steps">
    <el-step title="已提交" description="稿件已提交至期刊" />
    <el-step title="初审中" description="编辑部格式与选题初审" />
    <el-step title="外审中" description="同行专家评审" />
    <el-step title="修改中" description="按意见修改并重投" />
    <el-step :title="endTitle" :description="endDesc" />
  </el-steps>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const active = computed(() => {
  switch (props.status) {
    case 'submitted':
      return 1
    case 'initial_review':
      return 2
    case 'external_review':
      return 3
    case 'revision':
      return 4
    case 'accepted':
      return 5
    case 'rejected':
      return 4
    default:
      return 0
  }
})

const endTitle = computed(() => (props.status === 'rejected' ? '已拒稿' : '已录用'))
const endDesc = computed(() =>
  props.status === 'rejected' ? '稿件未通过评审' : '稿件被期刊收录'
)
</script>
