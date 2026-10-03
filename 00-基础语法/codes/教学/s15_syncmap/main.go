// s15_syncmap/main.go —— 对应教学文稿 08 并发编程（三）sync.Map
//
// 运行：go run ./s15_syncmap
// 竞态检测：go run -race ./s15_syncmap
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// map 不是并发安全的，并发读写会输出：
	// fatal error: concurrent map read and map write

	// ==================== Store / Load ====================
	demo := sync.Map{}
	go func() {
		for j := 0; j < 1000; j++ {
			demo.Store(j, j)
		}
	}()
	go func() {
		for j := 0; j < 1000; j++ {
			demo.Load(j) // 返回 (value, ok)
		}
	}()
	time.Sleep(time.Second)

	// ==================== 计算长度 ====================
	normal := make(map[int]int)
	for j := 0; j < 1000; j++ {
		normal[j] = j
	}
	fmt.Println("普通 map 的 len：", len(normal))

	// sync.Map 没有 len()，只能用 Range 计数
	lens := 0
	demo.Range(func(key, value interface{}) bool {
		lens++
		return true
	})
	fmt.Println("sync.Map 的 len：", lens)

	// ==================== 其余方法 ====================
	// 注意：Load / LoadOrStore / LoadAndDelete 都返回两个值，
	// 不能直接作为 Println 的参数传入
	m := sync.Map{}
	m.Store("a", 1)
	m.Store("b", 2)

	v1, ok1 := m.Load("a")
	fmt.Println("Load(a)          =", v1, ok1)

	actual, loaded := m.LoadOrStore("a", 100) // 已存在，返回旧值
	fmt.Println("LoadOrStore(a,100)=", actual, loaded)

	actual2, loaded2 := m.LoadOrStore("c", 300) // 不存在，存入新值
	fmt.Println("LoadOrStore(c,300)=", actual2, loaded2)

	m.Range(func(k, v interface{}) bool {
		fmt.Printf("  Range: %v = %v\n", k, v)
		return true
	})

	deleted, existed := m.LoadAndDelete("a")
	fmt.Println("LoadAndDelete(a) =", deleted, existed)

	deleted2, existed2 := m.LoadAndDelete("zzz")
	fmt.Println("LoadAndDelete(zzz)=", deleted2, existed2)

	m.Delete("b")
	fmt.Println("剩余 key 数：", countKeys(&m))

	// ==================== 对比 RWMutex + map ====================
	sm := NewSafeMap()
	sm.Set("x", 1)
	v, ok := sm.Get("x")
	fmt.Println("SafeMap.Get(x) =", v, ok, "Len =", sm.Len())

	raceDemo()

	// 复现 map 并发读写问题：取消下面这行的注释后运行，
	// 会输出 fatal error: concurrent map read and map write 并终止程序
	// raceMap()
}

// ==================== 辅助函数 ====================

func countKeys(m *sync.Map) int {
	n := 0
	m.Range(func(k, v interface{}) bool {
		n++
		return true
	})
	return n
}

// raceMap 复现「普通 map 并发读写」的问题。
// 用法：go run ./s15_syncmap/raceMap
// 输出：fatal error: concurrent map read and map write
func raceMap() {
	demo := make(map[int]int)
	go func() {
		for j := 0; j < 1000; j++ {
			demo[j] = j
		}
	}()
	go func() {
		for j := 0; j < 1000; j++ {
			_ = demo[j]
		}
	}()
	time.Sleep(time.Second)
	fmt.Println("不会执行到这里")
}

// raceDemo 用 WaitGroup + 互斥锁保护共享计数器，配合 -race 验证无竞态
func raceDemo() {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println("并发累加结果：", counter)
}

// SafeMap 是绝大多数场景的推荐写法：
// key 频繁增删、需要 len()、类型固定时，用 RWMutex + map
type SafeMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func NewSafeMap() *SafeMap { return &SafeMap{m: make(map[string]int)} }

func (s *SafeMap) Set(k string, v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[k] = v
}

func (s *SafeMap) Get(k string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[k]
	return v, ok
}

func (s *SafeMap) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m) // 有锁后可以直接 len
}
