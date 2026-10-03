// s04_slice/main.go —— 对应教学文稿 03 数组、切片与 Map（二）切片
//
// 运行：go run ./s04_slice
package main

import "fmt"

// 删除下标为 i 的元素
func deleteElem(s []int, i int) []int {
	return append(s[:i], s[i+1:]...)
}

// 保留满足条件的元素（复用底层数组，零分配）
func filter(s []int, keep func(int) bool) []int {
	res := s[:0]
	for _, v := range s {
		if keep(v) {
			res = append(res, v)
		}
	}
	return res
}

// 安全的删除：显式复制，不污染原有引用者
func deleteElemSafe(s []int, i int) []int {
	res := make([]int, 0, len(s)-1)
	res = append(res, s[:i]...)
	res = append(res, s[i+1:]...)
	return res
}

func main() {
	// ==================== 声明切片 ====================
	var sli1 []int // nil 切片
	fmt.Printf("nil 切片：  len=%d cap=%d slice=%v nil=%v\n", len(sli1), cap(sli1), sli1, sli1 == nil)

	sli2 := []int{}              // 空切片（非 nil）
	sli3 := []int{1, 2, 3, 4, 5} // 字面量
	sli4 := make([]int, 5)       // 长度 5，容量 5
	sli5 := make([]int, 5, 9)    // 长度 5，容量 9（预留空间）
	sli6 := sli3[1:3]            // 从已有切片切出

	fmt.Printf("空切片：    len=%d cap=%d slice=%v nil=%v\n", len(sli2), cap(sli2), sli2, sli2 == nil)
	fmt.Printf("字面量：    len=%d cap=%d slice=%v\n", len(sli3), cap(sli3), sli3)
	fmt.Printf("make(5)：   len=%d cap=%d slice=%v\n", len(sli4), cap(sli4), sli4)
	fmt.Printf("make(5,9)： len=%d cap=%d slice=%v\n", len(sli5), cap(sli5), sli5)
	fmt.Printf("切出：      len=%d cap=%d slice=%v\n", len(sli6), cap(sli6), sli6)

	// ==================== 截取切片 ====================
	sli := []int{1, 2, 3, 4, 5, 6}
	fmt.Println("\n--- 截取 ---")
	fmt.Println("sli[1]     ==", sli[1])
	fmt.Println("sli[:]     ==", sli[:])
	fmt.Println("sli[1:]    ==", sli[1:])
	fmt.Println("sli[:4]    ==", sli[:4])
	fmt.Println("sli[0:3]   ==", sli[0:3])
	fmt.Println("sli[0:3:4] ==", sli[0:3:4])

	// ==================== 三索引切片：容量隔离 ====================
	fmt.Println("\n--- 三索引切片防止数据污染 ---")
	s := make([]int, 3, 10)
	s[0], s[1], s[2] = 1, 2, 3

	sub := s[0:3:3] // 容量限制为 3
	sub = append(sub, 99)
	fmt.Println("原切片 s：", s)     // [1 2 3 0 0 0 0 0 0 0] —— 未被污染
	fmt.Println("子切片 sub：", sub) // [1 2 3 99]

	// 对比：不限制容量就会污染原切片
	s2 := make([]int, 3, 10)
	s2[0], s2[1], s2[2] = 1, 2, 3
	sub2 := s2[0:3] // 容量仍为 10
	sub2 = append(sub2, 99)
	fmt.Println("原切片 s2：", s2) // [1 2 3 99 ...] —— 被污染
	fmt.Println("子切片 sub2：", sub2)

	// ==================== 追加切片 ====================
	fmt.Println("\n--- 追加：观察扩容 ---")
	sli7 := []int{4, 5, 6} // len=3 cap=3
	for i := 7; i <= 10; i++ {
		sli7 = append(sli7, i)
		fmt.Printf("append %2d 后：len=%d cap=%d slice=%v\n", i, len(sli7), cap(sli7), sli7)
	}

	// 预分配：一次到位，不再扩容
	pre := make([]int, 0, 100)
	pre = append(pre, 1, 2, 3)
	fmt.Printf("预分配：len=%d cap=%d\n", len(pre), cap(pre))

	// ==================== 删除切片 ====================
	fmt.Println("\n--- 删除 ---")
	del := []int{1, 2, 3, 4, 5, 6, 7, 8}

	// 删除尾部 2 个
	tail := del[:len(del)-2]
	fmt.Println("删除尾部 2 个：", tail)

	// 删除开头 2 个
	head := del[2:]
	fmt.Println("删除开头 2 个：", head)

	// 删除中间 2 个（下标 3、4）
	mid := make([]int, len(del))
	copy(mid, del) // 先复制，避免污染原切片
	mid = append(mid[:3], mid[3+2:]...)
	fmt.Println("删除中间 2 个：", mid)

	// 封装成函数
	fmt.Println("deleteElem：", deleteElem([]int{1, 2, 3, 4, 5}, 1))
	fmt.Println("deleteElemSafe：", deleteElemSafe([]int{1, 2, 3, 4, 5}, 1))
	fmt.Println("filter 保留偶数：", filter([]int{1, 2, 3, 4, 5, 6}, func(v int) bool { return v%2 == 0 }))

	// ==================== 复制 ====================
	fmt.Println("\n--- 复制 ---")
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	n := copy(dst, src)
	fmt.Println("copy 返回复制的个数：", n, "dst：", dst)
}
