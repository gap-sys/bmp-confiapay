package auth

import (
	"cobranca-bmp/config"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Gera um token que será passado ao MBP para obtenção do token de acesso a API deles.O token deve ser assinado com  RS256.
func GenerateToken(tokenID, clientID, aud, privateKey string) (string, error) {
	key, err := GetKey(privateKey)
	if err != nil {
		return "", err
	}
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	claims := make(jwt.MapClaims)
	claims["jti"] = tokenID
	claims["sub"] = clientID
	claims["iat"] = time.Now().In(loc).Unix()
	claims["nbf"] = time.Now().In(loc).Unix()
	claims["exp"] = time.Now().In(loc).Add(config.REDIS_TOKEN_EXP).Unix()
	claims["iss"] = clientID
	claims["aud"] = config.JWT_AUD
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return signedToken, nil

}

// Gera um objeto do tipo  *rsa.PrivateKey, que é obrigatório para assinar tokens JWT com algoritmo RS256.
func GetKey(privateKey string) (*rsa.PrivateKey, error) {
	keystring := privateKey
	block, _ := pem.Decode([]byte(keystring))
	if block == nil {
		return nil, fmt.Errorf("chave privada inválida")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err != nil {
		return nil, err

	}
	key, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("chave privada deve ser do tipo rsa.PrivateKey")
	}
	return key, nil

}
