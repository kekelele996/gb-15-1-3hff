package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
)

// Response 统一响应体：{ code, message, data }。
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: constants.CodeOK, Message: constants.MsgOK, Data: data})
}

// Fail 失败响应（中断后续 handler）。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Response{Code: code, Message: message})
}
