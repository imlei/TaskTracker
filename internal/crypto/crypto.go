package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const keyFileName = "encryption.key"

// LoadOrCreateKey 从 dataDir 加载或创建 32 字节 AES 密钥
func LoadOrCreateKey(dataDir string) ([]byte, error) {
	keyPath := filepath.Join(dataDir, keyFileName)
	data, err := os.ReadFile(keyPath)
	if err == nil {
		key, err := hex.DecodeString(string(data))
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, err
	}
	return key, nil
}

// Encrypt 使用 AES-256-GCM 加密，返回 hex 编码的 nonce+ciphertext
func Encrypt(key, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return "enc:" + hex.EncodeToString(sealed), nil
}

// Decrypt 解密 Encrypt 产生的密文；如果不是加密格式（无 enc: 前缀），原样返回（兼容旧明文）
func Decrypt(key []byte, ciphertext string) (string, error) {
	if len(ciphertext) == 0 {
		return "", nil
	}
	// 兼容旧明文数据
	if len(ciphertext) < 4 || ciphertext[:4] != "enc:" {
		return ciphertext, nil
	}
	data, err := hex.DecodeString(ciphertext[4:])
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
