package constants

// 审计动作枚举。
const (
	AuditActionRegister       = "register"
	AuditActionLogin          = "login"
	AuditActionCreatePaper    = "create_paper"
	AuditActionUpdatePaper    = "update_paper"
	AuditActionInitialReview  = "initial_review"
	AuditActionAssignReviewer = "assign_reviewer"
	AuditActionRespondReview  = "respond_review"
	AuditActionSubmitReview   = "submit_review"
	AuditActionRevisePaper    = "revise_paper"
	AuditActionFinalDecision  = "final_decision"
	AuditActionUploadFile     = "upload_file"
	AuditActionRunPlagiarism  = "run_plagiarism"
)

// AuditActionList 全部审计动作。
var AuditActionList = []string{
	AuditActionRegister,
	AuditActionLogin,
	AuditActionCreatePaper,
	AuditActionUpdatePaper,
	AuditActionInitialReview,
	AuditActionAssignReviewer,
	AuditActionRespondReview,
	AuditActionSubmitReview,
	AuditActionRevisePaper,
	AuditActionFinalDecision,
	AuditActionUploadFile,
	AuditActionRunPlagiarism,
}
