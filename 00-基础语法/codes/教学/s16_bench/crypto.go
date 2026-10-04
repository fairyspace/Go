// s16_bench/crypto.go —— 对应教学文稿 09 性能调优（一）（二）被测代码
//
// 运行基准测试：go test -bench . -benchmem ./s16_bench
package bench

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ---------- MD5 ----------

type md5Enc struct{}

func New() *md5Enc { return &md5Enc{} }

func (m *md5Enc) Encrypt(str string) string {
	s := md5.Sum([]byte(str))
	return hex.EncodeToString(s[:])
}

// ---------- AES ----------

// aesEnc 演示「缓存重量级对象」的正确写法：
// cipher.Block 的创建（密钥扩展）成本较高，应缓存复用（并发安全）；
// 而 CBC 模式对象（CBCEncrypter/CBCDecrypter）有内部游标状态，每次必须新建。
type aesEnc struct {
	block cipher.Block // 缓存，NewCipher 只调用一次
	iv    []byte
}

// NewAES key 和 iv 长度必须都是 16
func NewAES(key, iv string) (*aesEnc, error) {
	if len(key) != 16 || len(iv) != 16 {
		return nil, errors.New("key 和 iv 长度必须都是 16")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	return &aesEnc{block: block, iv: []byte(iv)}, nil
}

// Encrypt AES-CBC 加密
func (a *aesEnc) Encrypt(str string) (string, error) {
	blockSize := a.block.BlockSize()

	// PKCS#7 填充
	padding := blockSize - len(str)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	plaintext := append([]byte(str), padText...)

	ciphertext := make([]byte, blockSize+len(plaintext))
	iv := ciphertext[:blockSize]
	copy(iv, a.iv)

	mode := cipher.NewCBCEncrypter(a.block, iv) // 每次新建：有内部状态，不可复用
	mode.CryptBlocks(ciphertext[blockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES-CBC 解密
func (a *aesEnc) Decrypt(str string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}

	blockSize := a.block.BlockSize()
	if len(ciphertext) < blockSize {
		return "", errors.New("密文长度不足")
	}

	iv := ciphertext[:blockSize]
	ciphertext = ciphertext[blockSize:]

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(a.block, iv) // 每次新建：有内部状态，不可复用
	mode.CryptBlocks(plaintext, ciphertext)

	if len(plaintext) == 0 {
		return "", nil
	}
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) {
		return "", errors.New("填充无效")
	}
	return string(plaintext[:len(plaintext)-padding]), nil
}

// ---------- JWT (HS256) ----------

type token struct {
	secret []byte
}

func NewToken(secret string) *token {
	return &token{secret: []byte(secret)}
}

// Sign 简化版：base64(header).base64(payload).base64(hmac)
func (t *token) Sign(userID int64, username string) (string, error) {
	header := base64URL([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64URL([]byte(`{"userId":` + strconv.FormatInt(userID, 10) + `,"username":"` + username + `"}`))

	signing := header + "." + payload
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(signing))

	return signing + "." + base64URL(mac.Sum(nil)), nil
}

// Parse 验证签名
func (t *token) Parse(tokenString string) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return errors.New("token 格式错误")
	}

	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if base64URL(mac.Sum(nil)) != parts[2] {
		return errors.New("签名不匹配")
	}
	return nil
}

func base64URL(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

// ---------- 字符串拼接：三种写法 ----------

// StringOp1 用 + 拼接
func StringOp1(n int) string {
	str := ""
	for i := 0; i < n; i++ {
		str += "golang"
	}
	return str
}

// StringOp2 用 fmt.Sprintf 拼接
func StringOp2(n int) string {
	str := ""
	for i := 0; i < n; i++ {
		str = fmt.Sprintf("%s%s", str, "golang")
	}
	return str
}

// StringOp3 用 strings.Builder（推荐）
func StringOp3(n int) string {
	var sb strings.Builder
	sb.Grow(n * len("golang"))
	for i := 0; i < n; i++ {
		sb.WriteString("golang")
	}
	return sb.String()
}
