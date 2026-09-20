package constants

// 评审等级枚举：录用/小修后录用/大修后重审/拒稿。
const (
	ReviewDecisionAccept        = "accept"
	ReviewDecisionMinorRevision = "minor_revision"
	ReviewDecisionMajorRevision = "major_revision"
	ReviewDecisionReject        = "reject"
)

// ReviewDecisionList 全部评审等级。
var ReviewDecisionList = []string{
	ReviewDecisionAccept,
	ReviewDecisionMinorRevision,
	ReviewDecisionMajorRevision,
	ReviewDecisionReject,
}
