package handler

import (
	"github.com/gin-gonic/gin"
)

// RequestContext 请求上下文信息
// 封装从 Gin Context 中提取的用户信息和请求元数据
type RequestContext struct {
	UserID    uint   // 用户 ID
	Username  string // 用户名
	ClientIP  string // 客户端 IP
	UserAgent string // User-Agent
}

// GetRequestContext 从 Gin Context 中提取请求上下文信息
// 用于审计日志记录和操作追踪
func GetRequestContext(c *gin.Context) *RequestContext {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	// 类型断言，处理可能的 nil 值
	var uid uint
	var uname string

	if userID != nil {
		uid = userID.(uint)
	}
	if username != nil {
		uname = username.(string)
	}

	return &RequestContext{
		UserID:    uid,
		Username:  uname,
		ClientIP:  c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
}
