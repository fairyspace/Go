// s02_var/main.go —— 对应教学文稿 02 变量、常量与数据类型
//
// 运行：go run ./s02_var
package main

import "fmt"

// iota 枚举：每个 const 块内从 0 开始，每出现一次自动加 1
type Weekday int

const (
	Sunday    Weekday = iota // 0
	Monday                   // 1，重复上一行表达式
	Tuesday                  // 2
	Wednesday                // 3
)

func (d Weekday) String() string {
	return [...]string{"Sunday", "Monday", "Tuesday", "Wednesday"}[d]
}

// 配合位运算定义权限位
const (
	PermRead  = 1 << iota // 0001
	PermWrite             // 0010
	PermExec              // 0100
)

func main() {
	// ==================== 常量声明 ====================
	const name string = "Tom"
	fmt.Println(name)

	const age = 30 // 自动推断为 int
	fmt.Println(age)

	const name1, name2 string = "Tom", "Jay"
	fmt.Println(name1, name2)

	const name3, age1 = "Tom", 30
	fmt.Println(name3, age1)

	// ==================== 变量声明 ====================
	// 三种声明方式
	var age2 uint8 = 31 // 第一种：显式类型
	var age3 = 32       // 第二种：自动推断
	age4 := 33          // 第三种：短变量声明（变量必须未声明过）
	fmt.Println(age2, age3, age4)

	// 多变量声明
	var age5, age6, age7 int = 31, 32, 33
	fmt.Println(age5, age6, age7)

	var name4, age8 = "Tom", 30
	fmt.Println(name4, age8)

	name5, isBoy, height := "Jay", true, 180.66
	fmt.Println(name5, isBoy, height)

	// ==================== 零值 ====================
	var s string
	var b bool
	var p *int
	fmt.Printf("零值：%q %v %v\n", s, b, p)

	// ==================== 变量交换与丢弃 ====================
	a, c := 1, 2
	a, c = c, a // 直接交换，无需临时变量
	fmt.Println("交换后：", a, c)

	_, _, d := 1, 2, 3 // 用空白符 _ 丢弃不要的值
	fmt.Println(d)

	// ==================== iota 枚举 ====================
	fmt.Println(Sunday, Monday, Tuesday, Wednesday) // Sunday Monday Tuesday Wednesday

	flag := PermRead | PermWrite
	fmt.Printf("权限位：%b，读=%v，写=%v\n", flag, flag&PermRead != 0, flag&PermWrite != 0)

	// ==================== 四种输出方法 ====================
	fmt.Print("输出到控制台不换行")
	fmt.Println("---")
	fmt.Println("输出到控制台并换行")
	fmt.Printf("name=%s,age=%d\n", "Tom", 30)
	fmt.Printf("name=%s,age=%d,height=%v\n", "Tom", 30, fmt.Sprintf("%.2f", 180.567))

	// Sprintf 单独使用：格式化但不输出
	s2 := fmt.Sprintf("%s-%d", "Tom", 30)
	fmt.Println("Sprintf:", s2)

	// ==================== 类型断言 ====================
	var val interface{} = "go"
	if str, ok := val.(string); ok {
		fmt.Println("断言成功：", str)
	}
}
