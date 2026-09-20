import {
  PAPER_STATUS_MAP,
  REVIEW_STATUS_MAP,
  REVIEW_DECISION_MAP,
  ROLE_MAP,
  PLAGIARISM_STATUS_MAP,
  SUBJECT_MAP
} from '../constants'

export function formatTime(t?: string | null): string {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 16)
}

export function paperStatusText(s: string): string {
  return PAPER_STATUS_MAP[s] || s
}

export function reviewStatusText(s: string): string {
  return REVIEW_STATUS_MAP[s] || s
}

export function reviewDecisionText(d: string): string {
  return REVIEW_DECISION_MAP[d] || d
}

export function roleText(r: string): string {
  return ROLE_MAP[r] || r
}

export function plagiarismStatusText(s: string): string {
  return PLAGIARISM_STATUS_MAP[s] || s
}

export function subjectText(s: string): string {
  return SUBJECT_MAP[s] || s
}

export function formatPercent(v: number): string {
  return `${v.toFixed(1)}%`
}

export function formatSize(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let n = bytes
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return `${n.toFixed(1)} ${units[i]}`
}
