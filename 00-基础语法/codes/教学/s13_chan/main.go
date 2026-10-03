// s13_chan/main.go —— 对应教学文稿 08 并发编程（一）goroutine 与 channel
//
// 运行：go run ./s13_chan
package main

import (
	"fmt"
	"time"
)

func main() {
	// ==================== go 关键字 ====================
	fmt.Println("=== go 关键字 ===")
	fmt.Println("main start")
	go func() {
		fmt.Println("goroutine")
	}()
	time.Sleep(1 * time.Second) // 让主线程等待
	fmt.Println("main end")

	// ==================== 声明 chan ====================
	fmt.Println("\n=== 声明 chan ===")
	ch1 := make(chan string)     // 不带缓冲
	ch2 := make(chan string, 10) // 带 10 个缓冲
	ch3 := make(<-chan string)   // 只读通道
	ch4 := make(chan<- string)   // 只写通道
	fmt.Println(len(ch1), cap(ch1), len(ch2), cap(ch2))
	fmt.Printf("只读通道 %T，只写通道 %T\n", ch3, ch4)

	// ==================== 写入 / 读取 / 关闭 ====================
	fmt.Println("\n=== 写入读取关闭 ===")
	ch := make(chan string, 2)
	ch <- "a"
	ch <- "b"
	fmt.Println("读取：", <-ch)

	val, ok := <-ch
	fmt.Printf("带 ok 读取：%v，ok=%v\n", val, ok)

	close(ch)
	// ch <- "c"  // panic: send on closed channel
	// close(ch)  // panic: close of closed channel
	val2, ok2 := <-ch
	fmt.Printf("关闭后读取：%q，ok=%v（数据已取完，ok 为 false）\n", val2, ok2)

	// ==================== 示例：无缓冲导致死锁 ====================
	fmt.Println("\n=== 无缓冲 chan 死锁（已注释，避免中断程序）===")
	// 下面这段会输出 fatal error: all goroutines are asleep - deadlock!
	// 无缓冲的 chan，写入后立即阻塞，main 无法继续派发接收方
	//
	// func main() {
	// 	fmt.Println("main start")
	// 	ch := make(chan string)
	// 	ch <- "a"          // 入 chan，无缓冲，写入即阻塞
	// 	go func() {
	// 		val := <-ch      // 出 chan
	// 		fmt.Println(val)
	// 	}()
	// 	fmt.Println("main end")
	// }
	// 输出：main start / fatal error: all goroutines are asleep - deadlock!

	// ==================== 示例：有缓冲 ====================
	fmt.Println("=== 有缓冲 chan ===")
	ch5 := make(chan string, 1)
	ch5 <- "a" // 不阻塞
	fmt.Println("写入后 len =", len(ch5), "cap =", cap(ch5))
	fmt.Println("主线程直接读取：", <-ch5)

	// ==================== 示例：正确的生产者消费者 ====================
	fmt.Println("\n=== 生产者-消费者 ===")
	correctProducerConsumer()

	// ==================== select 多路复用 ====================
	fmt.Println("\n=== select ===")
	selectDemo()

	// ==================== Worker Pool ====================
	fmt.Println("\n=== Worker Pool ===")
	workerPool()
}

func producer(ch chan string) {
	fmt.Println("producer start")
	ch <- "a"
	ch <- "b"
	ch <- "c"
	ch <- "d" // 缓冲区为 3，第 4 次写入会阻塞
	fmt.Println("producer end")
}

func customer(ch chan string) {
	// for range 会在 channel 被 close 时自动结束循环
	for v := range ch {
		fmt.Println("消费：", v)
	}
}

func correctProducerConsumer() {
	ch := make(chan string, 3)
	done := make(chan bool)

	go func() {
		producer(ch)
		close(ch) // 发送方负责关闭，通知 consumer 收工
	}()

	go func() {
		customer(ch)
		done <- true
	}()

	<-done // 用 channel 等待，而不是 time.Sleep
	fmt.Println("main end")
}

func selectDemo() {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	ch3 := make(chan string, 1)

	ch1 <- "from ch1"

	select {
	case v := <-ch1:
		fmt.Println("ch1:", v)
	case v := <-ch2:
		fmt.Println("ch2:", v)
	case ch3 <- "x":
		fmt.Println("写入 ch3")
	case <-time.After(1 * time.Second):
		fmt.Println("超时")
	}

	// 全部阻塞 + default 分支
	select {
	case v := <-ch2:
		fmt.Println("ch2:", v)
	default:
		fmt.Println("没有任何 channel 就绪，走 default")
	}
}

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		results <- j * j
	}
}

func workerPool() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	// 方向化 channel：worker 只读 jobs、只写 results，消除误用
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// 派发任务
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs) // 关键：通知所有 worker 收工

	// 收集结果
	sum := 0
	for i := 0; i < 9; i++ {
		sum += <-results
	}
	fmt.Println("sum =", sum) // 1+4+9+...+81 = 285
}
