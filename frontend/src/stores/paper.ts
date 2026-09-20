import { defineStore } from 'pinia'
import { createPaper, listMyPapers } from '../api/paper'
import type { Paper, PageResult } from '../api/types'

export const usePaperStore = defineStore('paper', {
  state: () => ({
    papers: [] as Paper[],
    total: 0,
    loading: false
  }),
  actions: {
    async fetchMine(params: Record<string, unknown> = {}) {
      this.loading = true
      try {
        const res: PageResult<Paper> = await listMyPapers({ page: 1, size: 20, ...params })
        this.papers = res.items
        this.total = res.total
      } finally {
        this.loading = false
      }
    },
    async create(data: Record<string, unknown>) {
      return createPaper(data)
    }
  }
})
