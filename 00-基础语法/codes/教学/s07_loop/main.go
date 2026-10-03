// s07_loop/main.go —— 对应教学文稿 05 循环与流程控制
//
// 运行：go run ./s07_loop
package main

import "fmt"

func main() {
	// ==================== 循环 array ====================
	person := [3]string{"Tom", "Aaron", "John"}
	fmt.Printf("len=%d cap=%d array=%v\n", len(person), cap(person), person)

	for k, v := range person { // 方式一：range 拿键值
		fmt.Printf("person[%d]: %s\n", k, v)
	}
	for i := range person { // 方式二：range 拿下标
		fmt.Printf("person[%d]: %s\n", i, person[i])
	}
	for i := 0; i < len(person); i++ { // 方式三：标准三段式
		fmt.Printf("person[%d]: %s\n", i, person[i])
	}
	for _, name := range person { // 空白符
		fmt.Println("name :", name)
	}

	// ==================== 循环 slice ====================
	fmt.Println("\n--- slice ---")
	sl := []string{"Tom", "Aaron", "John"}
	fmt.Printf("len=%d cap=%d slice=%v\n", len(sl), cap(sl), sl)

	for k, v := range sl {
		fmt.Printf("person[%d]: %s\n", k, v)
	}

	// range 的 value 是副本
	s := []int{1, 2, 3}
	for _, v := range s {
		v = 99 // 只改副本
		fmt.Println("副本 v =", v)
	}
	fmt.Println("改副本后：", s) // [1 2 3]

	for i := range s {
		s[i] = 99 // 改原切片
	}
	fmt.Println("改原切片：", s) // [99 99 99]

	// ==================== 循环 map ====================
	fmt.Println("\n--- map ---")
	personMap := map[int]string{
		1: "Tom",
		2: "Aaron",
		3: "John",
	}
	for k, v := range personMap {
		fmt.Printf("person[%d]: %s\n", k, v)
	}
	for i := range personMap {
		fmt.Printf("person[%d]: %s\n", i, personMap[i])
	}
	for _, name := range personMap {
		fmt.Println("name :", name)
	}

	// ==================== break ====================
	fmt.Println("\n--- break ---")
	for i := 1; i <= 10; i++ {
		if i == 6 {
			break // 跳出当前循环
		}
		fmt.Println("i =", i)
	}

	// break + 标签跳出多层
	fmt.Println("双层循环 break outer：")
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				fmt.Printf("跳出于 i=%d j=%d\n", i, j)
				break outer
			}
		}
	}

	// ==================== continue ====================
	fmt.Println("\n--- continue ---")
	for i := 1; i <= 10; i++ {
		if i == 6 {
			continue // 跳过本次
		}
		fmt.Println("i =", i)
	}

	// ==================== goto ====================
	fmt.Println("\n--- goto ---")
	fmt.Println("begin")
	for i := 1; i <= 10; i++ {
		if i == 6 {
			goto END
		}
		fmt.Println("i =", i)
	}
END:
	fmt.Println("end")

	// ==================== switch ====================
	fmt.Println("\n--- switch ---")
	for i := 1; i <= 7; i++ {
		fmt.Printf("当 i = %d 时：", i)
		switch i {
		case 1:
			fmt.Println("输出 i =", 1)
		case 2:
			fmt.Println("输出 i =", 2)
		case 3:
			fmt.Println("输出 i =", 3)
			fallthrough // 穿透一层
		case 4, 5, 6: // 多值 case
			fmt.Println("输出 i = 4 or 5 or 6")
		default:
			fmt.Println("输出 i = xxx")
		}
	}

	// 无表达式 switch
	fmt.Println("\n--- 无表达式 switch ---")
	score := 85
	var grade string
	switch {
	case score >= 90:
		grade = "A"
	case score >= 60:
		grade = "B"
	default:
		grade = "C"
	}
	fmt.Println("grade =", grade)

	// 类型 switch
	fmt.Println("\n--- 类型 switch ---")
	describe := func(i interface{}) {
		switch v := i.(type) {
		case int:
			fmt.Println("int:", v)
		case string:
			fmt.Println("string 长度:", len(v))
		case []int:
			fmt.Println("[]int 长度:", len(v))
		case error:
			fmt.Println("error:", v.Error())
		case nil:
			fmt.Println("nil")
		default:
			fmt.Println("未知类型")
		}
	}
	describe(1)
	describe("go")
	describe([]int{1, 2, 3})
	describe(nil)
	describe(3.14)
}
