package constants

// 接口/日志/错误提示文案（集中管理，service/handler 拼接业务信息后引用）。
const (
	MsgOK            = "ok"
	MsgSuccess       = "success"
	MsgUnauthorized  = "未登录或登录已过期，请重新登录"
	MsgForbidden     = "无权访问该资源"
	MsgBadRequest    = "请求参数错误"
	MsgInternalError = "服务器内部错误"
	MsgNotFound      = "资源不存在"
	MsgUserExists    = "用户名已被占用"
	MsgInvalidCredential = "用户名或密码错误"
	MsgPaperNotFound = "论文不存在"
	MsgReviewNotFound = "审稿任务不存在"
	MsgFileTooLarge  = "文件大小超过限制（50MB）"
	MsgFileTypeNotAllowed = "仅支持 PDF / Word 格式文件"
	MsgRateLimited   = "请求过于频繁，请稍后再试"
)
