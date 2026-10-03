// s08_defer/main.go —— 对应教学文稿 06 函数、闭包与 defer（四）defer 的五个坑
//
// 运行：go run ./s08_defer
package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func calc(index string, a, b int) int {
	ret := a + b
	fmt.Println(index, a, b, ret)
	return ret
}

// ============ 经典题 ============

func quiz() {
	x := 1
	y := 2
	defer calc("A", x, calc("B", x, y))
	x = 3
	defer calc("C", x, calc("D", x, y))
	y = 4
}

// ============ 坑 1：执行顺序是 LIFO ============

func pitfallLIFO() {
	defer fmt.Println("1")
	defer fmt.Println("2")
	defer fmt.Println("3")

	fmt.Println("main")
}

// ============ 坑 2：闭包 vs 传参 ============

func pitfallClosure() {
	fmt.Println("--- defer fmt.Println(a+b) ---")
	{
		var a = 1
		var b = 2
		defer fmt.Println(a + b) // 参数在 defer 时就确定了
		a = 2
		fmt.Println("main")
	}

	fmt.Println("--- defer func(){...}() ---")
	{
		var a = 1
		var b = 2
		defer func() {
			fmt.Println(a + b) // 函数体内变量在执行时才确定
		}()
		a = 2
		fmt.Println("main")
	}

	fmt.Println("--- defer func(a,b){...}(a,b) ---")
	{
		var a = 1
		var b = 2
		defer func(a int, b int) {
			fmt.Println(a + b) // 显式传参 → 值复制
		}(a, b)
		a = 2
		fmt.Println("main")
	}
}

// ============ 坑 3：Return 不是原子操作 ============

func t1() int {
	a := 1
	defer func() { a++ }()
	return a // 输出 1
}

func t2() (a int) {
	defer func() { a++ }()
	return 1 // 输出 2
}

func t3() (b int) {
	a := 1
	defer func() { a++ }() // 修改的不是返回值 b
	return 1               // 输出 1
}

func t4() (a int) {
	defer func(a int) { a++ }(a) // 值传递，改的是副本
	return 1                     // 输出 1
}

// ============ 坑 4：os.Exit 不执行 defer ============

func pitfallExit() {
	defer fmt.Println("1")
	fmt.Println("main")
	_ = os.Exit // 取消注释 os.Exit(0) 可看到 defer 不会执行
}

// ============ 坑 5：defer 只对当前协程有效 ============

func GoA() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("GoA panic:" + fmt.Sprintf("%s", err))
		}
	}()
	go GoB()
}

func GoB() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("GoB panic:" + fmt.Sprintf("%s", err)) // GoB 自己的 recover 才能捕获
		}
	}()
	panic("error")
}

// ============ 正确用法：资源释放 / 锁 / 对象池 ============

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func (s *SafeMap) Set(k string, v int) {
	s.mu.Lock()
	defer s.mu.Unlock() // 保证解锁一定执行
	s.m[k] = v
}

func (s *SafeMap) Get(k string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[k]
	return v, ok
}

// ============ 循环内 defer 的坑 ============

func loopDefer() {
	items := []int{1, 2, 3}
	fmt.Println("--- 正确：包成函数，每轮立即释放 ---")
	for _, v := range items {
		func() {
			defer fmt.Println("defer in loop:", v)
			fmt.Println("handle:", v)
		}()
	}
}

func main() {
	fmt.Println("=== 经典题 ===")
	quiz()

	fmt.Println("\n=== 坑 1：LIFO ===")
	pitfallLIFO()

	fmt.Println("\n=== 坑 2：闭包 vs 传参 ===")
	pitfallClosure()

	fmt.Println("\n=== 坑 3：Return 非原子 ===")
	fmt.Println("t1 =", t1())
	fmt.Println("t2 =", t2())
	fmt.Println("t3 =", t3())
	fmt.Println("t4 =", t4())

	fmt.Println("\n=== 坑 4：os.Exit ===")
	pitfallExit()

	fmt.Println("\n=== 坑 5：只对当前协程有效 ===")
	GoA()
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== 正确用法 ===")
	sm := &SafeMap{m: make(map[string]int)}
	sm.Set("a", 1)
	v, ok := sm.Get("a")
	fmt.Println("SafeMap.Get(a) =", v, ok)

	loopDefer()

	fmt.Println("\n=== 答案解析 ===")
	fmt.Println("经典题输出顺序：B D C A")
	fmt.Println("B 1 2 3 / D 3 2 5 / C 3 5 8 / A 1 3 4")
	fmt.Println("原因：defer LIFO 执行，且参数在 defer 定义时求值（值复制）")
}
