package constants

// 通用响应码。
const (
	CodeOK           = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeConflict     = 40900
	CodeValidation   = 42200
	CodeInternal     = 50000
	CodeTooManyRequests = 42900
)

// 业务错误码（集中维护，message 由各 service/handler 拼接）。
const (
	ErrUserExists          = 40901
	ErrUserNotFound        = 40401
	ErrInvalidCredential   = 40101
	ErrInvalidToken        = 40102
	ErrPermissionDenied    = 40301
	ErrRoleNotAllowed      = 40302
	ErrPaperNotFound       = 40402
	ErrPaperStatusNotAllowed = 40902
	ErrPaperTitleExists    = 40904
	ErrReviewNotFound      = 40403
	ErrReviewNotAllowed    = 40903
	ErrRevisionNotFound    = 40406
	ErrPlagiarismNotFound  = 40404
	ErrAuditNotFound       = 40405
	ErrFileUploadFailed    = 50001
	ErrStorageUnavailable  = 50002
	ErrSubjectNotAllowed   = 42201
)

// 通用业务错误别名（service/handler 引用）。
const (
	ErrBadRequest = CodeBadRequest
	ErrConflict   = CodeConflict
	ErrInternal   = CodeInternal
)
