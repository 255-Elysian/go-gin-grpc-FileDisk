package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

var HashSalt = "123456abcdef"

// GetHash 进行加盐哈希
func GetHash(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	hash.Write([]byte(HashSalt))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
