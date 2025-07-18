package token

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type RS struct {
	PublicKey  string
	PrivateKey string
}

// Encode 编码，使用私钥对 claims 生成签名
func (rs *RS) Encode(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	pKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(rs.PrivateKey))
	if err != nil {
		fmt.Println("私钥解析失败：", err.Error())
		return "", err
	}
	signature, err := token.SignedString(pKey)
	if err != nil {
		fmt.Println("签名失败：", err.Error())
	}
	return signature, nil
}

// Decode 解码，用公钥验证签名是否合法
func (rs *RS) Decode(signature string, claims jwt.Claims) error {
	_, err := jwt.ParseWithClaims(signature, claims, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM([]byte(rs.PublicKey))
	})
	if err != nil {
		fmt.Println("解码失败：", err.Error())
	}
	return err
}

// IssueRS 签发用户Token，生成 RS256 JWT
func IssueRS(userID uint, expTime time.Time) (string, error) {
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	signature, err := Rs.Encode(claims)
	if err != nil {
		fmt.Println(err.Error())
	}
	return signature, err
}

// CheckRS 校验 JWT 签名
func CheckRS(signature string) error {
	claims := UserClaims{}
	err := Rs.Decode(signature, &claims)
	return err
}
