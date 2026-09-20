import request from '../utils/request'
import type { PageResult, ReviewItem, ReviewSummary } from './types'

export function listMyReviews(params: Record<string, unknown>) {
  return request.get('/reviews/mine', { params }) as Promise<PageResult<ReviewItem>>
}

export function listPaperReviews(paperId: number | string) {
  return request.get(`/reviews/paper/${paperId}`) as Promise<ReviewItem[]>
}

export function getReviewSummary(paperId: number | string) {
  return request.get(`/reviews/paper/${paperId}/summary`) as Promise<ReviewSummary>
}

export function respondReview(id: number | string, data: { accept: boolean }) {
  return request.post(`/reviews/${id}/respond`, data) as Promise<{ review_id: number; accepted: boolean }>
}

export function submitReview(id: number | string, data: Record<string, unknown>) {
  return request.post(`/reviews/${id}/submit`, data) as Promise<{ review_id: number; decision: string }>
}

export function assignReviewer(paperId: number | string, reviewerId: number) {
  return request.post(`/papers/${paperId}/reviewers`, { reviewer_id: reviewerId }) as Promise<{ review_id: number }>
}
