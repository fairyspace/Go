// s17_pool/main.go —— 对应教学文稿 09 性能调优（四）sync.Pool
//
// 运行：go run ./s17_pool
package main

import (
	"fmt"
	"sync"
)

// student 是临时对象，可以用 sync.Pool 复用
type student struct {
	Name string
	Age  int
	// Tags 故意保留大字段，用于说明「Put 前必须重置」的重要性
	Tags []string
}

var studentPool = &sync.Pool{
	New: func() interface{} {
		return new(student)
	},
}

// New 从池中取出对象
func New(name string, age int) *student {
	stu := studentPool.Get().(*student)
	stu.Name = name
	stu.Age = age
	return stu
}

// Release 重置后归还池中
func Release(stu *student) {
	stu.Name = ""
	stu.Age = 0
	stu.Tags = nil // 关键：不重置会残留大对象，导致内存泄漏
	studentPool.Put(stu)
}

func handle(name string, age int) {
	stu := New(name, age)
	defer Release(stu) // 用完立刻归还，defer 保证异常路径也归还

	stu.Tags = append(stu.Tags, "读书", "爬山")
	fmt.Printf("处理中：%s(%d) tags=%v\n", stu.Name, stu.Age, stu.Tags)
}

func main() {
	fmt.Println("=== sync.Pool 复用临时对象 ===")
	for i := 0; i < 3; i++ {
		handle("Tom", 30)
	}
	// 每次 handle 结束后对象都归还池中，下次 New 会复用同一个对象

	fmt.Println("\n=== 池不是连接池 ===")
	fmt.Println("sync.Pool 只能存临时对象，不能存 socket 长连接、数据库连接池。")
	fmt.Println("原因：Go 1.13 起 Pool 采用 victim cache —— GC 开始时把主缓存挪入 victim 区，")
	fmt.Println("      上一轮的 victim 才真正丢弃，所以对象最多两个 GC 周期内被回收。")
	fmt.Println("      池不保证对象长期可用，因此不能当连接池用。")

	fmt.Println("\n=== 性能对比 ===")
	fmt.Println("微基准不要用 time.Now() 手写：短代码可能被编译器优化掉，也没有 allocs/op 统计。")
	fmt.Println("正确做法见 pool_test.go：")
	fmt.Println("  go test -bench . -benchmem ./s17_pool")
	fmt.Println("（Pool 的收益主要在减少 GC 压力，要看 allocs/op 而不是 ns/op）")
}
