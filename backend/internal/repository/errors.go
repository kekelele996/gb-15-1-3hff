package repository

import "errors"

// ErrNotFound 仓储层哨兵错误。
var ErrNotFound = errors.New("record not found")

// ErrConflict 仓储层冲突错误。
var ErrConflict = errors.New("record conflict")
