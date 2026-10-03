// s14_waitgroup/main.go —— 对应教学文稿 08 并发编程（二）sync.WaitGroup
//
// 运行：go run ./s14_waitgroup
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// ==================== 正确使用 ====================
	fmt.Println("=== 正确使用 ===")
	correct()

	// ==================== 坑 1：必须传引用类型 ====================
	fmt.Println("\n=== 坑 1：必须传引用类型 ===")
	fmt.Println("正确：go handlerTask1(&wg)")
	fmt.Println("错误：go handlerTask1(wg)")
	fmt.Println("原因：sync.WaitGroup 含 noCopy 标记，Go 1.20+ 传值会 panic：")
	fmt.Println("      sync: WaitGroup is copied before use")

	// ==================== 坑 2：Add 不能放在 go 内部 ====================
	fmt.Println("\n=== 坑 2：wg.Add() 不能放在 go 语句内部 ===")
	fmt.Println("错误写法：")
	fmt.Println("  go handlerTask1(&wg)")
	fmt.Println("  wg.Wait()")
	fmt.Println("  func handlerTask1(wg *sync.WaitGroup) { wg.Add(1); defer wg.Done() }")
	fmt.Println("原因：可能 Wait 已经返回，任务还没开始主流程就继续了")
	fmt.Println("正确做法：先 wg.Add(3)，再启动 3 个 goroutine，最后 wg.Wait()")

	// ==================== 坑 3：Add 与 Done 计数一致 ====================
	fmt.Println("\n=== 坑 3：Add 与 Done 计数必须一致 ===")
	fmt.Println("wg.Done() 就是执行 wg.Add(-1)")
	fmt.Println("计数归零时 Wait() 返回，计数为负会 panic")

	// ==================== 局限：需要通知机制时用 channel/context ====================
	fmt.Println("\n=== 使用局限 ===")
	fmt.Println("WaitGroup 仅适用于「等待全部子任务执行完毕」")
	fmt.Println("若需求是「第一个子任务失败时通知其他子任务停止」，需用 channel 或 context：")
	limitation()

	// ==================== 变体：收集结果 ====================
	fmt.Println("\n=== 变体：并发计算并收集结果 ===")
	collectResults()
}

func correct() {
	var wg sync.WaitGroup

	wg.Add(3) // 必须在 go 之前

	go handlerTask1(&wg)
	go handlerTask2(&wg)
	go handlerTask3(&wg)

	wg.Wait() // 等待全部完成

	fmt.Println("全部任务执行完毕.")
}

func handlerTask1(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("执行任务 1")
}

func handlerTask2(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("执行任务 2")
}

func handlerTask3(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("执行任务 3")
}

func limitation() {
	ctx, cancel := contextWithCancel()
	defer cancel() // 保证所有子任务都能收到取消信号

	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("任务被取消")
		case r := <-time.After(50 * time.Millisecond):
			fmt.Println("成功:", r)
		}
	}()

	time.Sleep(100 * time.Millisecond)
}

func collectResults() {
	tasks := []int{1, 2, 3, 4, 5}

	results := make(chan int, len(tasks))
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results <- n * n
		}(t)
	}

	wg.Wait()
	close(results)

	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("结果总和：", sum) // 1+4+9+16+25 = 55
}
