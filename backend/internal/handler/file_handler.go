package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

const maxUploadSize = 50 << 20 // 50MB

// FileHandler 文件上传处理器。
type FileHandler struct {
	storage  *service.StorageService
	auditSvc *service.AuditLogService
	logger   *slog.Logger
}

// NewFileHandler 构造文件上传处理器。
func NewFileHandler(storage *service.StorageService, auditSvc *service.AuditLogService, logger *slog.Logger) *FileHandler {
	return &FileHandler{storage: storage, auditSvc: auditSvc, logger: logger}
}

// Upload 上传论文文件到 MinIO，返回对象 key。
func (h *FileHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "上传失败：缺少文件字段 file")
		return
	}
	defer file.Close()
	if header.Size > maxUploadSize {
		util.Fail(c, http.StatusBadRequest, constants.ErrFileUploadFailed, constants.MsgFileTooLarge)
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
		util.Fail(c, http.StatusBadRequest, constants.ErrFileUploadFailed, constants.MsgFileTypeNotAllowed)
		return
	}
	key := fmt.Sprintf("papers/%s%s", uuid.NewString(), ext)
	if err := h.storage.Upload(c.Request.Context(), key, file, header.Size, header.Header.Get("Content-Type")); err != nil {
		h.logger.Error("upload file failed", "error", err)
		util.Fail(c, http.StatusInternalServerError, constants.ErrStorageUnavailable, "上传失败：对象存储不可用")
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionUploadFile, "file", key, "上传文件："+header.Filename, c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit upload failed", "error", e)
	}
	util.OK(c, gin.H{"key": key, "name": header.Filename, "size": header.Size})
}
