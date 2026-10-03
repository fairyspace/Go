// s17_pool/main.go —— 对应教学文稿 09 性能调优（四）sync.Pool
//
// 运行：go run ./s17_pool
package main

import (
	"fmt"
	"sync"
	"time"
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
	fmt.Println("原因：GC 会随时清空 pool（每轮 GC 都会），不能依赖它做连接复用。")

	fmt.Println("\n=== 性能对比 ===")
	const n = 100000

	// 每次都 new
	start := time.Now()
	for i := 0; i < n; i++ {
		s := &student{Name: "Tom", Age: 30}
		_ = s
	}
	fmt.Printf("每次 new：      %v\n", time.Since(start))

	// 用 Pool 复用
	start = time.Now()
	for i := 0; i < n; i++ {
		s := New("Tom", 30)
		Release(s)
	}
	fmt.Printf("Pool 复用：     %v\n", time.Since(start))
	fmt.Println("（Pool 的收益主要在减少 GC 压力，需配合 -benchmem 看 allocs/op）")
}
