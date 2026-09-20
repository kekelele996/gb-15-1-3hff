package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/paperflow/paperflow/internal/constants"
)

// FormatTime 格式化时间。
func FormatTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

// FormatPaperStatus 论文状态中文文本。
func FormatPaperStatus(status string) string {
	switch status {
	case constants.PaperStatusSubmitted:
		return "已提交"
	case constants.PaperStatusInitialReview:
		return "初审中"
	case constants.PaperStatusExternalReview:
		return "外审中"
	case constants.PaperStatusRevision:
		return "修改中"
	case constants.PaperStatusAccepted:
		return "已录用"
	case constants.PaperStatusRejected:
		return "已拒稿"
	default:
		return status
	}
}

// FormatReviewStatus 审稿状态中文文本。
func FormatReviewStatus(status string) string {
	switch status {
	case constants.ReviewStatusInvited:
		return "待接受"
	case constants.ReviewStatusAccepted:
		return "审稿中"
	case constants.ReviewStatusDeclined:
		return "已婉拒"
	case constants.ReviewStatusCompleted:
		return "已完成"
	default:
		return status
	}
}

// FormatReviewDecision 评审等级中文文本。
func FormatReviewDecision(decision string) string {
	switch decision {
	case constants.ReviewDecisionAccept:
		return "录用"
	case constants.ReviewDecisionMinorRevision:
		return "小修后录用"
	case constants.ReviewDecisionMajorRevision:
		return "大修后重审"
	case constants.ReviewDecisionReject:
		return "拒稿"
	default:
		return decision
	}
}

// FormatRole 角色中文文本。
func FormatRole(role string) string {
	switch role {
	case constants.RoleAdmin:
		return "管理员"
	case constants.RoleEditor:
		return "编辑"
	case constants.RoleReviewer:
		return "审稿人"
	case constants.RoleAuthor:
		return "作者"
	default:
		return role
	}
}

// FormatPlagiarismStatus 查重状态中文文本。
func FormatPlagiarismStatus(status string) string {
	switch status {
	case constants.PlagiarismStatusPending:
		return "检测中"
	case constants.PlagiarismStatusCompleted:
		return "已完成"
	case constants.PlagiarismStatusFailed:
		return "检测失败"
	default:
		return status
	}
}

// FormatPercent 百分比文本。
func FormatPercent(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

// FormatEntityName 实体名中文文本。
func FormatEntityName(entity string) string {
	switch entity {
	case "paper":
		return "论文"
	case "review":
		return "审稿"
	case "revision":
		return "修改"
	case "user":
		return "用户"
	case "plagiarism":
		return "查重"
	case "audit":
		return "审计"
	default:
		return entity
	}
}

// SubjectText 学科分类中文文本。
func SubjectText(subject string) string {
	switch subject {
	case "computer":
		return "计算机科学"
	case "mathematics":
		return "数学"
	case "physics":
		return "物理学"
	case "biology":
		return "生物学"
	case "medicine":
		return "医学"
	case "economics":
		return "经济学"
	case "management":
		return "管理学"
	case "education":
		return "教育学"
	case "literature":
		return "文学"
	case "engineering":
		return "工程技术"
	default:
		return subject
	}
}

// SplitKeywords 拆分关键词。
func SplitKeywords(keywords string) []string {
	parts := strings.Split(keywords, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
