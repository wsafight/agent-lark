package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const encPrefix = "enc:"

func masterKeyPath() string {
	return filepath.Join(HomeDir(), "master.key")
}

// loadOrCreateMasterKey reads the AES-256 key from ~/.agent-lark/master.key,
// creating a random one if it does not exist.
func loadOrCreateMasterKey() ([]byte, error) {
	path := masterKeyPath()
	if data, err := os.ReadFile(path); err == nil {
		key, decErr := hex.DecodeString(strings.TrimSpace(string(data)))
		if decErr == nil && len(key) == 32 {
			return key, nil
		}
		// corrupt — regenerate
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("生成加密密钥失败：%w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, fmt.Errorf("保存加密密钥失败：%w", err)
	}
	return key, nil
}

// encryptField encrypts a plaintext string with AES-256-GCM and returns
// "enc:<base64(nonce+ciphertext)>". Empty input returns empty output.
func encryptField(key []byte, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
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
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptField decrypts a value produced by encryptField. Values that do not
// start with the "enc:" prefix are returned unchanged (backward compatibility
// with old plaintext configs).
func decryptField(key []byte, encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	if !strings.HasPrefix(encoded, encPrefix) {
		return encoded, nil // old plaintext — pass through
	}
	data, err := base64.StdEncoding.DecodeString(encoded[len(encPrefix):])
	if err != nil {
		return "", fmt.Errorf("解码失败：%w", err)
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
		return "", errors.New("密文格式错误")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("解密失败：%w", err)
	}
	return string(plaintext), nil
}
