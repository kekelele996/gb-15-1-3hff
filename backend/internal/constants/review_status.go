package constants

// 审稿任务状态枚举。
const (
	ReviewStatusInvited   = "invited"
	ReviewStatusAccepted  = "accepted"
	ReviewStatusDeclined  = "declined"
	ReviewStatusCompleted = "completed"
)

// ReviewStatusList 全部审稿状态。
var ReviewStatusList = []string{
	ReviewStatusInvited,
	ReviewStatusAccepted,
	ReviewStatusDeclined,
	ReviewStatusCompleted,
}
