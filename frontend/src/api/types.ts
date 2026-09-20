export interface UserSummary {
  id: number
  username: string
  email: string
  real_name: string
  institution: string
  role: string
}

export interface Paper {
  id: number
  title: string
  abstract: string
  keywords: string
  subject: string
  authors_meta: string
  file_name: string
  file_key: string
  status: string
  version: number
  submitter_id: number
  submitter?: UserSummary
  initial_review_comment: string
  final_comment: string
  final_decision: string
  similarity: number
  created_at: string
  updated_at: string
  reviews?: ReviewItem[]
  revisions?: RevisionItem[]
}

export interface ReviewItem {
  id: number
  paper_id: number
  paper?: { id: number; title: string; subject: string }
  reviewer_id: number
  reviewer?: UserSummary
  status: string
  decision: string
  comments: string
  confidential_comments: string
  due_date: string
  completed_at: string
  created_at: string
}

export interface RevisionItem {
  id: number
  paper_id: number
  version: number
  file_name: string
  file_key: string
  response_letter: string
  submitted_by_id: number
  submitted_by?: UserSummary
  created_at: string
}

export interface PlagiarismResult {
  id: number
  paper_id: number
  similarity: number
  status: string
  report: string
  checked_at: string
}

export interface PageResult<T> {
  total: number
  page: number
  size: number
  items: T[]
}
