package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	// 固定 32 字节测试密钥
	return []byte("0123456789abcdef0123456789abcdef")
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey(t)
	tests := []struct {
		name      string
		plaintext string
	}{
		{"普通英文字符串", "hello world"},
		{"中文字符串", "飞书多维表格测试"},
		{"长字符串", strings.Repeat("abcdefgh", 1000)},
		{"含特殊字符", "key=value&foo=bar\nnewline\ttab"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := encryptField(key, tc.plaintext)
			if err != nil {
				t.Fatalf("加密失败：%v", err)
			}
			if !strings.HasPrefix(encrypted, encPrefix) {
				t.Fatalf("密文缺少 %q 前缀：%q", encPrefix, encrypted)
			}

			decrypted, err := decryptField(key, encrypted)
			if err != nil {
				t.Fatalf("解密失败：%v", err)
			}
			if decrypted != tc.plaintext {
				t.Errorf("往返不一致：got %q, want %q", decrypted, tc.plaintext)
			}
		})
	}
}

func TestEncryptFieldEmpty(t *testing.T) {
	key := testKey(t)
	out, err := encryptField(key, "")
	if err != nil {
		t.Fatalf("意外错误：%v", err)
	}
	if out != "" {
		t.Errorf("空字符串加密应返回空字符串，got %q", out)
	}
}

func TestDecryptFieldPlaintextPassthrough(t *testing.T) {
	key := testKey(t)
	tests := []struct {
		name  string
		input string
	}{
		{"普通明文", "cli_abcdef123456"},
		{"空字符串", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := decryptField(key, tc.input)
			if err != nil {
				t.Fatalf("意外错误：%v", err)
			}
			if out != tc.input {
				t.Errorf("明文透传失败：got %q, want %q", out, tc.input)
			}
		})
	}
}

func TestDecryptFieldInvalidBase64(t *testing.T) {
	key := testKey(t)
	_, err := decryptField(key, encPrefix+"!!!not-base64!!!")
	if err == nil {
		t.Fatal("期望 base64 解码错误，但未返回错误")
	}
}

func TestDecryptFieldTamperedCiphertext(t *testing.T) {
	key := testKey(t)
	encrypted, err := encryptField(key, "secret")
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	// 篡改 base64 数据的最后一个字节
	b64 := encrypted[len(encPrefix):]
	data, _ := base64.StdEncoding.DecodeString(b64)
	data[len(data)-1] ^= 0xff
	tampered := encPrefix + base64.StdEncoding.EncodeToString(data)

	_, err = decryptField(key, tampered)
	if err == nil {
		t.Fatal("期望篡改密文解密失败，但未返回错误")
	}
}

func TestDecryptFieldShortData(t *testing.T) {
	key := testKey(t)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	// 构造比 nonce size 短的数据
	short := make([]byte, gcm.NonceSize()-1)
	encoded := encPrefix + base64.StdEncoding.EncodeToString(short)

	_, err := decryptField(key, encoded)
	if err == nil {
		t.Fatal("期望短数据解密失败，但未返回错误")
	}
}
