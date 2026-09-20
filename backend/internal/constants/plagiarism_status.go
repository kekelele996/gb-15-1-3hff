package constants

// 查重检测状态枚举。
const (
	PlagiarismStatusPending   = "pending"
	PlagiarismStatusCompleted = "completed"
	PlagiarismStatusFailed    = "failed"
)

// PlagiarismStatusList 全部查重状态。
var PlagiarismStatusList = []string{
	PlagiarismStatusPending,
	PlagiarismStatusCompleted,
	PlagiarismStatusFailed,
}
