// demo_25.go
package main

import (
	"fmt"
	"time"
)

func main() {
	str := "abcdef"
	fmt.Printf("MD5(%s): %s\n", str, MD5("abcdef"))
	fmt.Printf("current time str : %s\n", getTimeStr1())
	fmt.Printf("current time unix : %d\n", getTimeInt1())
}

// 获取当前时间字符串
func getTimeStr1() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 获取当前时间戳
func getTimeInt1() int64 {
	return time.Now().Unix()
}
