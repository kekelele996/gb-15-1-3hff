import { defineStore } from 'pinia'
import { listMyReviews, respondReview, submitReview } from '../api/review'
import type { PageResult, ReviewItem } from '../api/types'

export const useReviewStore = defineStore('review', {
  state: () => ({
    reviews: [] as ReviewItem[],
    total: 0,
    loading: false
  }),
  actions: {
    async fetchMine(params: Record<string, unknown> = {}) {
      this.loading = true
      try {
        const res: PageResult<ReviewItem> = await listMyReviews({ page: 1, size: 20, ...params })
        this.reviews = res.items
        this.total = res.total
      } finally {
        this.loading = false
      }
    },
    async respond(id: number, accept: boolean) {
      await respondReview(id, { accept })
    },
    async submit(id: number, data: Record<string, unknown>) {
      await submitReview(id, data)
    }
  }
})
