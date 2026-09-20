package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 生成密码哈希。
func HashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

// CheckPassword 校验明文密码与哈希是否一致。
func CheckPassword(hashed, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pwd)) == nil
}
