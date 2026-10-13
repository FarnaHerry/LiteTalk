// Package auth 账号安全：bcrypt 密码、JWT 访问令牌、Redis 刷新令牌旋转、邮箱验证码。
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost 按安全要点 §8 固定 12。
const bcryptCost = 12

// HashPassword bcrypt(12) 哈希。
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	return string(b), err
}

// CheckPassword 校验密码。
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// ErrInvalidToken 令牌非法或过期。
var ErrInvalidToken = errors.New("invalid token")

// RandomToken 生成 n 字节随机串（hex 编码），用于刷新令牌。
func RandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken 刷新令牌落库存哈希（SHA-256），明文只出现在响应/请求中。
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// RandomCode 生成 6 位数字验证码。
func RandomCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return fmt.Sprintf("%06d", n%1000000), nil
}
