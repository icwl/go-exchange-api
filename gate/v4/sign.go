package gate

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func Sign(method, path, query string, body []byte, timestamp, secret string) string {
	h := sha512.New()
	if body != nil {
		h.Write(body)
	}
	payload := hex.EncodeToString(h.Sum(nil))
	s := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, path, query, payload, timestamp)
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(s))
	return hex.EncodeToString(mac.Sum(nil))
}
