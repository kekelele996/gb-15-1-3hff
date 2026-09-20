package constants

// 审稿任务状态枚举。
const (
	ReviewStatusInvited   = "invited"
	ReviewStatusAccepted  = "accepted"
	ReviewStatusDeclined  = "declined"
	ReviewStatusCompleted = "completed"
	ReviewStatusExpired   = "expired"
)

// ReviewStatusList 全部审稿状态。
var ReviewStatusList = []string{
	ReviewStatusInvited,
	ReviewStatusAccepted,
	ReviewStatusDeclined,
	ReviewStatusCompleted,
	ReviewStatusExpired,
}

// ReviewActiveStatuses 仍处有效期、可被审稿人响应的审稿状态。
var ReviewActiveStatuses = []string{
	ReviewStatusInvited,
	ReviewStatusAccepted,
}

// FinalDecisionMinCompletedReviewers 终审所需的最少完成评审的有效审稿人数量。
const FinalDecisionMinCompletedReviewers = 2
