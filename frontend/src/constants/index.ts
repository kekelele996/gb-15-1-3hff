// 与后端 internal/constants 对应共享枚举
export const ROLE_MAP: Record<string, string> = {
  admin: '管理员',
  editor: '编辑',
  reviewer: '审稿人',
  author: '作者'
}

export const PAPER_STATUS_MAP: Record<string, string> = {
  submitted: '已提交',
  initial_review: '初审中',
  external_review: '外审中',
  revision: '修改中',
  accepted: '已录用',
  rejected: '已拒稿'
}

export const PAPER_STATUS_ORDER = [
  'submitted',
  'initial_review',
  'external_review',
  'revision',
  'accepted'
]

export const REVIEW_STATUS_MAP: Record<string, string> = {
  invited: '待接受',
  accepted: '审稿中',
  declined: '已婉拒',
  completed: '已完成'
}

export const REVIEW_DECISION_MAP: Record<string, string> = {
  accept: '录用',
  minor_revision: '小修后录用',
  major_revision: '大修后重审',
  reject: '拒稿'
}

export const PLAGIARISM_STATUS_MAP: Record<string, string> = {
  pending: '检测中',
  completed: '已完成',
  failed: '检测失败'
}

export const SUBJECTS = [
  'computer',
  'mathematics',
  'physics',
  'biology',
  'medicine',
  'economics',
  'management',
  'education',
  'literature',
  'engineering'
]

export const SUBJECT_MAP: Record<string, string> = {
  computer: '计算机科学',
  mathematics: '数学',
  physics: '物理学',
  biology: '生物学',
  medicine: '医学',
  economics: '经济学',
  management: '管理学',
  education: '教育学',
  literature: '文学',
  engineering: '工程技术'
}
