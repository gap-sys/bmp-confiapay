package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generateRedisKey gera uma chave única para o Redis baseada na requisição
func GenerateIdempotencyKey(idempotencyKey string) string {
	hash := sha256.Sum256([]byte(idempotencyKey))
	return fmt.Sprintf("idempotency:%s", hex.EncodeToString(hash[:]))
}
