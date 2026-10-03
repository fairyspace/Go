// s03_array/main.go —— 对应教学文稿 03 数组、切片与 Map（一）数组
//
// 运行：go run ./s03_array
package main

import "fmt"

func modifyArr(a [5]int) {
	a[1] = 20 // 传值：只改副本
}

func modifyArrPtr(a *[5]int) {
	a[1] = 20 // 传地址：改动生效
}

func main() {
	// ==================== 声明数组 ====================
	// 一维数组
	var arr1 [5]int
	fmt.Println("零值数组：", arr1) // [0 0 0 0 0]

	var arr2 = [5]int{1, 2, 3, 4, 5}
	arr3 := [5]int{1, 2, 3, 4, 5}
	arr4 := [...]int{1, 2, 3, 4, 5, 6} // [...] 自动推断长度
	arr5 := [5]int{0: 3, 1: 5, 4: 6}   // 下标初始化，未指定的取零值

	// 二维数组
	var arr6 = [3][5]int{{1, 2, 3, 4, 5}, {9, 8, 7, 6, 5}, {3, 4, 5, 6, 7}}
	arr7 := [3][5]int{{1, 2, 3, 4, 5}, {9, 8, 7, 6, 5}, {3, 4, 5, 6, 7}}
	arr8 := [...][5]int{{1, 2, 3, 4, 5}, {9, 8, 7, 6, 5}, {0: 3, 1: 5, 4: 6}}

	fmt.Println("arr2:", arr2)
	fmt.Println("arr3:", arr3)
	fmt.Println("arr4:", arr4)
	fmt.Println("arr5:", arr5) // [3 5 0 0 6]
	fmt.Println("arr6:", arr6)
	fmt.Println("arr7:", arr7)
	fmt.Println("arr8:", arr8)

	// 数组的 len() 与 cap() 始终相同
	fmt.Printf("len=%d cap=%d\n", len(arr2), cap(arr2))

	// ==================== 注意事项一：长度不可变 ====================
	// 取消下面两行的注释可看到报错：
	// invalid array index 5 (out of bounds for 5-element array)
	// arr2[5] = 6

	// ==================== 注意事项二：值类型 ====================
	arr := [5]int{1, 2, 3, 4, 5}

	modifyArr(arr)           // 传值：函数内改动不影响外部
	fmt.Println("传值后：", arr) // [1 2 3 4 5]

	modifyArrPtr(&arr)       // 传地址：改动生效
	fmt.Println("传址后：", arr) // [1 20 3 4 5]

	// ==================== 注意事项三：类型必须完全一致 ====================
	// 取消下面这行注释可看到报错：
	// cannot use arr (type [5]int) as type [6]int in assignment
	// var arrOther [6]int = arr

	var arrSame [5]int = arr // 长度与元素类型都相同，OK
	fmt.Println("同类型赋值：", arrSame)
}
