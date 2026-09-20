package util

import "github.com/gin-gonic/gin"

const (
	ctxUserID    = "auth_user_id"
	ctxUsername  = "auth_username"
	ctxRole      = "auth_role"
	ctxRequestID = "request_id"
)

// SetAuthUser 写入当前认证用户信息。
func SetAuthUser(c *gin.Context, id uint, username, role string) {
	c.Set(ctxUserID, id)
	c.Set(ctxUsername, username)
	c.Set(ctxRole, role)
}

// GetUserID 获取当前用户 ID。
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(ctxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUsername 获取当前用户名。
func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(ctxUsername); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetRole 获取当前用户角色。
func GetRole(c *gin.Context) string {
	if v, ok := c.Get(ctxRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetRequestID 获取请求 ID。
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
