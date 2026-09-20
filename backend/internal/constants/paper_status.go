package constants

// 论文状态机枚举：已提交→初审中→外审中→修改中→已录用/已拒稿。
const (
	PaperStatusSubmitted      = "submitted"
	PaperStatusInitialReview  = "initial_review"
	PaperStatusExternalReview = "external_review"
	PaperStatusRevision       = "revision"
	PaperStatusAccepted       = "accepted"
	PaperStatusRejected       = "rejected"
)

// PaperStatusList 全部论文状态。
var PaperStatusList = []string{
	PaperStatusSubmitted,
	PaperStatusInitialReview,
	PaperStatusExternalReview,
	PaperStatusRevision,
	PaperStatusAccepted,
	PaperStatusRejected,
}
