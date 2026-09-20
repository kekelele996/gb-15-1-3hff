import request from '../utils/request'
import type { PageResult, Paper, PlagiarismResult, ReviewItem, RevisionItem } from './types'

export function createPaper(data: Record<string, unknown>) {
  return request.post('/papers', data) as Promise<Paper>
}

export function listMyPapers(params: Record<string, unknown>) {
  return request.get('/papers/mine', { params }) as Promise<PageResult<Paper>>
}

export function listPapers(params: Record<string, unknown>) {
  return request.get('/papers', { params }) as Promise<PageResult<Paper>>
}

export function getPaper(id: number | string) {
  return request.get(`/papers/${id}`) as Promise<Paper>
}

export function updatePaper(id: number | string, data: Record<string, unknown>) {
  return request.put(`/papers/${id}`, data) as Promise<Paper>
}

export function initialReview(id: number | string, data: Record<string, unknown>) {
  return request.post(`/papers/${id}/initial-review`, data) as Promise<Paper>
}

export function finalDecision(id: number | string, data: Record<string, unknown>) {
  return request.post(`/papers/${id}/final-decision`, data) as Promise<Paper>
}

export function revisePaper(id: number | string, data: Record<string, unknown>) {
  return request.post(`/papers/${id}/revise`, data) as Promise<Paper>
}

export function searchLibrary(params: Record<string, unknown>) {
  return request.get('/library/papers', { params }) as Promise<PageResult<Paper>>
}

export function listPaperReviews(paperId: number | string) {
  return request.get(`/reviews/paper/${paperId}`) as Promise<ReviewItem[]>
}

export function listPaperRevisions(paperId: number | string) {
  return request.get(`/papers/${paperId}/revisions`) as Promise<RevisionItem[]>
}

export function getPlagiarism(paperId: number | string) {
  return request.get(`/papers/${paperId}/plagiarism`) as Promise<PlagiarismResult>
}

export function rerunPlagiarism(paperId: number | string) {
  return request.post(`/papers/${paperId}/plagiarism/rerun`) as Promise<PlagiarismResult>
}

export function listReviewers() {
  return request.get('/users/reviewers') as Promise<Array<{ id: number; real_name: string; username: string }>>
}
