// s17_escape/main.go —— 对应教学文稿 09 性能调优（三）逃逸分析
//
// 分析命令：
//
//	go build -gcflags '-m -l' ./s17_escape
package main

import "fmt"

// 场景 01：interface 字段赋值会逃逸
type StudentInterface struct {
	Name interface{}
}

// 优化：使用固定类型，不逃逸
type StudentString struct {
	Name string
}

func escapeInterface() {
	stu := new(StudentInterface)
	stu.Name = "tom" // "tom" escapes to heap
	fmt.Println(stu.Name)
}

func noEscapeString() {
	stu := new(StudentString)
	stu.Name = "tom" // 不逃逸
	fmt.Println(stu.Name)
}

// 场景 02：返回指针会逃逸
func GetStudent() *StudentString {
	stu := new(StudentString)
	stu.Name = "tom"
	return stu // escapes to heap
}

// 场景 03：栈空间不足会逃逸
func bigSlice() {
	nums := make([]int, 10000, 10000) // escapes to heap
	for i := range nums {
		nums[i] = i
	}
	fmt.Println("len:", len(nums))
}

// 优化：容量足够小，可留在栈上
func smallSlice() {
	nums := make([]int, 10) // does not escape
	for i := range nums {
		nums[i] = i
	}
	fmt.Println("len:", len(nums))
}

func main() {
	fmt.Println("=== 场景 01：interface 字段 ===")
	escapeInterface()
	noEscapeString()

	fmt.Println("\n=== 场景 02：返回指针 ===")
	fmt.Println(GetStudent().Name)

	fmt.Println("\n=== 场景 03：栈空间不足 ===")
	bigSlice()
	smallSlice()

	fmt.Println("\n用以下命令查看逃逸分析结果：")
	fmt.Println("  go build -gcflags '-m -l' ./s17_escape")
}
