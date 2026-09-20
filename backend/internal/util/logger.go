package util

import (
	"log/slog"
	"os"
	"strings"
)

var globalLogger *slog.Logger

// NewLogger 依据日志级别创建 JSON 结构化 slog logger。
func NewLogger(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
}

// SetLogger 设置全局 logger，供各层复用。
func SetLogger(l *slog.Logger) {
	globalLogger = l
}

// Logger 返回全局 logger。
func Logger() *slog.Logger {
	if globalLogger == nil {
		globalLogger = slog.Default()
	}
	return globalLogger
}

// Info 输出 info 级别日志。
func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

// Warn 输出 warn 级别日志。
func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}

// Debug 输出 debug 级别日志。
func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

// Error 输出 error 级别日志。
func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}
