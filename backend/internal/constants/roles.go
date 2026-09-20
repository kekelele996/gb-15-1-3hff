package constants

// 角色枚举：admin/editor/reviewer/author。
const (
	RoleAdmin    = "admin"
	RoleEditor   = "editor"
	RoleReviewer = "reviewer"
	RoleAuthor   = "author"
)

// AllRoles 全部角色列表。
var AllRoles = []string{RoleAdmin, RoleEditor, RoleReviewer, RoleAuthor}
