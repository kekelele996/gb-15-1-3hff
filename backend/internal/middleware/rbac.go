package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/util"
)

// RequireRoles 校验当前用户角色是否在允许集合内。
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := util.GetRole(c)
		if _, ok := allowed[role]; !ok {
			util.Fail(c, http.StatusForbidden, constants.ErrPermissionDenied,
				"权限不足：当前角色 "+util.FormatRole(role)+" 无权执行该操作")
			return
		}
		c.Next()
	}
}
