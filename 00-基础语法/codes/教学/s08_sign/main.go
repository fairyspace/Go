// s08_sign/main.go —— 对应教学文稿 06 函数、闭包与 defer（四）实战：生成签名
//
// 运行：go run ./s08_sign
package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
)

// MD5 方法
func MD5(str string) string {
	s := md5.New()
	s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

// 生成签名
//
// 签名四要素（详见第 09 篇）：
//   - 可变性：每次签名不同
//   - 时效性：带时间戳，过期作废
//   - 唯一性：带随机 nonce
//   - 完整性：MD5 摘要防篡改
//
// 此示例实现了「唯一性 + 完整性」，生产环境请补齐时效性。
func createSign(params map[string]interface{}, secret string) string {
	// 1. 取出所有 key
	var key []string
	for k := range params {
		key = append(key, k)
	}

	// 2. 排序（map 遍历无序，必须排序）
	sort.Strings(key)

	// 3. 拼接字符串
	var str = ""
	for i := 0; i < len(key); i++ {
		if i == 0 {
			str = fmt.Sprintf("%v=%v", key[i], params[key[i]])
		} else {
			str = str + fmt.Sprintf("&xl_%v=%v", key[i], params[key[i]])
		}
	}

	// 4. 双重 MD5
	return MD5(MD5(str) + MD5(secret))
}

func main() {
	params := map[string]interface{}{
		"name": "Tom",
		"pwd":  "123456",
		"age":  30,
	}
	fmt.Printf("sign : %s\n", createSign(params, "123456789"))

	// 同样参数、同样 secret，签名固定（可被重放）
	fmt.Printf("sign : %s\n", createSign(params, "123456789"))

	// secret 不同，签名不同
	fmt.Printf("sign : %s\n", createSign(params, "other-secret"))
}
