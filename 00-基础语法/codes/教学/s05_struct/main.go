// s05_struct/main.go —— 对应教学文稿 04 结构体与 JSON（一）结构体
//
// 运行：go run ./s05_struct
package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name string
	Age  int
}

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type Address struct {
	City string
	Zip  string
}

type User struct {
	Name    string
	Age     int
	Address Address  // 值嵌套（拷贝）
	Ptr     *Address // 指针嵌套（共享）
}

// 值接收者：只读
func (a Address) GetCity() string {
	return a.City
}

// 指针接收者：修改生效
func (a *Address) SetCity(city string) {
	a.City = city
}

type Counter struct{ n int }

func (c *Counter) Add()    { c.n++ }
func (c Counter) Get() int { return c.n }

func setData(res *Result) { // 指针参数
	res.Code = 500
	res.Message = "fail"
}

func toJson(res *Result) {
	jsons, errs := json.Marshal(res)
	if errs != nil {
		fmt.Println("json marshal error:", errs)
	}
	fmt.Println("json data :", string(jsons))
}

func main() {
	// ==================== 声明结构体 ====================
	// 零值演示：未赋值的字段取各自零值（要在赋值之前打印才看得到）
	fmt.Printf("零值结构体：%+v\n", Person{}) // {Name: Age:0}

	var p1 Person // 声明即取零值
	p1.Name = "Tom"
	p1.Age = 30
	fmt.Println("p1 =", p1) // {Tom 30}

	p2 := Person{Name: "Burke", Age: 31} // 字段名初始化
	p3 := Person{Name: "Aaron"}          // 部分初始化：未写的字段取零值（Age = 0）
	fmt.Println("p2 =", p2, "p3 =", p3)

	// 匿名结构体：常用于临时数据、测试用例
	p4 := struct {
		Name string
		Age  int
	}{Name: "匿名", Age: 33}
	fmt.Println("p4 =", p4)

	// ==================== 嵌套 ====================
	u := User{
		Name:    "Tom",
		Address: Address{City: "杭州", Zip: "310000"},
	}
	fmt.Println("嵌套取值：", u.Address.City, u.Address.GetCity())

	u.Ptr = &Address{City: "上海", Zip: "200000"}
	u.Ptr.SetCity("北京")
	fmt.Println("指针嵌套取值：", u.Ptr.City)

	// 结构体可整体比较（要求所有字段可比较）
	a := Address{City: "杭州", Zip: "310000"}
	b := Address{City: "杭州", Zip: "310000"}
	fmt.Println("结构体比较：", a == b)

	// ==================== 接收者 ====================
	var c Counter
	c.Add() // Go 会自动取地址，等价于 (&c).Add()
	fmt.Println("Counter：", c.Get())

	// ==================== 值传递 vs 指针传递 ====================
	fmt.Println("\n--- 传值 vs 传指针 ---")
	res := Result{Code: 200, Message: "success"}
	toJson(&res)
	setData(&res) // 传指针，修改生效
	toJson(&res)

	// ==================== JSON 序列化 ====================
	fmt.Println("\n--- JSON ---")
	jsons, errs := json.Marshal(res)
	if errs != nil {
		fmt.Println("json marshal error:", errs)
	}
	fmt.Println("序列化：", string(jsons))

	// 格式化输出
	pretty, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println("格式化：\n" + string(pretty))

	// omitempty：零值时省略
	type Optional struct {
		Name  string `json:"name"`
		Age   int    `json:"age,omitempty"`
		Skip  string `json:"-"`
		Empty string `json:"empty,omitempty"`
	}
	o, _ := json.Marshal(Optional{Name: "Tom"})
	fmt.Println("omitempty：", string(o)) // {"name":"Tom"}

	// 反序列化
	var res2 Result
	if err := json.Unmarshal(jsons, &res2); err != nil {
		fmt.Println("json unmarshal error:", err)
	}
	fmt.Printf("反序列化：%+v\n", res2)
}
