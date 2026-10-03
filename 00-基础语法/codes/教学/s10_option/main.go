// s10_option/main.go —— 对应教学文稿 07 接口与 Options 模式（二）
//
// 运行：go run ./s10_option
package main

import (
	"demo/s10_option/friend"
	"fmt"
)

func main() {
	// ==================== Options 模式 ====================
	friends, err := friend.Find("附近的人",
		friend.WithSex(1),
		friend.WithAge(30),
		friend.WithHeight(160),
		friend.WithWeight(55),
		friend.WithHobby("爬山"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("--- Options 模式 ---")
	fmt.Println(friends)

	// 只传部分选项，其余走零值
	friends2, _ := friend.Find("附近的人", friend.WithSex(2), friend.WithAge(25))
	fmt.Println("--- 只传部分选项 ---")
	fmt.Println(friends2)

	// 不传任何选项
	friends3, _ := friend.Find("附近的人")
	fmt.Println("--- 不传选项 ---")
	fmt.Println(friends3)

	// ==================== 对比：结构体参数 ====================
	friends4, _ := friend.FindWithStruct("附近的人", friend.FindOptions{
		Sex:   1,
		Age:   30,
		Hobby: "爬山",
	})
	fmt.Println("--- 结构体参数 ---")
	fmt.Println(friends4)
}
