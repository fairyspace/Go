// s09_interface/main.go —— 对应教学文稿 07 接口与 Options 模式（一）
//
// 运行：go run ./s09_interface
package main

import (
	"demo/s09_interface/study"
	"fmt"
)

func main() {
	// ==================== 结构体实现接口 ====================
	name := "Tom"
	s, err := study.New(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(s.Listen("english"))
	fmt.Println(s.Speak("english"))
	fmt.Println(s.Read("english"))
	fmt.Println(s.Write("english"))

	// 参数校验
	_, err = study.New("")
	fmt.Println("空名字校验：", err)

	// ==================== 类型断言 ====================
	// 单返回值：失败会 panic
	sw := s.(interface{ Write(msg string) string })
	fmt.Println("断言后：", sw.Write("music"))

	// 双返回值：安全
	sw2, ok := s.(interface{ Write(msg string) string })
	if ok {
		fmt.Println("安全断言：", sw2.Write("song"))
	}

	// ==================== 类型 switch ====================
	switch v := s.(type) {
	case interface{ Listen(string) string }:
		fmt.Println("类型 switch：", v.Listen("music"))
	default:
		fmt.Println("未知类型")
	}

	// ==================== 接口组合 ====================
	f := &study.File{Data: []byte("hello")}
	var rw study.ReadWriter = f

	buf := make([]byte, 5)
	n, _ := rw.Read(buf)
	fmt.Println("组合接口 Read：", n, string(buf[:n]))

	n, _ = rw.Write([]byte(" world"))
	fmt.Println("组合接口 Write：", n)
	n, _ = rw.Read(buf)
	fmt.Println("再次 Read：", string(buf[:n]))

	// ==================== 空接口 ====================
	var v interface{}
	v = 1
	v = "string"
	v = []int{1, 2}

	if s2, ok := v.(string); ok {
		fmt.Println("空接口断言：", s2)
	} else {
		fmt.Printf("不是 string，实际类型：%T\n", v)
	}
}
