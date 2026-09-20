import request from '../utils/request'

export interface Overview {
  total: number
  submitted: number
  initial_review: number
  external_review: number
  revision: number
  accepted: number
  rejected: number
  acceptance_rate: number
  avg_review_days: number
}

export interface TrendPoint {
  day: string
  count: number
}

export interface SubjectCount {
  subject: string
  count: number
}

export interface ReviewerLoad {
  reviewer_id: number
  reviewer_name: string
  total: number
  completed: number
}

export function getOverview() {
  return request.get('/statistics/overview') as Promise<Overview>
}

export function getTrend(days = 30) {
  return request.get('/statistics/trend', { params: { days } }) as Promise<TrendPoint[]>
}

export function getSubjects() {
  return request.get('/statistics/subjects') as Promise<SubjectCount[]>
}

export function getReviewerWorkload() {
  return request.get('/statistics/reviewers') as Promise<ReviewerLoad[]>
}
