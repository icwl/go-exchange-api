package coinex

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// 认证签名
func Sign(method, path, body, timestamp, secret string) string {
	preparedStr := method + path + body + timestamp + secret
	hash := sha256.Sum256([]byte(preparedStr))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}
