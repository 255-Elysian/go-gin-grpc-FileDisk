package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// RsaDecode 解密
/*
读取内存中的 私钥

解码 Base64 密文

使用 RSA 私钥进行解密，还原为原始明文
*/
func RsaDecode(encryptedData string) (string, error) {
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	block, _ := pem.Decode([]byte(KeyForPwd.PrivateKey))
	privateKey, parseError := x509.ParsePKCS8PrivateKey(block.Bytes)
	if parseError != nil {
		fmt.Println(parseError.Error())
		return "", errors.New("解析私钥失败")
	}
	originalData, encryptError := rsa.DecryptPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), encryptedDecodeBytes)
	if encryptError != nil {
		fmt.Println(encryptError.Error())
	}
	return string(originalData), encryptError
}

// RsaEncode 加密
/*
读取内存中的 公钥

使用 RSA 算法加密明文数据（如密码）

返回一个 Base64 编码的密文
*/
func RsaEncode(plainData string) (string, error) {
	// 解析公钥
	block, _ := pem.Decode([]byte(KeyForPwd.PublicKey))
	if block == nil {
		return "", errors.New("解析公钥失败")
	}

	// 尝试使用ParsePKIXPublicKey解析公钥
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// 对数据进行RSA加密
	encryptedData, encryptError := rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), []byte(plainData))
	if encryptError != nil {
		return "", encryptError
	}

	// 对加密后的数据进行Base64编码
	encryptedDataBase64 := base64.StdEncoding.EncodeToString(encryptedData)
	return encryptedDataBase64, nil
}
