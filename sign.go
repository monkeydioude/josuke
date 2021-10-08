package josuke

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

// returns a hexadecimal HMAC SHA 256 digest
func hmacSha256(secret []byte, data string) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}