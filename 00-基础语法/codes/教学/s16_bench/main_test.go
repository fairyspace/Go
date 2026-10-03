// s16_bench/crypto_test.go —— 对应教学文稿 09 性能调优（一）（二）基准测试
//
// 运行：go test -bench . -benchmem ./s16_bench
package bench

import (
	"bytes"
	"fmt"
	"testing"
)

// ---------- 一、签名算法基准测试 ----------

func BenchmarkMD5(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New().Encrypt("123456")
	}
}

func BenchmarkAES(b *testing.B) {
	b.ResetTimer()
	aes, err := NewAES("0123456789abcdef", "fedcba9876543210")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		encryptString, _ := aes.Encrypt("123456")
		aes.Decrypt(encryptString)
	}
}

func BenchmarkJWT(b *testing.B) {
	b.ResetTimer()
	token := NewToken("secret")
	for i := 0; i < b.N; i++ {
		tokenString, _ := token.Sign(123456789, "xinliangnote")
		token.Parse(tokenString)
	}
}

// ---------- 二、字符串拼接：三种写法 ----------

func BenchmarkStringOp1(b *testing.B) { // +
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringOp1(1000)
	}
}

func BenchmarkStringOp2(b *testing.B) { // fmt.Sprintf
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringOp2(1000)
	}
}

func BenchmarkStringOp3(b *testing.B) { // strings.Builder
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringOp3(1000)
	}
}

// ---------- 三、map[string]interface{} vs 临时 Struct ----------

type demoStruct struct {
	Name string
	Age  int
}

func BenchmarkMapOperation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var demo = map[string]interface{}{}
		demo["Name"] = "Tom"
		demo["Age"] = 30
	}
}

func BenchmarkStructOperation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var demo demoStruct
		demo.Name = "Tom"
		demo.Age = 30
	}
}

// ---------- 正确性验证（非基准） ----------

func TestStringOp(t *testing.T) {
	want := ""
	for i := 0; i < 10; i++ {
		want += "golang"
	}
	for name, got := range map[string]string{
		"Op1 +":               StringOp1(10),
		"Op2 fmt.Sprintf":     StringOp2(10),
		"Op3 strings.Builder": StringOp3(10),
	} {
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestAESRoundTrip(t *testing.T) {
	a, err := NewAES("0123456789abcdef", "fedcba9876543210")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := a.Encrypt("Hello, Go World")
	if err != nil {
		t.Fatal(err)
	}
	dec, err := a.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "Hello, Go World" {
		t.Errorf("解密结果 = %q", dec)
	}
}

func TestJWTRoundTrip(t *testing.T) {
	tk := NewToken("secret")
	s, err := tk.Sign(123456789, "xinliangnote")
	if err != nil {
		t.Fatal(err)
	}
	if err := tk.Parse(s); err != nil {
		t.Errorf("验证失败: %v", err)
	}
	// 错误签名应验证失败
	if err := tk.Parse(s + "x"); err == nil {
		t.Error("被篡改的 token 应验证失败")
	}
}

var _ = fmt.Sprintf
var _ = bytes.NewBuffer
