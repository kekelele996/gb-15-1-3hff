<template>
  <el-tag :type="tagType" effect="light" round>{{ text }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  PAPER_STATUS_MAP,
  REVIEW_STATUS_MAP,
  REVIEW_DECISION_MAP,
  ROLE_MAP,
  PLAGIARISM_STATUS_MAP
} from '../constants'

const props = defineProps<{ status: string; kind?: 'paper' | 'review' | 'decision' | 'role' | 'plagiarism' }>()

const map = computed(() => {
  switch (props.kind || 'paper') {
    case 'review':
      return REVIEW_STATUS_MAP
    case 'decision':
      return REVIEW_DECISION_MAP
    case 'role':
      return ROLE_MAP
    case 'plagiarism':
      return PLAGIARISM_STATUS_MAP
    default:
      return PAPER_STATUS_MAP
  }
})

const text = computed(() => map.value[props.status] || props.status)

const tagType = computed(() => {
  if (props.kind === 'decision') {
    if (props.status === 'accept') return 'success'
    if (props.status === 'reject') return 'danger'
    return 'warning'
  }
  switch (props.status) {
    case 'accepted':
    case 'completed':
      return 'success'
    case 'rejected':
    case 'declined':
    case 'failed':
      return 'danger'
    case 'submitted':
    case 'invited':
    case 'pending':
      return 'info'
    default:
      return 'warning'
  }
})
</script>
