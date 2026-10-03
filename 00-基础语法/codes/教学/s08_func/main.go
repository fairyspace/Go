// s08_func/main.go —— 对应教学文稿 06 函数、闭包与 defer（一 ~ 三）
//
// 运行：go run ./s08_func
package main

import (
	"fmt"
	"time"
)

type Result struct {
	Code    int
	Message string
}

// 值传递：复制一份，改动不影响外部
func setDataByValue(res Result) {
	res.Code = 500
}

// 引用传递：传地址，改动生效
func setDataByPtr(res *Result) {
	res.Code = 500
}

// 命名返回值 + 裸 return
func getSum(a, b int) (sum int) {
	sum = a + b
	return
}

// 错误处理惯例：error 作为最后一个返回值
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为 0")
	}
	return a / b, nil
}

// 不定参数传递
func Add(a int, args ...int) (result int) {
	result += a
	for _, arg := range args {
		result += arg
	}
	return
}

func sum(nums ...int) int {
	s := 0
	for _, n := range nums {
		s += n
	}
	return s
}

func main() {
	// ==================== 值传递 vs 引用传递 ====================
	fmt.Println("--- 传参 ---")
	res := Result{Code: 200, Message: "success"}
	setDataByValue(res)
	fmt.Println("传值后：", res) // 200，未改变

	setDataByPtr(&res)
	fmt.Println("传址后：", res) // 500，已改变

	// ==================== 命名返回值 ====================
	fmt.Println("\n--- 命名返回值 ---")
	fmt.Println("getSum(1, 2) =", getSum(1, 2))

	// ==================== 错误处理 ====================
	fmt.Println("\n--- 错误处理 ---")
	r, err := divide(10, 0)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(r)

	// ==================== 不定参数 ====================
	fmt.Println("\n--- 不定参数 ---")
	fmt.Println("Add(1, 2, 3) =", Add(1, 2, 3))
	fmt.Println("Add(1) =", Add(1))
	fmt.Println("sum(1, 2, 3) =", sum(1, 2, 3))
	nums := []int{1, 2, 3}
	fmt.Println("sum(nums...) =", sum(nums...)) // 已有切片用 ... 展开传

	// 匿名函数接收不定参数
	f := func(prefix string, args ...interface{}) {
		fmt.Println(append([]interface{}{prefix}, args...)...)
	}
	f("参数：", 1, "a", true)

	// ==================== 闭包 ====================
	fmt.Println("\n--- 闭包 ---")
	var a = 1
	var b = 2
	add := func() {
		fmt.Println("a+b =", a+b)
	}
	a = 3
	add() // 输出 4，而不是 3 —— 闭包捕获的是变量本身

	// 闭包 + 计数器
	newCounter := func() func() int {
		i := 0
		return func() int {
			i++
			return i
		}
	}
	c := newCounter()
	fmt.Println(c(), c(), c())

	// ==================== 工具函数 ====================
	fmt.Println("\n--- 工具函数 ---")
	fmt.Println("当前时间：", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("时间戳：", time.Now().Unix())
}
