package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"
)

func evpBytesToKey(password []byte, salt []byte, keyLen int) []byte {
	var key = make([]byte, 0, keyLen)
	var digest []byte
	for len(key) < keyLen {
		data := append(digest, password...)
		data = append(data, salt...)
		sum := md5.Sum(data)
		digest = sum[:]
		key = append(key, digest...)
	}
	return key[:keyLen]
}

func EncryptHash(originalHash string, secret string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	location, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(location)

	// Timestamp com precisão de milissegundos no formato desejado
	createdAt := now.Format("2006-01-02 15:04:05")

	// created_at em milissegundos desde 1970
	formattedTimestamp := now.UnixNano() / int64(time.Millisecond)

	payload := map[string]any{
		"hash":       originalHash,
		"timestamp":  formattedTimestamp,
		"created_at": createdAt,
	}

	plain, _ := json.Marshal(payload)
	keyIv := evpBytesToKey([]byte(secret), salt, 48)
	key := keyIv[:32]
	iv := keyIv[32:]

	block, _ := aes.NewCipher(key)
	mode := cipher.NewCBCEncrypter(block, iv)

	padding := aes.BlockSize - len(plain)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	padded := append(plain, padtext...)

	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	prefix := append([]byte("Salted__"), salt...)
	full := append(prefix, ciphertext...)

	return base64.StdEncoding.EncodeToString(full), nil
}
