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

// ReviewOpenStatusList 仍可回应/提交的开放状态（未超期时视为有效任务）。
var ReviewOpenStatusList = []string{
	ReviewStatusInvited,
	ReviewStatusAccepted,
}

// ReviewActiveStatusList 判定“同一人不能同时存在两条有效任务”时占位的状态。
var ReviewActiveStatusList = []string{
	ReviewStatusInvited,
	ReviewStatusAccepted,
	ReviewStatusCompleted,
}
